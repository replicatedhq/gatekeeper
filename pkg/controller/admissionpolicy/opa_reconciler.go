package admissionpolicy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	policiesv1alpha2 "github.com/replicatedhq/gatekeeper/pkg/apis/policies/v1alpha2"
	gatekeepertls "github.com/replicatedhq/gatekeeper/pkg/tls"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *ReconcileAdmissionPolicy) ensureOPARunning(instance *policiesv1alpha2.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "ensureOPARunning"))
	debug.Log("event", "ensure opa instance running")

	// Create the opa instance
	if err := r.reconcileOpenPolicyAgentSecret(instance); err != nil {
		return errors.Wrap(err, "reconcile opa secret")
	}
	if err := r.reconcileOpenPolicyAgentService(instance); err != nil {
		return errors.Wrap(err, "reconcile opa service")
	}
	if err := r.reconcileOpenPolicyAgentDeployment(instance); err != nil {
		return errors.Wrap(err, "reconcile opa deployment")
	}
	if err := r.waitForOpenPolicyAgentDeploymentReady(instance); err != nil {
		return errors.Wrap(err, "is opa ready")
	}

	return nil
}

// reconcileOpenPolicyAgentSecret will create a ca and cert secret for the instance
func (r *ReconcileAdmissionPolicy) reconcileOpenPolicyAgentSecret(instance *policiesv1alpha2.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.reconcileOpenPolicyAgentSecret"))
	debug.Log("event", "reconciling opa secret")

	secretName, err := opaSecretName()
	if err != nil {
		return errors.Wrap(err, "get secret name")
	}
	serviceName, err := opaServiceName()
	if err != nil {
		return errors.Wrap(err, "get service name")
	}

	found := &corev1.Secret{}
	err = r.Get(context.TODO(), secretName, found)
	if err != nil && apierrors.IsNotFound(err) {
		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName.Name,
				Namespace: secretName.Namespace,
			},
			Data: map[string][]byte{},
		}

		// Create CA
		caCert, caKey, err := gatekeepertls.CreateCertificateAuthority(r.Logger)
		if err != nil {
			return errors.Wrap(err, "create cert authority")
		}

		secret.Data["ca.crt"] = caCert
		secret.Data["ca.key"] = caKey

		// Create Cert
		cert, key, err := gatekeepertls.CreateCertFromCA(r.Logger, serviceName, caCert, caKey)
		if err != nil {
			return errors.Wrap(err, "create cert from ca")
		}

		secret.Data["tls.crt"] = cert
		secret.Data["tls.key"] = key

		debug.Log("event", "create tls secret")
		err = r.Create(context.TODO(), secret)
		if err != nil {
			return errors.Wrap(err, "create secret")
		}
	} else if err != nil {
		return errors.Wrap(err, "get secret")
	}

	// TODO reconcile isn't simply a deepequal here, we need to check the tls cert expiration

	return nil
}

func (r *ReconcileAdmissionPolicy) reconcileOpenPolicyAgentService(instance *policiesv1alpha2.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.reconcileOpenPolicyAgentService"))
	debug.Log("event", "reconciling opa service")

	serviceName, err := opaSecretName()
	if err != nil {
		return errors.Wrap(err, "get service name")
	}
	deploymentName, err := opaDeploymentName()
	if err != nil {
		return errors.Wrap(err, "get deployment name")
	}
	secretName, err := opaSecretName()
	if err != nil {
		return errors.Wrap(err, "get secret name")
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName.Name,
			Namespace: serviceName.Namespace,
			Labels: map[string]string{
				"app":               "gatekeeper",
				"role":              "openpolicyagent",
				"caBundleSecret":    secretName.Name,
				"caBundleNamespace": secretName.Namespace,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:     "https",
					Protocol: corev1.ProtocolTCP,
					Port:     443,
					TargetPort: intstr.IntOrString{
						IntVal: 443,
					},
				},
			},
			Selector: map[string]string{
				"gatekeeper": deploymentName.Name,
				"controller": "openPolicyAgent",
			},
		},
	}

	foundService := &corev1.Service{}
	err = r.Get(context.TODO(), serviceName, foundService)
	if err != nil && apierrors.IsNotFound(err) {
		debug.Log("event", "creating service")

		err := r.Create(context.TODO(), service)
		if err != nil {
			return errors.Wrap(err, "create service")
		}
	} else if err != nil {
		return errors.Wrap(err, "get service")
	}

	// TODO compare ignoring the generated fields

	return nil
}

func (r *ReconcileAdmissionPolicy) waitForOpenPolicyAgentDeploymentReady(instance *policiesv1alpha2.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.waitForOpenPolicyAgentDeploymentReady"))

	debug.Log("event", "waiting for opa to report ready")

	abortAt := time.Now().Add(time.Minute * 2)
	for {
		if time.Now().After(abortAt) {
			return errors.Wrap(fmt.Errorf("timeout waiting for opa to be ready"), "waiting for opa")
		}

		isReady, err := r.isOpenPolicyAgentDeploymentReady(instance)
		if err != nil {
			return errors.Wrap(err, "is opa ready")
		}

		if isReady {
			return nil
		}
	}
}

func (r *ReconcileAdmissionPolicy) isOpenPolicyAgentDeploymentReady(instance *policiesv1alpha2.AdmissionPolicy) (bool, error) {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.isOpenPolicyAgentDeploymentReady"))

	secretName, err := opaSecretName()
	if err != nil {
		return false, errors.Wrap(err, "get secret name")
	}
	serviceName, err := opaServiceName()
	if err != nil {
		return false, errors.Wrap(err, "get service name")
	}
	deploymentName, err := opaDeploymentName()
	if err != nil {
		return false, errors.Wrap(err, "get deployment name")
	}

	foundSecret := &corev1.Secret{}
	err = r.Get(context.TODO(), secretName, foundSecret)
	if err != nil {
		return false, errors.Wrap(err, "get secret")
	}

	foundDeployment := &appsv1.Deployment{}
	err = r.Get(context.TODO(), deploymentName, foundDeployment)
	if err != nil && apierrors.IsNotFound(err) {
		// ignore this, the dpeloyment will be created eventually
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "get deployment to check status")
	}

	if foundDeployment.Status.AvailableReplicas == 0 {
		return false, nil
	}

	debug.Log("event", "deploying main status")

	mainPolicy := `package system

import data.kubernetes.admission

main = {
	"apiVersion": "admission.k8s.io/v1beta1",
	"kind": "AdmissionReview",
	"response": response,
}

default response = {"allowed": true}

response = {
	"allowed": false,
	"status": {
	 	"reason": reason,
	},
} {
	reason = concat(", ", admission.deny)
	reason != ""
}`

	uri := fmt.Sprintf("https://%s.%s.svc/v1/policies/main", serviceName.Name, serviceName.Namespace)

	rootCAs, err := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if ok := rootCAs.AppendCertsFromPEM(foundSecret.Data["ca.crt"]); !ok {
		return false, errors.Wrapf(err, "append ca cert")
	}
	config := &tls.Config{
		RootCAs: rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: config}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second,
	}
	req, err := http.NewRequest("PUT", uri, strings.NewReader(mainPolicy))
	if err != nil {
		return false, errors.Wrap(err, "create request main policy")
	}
	req.ContentLength = int64(len(mainPolicy))
	resp, err := client.Do(req)
	if err != nil {
		return false, errors.Wrap(err, "create main policy")
	}

	debug.Log("event", "main status deployment result", "status code", resp.StatusCode)

	return resp.StatusCode == http.StatusOK, nil
}

func (r *ReconcileAdmissionPolicy) reconcileOpenPolicyAgentDeployment(instance *policiesv1alpha2.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.reconcileOpenPolicyAgentDeployment"))

	debug.Log("event", "reconciling opa deployment")

	secretName, err := opaSecretName()
	if err != nil {
		return errors.Wrap(err, "get secret name")
	}
	deploymentName, err := opaDeploymentName()
	if err != nil {
		return errors.Wrap(err, "get deployment name")
	}

	// Deployment
	labels := map[string]string{
		"gatekeeper": deploymentName.Name,
		"controller": "openPolicyAgent",
	}

	replicas := int32(1)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName.Name,
			Namespace: deploymentName.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   deploymentName.Name,
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "opa",
							Image: "openpolicyagent/opa:0.10.1",
							Args: []string{
								"run",
								"--server",
								"--tls-cert-file=/certs/tls.crt",
								"--tls-private-key-file=/certs/tls.key",
								"--addr=0.0.0.0:443",
								"--addr=http://127.0.0.1:8181",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									ReadOnly:  true,
									MountPath: "/certs",
									Name:      "tls",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "https",
									ContainerPort: 443,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "tls",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secretName.Name,
								},
							},
						},
					},
				},
			},
		},
	}

	foundDeployment := &appsv1.Deployment{}
	err = r.Get(context.TODO(), deploymentName, foundDeployment)
	if err != nil && apierrors.IsNotFound(err) {
		debug.Log("event", "creating deployment")

		err := r.Create(context.TODO(), deployment)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// TODO compare ignoring the generated fields

	return nil
}

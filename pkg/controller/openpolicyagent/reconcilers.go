package openpolicyagent

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
	controllersv1alpha1 "github.com/replicatedhq/gatekeeper/pkg/apis/controllers/v1alpha1"
	gatekeepertls "github.com/replicatedhq/gatekeeper/pkg/tls"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	serviceNamePrefix = "gatekeeper-opa"
	secretNamePrefix  = "gatekeeper"
)

func (r *ReconcileOpenPolicyAgent) reconcileOpenPolicyAgent(instance *controllersv1alpha1.OpenPolicyAgent) error {
	// ignore is defaulted to enabled
	ignoreEnabled := true
	if instance.Spec.EnabledFailureModes != nil {
		ignoreEnabled = instance.Spec.EnabledFailureModes.Ignore
	}

	if ignoreEnabled {
		if err := r.reconcileOpenPolicyAgentWithFailurePolicy(instance, "Ignore"); err != nil {
			return errors.Wrap(err, "reconcile opa with failurepolicy")
		}
	} else {
		if err := r.deleteOpenPolicyAgentWithFailurePolicy(instance, "Ignore"); err != nil {
			return errors.Wrap(err, "delete opa with failure policy")
		}
	}

	// failure mode is not default
	failEnabled := false
	if instance.Spec.EnabledFailureModes != nil {
		failEnabled = instance.Spec.EnabledFailureModes.Fail
	}

	if failEnabled {
		if err := r.reconcileOpenPolicyAgentWithFailurePolicy(instance, "Fail"); err != nil {
			return errors.Wrap(err, "reconcile opa with failurepolicy")
		}
	} else {
		if err := r.deleteOpenPolicyAgentWithFailurePolicy(instance, "Fail"); err != nil {
			return errors.Wrap(err, "delete opa with failure policy")
		}
	}

	return nil
}

func (r *ReconcileOpenPolicyAgent) deleteOpenPolicyAgentWithFailurePolicy(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	debug := level.Info(log.With(r.Logger, "method", "deleteOpenPolicyAgentWithFailurePolicy"))
	debug.Log("event", "ensuring open policy agent is deleted", "name", instance.Name, "failurePolicy", failurePolicy)

	if err := r.deleteOpenPolicyAgentValidatingWebhook(instance, failurePolicy); err != nil {
		return errors.Wrap(err, "delete validating webhook")
	}

	if err := r.deleteOpenPolicyAgentDeployment(instance, failurePolicy); err != nil {
		return errors.Wrap(err, "delete opa deployment")
	}

	if err := r.deleteOpenPolicyAgentService(instance, failurePolicy); err != nil {
		return errors.Wrap(err, "delete opa service")
	}

	if err := r.deleteOpenPolicyAgentSecret(instance, failurePolicy); err != nil {
		return errors.Wrap(err, "delete tls with failure policy")
	}

	return nil
}

func (r *ReconcileOpenPolicyAgent) reconcileOpenPolicyAgentWithFailurePolicy(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	debug := level.Info(log.With(r.Logger, "method", "reconcileOpenPolicyAgentWithFailurePolicy"))
	debug.Log("event", "ensuring open policy agent is present", "name", instance.Name, "failurePolicy", failurePolicy)

	if err := r.reconcileOpenPolicyAgentSecret(instance, failurePolicy); err != nil {
		return errors.Wrap(err, "reconcile tls with failurepolicy")
	}

	if err := r.reconcileOpenPolicyAgentService(instance, failurePolicy); err != nil {
		return errors.Wrap(err, "reconcile opa service")
	}

	if err := r.reconcileOpenPolicyAgentDeployment(instance, failurePolicy); err != nil {
		return errors.Wrap(err, "reconcile opa deployment")
	}

	reconcileValidatingWebhook, err := r.isOpenPolicyAgentDeploymentReady(instance, failurePolicy)
	if err != nil {
		return errors.Wrap(err, "check openpolicy agent deployment status")
	}

	if reconcileValidatingWebhook {
		if err := r.reconcileOpenPolicyAgentValidatingWebhook(instance, failurePolicy); err != nil {
			return errors.Wrap(err, "reconcile opa validatingwebhook")
		}
	}

	return nil
}

func (r *ReconcileOpenPolicyAgent) deleteOpenPolicyAgentSecret(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	secretName := strings.ToLower(fmt.Sprintf("%s-%s", secretNamePrefix, failurePolicy))

	foundSecret := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, foundSecret)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "get secret")
	} else if err != nil && apierrors.IsNotFound(err) {
		return nil
	}

	err = r.Delete(context.TODO(), foundSecret)
	if err != nil {
		return errors.Wrap(err, "delete service")
	}

	return nil
}

// reconcileTLSWithFailurePolicy will create a ca and cert secret for the instance with the failure policy
func (r *ReconcileOpenPolicyAgent) reconcileOpenPolicyAgentSecret(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	debug := level.Info(log.With(r.Logger, "method", "reconcileOpenPolicyAgent.reconcileTLS"))

	secretName := strings.ToLower(fmt.Sprintf("%s-%s", secretNamePrefix, failurePolicy))

	found := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		secret := &corev1.Secret{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "Secret",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: instance.Namespace,
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
		cert, key, err := gatekeepertls.CreateCertFromCA(r.Logger, instance.Namespace, strings.ToLower(fmt.Sprintf("%s-%s", serviceNamePrefix, failurePolicy)), caCert, caKey)
		if err != nil {
			return errors.Wrap(err, "create cert from ca")
		}

		secret.Data["tls.crt"] = cert
		secret.Data["tls.key"] = key

		if err := controllerutil.SetControllerReference(instance, secret, r.scheme); err != nil {
			return err
		}

		debug.Log("event", "create tls secret")
		err = r.Create(context.TODO(), secret)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// TODO reconcile isn't simply a deepequal here, we need to check the tls cert expiration

	return nil
}

func (r *ReconcileOpenPolicyAgent) deleteOpenPolicyAgentService(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	serviceName := strings.ToLower(fmt.Sprintf("%s-%s", serviceNamePrefix, failurePolicy))

	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: instance.Namespace}, foundService)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "get service")
	} else if err != nil && apierrors.IsNotFound(err) {
		return nil
	}

	err = r.Delete(context.TODO(), foundService)
	if err != nil {
		return errors.Wrap(err, "delete service")
	}

	return nil
}

func (r *ReconcileOpenPolicyAgent) reconcileOpenPolicyAgentService(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	debug := level.Info(log.With(r.Logger, "method", "reconcileOpenPolicyAgent.reconcileOpenPolicyAgentService"))

	serviceName := strings.ToLower(fmt.Sprintf("%s-%s", serviceNamePrefix, failurePolicy))

	// Service
	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: instance.Namespace,
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
				"gatekeeper":    instance.Name,
				"controller":    "openPolicyAgent",
				"failurePolicy": failurePolicy,
			},
		},
	}

	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: serviceName, Namespace: instance.Namespace}, foundService)
	if err != nil && apierrors.IsNotFound(err) {
		debug.Log("event", "creating service")

		if err := controllerutil.SetControllerReference(instance, service, r.scheme); err != nil {
			return err
		}

		err := r.Create(context.TODO(), service)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// TODO compare ignoring the generated fields

	return nil
}

func (r *ReconcileOpenPolicyAgent) isOpenPolicyAgentDeploymentReady(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) (bool, error) {
	debug := level.Info(log.With(r.Logger, "method", "reconcileOpenPolicyAgent.isOpenPolicyAgentDeploymentReady"))

	deploymentName := strings.ToLower(fmt.Sprintf("%s-%s", instance.Name, failurePolicy))
	secretName := strings.ToLower(fmt.Sprintf("%s-%s", secretNamePrefix, failurePolicy))

	foundSecret := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, foundSecret)
	if err != nil {
		return false, errors.Wrap(err, "get secret")
	}

	foundDeployment := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: instance.Namespace}, foundDeployment)
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

	serviceName := strings.ToLower(fmt.Sprintf("%s-%s", serviceNamePrefix, failurePolicy))
	uri := fmt.Sprintf("https://%s.%s.svc/v1/policies/main", serviceName, instance.Namespace)

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

func (r *ReconcileOpenPolicyAgent) deleteOpenPolicyAgentDeployment(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	deploymentName := strings.ToLower(fmt.Sprintf("%s-%s", instance.Name, failurePolicy))

	foundDeployment := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: instance.Namespace}, foundDeployment)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "get deployment")
	} else if err != nil && apierrors.IsNotFound(err) {
		return nil
	}

	err = r.Delete(context.TODO(), foundDeployment)
	if err != nil {
		return errors.Wrap(err, "delete deployment")
	}

	return nil
}

func (r *ReconcileOpenPolicyAgent) reconcileOpenPolicyAgentDeployment(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	debug := level.Info(log.With(r.Logger, "method", "reconcileOpenPolicyAgent.reconcileOpenPolicyAgentDeployment"))

	deploymentName := strings.ToLower(fmt.Sprintf("%s-%s", instance.Name, failurePolicy))
	secretName := strings.ToLower(fmt.Sprintf("%s-%s", secretNamePrefix, failurePolicy))

	// Deployment
	labels := map[string]string{
		"app":           deploymentName,
		"gatekeeper":    instance.Name,
		"controller":    "openPolicyAgent",
		"failurePolicy": failurePolicy,
	}

	replicas := int32(1)

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   deploymentName,
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
									SecretName: secretName,
								},
							},
						},
					},
				},
			},
		},
	}

	foundDeployment := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: instance.Namespace}, foundDeployment)
	if err != nil && apierrors.IsNotFound(err) {
		debug.Log("event", "creating deployment")

		if err := controllerutil.SetControllerReference(instance, deployment, r.scheme); err != nil {
			return err
		}

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

func (r *ReconcileOpenPolicyAgent) deleteOpenPolicyAgentValidatingWebhook(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	webhookName := strings.ToLower(fmt.Sprintf("%s-%s", instance.Name, failurePolicy))

	foundValidatingWebhookConfiguration := &admissionregistrationv1beta1.ValidatingWebhookConfiguration{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: webhookName, Namespace: ""}, foundValidatingWebhookConfiguration)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "get validating webhook")
	} else if err != nil && apierrors.IsNotFound(err) {
		return nil
	}

	err = r.Delete(context.TODO(), foundValidatingWebhookConfiguration)
	if err != nil {
		return errors.Wrap(err, "delete validating webhook")
	}

	return nil
}

func (r *ReconcileOpenPolicyAgent) reconcileOpenPolicyAgentValidatingWebhook(instance *controllersv1alpha1.OpenPolicyAgent, failurePolicy string) error {
	debug := level.Info(log.With(r.Logger, "method", "reconcileOpenPolicyAgent.reconcileOpaValidatingWebhook"))

	secretName := strings.ToLower(fmt.Sprintf("%s-%s", secretNamePrefix, failurePolicy))
	serviceNmae := strings.ToLower(fmt.Sprintf("%s-%s", serviceNamePrefix, failurePolicy))
	webhookName := strings.ToLower(fmt.Sprintf("%s-%s", instance.Name, failurePolicy))
	var policy admissionregistrationv1beta1.FailurePolicyType

	if failurePolicy == "Ignore" {
		policy = admissionregistrationv1beta1.Ignore
	} else if failurePolicy == "Fail" {
		policy = admissionregistrationv1beta1.Fail
	}

	// Read the CA bundle from the secret
	tlsSecret := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, tlsSecret)
	if err != nil {
		return errors.Wrap(err, "get tls secret")
	}

	validatingWebhookConfiguration := &admissionregistrationv1beta1.ValidatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ValidatingWebhookConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []admissionregistrationv1beta1.Webhook{
			{
				Name:          "validating-webhook.openpolicyagent.org",
				FailurePolicy: &policy,
				Rules: []admissionregistrationv1beta1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1beta1.OperationType{admissionregistrationv1beta1.Create, admissionregistrationv1beta1.Update},
						Rule: admissionregistrationv1beta1.Rule{
							APIGroups:   []string{"*"},
							APIVersions: []string{"*"},
							Resources:   []string{"*"},
						},
					},
				},
				ClientConfig: admissionregistrationv1beta1.WebhookClientConfig{
					Service: &admissionregistrationv1beta1.ServiceReference{
						Namespace: instance.Namespace,
						Name:      serviceNmae,
					},
					CABundle: tlsSecret.Data["ca.crt"],
				},
			},
		},
	}

	foundValidatingWebhookConfiguration := &admissionregistrationv1beta1.ValidatingWebhookConfiguration{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: webhookName, Namespace: ""}, foundValidatingWebhookConfiguration)
	if err != nil && apierrors.IsNotFound(err) {
		debug.Log("event", "creating validataing webhook configuration")

		if err := controllerutil.SetControllerReference(instance, validatingWebhookConfiguration, r.scheme); err != nil {
			return err
		}

		err := r.Create(context.TODO(), validatingWebhookConfiguration)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// TODO compare ignoring the generated fields

	return nil
}

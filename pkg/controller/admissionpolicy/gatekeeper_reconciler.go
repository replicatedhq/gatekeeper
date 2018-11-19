package admissionpolicy

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	gatekeepertls "github.com/replicatedhq/gatekeeper/pkg/tls"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *ReconcileAdmissionPolicy) ensureProxyRunning() error {
	debug := level.Info(log.With(r.Logger, "method", "ensureProxyRunning"))
	debug.Log("event", "ensure proxy running")

	if err := r.reconcileGatekeeperSecret(); err != nil {
		return errors.Wrap(err, "reconcile gaatekeeper secret")
	}
	if err := r.reconcileGatekeeperService(); err != nil {
		return errors.Wrap(err, "reconcile gatekeeper service")
	}
	if err := r.reconcileGatekeeperDeployment(); err != nil {
		return errors.Wrap(err, "reconcile gatekeeper deployment")
	}
	if err := r.waitForGatekeeperDeploymentReady(); err != nil {
		return errors.Wrap(err, "is gatekeeper ready")
	}
	if err := r.reconcileGatekeeperValidatingWebhook(); err != nil {
		return errors.Wrap(err, "reconsile webhook")
	}

	return nil
}

func (r *ReconcileAdmissionPolicy) reconcileGatekeeperSecret() error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.reconcileGatekeeperSecret"))
	debug.Log("event", "reconciling gatekeeper secret")

	secretName, err := gatekeeperSecretName()
	if err != nil {
		return errors.Wrap(err, "get secret name")
	}
	serviceName, err := gatekeeperServiceName()
	if err != nil {
		return errors.Wrap(err, "get service name")
	}

	found := &corev1.Secret{}
	err = r.Get(context.TODO(), secretName, found)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "get secret")
	} else if apierrors.IsNotFound(err) {
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
	}

	return nil
}

func (r *ReconcileAdmissionPolicy) reconcileGatekeeperService() error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.reconcileGatekeeperService"))
	debug.Log("event", "reconciling gatekeeper service")

	serviceName, err := gatekeeperServiceName()
	if err != nil {
		return errors.Wrap(err, "get service name")
	}

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName.Name,
			Namespace: serviceName.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Name:     "https",
					Protocol: corev1.ProtocolTCP,
					Port:     443,
					TargetPort: intstr.IntOrString{
						IntVal: 8000,
					},
				},
			},
			Selector: map[string]string{
				"app":        "gatekeeper",
				"controller": "gatekeeper",
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

func (r *ReconcileAdmissionPolicy) reconcileGatekeeperDeployment() error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.reconcileGatekeeperDeployment"))
	debug.Log("event", "reconciling opa deployment")

	secretName, err := gatekeeperSecretName()
	if err != nil {
		return errors.Wrap(err, "get secret name")
	}
	deploymentName, err := gatekeeperDeploymentName()
	if err != nil {
		return errors.Wrap(err, "get deployment name")
	}

	// Deployment
	labels := map[string]string{
		"app":        "gatekeeper",
		"controller": "gatekeeper",
	}

	replicas := int32(1)

	// Get the image name from the manager
	// bonus side effect, the logs get picked up in skaffold from this
	imageName, err := r.getManagerImageName()
	if err != nil {
		return errors.Wrap(err, "get manager image name")
	}

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
							Name:            "gatekeeper",
							Image:           imageName,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/root/gatekeeper",
							},
							Args: []string{
								"proxy",
								"--tls-cert-file=/certs/tls.crt",
								"--tls-key-file=/certs/tls.key",
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
									ContainerPort: 8000,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "LOG_LEVEL",
									Value: os.Getenv("LOG_LEVEL"),
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

func (r *ReconcileAdmissionPolicy) waitForGatekeeperDeploymentReady() error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.waitForGatekeeperDeploymentReady"))
	debug.Log("event", "waiting for gatekeeper to report ready")

	abortAt := time.Now().Add(time.Minute * 2)
	for {
		if time.Now().After(abortAt) {
			return errors.Wrap(fmt.Errorf("timeout waiting for gatekeeper to be ready"), "waiting for gatekeeper")
		}

		isReady, err := r.isGatekeeperDeploymentReady()
		if err != nil {
			return errors.Wrap(err, "is gatekeeper ready")
		}

		if isReady {
			return nil
		}
	}
}
func (r *ReconcileAdmissionPolicy) isGatekeeperDeploymentReady() (bool, error) {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.isGatekeeperDeploymentReady"))
	debug.Log("event", "check if gatekeeper to report ready")

	secretName, err := gatekeeperSecretName()
	if err != nil {
		return false, errors.Wrap(err, "get secret name")
	}
	deploymentName, err := gatekeeperDeploymentName()
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

	return foundDeployment.Status.AvailableReplicas > 0, nil
}

func (r *ReconcileAdmissionPolicy) reconcileGatekeeperValidatingWebhook() error {
	debug := level.Info(log.With(r.Logger, "method", "ReconcileAdmissionPolicy.reconcileGatekeeperValidatingWebhook"))
	debug.Log("event", "reconciling opa validatingwebhook")

	secretName, err := gatekeeperSecretName()
	if err != nil {
		return errors.Wrap(err, "get secret name")
	}
	serviceName, err := gatekeeperServiceName()
	if err != nil {
		return errors.Wrap(err, "get service name")
	}
	webhookName := getkeeperWebhookName()

	policy := admissionregistrationv1beta1.Fail

	// Read the CA bundle from the secret
	tlsSecret := &corev1.Secret{}
	err = r.Get(context.TODO(), secretName, tlsSecret)
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
				Name:          "validating-webhook.gatekeeper.sh",
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
						Namespace: serviceName.Namespace,
						Name:      serviceName.Name,
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

func (r *ReconcileAdmissionPolicy) getManagerImageName() (string, error) {
	managerName, err := managerName()
	if err != nil {
		return "", errors.Wrap(err, "get manager name")
	}

	statefulset := &appsv1.StatefulSet{}
	if err := r.Get(context.TODO(), managerName, statefulset); err != nil {
		return "", errors.Wrap(err, "get stateful set")
	}

	// this is brittle
	return statefulset.Spec.Template.Spec.Containers[0].Image, nil
}

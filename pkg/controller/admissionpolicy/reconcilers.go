package admissionpolicy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/pkg/errors"
	controllersv1alpha1 "github.com/replicatedhq/gatekeeper/pkg/apis/controllers/v1alpha1"
	policiesv1alpha1 "github.com/replicatedhq/gatekeeper/pkg/apis/policies/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	serviceNamePrefix    = "gatekeeper-opa"
	secretNamePrefix     = "gatekeeper"
	deploymentNamePrefix = "gatekeeper"
)

func (r *ReconcileAdmissionPolicy) reconcileAdmissionPolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	if err := r.validatePolicy(instance); err != nil {
		return errors.Wrap(err, "validate policy")
	}

	// Deploy any required OPA instances referenced by this policy
	if err := r.ensureOPARunningForPolicy(instance); err != nil {
		return errors.Wrap(err, "ensure opa running for policy")
	}

	if err := r.applyPolicy(instance); err != nil {
		return errors.Wrap(err, "apply policy")
	}

	return nil
}

func (r *ReconcileAdmissionPolicy) validatePolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	// TODO
	return nil
}

func (r *ReconcileAdmissionPolicy) ensureOPARunningForPolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "ensureOPARunningForPolicy"))
	debug.Log("event", "ensure opa instance running", "failurePolicy", instance.Spec.FailurePolicy)

	deploymentName := strings.ToLower(fmt.Sprintf("%s-%s", deploymentNamePrefix, instance.Spec.FailurePolicy))

	foundDeployment := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: instance.Namespace}, foundDeployment)
	if err != nil && !apierrors.IsNotFound(err) {
		return errors.Wrap(err, "find deployment for opa instance")
	}

	if err == nil {
		if foundDeployment.Status.AvailableReplicas > 0 {
			return nil
		}

		abortAt := time.Now().Add(time.Minute * 2)
		for {
			if time.Now().After(abortAt) {
				return errors.Wrap(fmt.Errorf("timeout waiting for opa to be ready"), "waiting for opa")
			}

			foundDeployment := &appsv1.Deployment{}
			err := r.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: instance.Namespace}, foundDeployment)
			if err != nil && !apierrors.IsNotFound(err) {
				return errors.Wrap(err, "find deployment for opa instance")
			}

			// If the deployment isn't found, continue and check again
			if apierrors.IsNotFound(err) {
				continue
			}

			if foundDeployment.Status.AvailableReplicas > 0 {
				return nil
			}

		}
	}

	// Create the opa instance...
	// ... using the _other_ controller, that we probably shouldn't assume exists, but we do for now

	openPolicyAgent := &controllersv1alpha1.OpenPolicyAgent{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1alpha1",
			Kind:       "OpenPolicyAgent",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentNamePrefix, // the controller will add a suffix
			Namespace: instance.Namespace,
		},
		Spec: controllersv1alpha1.OpenPolicyAgentSpec{
			Name: deploymentName,
			EnabledFailureModes: &controllersv1alpha1.OpenPolicyAgentEnabledFailureModes{
				Ignore: instance.Spec.FailurePolicy == "Ignore",
				Fail:   instance.Spec.FailurePolicy == "Fail",
			},
		},
	}
	if err = r.Create(context.TODO(), openPolicyAgent); err != nil {
		return errors.Wrap(err, "creating open policy agent")
	}

	// block until this has been created
	abortAt := time.Now().Add(time.Minute * 2)
	for {
		if time.Now().After(abortAt) {
			return errors.Wrap(fmt.Errorf("timeout waiting for opa to be created"), "waiting for opa")
		}

		debug.Log("event", "polling for deployment to be ready")
		foundDeployment := &appsv1.Deployment{}
		err := r.Get(context.TODO(), types.NamespacedName{Name: deploymentName, Namespace: instance.Namespace}, foundDeployment)
		if err != nil && !apierrors.IsNotFound(err) {
			return errors.Wrap(err, "find deployment for opa instance")
		}

		// If the deployment isn't found, continue and check again
		if apierrors.IsNotFound(err) {
			time.Sleep(time.Second)
			continue
		}

		debug.Log("event", "found deployment", "available replicas", foundDeployment.Status.AvailableReplicas)
		if foundDeployment.Status.AvailableReplicas > 0 {
			return nil
		}

		time.Sleep(time.Second)
	}
}

func (r *ReconcileAdmissionPolicy) buildOpaURIForFailurePolicy(instance *policiesv1alpha1.AdmissionPolicy) string {
	suffix := strings.ToLower(instance.Spec.FailurePolicy)
	serviceName := fmt.Sprintf("%s-%s.%s.svc", serviceNamePrefix, suffix, instance.Namespace)

	return fmt.Sprintf("https://%s/v1/policies/%s", serviceName, instance.Spec.Name)
}

func (r *ReconcileAdmissionPolicy) applyPolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "applyPolicy"))
	debug.Log("event", "applyPolicy", "name", instance.Name)

	opaURI := r.buildOpaURIForFailurePolicy(instance)

	// Get the CA from the secret so we can communicate
	secretName := strings.ToLower(fmt.Sprintf("%s-%s", secretNamePrefix, instance.Spec.FailurePolicy))
	foundSecret := &corev1.Secret{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: instance.Namespace}, foundSecret)
	if err != nil {
		return errors.Wrap(err, "get secret")
	}

	rootCAs, err := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if ok := rootCAs.AppendCertsFromPEM(foundSecret.Data["ca.crt"]); !ok {
		return errors.Wrapf(err, "append ca cert")
	}
	config := &tls.Config{
		RootCAs: rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: config}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second,
	}
	req, err := http.NewRequest("PUT", opaURI, strings.NewReader(instance.Spec.Policy))
	if err != nil {
		return errors.Wrap(err, "create request policy")
	}
	req.ContentLength = int64(len(instance.Spec.Policy))
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "create policy")
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read error body")
		}

		level.Warn(r.Logger).Log("message", string(body))
		return errors.Wrap(fmt.Errorf("unexpected status code %d", resp.StatusCode), "create policy response")
	}

	return nil
}

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
	policiesv1alpha1 "github.com/replicatedhq/gatekeeper/pkg/apis/policies/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func (r *ReconcileAdmissionPolicy) reconcileAdmissionPolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	if err := r.validatePolicy(instance); err != nil {
		return errors.Wrap(err, "validate policy")
	}

	// Ensure the Gatekeeper proxy is running
	if err := r.ensureProxyRunning(); err != nil {
		return errors.Wrap(err, "ensure gatekeeper proxy running")
	}

	// Deploy any required OPA instances referenced by this policy
	if err := r.ensureOPARunningForPolicy(instance); err != nil {
		return errors.Wrap(err, "ensure opa running for policy")
	}

	// Reconcile the actual policy
	if err := r.applyPolicy(instance); err != nil {
		return errors.Wrap(err, "apply policy")
	}

	return nil
}

func (r *ReconcileAdmissionPolicy) validatePolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	// TODO
	return nil
}

func (r *ReconcileAdmissionPolicy) buildOPAUri(instance *policiesv1alpha1.AdmissionPolicy) (string, error) {
	serviceName, err := opaServiceName(instance.Spec.FailurePolicy)
	if err != nil {
		return "", errors.Wrap(err, "get service name")
	}
	return fmt.Sprintf("https://%s.%s.svc/v1/policies/%s", serviceName.Name, serviceName.Namespace, instance.Spec.Name), nil
}

func (r *ReconcileAdmissionPolicy) applyPolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	debug := level.Info(log.With(r.Logger, "method", "applyPolicy"))
	debug.Log("event", "applyPolicy", "name", instance.Name)

	opaURI, err := r.buildOPAUri(instance)
	if err != nil {
		return errors.Wrap(err, "get url")
	}

	// Get the CA from the secret so we can communicate
	secretName, err := opaSecretName(instance.Spec.FailurePolicy)
	if err != nil {
		return errors.Wrap(err, "get secret name")
	}

	foundSecret := &corev1.Secret{}
	err = r.Get(context.TODO(), secretName, foundSecret)
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

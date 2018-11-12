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

	"github.com/go-kit/kit/log/level"

	"github.com/pkg/errors"
	policiesv1alpha1 "github.com/replicatedhq/gatekeeper/pkg/apis/policies/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	serviceNamePrefix = "gatekeeper-opa"
	secretNamePrefix  = "gatekeeper"
)

func (r *ReconcileAdmissionPolicy) reconcileAdmissionPolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
	if err := r.validatePolicy(instance); err != nil {
		return errors.Wrap(err, "validate policy")
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

func (r *ReconcileAdmissionPolicy) buildOpaURIForFailurePolicy(instance *policiesv1alpha1.AdmissionPolicy) string {
	suffix := strings.ToLower(instance.Spec.FailurePolicy)
	serviceName := fmt.Sprintf("%s-%s.%s.svc", serviceNamePrefix, suffix, instance.Namespace)

	return fmt.Sprintf("https://%s/v1/policies/%s", serviceName, instance.Spec.Name)
}

func (r *ReconcileAdmissionPolicy) applyPolicy(instance *policiesv1alpha1.AdmissionPolicy) error {
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

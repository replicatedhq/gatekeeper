package admissionpolicy

import (
	"fmt"
	"strings"

	confighelper "admiralty.io/multicluster-service-account/pkg/config"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	serviceNameFormat    = "gatekeeper-opa-%s"
	secretNameFormat     = "gatekeeper-opa-%s"
	deploymentNameFormat = "gatekeeper-opa-%s"
	webhookNameFormat    = "gatekeeper-opa-%s"
)

func opaSecretName(failurePolicy string) (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      strings.ToLower(fmt.Sprintf(secretNameFormat, failurePolicy)),
		Namespace: ns,
	}, nil
}

func opaServiceName(failurePolicy string) (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      strings.ToLower(fmt.Sprintf(serviceNameFormat, failurePolicy)),
		Namespace: ns,
	}, nil
}

func opaDeploymentName(failurePolicy string) (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      strings.ToLower(fmt.Sprintf(deploymentNameFormat, failurePolicy)),
		Namespace: ns,
	}, nil
}

func opaWebhookName(failurePolicy string) string {
	return strings.ToLower(fmt.Sprintf(webhookNameFormat, failurePolicy))
}

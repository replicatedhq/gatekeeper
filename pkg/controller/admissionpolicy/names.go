package admissionpolicy

import (
	"strings"

	confighelper "admiralty.io/multicluster-service-account/pkg/config"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	serviceNameFormat    = "gatekeeper-opa"
	secretNameFormat     = "gatekeeper-opa"
	deploymentNameFormat = "gatekeeper-opa"
	webhookNameFormat    = "gatekeeper-opa"
)

func managerName() (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      "gatekeeper-controller-manager",
		Namespace: ns,
	}, nil
}

func gatekeeperSecretName() (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      "gatekeeper-tls",
		Namespace: ns,
	}, nil
}

func opaSecretName() (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      strings.ToLower(secretNameFormat),
		Namespace: ns,
	}, nil
}

func gatekeeperServiceName() (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      "gatekeeper",
		Namespace: ns,
	}, nil
}

func opaServiceName() (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      strings.ToLower(serviceNameFormat),
		Namespace: ns,
	}, nil
}

func gatekeeperDeploymentName() (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      "gatekeeper",
		Namespace: ns,
	}, nil
}

func opaDeploymentName() (types.NamespacedName, error) {
	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return types.NamespacedName{}, errors.Wrap(err, "config and namespace")
	}

	return types.NamespacedName{
		Name:      strings.ToLower(deploymentNameFormat),
		Namespace: ns,
	}, nil
}

func getkeeperWebhookName() string {
	return "gatekeeper"
}

func opaWebhookName() string {
	return strings.ToLower(webhookNameFormat)
}

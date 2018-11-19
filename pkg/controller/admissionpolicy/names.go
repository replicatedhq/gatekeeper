package admissionpolicy

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/types"
)

const (
	serviceNameFormat    = "gatekeeper-opa-%s"
	secretNameFormat     = "gatekeeper-opa-%s"
	deploymentNameFormat = "gatekeeper-opa-%s"
	webhookNameFormat    = "gatekeeper-opa-%s"
)

func opaSecretName(failurePolicy string) types.NamespacedName {
	return types.NamespacedName{
		Name:      strings.ToLower(fmt.Sprintf(secretNameFormat, failurePolicy)),
		Namespace: os.Getenv("POD_NAMESPACE"),
	}
}

func opaServiceName(failurePolicy string) types.NamespacedName {
	return types.NamespacedName{
		Name:      strings.ToLower(fmt.Sprintf(serviceNameFormat, failurePolicy)),
		Namespace: os.Getenv("POD_NAMESPACE"),
	}
}

func opaDeploymentName(failurePolicy string) types.NamespacedName {
	return types.NamespacedName{
		Name:      strings.ToLower(fmt.Sprintf(deploymentNameFormat, failurePolicy)),
		Namespace: os.Getenv("POD_NAMESPACE"),
	}
}

func opaWebhookName(failurePolicy string) string {
	return strings.ToLower(fmt.Sprintf(webhookNameFormat, failurePolicy))
}

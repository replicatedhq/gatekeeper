package gatekeeper

import (
	"github.com/go-kit/kit/log"
	"github.com/replicatedhq/gatekeeper/pkg/client/gatekeeperclientset"
	"github.com/replicatedhq/gatekeeper/pkg/config"
	"k8s.io/client-go/kubernetes"
)

type Gatekeeper struct {
	Config *config.Config
	Logger log.Logger

	K8sClient           kubernetes.Interface
	GatekeeperK8sClient gatekeeperclientset.Interface
}

func NewGatekeeper(logger log.Logger, config *config.Config, k8sClient kubernetes.Interface, gatekeeperK8sClient gatekeeperclientset.Interface) (*Gatekeeper, error) {
	return &Gatekeeper{
		Config:              config,
		Logger:              logger,
		K8sClient:           k8sClient,
		GatekeeperK8sClient: gatekeeperK8sClient,
	}, nil
}

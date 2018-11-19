package proxy

import (
	"io"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/replicatedhq/gatekeeper/pkg/config"
	"github.com/replicatedhq/gatekeeper/pkg/kubernetes"
	"github.com/replicatedhq/gatekeeper/pkg/logger"
	"go.uber.org/dig"
)

func buildInjector(c *config.Config, out io.Writer) (*dig.Container, error) {
	providers := []interface{}{
		func() *config.Config {
			return c
		},
		func() io.Writer {
			return out
		},
		logger.New,

		kubernetes.NewClient,
		NewProxy,
	}

	container := dig.New()

	for _, provider := range providers {
		err := container.Provide(provider)
		if err != nil {
			return nil, errors.Wrap(err, "register providers")
		}
	}

	return container, nil
}

func Get(c *config.Config, out io.Writer) (*GatekeeperProxy, error) {
	debug := log.With(level.Debug(logger.New()), "component", "injector", "phase", "instance.get")

	debug.Log("event", "injector.build")
	injector, err := buildInjector(c, out)
	if err != nil {
		debug.Log("event", "injector.build.fail", "error", err)
		return nil, errors.Wrap(err, "build injector")
	}

	var gatekeeperProxy *GatekeeperProxy

	debug.Log("event", "injector.invoke")
	if err := injector.Invoke(func(g *GatekeeperProxy) {
		debug.Log("event", "injector.invoke.resolve")
		gatekeeperProxy = g
	}); err != nil {
		debug.Log("event", "injector.invoke.fail", "err", err)
		return nil, errors.Wrap(err, "resolve deps")
	}

	return gatekeeperProxy, nil
}

package proxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	confighelper "admiralty.io/multicluster-service-account/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/replicatedhq/gatekeeper/pkg/config"
	"github.com/spf13/viper"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type GatekeeperProxy struct {
	Config    *config.Config
	Logger    log.Logger
	K8sClient kubernetes.Interface
}

type UpstreamValidatingWebhook struct {
	CABundle []byte
	URI      string
}

func NewProxy(logger log.Logger, config *config.Config, k8sclient kubernetes.Interface) (*GatekeeperProxy, error) {
	return &GatekeeperProxy{
		Config:    config,
		Logger:    logger,
		K8sClient: k8sclient,
	}, nil
}

func (p *GatekeeperProxy) Serve(ctx context.Context) error {
	debug := level.Info(log.With(p.Logger, "method", "GatekeeperProxy.Serve"))

	v := viper.GetViper()

	g := gin.New()
	p.configureRoutes(g)

	errCh := make(chan error)
	if v.GetBool("enable-tls") {
		debug.Log("event", "listen and serve tls")
		errCh <- http.ListenAndServeTLS(p.Config.ProxyAddress, v.GetString("tls-cert-file"), v.GetString("tls-key-file"), g)
	} else {
		debug.Log("event", "listen and serve")
		server := http.Server{Addr: p.Config.ProxyAddress, Handler: g}
		errCh <- server.ListenAndServe()
	}

	return nil
}

func (p *GatekeeperProxy) configureRoutes(g *gin.Engine) {
	root := g.Group("/")

	root.GET("/healthz", p.Healthz)
	root.POST("/", p.AdmissionRequest)
}

func (p *GatekeeperProxy) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		// TODO
	})
}

func (p GatekeeperProxy) AdmissionRequest(c *gin.Context) {
	warn := level.Warn(log.With(p.Logger, "method", "GatekeeperProxy.AdmissionRequest"))
	debug := level.Info(log.With(p.Logger, "method", "GatekeeperProxy.AdmissionRequest"))
	debug.Log("event", "admission request")

	admissionRequest := admissionv1beta1.AdmissionReview{}
	if err := c.BindJSON(&admissionRequest); err != nil {
		level.Warn(p.Logger).Log("event", "unable to bind request", "err", err)
		return
	}

	b, err := json.Marshal(admissionRequest)
	if err != nil {
		level.Warn(p.Logger).Log("event", "unable to marshall request", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	body := string(b)

	// Send this to the OPAs that are configured
	upstreams, err := p.listGatekeeperManagedOPAs()
	if err != nil {
		level.Error(p.Logger).Log("event", "listGatekeeperManagedOPAs", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Some defaults
	allowed := true
	status := &metav1.Status{
		Status:  "Gatekeeper Reviewed",
		Message: "Reviewed by Gatekeeper",
		Code:    204,
	}

	decisions := make([]admissionv1beta1.AdmissionReview, 0, 0)

	for _, upstream := range upstreams {
		debug.Log("event", "sending upstream...", "upstream", upstream.URI)

		rootCAs, err := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		if ok := rootCAs.AppendCertsFromPEM(upstream.CABundle); !ok {
			warn.Log("event", "append certs from pem", "err", "not ok")
			continue
		}
		config := &tls.Config{
			RootCAs: rootCAs,
		}
		tr := &http.Transport{TLSClientConfig: config}
		client := &http.Client{
			Transport: tr,
			Timeout:   time.Second,
		}
		req, err := http.NewRequest("POST", upstream.URI, strings.NewReader(body))
		if err != nil {
			warn.Log("event", "build request to post to upstream", "err", err)
			continue
		}
		req.ContentLength = int64(len(body))
		resp, err := client.Do(req)
		if err != nil {
			warn.Log("event", "send request upstream", "err", err)
			continue
		}

		defer resp.Body.Close()
		var decision admissionv1beta1.AdmissionReview
		if err := json.NewDecoder(resp.Body).Decode(&decision); err != nil {
			warn.Log("event", "error decoding response", "err", err)
			continue
		}

		decisions = append(decisions, decision)

		if !decision.Response.Allowed {
			allowed = false
			status = decision.Response.Result
		}

		// TODO increment counters on the policy
		debug.Log("event", "admission decision", "allowed", allowed)
	}

	// Pick a response to send...

	response := admissionv1beta1.AdmissionReview{
		Response: &admissionv1beta1.AdmissionResponse{
			Allowed: allowed,
			Result:  status,
		},
	}

	c.JSON(http.StatusOK, response)
}

// TODO this should be cached and an informer
func (p GatekeeperProxy) listGatekeeperManagedOPAs() ([]*UpstreamValidatingWebhook, error) {
	debug := level.Info(log.With(p.Logger, "method", "GatekeeperProxy.listGatekeeperManagedOPAs"))
	debug.Log("event", "list managed opas")

	_, ns, err := confighelper.ConfigAndNamespace()
	if err != nil {
		return nil, errors.Wrap(err, "config and namespace")
	}

	services, err := p.K8sClient.CoreV1().Services(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "list services")
	}

	result := make([]*UpstreamValidatingWebhook, 0, 0)
	for _, service := range services.Items {
		val, ok := service.Labels["app"]
		if !ok || val != "gatekeeper" {
			continue
		}
		val, ok = service.Labels["role"]
		if !ok || val != "openpolicyagent" {
			continue
		}

		caBundleSecretName, ok := service.Labels["caBundleSecret"]
		if !ok {
			debug.Log("event", "ignoring service", "service", service.Name, "reason", "no bundle secret")
			continue
		}
		caBundleSecretNamespace, ok := service.Labels["caBundleNamespace"]
		if !ok {
			debug.Log("event", "ignoring service", "service", service.Name, "reason", "no bundle secret namespace")
			continue
		}

		secret, err := p.K8sClient.CoreV1().Secrets(caBundleSecretNamespace).Get(caBundleSecretName, metav1.GetOptions{})
		if err != nil {
			debug.Log("event", "ignoring service", "service", service.Name, "reason", "cannot read secret", "err", err)
			continue
		}

		upstream := UpstreamValidatingWebhook{
			URI:      fmt.Sprintf("https://%s.%s.svc", service.Name, ns),
			CABundle: secret.Data["ca.crt"],
		}

		result = append(result, &upstream)
	}

	return result, nil
}

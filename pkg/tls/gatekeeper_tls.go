package gatekeepertls

import (
	"fmt"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/signer"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
)

// CreateCertificateAuthority will create a new CA and return the
// pem encoded cert, key and error
func CreateCertificateAuthority(logger log.Logger) ([]byte, []byte, error) {
	debug := level.Info(log.With(logger, "method", "worker.createCertificateAuthority"))

	req := csr.CertificateRequest{
		KeyRequest: &csr.KeyRequest{
			Algo: "rsa",
			Size: 2048,
		},
		CN: "gatekeeper_ca",
		Hosts: []string{
			"gatekeeper_ca",
		},
		CA: &csr.CAConfig{
			Expiry: "8760h",
		},
	}

	debug.Log("event", "create ca")
	cert, key, err := initca.New(&req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "initca")
	}

	return cert, key, nil
}

// CreateCertFromCA takes a certificate authority cert and key and generates a
// new cert, and uses the CA to sign it
func CreateCertFromCA(logger log.Logger, namespacedName types.NamespacedName, caCert []byte, caKey []byte) ([]byte, []byte, error) {
	// Parse the ca into an x509 object
	parsedCaCert, err := helpers.ParseCertificatePEM(caCert)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse cert pem")
	}
	parsedCaKey, err := helpers.ParsePrivateKeyPEM(caKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse key pem")
	}

	req := csr.CertificateRequest{
		KeyRequest: &csr.KeyRequest{
			Algo: "rsa",
			Size: 2048,
		},
		CN: fmt.Sprintf("%s.%s.svc", namespacedName.Name, namespacedName.Namespace),
		Hosts: []string{
			fmt.Sprintf("%s.%s.svc", namespacedName.Name, namespacedName.Namespace),
		},
	}
	certReq, key, err := csr.ParseRequest(&req)
	if err != nil {
		return nil, nil, errors.Wrap(err, "parse csr")
	}

	signReq := signer.NewStandardSigner(parsedCaKey, parsedCaCert, signer.DefaultSigAlgo(parsedCaKey))

	signedCert, err := signReq.Sign(
		fmt.Sprintf("%s.%s.svc", namespacedName.Name, namespacedName.Namespace),
		certReq,
		&signer.Subject{
			CN: fmt.Sprintf("%s.%s.svc", namespacedName.Name, namespacedName.Namespace),
			Hosts: []string{
				fmt.Sprintf("%s.%s.svc", namespacedName.Name, namespacedName.Namespace),
			},
		},
		fmt.Sprintf("%s.%s.svc", namespacedName.Name, namespacedName.Namespace))

	if err != nil {
		return nil, nil, errors.Wrap(err, "sign")
	}

	return signedCert, key, nil
}

/*
Copyright 2018 Replicated.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gatekeepertls

import (
	"crypto/tls"
	"testing"

	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/revoke"
	"github.com/onsi/gomega"
	"github.com/replicatedhq/gatekeeper/pkg/logger"
	"k8s.io/apimachinery/pkg/types"
)

func TestCreateCertificateAuthority(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	logger := logger.New()
	ca, key, err := CreateCertificateAuthority(logger)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	parsedCA, err := helpers.ParseCertificatePEM(ca)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(parsedCA.IsCA).To(gomega.BeTrue())

	revoked, ok := revoke.VerifyCertificate(parsedCA)
	g.Expect(revoked).To(gomega.BeFalse())
	g.Expect(ok).To(gomega.BeTrue())

	_, err = helpers.ParsePrivateKeyPEM(key)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	_, err = tls.X509KeyPair(ca, key)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func TestCreateCertFromCA(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	logger := logger.New()
	caCert, caKey, err := CreateCertificateAuthority(logger)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	parsedCA, err := helpers.ParseCertificatePEM(caCert)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	cert, _, err := CreateCertFromCA(logger, types.NamespacedName{Namespace: "namesspace", Name: "name"}, caCert, caKey)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	parsedCert, err := helpers.ParseCertificatePEM(cert)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(parsedCert.IsCA).To(gomega.BeFalse())

	err = parsedCert.CheckSignatureFrom(parsedCA)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

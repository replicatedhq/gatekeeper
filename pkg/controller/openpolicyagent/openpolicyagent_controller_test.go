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

package openpolicyagent

import (
	"testing"
	"time"

	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/revoke"
	"github.com/onsi/gomega"
	controllersv1alpha1 "github.com/replicatedhq/gatekeeper/pkg/apis/controllers/v1alpha1"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var c client.Client

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: "default"}}
var secretKey = types.NamespacedName{Name: "gatekeeper-ignore", Namespace: "default"}

const timeout = time.Second * 5

func TestReconcileIgnoreOnly(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	instance := &controllersv1alpha1.OpenPolicyAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: controllersv1alpha1.OpenPolicyAgentSpec{
			Name: "name",
			EnabledFailureModes: &controllersv1alpha1.OpenPolicyAgentEnabledFailureModes{
				Ignore: true,
				Fail:   false,
			},
		},
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c = mgr.GetClient()

	recFn, requests := SetupTestReconcile(newReconciler(mgr))
	g.Expect(add(mgr, recFn)).NotTo(gomega.HaveOccurred())

	stopMgr, mgrStopped := StartTestManager(mgr, g)

	defer func() {
		close(stopMgr)
		mgrStopped.Wait()
	}()

	// Create the OpenPolicyAgent object and expect the Reconcile and Deployment to be created
	err = c.Create(context.TODO(), instance)
	// The instance object may not be a valid object because it might be missing some required fields.
	// Please modify the instance object by adding required fields and then remove the following if statement.
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), instance)
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	secret := &corev1.Secret{}
	g.Eventually(func() error { return c.Get(context.TODO(), secretKey, secret) }, timeout).
		Should(gomega.Succeed())

	// Validate that the tls cert was signed by the ca cert that's in the secret
	parsedCaCert, err := helpers.ParseCertificatePEM(secret.Data["ca.crt"])
	g.Expect(err).NotTo(gomega.HaveOccurred())
	revoked, ok := revoke.VerifyCertificate(parsedCaCert)
	g.Expect(revoked).To(gomega.BeFalse())
	g.Expect(ok).To(gomega.BeTrue())

	parsedServerCert, err := helpers.ParseCertificatePEM(secret.Data["tls.crt"])
	g.Expect(err).NotTo(gomega.HaveOccurred())
	revoked, ok = revoke.VerifyCertificate(parsedServerCert)
	g.Expect(revoked).To(gomega.BeFalse())
	g.Expect(ok).To(gomega.BeTrue())

	// Delete the Secret and expect Reconcile to be called for Secret deletion
	g.Expect(c.Delete(context.TODO(), secret)).NotTo(gomega.HaveOccurred())
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	g.Eventually(func() error { return c.Get(context.TODO(), secretKey, secret) }, timeout).
		Should(gomega.Succeed())

	// Manually delete Secret since GC isn't enabled in the test control plane
	//g.Expect(c.Delete(context.TODO(), secret)).To(gomega.Succeed())

}

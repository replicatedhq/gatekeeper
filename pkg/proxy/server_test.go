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

package proxy

import (
	"testing"

	"github.com/onsi/gomega"
	"github.com/replicatedhq/gatekeeper/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpstreamFromObjects(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	service := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "gatekeeper",
				"role": "openpolicyagent",
			},
		},
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"ca.crt": []byte("asdasd"),
		},
	}

	result, err := upstreamFromObjects(logger.New(), service, secret)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(result.URI).To(gomega.Equal("https://foo.default.svc"))
}

func TestFilterServices(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	expectedService := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "expected",
			Namespace: "default",
			Labels: map[string]string{
				"app":  "gatekeeper",
				"role": "openpolicyagent",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "gatekeeper",
				"role": "openpolicyagent",
			},
		},
	}

	unexpectedService := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unexpected",
			Namespace: "default",
			Labels: map[string]string{
				"app":  "gatekeeper",
				"role": "gatekeeper",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":  "gatekeeper",
				"role": "gatekeeper",
			},
		},
	}

	randomService := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "random",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{},
		},
	}

	services := []corev1.Service{
		expectedService,
		unexpectedService,
		randomService,
	}

	filteredServices := filterServices(logger.New(), services)

	g.Expect(filteredServices).To(gomega.HaveLen(1))

	//g.Expect(filteredServices[0].ObjectMeta.Name).To(gomega.Equal("expected"))
}

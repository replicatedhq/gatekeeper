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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OpenPolicyAgentEnabledFailureModes struct {
	Ignore bool `json:"ignore"`
	Fail   bool `json:"fail"`
}

// OpenPolicyAgentSpec defines the desired state of OpenPolicyAgent
type OpenPolicyAgentSpec struct {
	Name                string                              `json:"name"`
	EnabledFailureModes *OpenPolicyAgentEnabledFailureModes `json:"enabledFailureModes"`
}

// OpenPolicyAgentStatus defines the observed state of OpenPolicyAgent
type OpenPolicyAgentStatus struct {
	MainPolicyDeployed bool `json:"mainPolicyDeployed"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenPolicyAgent is the Schema for the openpolicyagents API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type OpenPolicyAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenPolicyAgentSpec   `json:"spec,omitempty"`
	Status OpenPolicyAgentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OpenPolicyAgentList contains a list of OpenPolicyAgent
type OpenPolicyAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenPolicyAgent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenPolicyAgent{}, &OpenPolicyAgentList{})
}

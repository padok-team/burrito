/*
Copyright 2022.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TerraformRepositorySpec defines the desired state of TerraformRepository
type TerraformRepositorySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Repository          TerraformRepositoryRepository          `json:"repository,omitempty"`
	RemediationStrategy TerraformRepositoryRemediationStrategy `json:"remediationStrategy,omitempty"`
	OverrideRunnerSpec  OverrideRunnerSpec                     `json:"overrideRunnerSpec,omitempty"`
}

type TerraformRepositoryRepository struct {
	Url        string `json:"url,omitempty"`
	SecretName string `json:"secretName,omitempty"`
}

// TerraformRepositoryStatus defines the observed state of TerraformRepository
type TerraformRepositoryStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

type TerraformRepositoryRemediationStrategy struct {
	PlanOnDrift  bool `json:"planOnDrift,omitempty"`
	ApplyOnDrift bool `json:"applyOnDrift,omitempty"`
	PlanOnPush   bool `json:"planOnPush,omitempty"`
	ApplyOnPush  bool `json:"applyOnPush,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TerraformRepository is the Schema for the terraformrepositories API
type TerraformRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformRepositorySpec   `json:"spec,omitempty"`
	Status TerraformRepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TerraformRepositoryList contains a list of TerraformRepository
type TerraformRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TerraformRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TerraformRepository{}, &TerraformRepositoryList{})
}

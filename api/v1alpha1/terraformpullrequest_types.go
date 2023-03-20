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

// TerraformPullRequestSpec defines the desired state of TerraformPullRequest
type TerraformPullRequestSpec struct {
	Provider   string                   `json:"provider,omitempty"`
	Branch     string                   `json:"branch,omitempty"`
	Base       string                   `json:"base,omitempty"`
	ID         string                   `json:"id,omitempty"`
	Repository TerraformLayerRepository `json:"repository,omitempty"`
}

// TerraformPullRequestStatus defines the observed state of TerraformPullRequest
type TerraformPullRequestStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	State      string             `json:"state,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=pr;prs;pullrequest;pullrequests;
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ID",type=string,JSONPath=`.spec.id`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Base",type=string,JSONPath=`.spec.base`
// +kubebuilder:printcolumn:name="Branch",type=string,JSONPath=`.spec.branch`
// TerraformPullRequest is the Schema for the TerraformPullRequests API
type TerraformPullRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformPullRequestSpec   `json:"spec,omitempty"`
	Status TerraformPullRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TerraformPullRequestList contains a list of TerraformPullRequest
type TerraformPullRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TerraformPullRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TerraformPullRequest{}, &TerraformPullRequestList{})
}

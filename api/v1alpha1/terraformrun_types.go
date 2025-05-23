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

// TerraformRunSpec defines the desired state of TerraformRun
type TerraformRunSpec struct {
	Action   string            `json:"action,omitempty"`
	Artifact Artifact          `json:"artifact,omitempty"`
	Layer    TerraformRunLayer `json:"layer,omitempty"`
}

type Artifact struct {
	Run     string `json:"run,omitempty"`
	Attempt string `json:"attempt,omitempty"`
}

type TerraformRunLayer struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Revision  string `json:"revision,omitempty"`
}

// TerraformRunStatus defines the observed state of TerraformRun
type TerraformRunStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	State      string             `json:"state,omitempty"`
	Retries    int                `json:"retries"`
	LastRun    string             `json:"lastRun,omitempty"`
	Attempts   []Attempt          `json:"attempts,omitempty"`
	RunnerPod  string             `json:"runnerPod,omitempty"`
}

type Attempt struct {
	PodName      string `json:"podName"`
	Number       int    `json:"number"`
	LogsUploaded bool   `json:"logsUploaded,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=runs;run;tfruns;tfrun;
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Retries",type=integer,JSONPath=`.status.retries`
// +kubebuilder:printcolumn:name="Created On",type=string,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Runner Pod",type=string,JSONPath=`.status.runnerPod`
// TerraformRun is the Schema for the terraformRuns API
type TerraformRun struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TerraformRunSpec   `json:"spec,omitempty"`
	Status TerraformRunStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TerraformRunList contains a list of TerraformRun
type TerraformRunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TerraformRun `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TerraformRun{}, &TerraformRunList{})
}

// Associated functions

// Workaround needed for envtest which does not populate the TypeMeta structure
// See https://github.com/kubernetes-sigs/controller-runtime/issues/1870
func (run *TerraformRun) GetAPIVersion() string {
	if run.APIVersion == "" {
		return "config.terraform.padok.cloud/v1alpha1"
	}
	return run.APIVersion
}

// Same as above
func (run *TerraformRun) GetKind() string {
	if run.Kind == "" {
		return "TerraformRun"
	}
	return run.Kind
}

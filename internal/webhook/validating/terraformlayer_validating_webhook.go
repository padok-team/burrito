/*
Copyright 2026.

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

package validating

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

// TerraformLayerValidator validates TerraformLayer resources
type TerraformLayerValidator struct {
	client.Client
}

// NewTerraformLayerValidator creates a new validator for TerraformLayer
func NewTerraformLayerValidator(client client.Client) *TerraformLayerValidator {
	return &TerraformLayerValidator{Client: client}
}

//+kubebuilder:webhook:path=/validate-config-terraform-padok-cloud-v1alpha1-terraformlayer,mutating=false,failurePolicy=fail,sideEffects=None,groups=config.terraform.padok.cloud,resources=terraformlayers,verbs=create;update,versions=v1alpha1,name=vterraformlayer.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &TerraformLayerValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TerraformLayerValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	layer, ok := obj.(*configv1alpha1.TerraformLayer)
	if !ok {
		return nil, fmt.Errorf("expected a TerraformLayer object but got %T", obj)
	}
	return v.validateLayer(layer)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TerraformLayerValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	layer, ok := newObj.(*configv1alpha1.TerraformLayer)
	if !ok {
		return nil, fmt.Errorf("expected a TerraformLayer object but got %T", newObj)
	}
	return v.validateLayer(layer)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TerraformLayerValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// validateLayer validates the TerraformLayer spec
func (v *TerraformLayerValidator) validateLayer(layer *configv1alpha1.TerraformLayer) (admission.Warnings, error) {
	var allErrs field.ErrorList

	// Validate path and branch when repository is configured
	if layer.Spec.Repository.Name != "" {
		if layer.Spec.Path == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("path"), "path is required when repository is configured"))
		}
		if layer.Spec.Branch == "" {
			allErrs = append(allErrs, field.Required(field.NewPath("spec").Child("branch"), "branch is required when repository is configured"))
		}
	}

	// Validate that at least one of Terraform or OpenTofu is enabled
	terraformEnabled := layer.Spec.TerraformConfig.Enabled != nil && *layer.Spec.TerraformConfig.Enabled
	openTofuEnabled := layer.Spec.OpenTofuConfig.Enabled != nil && *layer.Spec.OpenTofuConfig.Enabled

	if !terraformEnabled && !openTofuEnabled {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("spec").Child("terraform"),
			"",
			"at least one of terraform.enabled or opentofu.enabled must be true",
		))
	}

	if len(allErrs) > 0 {
		return nil, errors.NewForbidden(schema.GroupResource{Resource: "terraformlayers"}, layer.Name, allErrs.ToAggregate())
	}

	return nil, nil
}
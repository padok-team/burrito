/*
Copyright 2024.

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

package validatingwebhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

// Validator implements controller-runtime's CustomValidator interface for
// Burrito CRDs. While primary validation lives in kubebuilder markers on the
// API types (see api/v1alpha1/*_types.go), this webhook enforces cross-field
// and business-rule checks that are difficult or verbose to express as CEL /
// XValidation expressions — such as validating the remediation strategy and
// interactions between options described in issue #42.
type Validator struct {
	decoder *admission.Decoder
}

// NewValidator returns a Validator that can be registered with a
// controller-runtime webhook server.
func NewValidator(scheme *runtime.Scheme) *Validator {
	return &Validator{
		decoder: admission.NewDecoder(scheme),
	}
}

// ValidateCreate implements webhook.CustomValidator.
func (v *Validator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return v.validate(obj, admission.OperationCreate)
}

// ValidateUpdate implements webhook.CustomValidator.
func (v *Validator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return v.validate(newObj, admission.OperationUpdate)
}

// ValidateDelete implements webhook.CustomValidator.
func (v *Validator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	// No delete-specific validations yet.
	return nil, nil
}

// validate dispatches to the correct per-type validator.
func (v *Validator) validate(obj runtime.Object, op admission.Operation) (admission.Warnings, error) {
	switch o := obj.(type) {
	case *configv1alpha1.TerraformLayer:
		return v.validateTerraformLayer(o)
	case *configv1alpha1.TerraformRepository:
		return v.validateTerraformRepository(o)
	case *configv1alpha1.TerraformRun:
		return nil, nil // pass
	case *configv1alpha1.TerraformPullRequest:
		return nil, nil // pass
	default:
		return nil, fmt.Errorf("unsupported type %T", obj)
	}
}

// validateTerraformLayer runs TerraformLayer-specific business-rule checks.
func (v *Validator) validateTerraformLayer(layer *configv1alpha1.TerraformLayer) (admission.Warnings, error) {
	var errs []error

	// RemediationStrategy.OnError.MaxRetries must be non-negative.
	if layer.Spec.RemediationStrategy.OnError.MaxRetries != nil {
		if *layer.Spec.RemediationStrategy.OnError.MaxRetries < 0 {
			errs = append(errs, fmt.Errorf(
				"spec.remediationStrategy.onError.maxRetries: must be >= 0, got %d",
				*layer.Spec.RemediationStrategy.OnError.MaxRetries,
			))
		}
	}

	// Additional target refs must not duplicate the main path.
	for _, ref := range layer.Spec.AdditionalTargetRefs {
		if ref == layer.Spec.Path {
			errs = append(errs, fmt.Errorf(
				"spec.additionalTargetRefs: %q duplicates spec.path", ref,
			))
			break
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("terraformlayer validation failed: %w", joinErrors(errs))
	}
	return nil, nil
}

// validateTerraformRepository runs TerraformRepository-specific validations.
func (v *Validator) validateTerraformRepository(repo *configv1alpha1.TerraformRepository) (admission.Warnings, error) {
	var errs []error

	// RemediationStrategy.OnError.MaxRetries must be non-negative.
	if repo.Spec.RemediationStrategy.OnError.MaxRetries != nil {
		if *repo.Spec.RemediationStrategy.OnError.MaxRetries < 0 {
			errs = append(errs, fmt.Errorf(
				"spec.remediationStrategy.onError.maxRetries: must be >= 0, got %d",
				*repo.Spec.RemediationStrategy.OnError.MaxRetries,
			))
		}
	}

	// SyncWindows must reference known actions.
	validActions := map[string]struct{}{
		"plan":  {},
		"apply": {},
	}
	for i, window := range repo.Spec.SyncWindows {
		for _, action := range window.Actions {
			if _, ok := validActions[action]; !ok {
				errs = append(errs, fmt.Errorf(
					"spec.syncWindows[%d].actions: unknown action %q, must be one of plan, apply",
					i, action,
				))
			}
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("terraformrepository validation failed: %w", joinErrors(errs))
	}
	return nil, nil
}

// joinErrors concatenates error messages for admission response clarity.
func joinErrors(errs []error) error {
	msg := ""
	for i, e := range errs {
		if i > 0 {
			msg += "; "
		}
		msg += e.Error()
	}
	return fmt.Errorf("%s", msg)
}

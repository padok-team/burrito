package utils

import (
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
)

func IsLayerPlanApproved(layer configv1alpha1.TerraformLayer) bool {
	if layer.Annotations[annotations.LastPlanApproved] == "true" {
		return true
	}
	return false
}

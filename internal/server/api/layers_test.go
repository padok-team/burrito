package api

import (
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetLayerState(t *testing.T) {
	baseAnnotations := map[string]string{
		annotations.LastPlanSum: "some-sum",
	}
	someCondition := []metav1.Condition{{
		Type:   "some-condition",
		Status: metav1.ConditionTrue,
		Reason: "some-reason",
	}}

	tests := []struct {
		name     string
		layer    configv1alpha1.TerraformLayer
		expected string
	}{
		{
			name: "no conditions is disabled",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{Annotations: baseAnnotations},
				Status:     configv1alpha1.TerraformLayerStatus{},
			},
			expected: "disabled",
		},
		{
			name: "max retries reached",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{Annotations: baseAnnotations},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: someCondition,
					State:      "MaxRetriesReached",
				},
			},
			expected: "retriesExhausted",
		},
		{
			name: "default state is success",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{Annotations: baseAnnotations},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: someCondition,
					State:      "Idle",
				},
			},
			expected: "success",
		},
	}

	a := &API{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := a.getLayerState(tt.layer); got != tt.expected {
				t.Errorf("getLayerState() = %q, want %q", got, tt.expected)
			}
		})
	}
}

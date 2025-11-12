package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

func TestInitMetrics(t *testing.T) {
	// Reset global metrics for testing
	Metrics = nil

	m := InitMetrics()

	assert.NotNil(t, m)
	assert.NotNil(t, m.LayerStatusGauge)
	assert.NotNil(t, m.LayerStateGauge)
	assert.NotNil(t, m.RepositoryStatusGauge)
	assert.NotNil(t, m.TotalLayers)
	assert.NotNil(t, m.TotalRepositories)
	assert.NotNil(t, m.TotalRuns)
	assert.NotNil(t, m.TotalPullRequests)

	// Verify we can get the metrics again
	m2 := GetMetrics()
	assert.Equal(t, m, m2)
}

func TestGetLayerUIStatus(t *testing.T) {
	tests := []struct {
		name     string
		layer    configv1alpha1.TerraformLayer
		expected string
	}{
		{
			name: "disabled - no conditions",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"runner.terraform.padok.cloud/plan-sum": "abc123",
					},
				},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: []metav1.Condition{},
				},
			},
			expected: "disabled",
		},
		{
			name: "success - ApplyNeeded with no changes",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"runner.terraform.padok.cloud/plan-sum": "abc123",
					},
				},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
					},
					State:      "ApplyNeeded",
					LastResult: "Plan: 0 to create, 0 to update, 0 to delete",
				},
			},
			expected: "success",
		},
		{
			name: "warning - ApplyNeeded with changes",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"runner.terraform.padok.cloud/plan-sum": "abc123",
					},
				},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
					},
					State:      "ApplyNeeded",
					LastResult: "Plan: 1 to create, 0 to update, 0 to delete",
				},
			},
			expected: "warning",
		},
		{
			name: "warning - PlanNeeded",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"runner.terraform.padok.cloud/plan-sum": "abc123",
					},
				},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
					},
					State: "PlanNeeded",
				},
			},
			expected: "warning",
		},
		{
			name: "running - IsRunning condition",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"runner.terraform.padok.cloud/plan-sum": "abc123",
					},
				},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: []metav1.Condition{
						{Type: "IsRunning", Status: metav1.ConditionTrue},
					},
					State: "PlanNeeded",
				},
			},
			expected: "running",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLayerUIStatus(tt.layer)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRepositoryStatus(t *testing.T) {
	tests := []struct {
		name     string
		repo     configv1alpha1.TerraformRepository
		expected string
	}{
		{
			name: "success - no conditions",
			repo: configv1alpha1.TerraformRepository{
				Status: configv1alpha1.TerraformRepositoryStatus{
					Conditions: []metav1.Condition{},
				},
			},
			expected: "success",
		},
		{
			name: "success - all conditions true",
			repo: configv1alpha1.TerraformRepository{
				Status: configv1alpha1.TerraformRepositoryStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
					},
				},
			},
			expected: "success",
		},
		{
			name: "error - condition false",
			repo: configv1alpha1.TerraformRepository{
				Status: configv1alpha1.TerraformRepositoryStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionFalse},
					},
				},
			},
			expected: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRepositoryStatus(tt.repo)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateLayerMetrics(t *testing.T) {
	// Get or initialize metrics (don't re-initialize if already set)
	if Metrics == nil {
		InitMetrics()
	}

	layer := configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-layer",
			Namespace: "test-namespace",
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Repository: configv1alpha1.TerraformLayerRepository{
				Name: "test-repo",
			},
		},
		Status: configv1alpha1.TerraformLayerStatus{
			Conditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionTrue},
			},
			State: "Idle",
		},
	}

	// Should not panic
	UpdateLayerMetrics(layer)

	// Verify metric was set (we can't easily verify the exact value without accessing the registry)
	// But we can verify it doesn't panic
	assert.NotNil(t, Metrics)
}

func TestUpdateRepositoryMetrics(t *testing.T) {
	// Get or initialize metrics (don't re-initialize if already set)
	if Metrics == nil {
		InitMetrics()
	}

	repo := configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "test-namespace",
		},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: "https://github.com/test/repo",
			},
		},
		Status: configv1alpha1.TerraformRepositoryStatus{
			Conditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionTrue},
			},
		},
	}

	// Should not panic
	UpdateRepositoryMetrics(repo)
	assert.NotNil(t, Metrics)
}

func TestMetricsRegistration(t *testing.T) {
	// Get existing metrics or create if needed
	var m *BurritoMetrics
	if Metrics != nil {
		m = Metrics
	} else {
		m = InitMetrics()
	}

	// Create a new registry for testing gathering
	registry := prometheus.NewRegistry()

	// Verify all metrics are non-nil
	assert.NotNil(t, m.LayerStatusGauge)
	assert.NotNil(t, m.LayerStateGauge)
	assert.NotNil(t, m.LayersByStatus)
	assert.NotNil(t, m.LayersByNamespace)
	assert.NotNil(t, m.RepositoryStatusGauge)
	assert.NotNil(t, m.RunsByAction)
	assert.NotNil(t, m.RunsByStatus)
	assert.NotNil(t, m.TotalLayers)
	assert.NotNil(t, m.TotalRepositories)
	assert.NotNil(t, m.TotalRuns)
	assert.NotNil(t, m.TotalPullRequests)
	assert.NotNil(t, m.ReconcileDuration)
	assert.NotNil(t, m.ReconcileTotal)
	assert.NotNil(t, m.ReconcileErrors)

	// Verify we can gather metrics without error
	_, err := registry.Gather()
	assert.NoError(t, err)
}

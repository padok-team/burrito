package v1alpha1_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

func TestGetTerraformEnabled(t *testing.T) {
	tt := []struct {
		name       string
		repository *configv1alpha1.TerraformRepository
		layer      *configv1alpha1.TerraformLayer
		expected   bool
	}{
		{"OnlyRepositoryEnabling",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			true,
		},
		{"OnlyLayerEnabling",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		// Edge case: Merging config from repository and layer should
		// never set several base execution tools to true at the same time.
		{"OverrideRepositoryTerraformWithLayerOpenTofu",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			false,
		},
		{"OverrideRepositoryOpentofuWithLayerTerraform",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetTerraformEnabled(tc.repository, tc.layer)
			if tc.expected != result {
				t.Errorf("different base exec enabled status computed: expected %t go %t", tc.expected, result)
			}
		})
	}
}

func TestGetOpenTofuEnabled(t *testing.T) {
	tt := []struct {
		name       string
		repository *configv1alpha1.TerraformRepository
		layer      *configv1alpha1.TerraformLayer
		expected   bool
	}{
		{"OnlyRepositoryEnabling",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			true,
		},
		{"OnlyLayerEnabling",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		// Edge case: Merging config from repository and layer should
		// never set several base execution tools to true at the same time.
		{"OverrideRepositoryTerraformWithLayerOpenTofu",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		{"OverrideRepositoryOpentofuWithLayerTerraform",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetOpenTofuEnabled(tc.repository, tc.layer)
			if tc.expected != result {
				t.Errorf("different base exec enabled status computed: expected %t go %t", tc.expected, result)
			}
		})
	}
}

func TestGetOpenTofuVersion(t *testing.T) {
	tt := []struct {
		name            string
		repository      *configv1alpha1.TerraformRepository
		layer           *configv1alpha1.TerraformLayer
		expectedVersion string
	}{
		{
			"OnlyRepositoryVersion",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Version: "1.0.1",
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			"1.0.1",
		},
		{
			"OnlyLayerVersion",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Version: "1.0.1",
					},
				},
			},
			"1.0.1",
		},
		{
			"OnlyLayerVersion",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Version: "1.0.1",
					},
				},
			},
			"1.0.1",
		},
		{
			"OverrideRepositoryWithLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Version: "1.0.1",
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OpenTofuConfig: configv1alpha1.OpenTofuConfig{
						Version: "1.0.6",
					},
				},
			},
			"1.0.6",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetOpenTofuVersion(tc.repository, tc.layer)
			if tc.expectedVersion != result {
				t.Errorf("different version computed: expected %s go %s", tc.expectedVersion, result)
			}
		})
	}
}

func TestGetTerraformVersion(t *testing.T) {
	tt := []struct {
		name            string
		repository      *configv1alpha1.TerraformRepository
		layer           *configv1alpha1.TerraformLayer
		expectedVersion string
	}{
		{
			"OnlyRepositoryVersion",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Version: "1.0.1",
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			"1.0.1",
		},
		{
			"OnlyLayerVersion",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Version: "1.0.1",
					},
				},
			},
			"1.0.1",
		},
		{
			"OverrideRepositoryWithLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Version: "1.0.1",
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						Version: "1.0.6",
					},
				},
			},
			"1.0.6",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetTerraformVersion(tc.repository, tc.layer)
			if tc.expectedVersion != result {
				t.Errorf("different version computed: expected %s go %s", tc.expectedVersion, result)
			}
		})
	}
}

func TestGetTerragruntVersion(t *testing.T) {
	tt := []struct {
		name            string
		repository      *configv1alpha1.TerraformRepository
		layer           *configv1alpha1.TerraformLayer
		expectedVersion string
	}{
		{
			"OnlyRepositoryVersion",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Version: "0.43.0",
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			"0.43.0",
		},
		{
			"OnlyLayerVersion",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Version: "0.43.0",
					},
				},
			},
			"0.43.0",
		},
		{
			"OverrideRepositoryWithLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Version: "0.43.0",
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Version: "0.45.0",
					},
				},
			},
			"0.45.0",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetTerragruntVersion(tc.repository, tc.layer)
			if tc.expectedVersion != result {
				t.Errorf("different version computed: expected %s go %s", tc.expectedVersion, result)
			}
		})
	}
}

func TestGetTerragruntEnabled(t *testing.T) {
	tt := []struct {
		name       string
		repository *configv1alpha1.TerraformRepository
		layer      *configv1alpha1.TerraformLayer
		expected   bool
	}{
		{
			"OnlyRepositoryEnabling",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			true,
		},
		{
			"OnlyLayerEnabling",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		{
			"DisabledInRepositoryEnabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Enabled: &[]bool{false}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		{
			"EnabledInRepositoryDisabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Enabled: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerragruntConfig: configv1alpha1.TerragruntConfig{
						Enabled: &[]bool{false}[0],
					},
				},
			},
			false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetTerragruntEnabled(tc.repository, tc.layer)
			if tc.expected != result {
				t.Errorf("different enabled status computed: expected %t go %t", tc.expected, result)
			}
		})
	}
}

func TestGetApplyWithoutPlanArtifactEnabled(t *testing.T) {
	tt := []struct {
		name       string
		repository *configv1alpha1.TerraformRepository
		layer      *configv1alpha1.TerraformLayer
		expected   bool
	}{
		{
			"DisabledInRepositoryEnabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						ApplyWithoutPlanArtifact: &[]bool{false}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						ApplyWithoutPlanArtifact: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		{
			"EnabledInRepositoryDisabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						ApplyWithoutPlanArtifact: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						ApplyWithoutPlanArtifact: &[]bool{false}[0],
					},
				},
			},
			false,
		},
		{
			"OnlyRepositoryEnabling",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						ApplyWithoutPlanArtifact: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			true,
		},
		{
			"OnlyLayerEnabling",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						ApplyWithoutPlanArtifact: &[]bool{true}[0],
					},
				},
			},
			true,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetApplyWithoutPlanArtifactEnabled(tc.repository, tc.layer)
			if tc.expected != result {
				t.Errorf("different enabled status computed: expected %t go %t", tc.expected, result)
			}
		})
	}
}

func TestGetAutoApplyEnabled(t *testing.T) {
	tt := []struct {
		name       string
		repository *configv1alpha1.TerraformRepository
		layer      *configv1alpha1.TerraformLayer
		expected   bool
	}{
		{
			"EnabledInRepositoryDisabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{false}[0],
					},
				},
			},
			false,
		},
		{
			"DisabledInRepositoryEnabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{false}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		{
			"EnabledInRepositoryEnabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{true}[0],
					},
				},
			},
			true,
		},
		{
			"DisabledInRepositoryDisabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{false}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{false}[0],
					},
				},
			},
			false,
		},
		{
			"OnlyRepositoryEnabling",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{true}[0],
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			true,
		},
		{
			"OnlyLayerEnabling",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: &[]bool{true}[0],
					},
				},
			},
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetAutoApplyEnabled(tc.repository, tc.layer)
			if tc.expected != result {
				t.Errorf("different enabled status computed: expected %t go %t", tc.expected, result)
			}
		})
	}
}

func TestOverrideRunnerSpec(t *testing.T) {
	tt := []struct {
		name         string
		repository   *configv1alpha1.TerraformRepository
		layer        *configv1alpha1.TerraformLayer
		expectedSpec configv1alpha1.OverrideRunnerSpec
	}{
		{
			"OverrideTolerations",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "only-exists-in-repository",
								Value:  "true",
								Effect: "NoSchedule",
							},
							{
								Key:    "does-not-exists-in-layer",
								Value:  "true",
								Effect: "NoSchedule",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "does-not-exists-in-layer",
								Value:  "false",
								Effect: "NoExecute",
							},
							{
								Key:    "only-exists-in-layer",
								Value:  "true",
								Effect: "NoSchedule",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:    "does-not-exists-in-layer",
						Value:  "false",
						Effect: "NoExecute",
					},
					{
						Key:    "only-exists-in-layer",
						Value:  "true",
						Effect: "NoSchedule",
					},
				},
			},
		},
		{
			"TolerationsOnlyInrepository",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "only-exists-in-repository",
								Value:  "true",
								Effect: "NoSchedule",
							},
							{
								Key:    "does-not-exists-in-layer",
								Value:  "true",
								Effect: "NoSchedule",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:    "only-exists-in-repository",
						Value:  "true",
						Effect: "NoSchedule",
					},
					{
						Key:    "does-not-exists-in-layer",
						Value:  "true",
						Effect: "NoSchedule",
					},
				},
			},
		},
		{
			"TolerationsOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "does-not-exists-in-layer",
								Value:  "false",
								Effect: "NoExecute",
							},
							{
								Key:    "only-exists-in-layer",
								Value:  "true",
								Effect: "NoSchedule",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:    "does-not-exists-in-layer",
						Value:  "false",
						Effect: "NoExecute",
					},
					{
						Key:    "only-exists-in-layer",
						Value:  "true",
						Effect: "NoSchedule",
					},
				},
			},
		},
		{
			"TolerationsWithSameKeyButDifferentValuesExistInBoth",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "exists-in-both",
								Value:  "true",
								Effect: "NoExecute",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "exists-in-both",
								Value:  "false",
								Effect: "NoExecute",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:    "exists-in-both",
						Value:  "false",
						Effect: "NoExecute",
					},
				},
			},
		},
		{
			"TolerationsWithSameKeyButDifferentValuesOnlyInRepository",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "same-key",
								Value:  "value-1",
								Effect: "NoExecute",
							},
							{
								Key:    "same-key",
								Value:  "value-2",
								Effect: "NoExecute",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:    "same-key",
						Value:  "value-1",
						Effect: "NoExecute",
					},
					{
						Key:    "same-key",
						Value:  "value-2",
						Effect: "NoExecute",
					},
				},
			},
		},
		{
			"TolerationsWithSameKeyButDifferentValuesOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Tolerations: []corev1.Toleration{
							{
								Key:    "same-key",
								Value:  "value-1",
								Effect: "NoExecute",
							},
							{
								Key:    "same-key",
								Value:  "value-2",
								Effect: "NoExecute",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:    "same-key",
						Value:  "value-1",
						Effect: "NoExecute",
					},
					{
						Key:    "same-key",
						Value:  "value-2",
						Effect: "NoExecute",
					},
				},
			},
		},
		{
			"ChooseImageNotSpecified",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{},
		},
		{
			"ChooseRepositoryImage",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Image: "test",
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				Image: "test",
			},
		},
		{
			"ChooseLayerImage",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Image: "test",
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Image: "test",
			},
		},
		{
			"OverrideRepositoryImageInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Image: "test",
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Image: "overrdie",
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Image: "overrdie",
			},
		},
		{
			"PullSecretsInRepository",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "test"},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "test"},
				},
			},
		},
		{
			"PullSecretsInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "test"},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "test"},
				},
			},
		},
		{
			"PullSecretsInBoth",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "repo"},
							{Name: "common"},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "layer"},
							{Name: "common"},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				ImagePullSecrets: []corev1.LocalObjectReference{
					{Name: "repo"},
					{Name: "common"},
					{Name: "layer"},
				},
			},
		},
		{
			"MergeNodeSelector",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						NodeSelector: map[string]string{"only-in-repo": "true", "exists-in-both": "false"},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						NodeSelector: map[string]string{"exists-in-both": "true", "only-in-layer": "true"},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				NodeSelector: map[string]string{"only-in-repo": "true", "exists-in-both": "true", "only-in-layer": "true"},
			},
		},
		{
			"NodeSelectorOnlyInRepo",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						NodeSelector: map[string]string{"only-in-repo": "true", "exists-in-both": "false"},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				NodeSelector: map[string]string{"only-in-repo": "true", "exists-in-both": "false"},
			},
		},
		{
			"NodeSelectorOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						NodeSelector: map[string]string{"exists-in-both": "true", "only-in-layer": "true"},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				NodeSelector: map[string]string{"exists-in-both": "true", "only-in-layer": "true"},
			},
		},
		{
			"ChooseRepositoryServiceAccount",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ServiceAccountName: "test",
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				ServiceAccountName: "test",
			},
		},
		{
			"ChooseLayerServiceAccount",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ServiceAccountName: "test",
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				ServiceAccountName: "test",
			},
		},
		{
			"OverrideRepositoryServiceAccountInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ServiceAccountName: "test",
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						ServiceAccountName: "overrdie",
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				ServiceAccountName: "overrdie",
			},
		},
		{
			"ChooseRepositoryResources",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    *resource.NewQuantity(1, resource.DecimalSI),
								"memory": *resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    *resource.NewQuantity(1, resource.DecimalSI),
						"memory": *resource.NewQuantity(2, resource.DecimalSI),
					},
				},
			},
		},
		{
			"ChooseLayerResources",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    *resource.NewQuantity(1, resource.DecimalSI),
								"memory": *resource.NewQuantity(2, resource.DecimalSI),
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    *resource.NewQuantity(1, resource.DecimalSI),
						"memory": *resource.NewQuantity(2, resource.DecimalSI),
					},
				},
			},
		},
		{
			"OverrideRepositoryResourcesLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    *resource.NewQuantity(2, resource.DecimalSI),
								"memory": *resource.NewQuantity(2, resource.DecimalSI),
							},
							Requests: corev1.ResourceList{
								"cpu": *resource.NewQuantity(1, resource.DecimalSI),
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu": *resource.NewQuantity(3, resource.DecimalSI),
							},
							Requests: corev1.ResourceList{
								"memory": *resource.NewQuantity(1, resource.DecimalSI),
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    *resource.NewQuantity(3, resource.DecimalSI),
						"memory": *resource.NewQuantity(2, resource.DecimalSI),
					},
					Requests: corev1.ResourceList{
						"cpu":    *resource.NewQuantity(1, resource.DecimalSI),
						"memory": *resource.NewQuantity(1, resource.DecimalSI),
					},
				},
			},
		},
		{
			"EnvOnlyInRepo",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Env: []corev1.EnvVar{
							{
								Name:  "ONLY_REPO",
								Value: "1",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				Env: []corev1.EnvVar{
					{
						Name:  "ONLY_REPO",
						Value: "1",
					},
				},
			},
		},
		{
			"EnvOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Env: []corev1.EnvVar{
							{
								Name:  "ONLY_LAYER",
								Value: "1",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Env: []corev1.EnvVar{
					{
						Name:  "ONLY_LAYER",
						Value: "1",
					},
				},
			},
		},
		{
			"EnvInBoth",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Env: []corev1.EnvVar{
							{
								Name:  "ONLY_REPO",
								Value: "1",
							},
							{
								Name:  "IN_BOTH",
								Value: "0",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Env: []corev1.EnvVar{
							{
								Name:  "ONLY_LAYER",
								Value: "1",
							},
							{
								Name:  "IN_BOTH",
								Value: "1",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Env: []corev1.EnvVar{
					{
						Name:  "ONLY_REPO",
						Value: "1",
					},
					{
						Name:  "IN_BOTH",
						Value: "1",
					},
					{
						Name:  "ONLY_LAYER",
						Value: "1",
					},
				},
			},
		},
		{
			"EnvFromOnlyInRepo",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "repo-cm"},
								},
							},
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "repo-secret"},
								},
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "repo-cm"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "repo-secret"},
						},
					},
				},
			},
		},
		{
			"EnvFromOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "layer-cm"},
								},
							},
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "layer-secret"},
								},
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "layer-cm"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "layer-secret"},
						},
					},
				},
			},
		},
		{
			"EnvFromInBoth",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "repo-cm"},
								},
							},
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "repo-secret"},
								},
							},
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "both-cm"},
								},
							},
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "both-secret"},
								},
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						EnvFrom: []corev1.EnvFromSource{
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "layer-cm"},
								},
							},
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "layer-secret"},
								},
							},
							{
								ConfigMapRef: &corev1.ConfigMapEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "both-cm"},
								},
							},
							{
								SecretRef: &corev1.SecretEnvSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "both-secret"},
								},
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				EnvFrom: []corev1.EnvFromSource{
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "repo-cm"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "repo-secret"},
						},
					},
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "layer-cm"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "layer-secret"},
						},
					},
					{
						ConfigMapRef: &corev1.ConfigMapEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "both-cm"},
						},
					},
					{
						SecretRef: &corev1.SecretEnvSource{
							LocalObjectReference: corev1.LocalObjectReference{Name: "both-secret"},
						},
					},
				},
			},
		},
		{
			"VolumesOnlyInRepo",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Volumes: []corev1.Volume{
							{Name: "only-repo"},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				Volumes: []corev1.Volume{
					{Name: "only-repo"},
				},
			},
		},
		{
			"VolumesOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Volumes: []corev1.Volume{
							{Name: "only-layer"},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Volumes: []corev1.Volume{
					{Name: "only-layer"},
				},
			},
		},
		{
			"VolumesInBoth",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Volumes: []corev1.Volume{
							{Name: "only-repo"},
							{
								Name: "both",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/repo/path",
									},
								},
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Volumes: []corev1.Volume{
							{Name: "only-layer"},
							{
								Name: "both",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/layer/path",
									},
								},
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Volumes: []corev1.Volume{
					{Name: "only-repo"},
					{
						Name: "both",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/layer/path",
							},
						},
					},
					{Name: "only-layer"},
				},
			},
		},
		{
			"VolumeMountsOnlyInRepo",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						VolumeMounts: []corev1.VolumeMount{
							{Name: "only-repo"},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				VolumeMounts: []corev1.VolumeMount{
					{Name: "only-repo"},
				},
			},
		},
		{
			"VolumeMountsOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						VolumeMounts: []corev1.VolumeMount{
							{Name: "only-layer"},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				VolumeMounts: []corev1.VolumeMount{
					{Name: "only-layer"},
				},
			},
		},
		{
			"VolumeMountsInBoth",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						VolumeMounts: []corev1.VolumeMount{
							{Name: "only-repo"},
							{
								Name:      "both",
								MountPath: "/repo/path",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						VolumeMounts: []corev1.VolumeMount{
							{Name: "only-layer"},
							{
								Name:      "both",
								MountPath: "/layer/path",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				VolumeMounts: []corev1.VolumeMount{
					{Name: "only-layer"},
					{
						Name:      "both",
						MountPath: "/layer/path",
					},
					{Name: "only-repo"},
				},
			},
		},
		{
			"MetadataOnlyInRepo",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Metadata: configv1alpha1.MetadataOverride{
							Annotations: map[string]string{
								"only-repo": "1",
							},
							Labels: map[string]string{
								"only-repo": "1",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.OverrideRunnerSpec{
				Metadata: configv1alpha1.MetadataOverride{
					Annotations: map[string]string{
						"only-repo": "1",
					},
					Labels: map[string]string{
						"only-repo": "1",
					},
				},
			},
		},
		{
			"MetadataOnlyInLayer",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Metadata: configv1alpha1.MetadataOverride{
							Annotations: map[string]string{
								"only-layer": "1",
							},
							Labels: map[string]string{
								"only-layer": "1",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Metadata: configv1alpha1.MetadataOverride{
					Annotations: map[string]string{
						"only-layer": "1",
					},
					Labels: map[string]string{
						"only-layer": "1",
					},
				},
			},
		},
		{
			"MetadataInBoth",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Metadata: configv1alpha1.MetadataOverride{
							Annotations: map[string]string{
								"only-repo": "1",
								"in-both":   "0",
							},
							Labels: map[string]string{
								"only-repo": "1",
								"in-both":   "0",
							},
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					OverrideRunnerSpec: configv1alpha1.OverrideRunnerSpec{
						Metadata: configv1alpha1.MetadataOverride{
							Annotations: map[string]string{
								"only-layer": "1",
								"in-both":    "1",
							},
							Labels: map[string]string{
								"only-layer": "1",
								"in-both":    "1",
							},
						},
					},
				},
			},
			configv1alpha1.OverrideRunnerSpec{
				Metadata: configv1alpha1.MetadataOverride{
					Annotations: map[string]string{
						"only-repo":  "1",
						"in-both":    "1",
						"only-layer": "1",
					},
					Labels: map[string]string{
						"only-repo":  "1",
						"in-both":    "1",
						"only-layer": "1",
					},
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetOverrideRunnerSpec(tc.repository, tc.layer)

			// Check Tolerations
			if len(result.Tolerations) != len(tc.expectedSpec.Tolerations) {
				t.Errorf("different tolerations size: got %d expected %d", len(result.Tolerations), len(tc.expectedSpec.Tolerations))
			}
			for i, tol := range result.Tolerations {
				if tol.Key != result.Tolerations[i].Key {
					t.Errorf("different tolerations key: got %s expected %s", result.Tolerations[i].Key, tol.Key)
				}
				if tol.Value != result.Tolerations[i].Value {
					t.Errorf("different tolerations value: got %s expected %s", result.Tolerations[i].Value, tol.Value)
				}
				if tol.Effect != result.Tolerations[i].Effect {
					t.Errorf("different tolerations effect: got %s expected %s", result.Tolerations[i].Effect, tol.Effect)
				}
			}

			// Check Image
			if tc.expectedSpec.Image != result.Image {
				t.Errorf("different images: got %s expect %s", result.Image, tc.expectedSpec.Image)
			}

			// Check ImagePullSecrets
			if len(result.ImagePullSecrets) != len(tc.expectedSpec.ImagePullSecrets) {
				t.Errorf("differents image pull secrets size: got %d expected %d", len(result.ImagePullSecrets), len(tc.expectedSpec.ImagePullSecrets))
			}
			for _, secret := range result.ImagePullSecrets {
				found := false
				for _, expected := range tc.expectedSpec.ImagePullSecrets {
					if secret.Name == expected.Name {
						found = true
					}
				}
				if !found {
					t.Errorf("image pull secret %s not found in expected list %v", secret.Name, tc.expectedSpec.ImagePullSecrets)
				}
			}

			//Check NodeSelector
			if len(tc.expectedSpec.NodeSelector) != len(result.NodeSelector) {
				t.Errorf("different size of node selector: got %d expected %d", len(result.NodeSelector), len(tc.expectedSpec.NodeSelector))
			}
			for k, v := range result.NodeSelector {
				if tc.expectedSpec.NodeSelector[k] != v {
					t.Errorf("different node selector value for label %s: got %s expected %s", k, v, tc.expectedSpec.NodeSelector[k])
				}
			}

			// Check ServiceAccountName
			if tc.expectedSpec.ServiceAccountName != result.ServiceAccountName {
				t.Errorf("different serivce account names: got %s expect %s", result.ServiceAccountName, tc.expectedSpec.ServiceAccountName)
			}

			// Check Resources
			for k, v := range result.Resources.Limits {
				if v != tc.expectedSpec.Resources.Limits[k] {
					t.Errorf("different limit value for %s: got %v expected %v", k, v, tc.expectedSpec.Resources.Limits[k])
				}
			}
			for k, v := range result.Resources.Requests {
				if v != tc.expectedSpec.Resources.Requests[k] {
					t.Errorf("different request value for %s: got %v expected %v", k, v, tc.expectedSpec.Resources.Requests[k])
				}
			}

			// Check Env
			if len(result.Env) != len(tc.expectedSpec.Env) {
				t.Errorf("differents env size: got %d expected %d", len(result.Env), len(tc.expectedSpec.Env))
			}
			for _, env := range result.Env {
				found := false
				for _, expected := range tc.expectedSpec.Env {
					if env.Name == expected.Name && env.Value == expected.Value {
						found = true
					}
				}
				if !found {
					t.Errorf("env %v not found in expected list %v", env, tc.expectedSpec.Env)
				}
			}

			// Check EnvFrom
			if len(result.EnvFrom) != len(tc.expectedSpec.EnvFrom) {
				t.Errorf("differents env from size: got %d expected %d", len(result.EnvFrom), len(tc.expectedSpec.EnvFrom))
			}
			for _, envFrom := range result.EnvFrom {
				found := false
				for _, expected := range tc.expectedSpec.EnvFrom {
					// We use two different if statements because, if we don't there might ba a nil pointer dereference
					if envFrom.ConfigMapRef != nil && expected.ConfigMapRef != nil {
						if envFrom.ConfigMapRef.LocalObjectReference.Name == expected.ConfigMapRef.LocalObjectReference.Name {
							found = true
						}
					}
					if envFrom.SecretRef != nil && expected.SecretRef != nil {
						if envFrom.SecretRef.LocalObjectReference.Name == expected.SecretRef.LocalObjectReference.Name {
							found = true
						}
					}
				}
				if !found {
					t.Errorf("env from %v not found in expected list %v", envFrom, tc.expectedSpec.EnvFrom)
				}
			}

			// Check Volumes
			if len(result.Volumes) != len(tc.expectedSpec.Volumes) {
				t.Errorf("differents volumes size: got %d expected %d", len(result.Volumes), len(tc.expectedSpec.Volumes))
			}
			for _, vol := range result.Volumes {
				found := false
				for _, expected := range tc.expectedSpec.Volumes {
					if vol.Name == expected.Name {
						found = true
					}
				}
				if !found {
					t.Errorf("volume %v not found in expected list %v", vol, tc.expectedSpec.Volumes)
				}
			}

			// Check VolumeMounts
			if len(result.VolumeMounts) != len(tc.expectedSpec.VolumeMounts) {
				t.Errorf("differents volume mounts size: got %d expected %d", len(result.VolumeMounts), len(tc.expectedSpec.VolumeMounts))
			}
			for _, vol := range result.VolumeMounts {
				found := false
				for _, expected := range tc.expectedSpec.VolumeMounts {
					// We only check for MountPath as it is enough to validate that layer config overrides the repo one
					if vol.Name == expected.Name && vol.MountPath == expected.MountPath {
						found = true
					}
				}
				if !found {
					t.Errorf("volume mount %v not found in expected list %v", vol, tc.expectedSpec.VolumeMounts)
				}
			}

			// Check Metadata.Annotations
			if len(result.Metadata.Annotations) != len(tc.expectedSpec.Metadata.Annotations) {
				t.Errorf("differents annotations size: got %d expected %d", len(result.Metadata.Annotations), len(tc.expectedSpec.Metadata.Annotations))
			}
			for k, v := range result.Metadata.Annotations {
				if tc.expectedSpec.Metadata.Annotations[k] != v {
					t.Errorf("different annotation value for key %s: expected %s got %s", k, tc.expectedSpec.Metadata.Annotations[k], v)
				}
			}

			// Check Metadata.v
			if len(result.Metadata.Labels) != len(tc.expectedSpec.Metadata.Labels) {
				t.Errorf("differents labels size: got %d expected %d", len(result.Metadata.Labels), len(tc.expectedSpec.Metadata.Labels))
			}
			for k, v := range result.Metadata.Labels {
				if tc.expectedSpec.Metadata.Labels[k] != v {
					t.Errorf("different label value for key %s: expected %s got %s", k, tc.expectedSpec.Metadata.Labels[k], v)
				}
			}
		})
	}
}

func intPointer(i int) *int {
	return &i
}

func TestGetHistoryPolicy(t *testing.T) {
	tt := []struct {
		name                  string
		repository            *configv1alpha1.TerraformRepository
		layer                 *configv1alpha1.TerraformLayer
		expectedHistoryPolicy configv1alpha1.RunHistoryPolicy
	}{
		{
			"OnlyRepositoryHistoryPolicy",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RunHistoryPolicy: configv1alpha1.RunHistoryPolicy{
						KeepLastRuns: intPointer(10),
					},
				},
			},
			&configv1alpha1.TerraformLayer{},
			configv1alpha1.RunHistoryPolicy{
				KeepLastRuns: intPointer(10),
			},
		},
		{
			"OnlyLayerHistoryPolicy",
			&configv1alpha1.TerraformRepository{},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RunHistoryPolicy: configv1alpha1.RunHistoryPolicy{
						KeepLastRuns: intPointer(10),
					},
				},
			},
			configv1alpha1.RunHistoryPolicy{
				KeepLastRuns: intPointer(10),
			},
		},
		{
			"OverrideRepositoryWithLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					RunHistoryPolicy: configv1alpha1.RunHistoryPolicy{
						KeepLastRuns: intPointer(10),
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					RunHistoryPolicy: configv1alpha1.RunHistoryPolicy{
						KeepLastRuns: intPointer(5),
					},
				},
			},
			configv1alpha1.RunHistoryPolicy{
				KeepLastRuns: intPointer(5),
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetRunHistoryPolicy(tc.repository, tc.layer)
			if *tc.expectedHistoryPolicy.KeepLastRuns != *result.KeepLastRuns {
				t.Errorf("different policy computed: expected %d got %d", *tc.expectedHistoryPolicy.KeepLastRuns, *result.KeepLastRuns)
			}
		})
	}
}

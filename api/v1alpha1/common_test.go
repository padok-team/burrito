package v1alpha1_test

import (
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

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
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Version: "0.43.0",
						},
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
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Version: "0.43.0",
						},
					},
				},
			},
			"0.43.0",
		},
		{
			"OverrideRepositoryWithLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Version: "0.43.0",
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Version: "0.45.0",
						},
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
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Enabled: &[]bool{true}[0],
						},
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
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Enabled: &[]bool{true}[0],
						},
					},
				},
			},
			true,
		},
		{
			"DisabledInRepositoryEnabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Enabled: &[]bool{false}[0],
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Enabled: &[]bool{true}[0],
						},
					},
				},
			},
			true,
		},
		{
			"EnabledInRepositoryDisabledInLayer",
			&configv1alpha1.TerraformRepository{
				Spec: configv1alpha1.TerraformRepositorySpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Enabled: &[]bool{true}[0],
						},
					},
				},
			},
			&configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					TerraformConfig: configv1alpha1.TerraformConfig{
						TerragruntConfig: configv1alpha1.TerragruntConfig{
							Enabled: &[]bool{false}[0],
						},
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

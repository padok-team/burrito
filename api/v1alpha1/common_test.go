package v1alpha1_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

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

func TestOverrideRunnerSpec(t *testing.T) {
	tt := []struct {
		name         string
		repository   *configv1alpha1.TerraformRepository
		layer        *configv1alpha1.TerraformLayer
		expectedSpec configv1alpha1.OverrideRunnerSpec
	}{
		{
			"MergeTolerations",
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
						Key:    "only-exists-in-repository",
						Value:  "true",
						Effect: "NoSchedule",
					},
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
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := configv1alpha1.GetOverrideRunnerSpec(tc.repository, tc.layer)

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
		})
	}
}

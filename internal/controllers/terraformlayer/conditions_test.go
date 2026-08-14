package terraformlayer

import (
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLayerFilesHaveChanged(t *testing.T) {
	tests := []struct {
		name         string
		layer        configv1alpha1.TerraformLayer
		changedFiles []string
		want         bool
	}{
		{
			name: "no changed files always triggers",
			layer: configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{Path: "terragrunt/random-pets/test"},
			},
			changedFiles: nil,
			want:         true,
		},
		{
			name: "change under layer path triggers",
			layer: configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{Path: "terragrunt/random-pets/test"},
			},
			changedFiles: []string{"terragrunt/random-pets/test/main.tf"},
			want:         true,
		},
		{
			name: "unrelated change does not trigger",
			layer: configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{Path: "terragrunt/random-pets/test"},
			},
			changedFiles: []string{"other/module/main.tf"},
			want:         false,
		},
		{
			name: "spec literal additional trigger path",
			layer: configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:                   "terragrunt/random-pets/test",
					AdditionalTriggerPaths: []string{"../../../modules/random-pets"},
				},
			},
			changedFiles: []string{"modules/random-pets/main.tf"},
			want:         true,
		},
		{
			name: "spec recursive glob additional trigger path",
			layer: configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:                   "terragrunt/random-pets/test",
					AdditionalTriggerPaths: []string{"../**/*.yaml"},
				},
			},
			changedFiles: []string{"terragrunt/random-pets/nested/dir/values.yaml"},
			want:         true,
		},
		{
			name: "spec glob does not match other extensions",
			layer: configv1alpha1.TerraformLayer{
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:                   "terragrunt/random-pets/test",
					AdditionalTriggerPaths: []string{"../**/*.yaml"},
				},
			},
			changedFiles: []string{"terragrunt/random-pets/nested/dir/main.tf"},
			want:         false,
		},
		{
			name: "deprecated annotation literal path still works",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotations.AdditionnalTriggerPaths: "../../../modules/random-pets",
					},
				},
				Spec: configv1alpha1.TerraformLayerSpec{Path: "terragrunt/random-pets/test"},
			},
			changedFiles: []string{"modules/random-pets/main.tf"},
			want:         true,
		},
		{
			name: "deprecated annotation supports glob too",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotations.AdditionnalTriggerPaths: "../**/*.yaml",
					},
				},
				Spec: configv1alpha1.TerraformLayerSpec{Path: "terragrunt/random-pets/test"},
			},
			changedFiles: []string{"terragrunt/random-pets/nested/dir/values.yaml"},
			want:         true,
		},
		{
			name: "spec field takes precedence, annotation ignored",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotations.AdditionnalTriggerPaths: "../../../modules/random-pets",
					},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:                   "terragrunt/random-pets/test",
					AdditionalTriggerPaths: []string{"../**/*.yaml"},
				},
			},
			changedFiles: []string{"modules/random-pets/main.tf"},
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LayerFilesHaveChanged(tt.layer, tt.changedFiles)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestLayerFilesHaveChangedLogsDeprecationForAnnotation(t *testing.T) {
	hook := logrustest.NewGlobal()
	layer := configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "random-pets-terragrunt",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.AdditionnalTriggerPaths: "../../../modules/random-pets",
			},
		},
		Spec: configv1alpha1.TerraformLayerSpec{Path: "terragrunt/random-pets/test"},
	}

	LayerFilesHaveChanged(layer, []string{"modules/random-pets/main.tf"})

	found := false
	for _, entry := range hook.AllEntries() {
		if entry.Level == logrus.WarnLevel && strings.HasPrefix(entry.Message, "DEPRECATED:") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a DEPRECATED warning log, got entries: %+v", hook.AllEntries())
	}
}

func TestLayerFilesHaveChangedNoDeprecationLogWhenSpecUsed(t *testing.T) {
	hook := logrustest.NewGlobal()
	layer := configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				annotations.AdditionnalTriggerPaths: "../../../modules/random-pets",
			},
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Path:                   "terragrunt/random-pets/test",
			AdditionalTriggerPaths: []string{"../../../modules/random-pets"},
		},
	}

	LayerFilesHaveChanged(layer, []string{"modules/random-pets/main.tf"})

	for _, entry := range hook.AllEntries() {
		if strings.HasPrefix(entry.Message, "DEPRECATED:") {
			t.Fatalf("did not expect DEPRECATED warning when spec field is set, got: %+v", entry)
		}
	}
}

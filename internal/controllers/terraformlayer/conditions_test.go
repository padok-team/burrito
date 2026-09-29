package terraformlayer

import (
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLayerFilesHaveChangedAdditionalTriggerPaths(t *testing.T) {
	tests := []struct {
		name  string
		paths string
		want  bool
	}{
		{"single relative path", "../modules/foo", true},
		{"spaces after commas are trimmed", "../docs, ../modules/foo", true},
		{"no matching path", "../docs, ../other", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer := configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{annotations.AdditionnalTriggerPaths: tt.paths}},
				Spec:       configv1alpha1.TerraformLayerSpec{Path: "layers/app"},
			}
			if got := LayerFilesHaveChanged(layer, []string{"layers/modules/foo/main.tf"}); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

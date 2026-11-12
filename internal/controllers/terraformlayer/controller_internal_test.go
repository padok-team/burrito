package terraformlayer

import (
	"testing"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestUpdateLatestRuns(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	history := func(names ...string) []configv1alpha1.TerraformLayerRun {
		runs := make([]configv1alpha1.TerraformLayerRun, 0, len(names))
		for i, name := range names {
			runs = append(runs, configv1alpha1.TerraformLayerRun{
				Name: name,
				Date: metav1.NewTime(base.Add(time.Duration(i) * time.Hour)),
			})
		}
		return runs
	}
	newRun := configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "new",
			CreationTimestamp: metav1.NewTime(base.Add(24 * time.Hour)),
		},
	}

	tests := []struct {
		name     string
		runs     []configv1alpha1.TerraformLayerRun
		keep     int
		expected []string
	}{
		{
			name:     "empty history",
			runs:     nil,
			keep:     5,
			expected: []string{"new"},
		},
		{
			name:     "history below the limit",
			runs:     history("a", "b"),
			keep:     5,
			expected: []string{"a", "b", "new"},
		},
		{
			name:     "history at the limit drops the oldest run",
			runs:     history("a", "b", "c"),
			keep:     3,
			expected: []string{"b", "c", "new"},
		},
		{
			name:     "history above the limit is trimmed down to the limit",
			runs:     history("a", "b", "c", "d", "e"),
			keep:     2,
			expected: []string{"e", "new"},
		},
		{
			name: "unordered history drops the oldest runs",
			runs: []configv1alpha1.TerraformLayerRun{
				{Name: "c", Date: metav1.NewTime(base.Add(2 * time.Hour))},
				{Name: "a", Date: metav1.NewTime(base)},
				{Name: "b", Date: metav1.NewTime(base.Add(time.Hour))},
			},
			keep:     2,
			expected: []string{"c", "new"},
		},
		{
			name:     "the new run is always kept",
			runs:     history("a", "b"),
			keep:     0,
			expected: []string{"new"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := append([]configv1alpha1.TerraformLayerRun(nil), tt.runs...)
			result := updateLatestRuns(tt.runs, newRun, tt.keep)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d runs, got %d: %v", len(tt.expected), len(result), result)
			}
			for i, name := range tt.expected {
				if result[i].Name != name {
					t.Errorf("run %d: expected %q, got %q", i, name, result[i].Name)
				}
			}
			for i := range input {
				if tt.runs[i].Name != input[i].Name {
					t.Errorf("input history was modified at index %d", i)
				}
			}
		})
	}
}

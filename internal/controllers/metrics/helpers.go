package metrics

import (
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetLayerStatus(layer configv1alpha1.TerraformLayer) string {
	state := "success"

	switch {
	case len(layer.Status.Conditions) == 0:
		state = "disabled"
	case layer.Status.State == "ApplyNeeded":
		if layer.Status.LastResult == "Plan: 0 to create, 0 to update, 0 to delete" {
			state = "success"
		} else {
			state = "warning"
		}
	case layer.Status.State == "PlanNeeded":
		state = "warning"
	}

	if layer.Annotations != nil {
		if layer.Annotations[annotations.LastPlanSum] == "" {
			state = "error"
		}
		if layer.Annotations[annotations.LastApplySum] != "" && layer.Annotations[annotations.LastApplySum] == "" {
			state = "error"
		}
	}

	for _, condition := range layer.Status.Conditions {
		if condition.Type == "IsRunning" && condition.Status == metav1.ConditionTrue {
			state = "running"
			break
		}
	}

	return state
}

func GetRepositoryStatus(repo configv1alpha1.TerraformRepository) string {
	// Return the repository state which can be "Synced" or "SyncNeeded"
	// This matches what kubectl shows in the State column
	if repo.Status.State != "" {
		return repo.Status.State
	}
	return "Synced" // default
}

func UpdateLayerMetrics(layer configv1alpha1.TerraformLayer) {
	m := GetMetrics()
	if m == nil {
		return
	}

	namespace := layer.Namespace
	layerName := layer.Name
	repositoryName := layer.Spec.Repository.Name
	status := GetLayerStatus(layer)

	// Clear all previous status metrics for this layer to avoid multiple statuses
	possibleStatuses := []string{"success", "error", "warning", "running", "disabled"}
	for _, s := range possibleStatuses {
		m.LayerStatusGauge.DeleteLabelValues(namespace, layerName, repositoryName, s)
	}

	// Update individual layer metrics
	// Set to 1 to indicate this layer exists with this status (status is identified by label)
	m.LayerStatusGauge.WithLabelValues(namespace, layerName, repositoryName, status).Set(1)
}

// UpdateRepositoryMetrics updates all metrics related to a specific repository
func UpdateRepositoryMetrics(repo configv1alpha1.TerraformRepository) {
	m := GetMetrics()
	if m == nil {
		return
	}

	namespace := repo.Namespace
	name := repo.Name
	url := repo.Spec.Repository.Url
	status := GetRepositoryStatus(repo)

	// Clear all previous status metrics for this repository to avoid multiple statuses
	possibleStatuses := []string{"Synced", "SyncNeeded"}
	for _, s := range possibleStatuses {
		m.RepositoryStatusGauge.DeleteLabelValues(namespace, name, url, s)
	}

	m.RepositoryStatusGauge.WithLabelValues(namespace, name, url, status).Set(1)
}

// UpdateRunMetrics updates all metrics related to runs
// This is used for aggregate metrics that are calculated periodically
func UpdateRunMetrics(run configv1alpha1.TerraformRun) {
	m := GetMetrics()
	if m == nil {
		return
	}
}

func RecordRunCreated(run configv1alpha1.TerraformRun) {
	m := GetMetrics()
	if m == nil {
		return
	}

	namespace := run.Namespace
	action := run.Spec.Action

	m.RunsCreatedTotal.WithLabelValues(namespace, action).Inc()
}

func RecordRunCompleted(run configv1alpha1.TerraformRun) {
	m := GetMetrics()
	if m == nil {
		return
	}

	namespace := run.Namespace
	action := run.Spec.Action

	m.RunsCompletedTotal.WithLabelValues(namespace, action).Inc()
}

func RecordRunFailed(run configv1alpha1.TerraformRun) {
	m := GetMetrics()
	if m == nil {
		return
	}

	namespace := run.Namespace
	action := run.Spec.Action

	m.RunsFailedTotal.WithLabelValues(namespace, action).Inc()
}

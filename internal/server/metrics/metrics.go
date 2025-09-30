package metrics

import (
	"context"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MetricsCollector collects and exposes Burrito metrics
type MetricsCollector struct {
	client   client.Client
	registry *prometheus.Registry

	// Layer metrics
	layerStatusGauge *prometheus.GaugeVec
	layerStateGauge  *prometheus.GaugeVec
	layerRunsTotal   *prometheus.CounterVec
	layerLastRunTime *prometheus.GaugeVec
	layerRunDuration *prometheus.GaugeVec
	layerConditions  *prometheus.GaugeVec

	// Repository metrics
	repositoryStatusGauge *prometheus.GaugeVec
	repositoriesTotal     prometheus.Gauge

	// Run metrics
	runStatusGauge *prometheus.GaugeVec
	runDuration    *prometheus.GaugeVec
	runRetries     *prometheus.GaugeVec

	// Pull Request metrics
	pullRequestStatusGauge *prometheus.GaugeVec
	pullRequestsTotal      prometheus.Gauge
}

// NewMetricsCollector creates a new metrics collector with its own registry
func NewMetricsCollector(client client.Client) *MetricsCollector {
	registry := prometheus.NewRegistry()

	layerStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_layer_status",
			Help: "Status of Terraform layers (0=disabled, 1=success, 2=warning, 3=error)",
		},
		[]string{"namespace", "name", "repository", "branch", "path", "status"},
	)

	layerStateGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_layer_state",
			Help: "State of Terraform layers (1=active for the given state)",
		},
		[]string{"namespace", "name", "repository", "branch", "path", "state"},
	)

	layerRunsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "burrito_terraform_layer_runs_total",
			Help: "Total number of runs for Terraform layers",
		},
		[]string{"namespace", "name", "repository", "branch", "path", "action"},
	)

	layerLastRunTime := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_layer_last_run_timestamp",
			Help: "Timestamp of the last run for Terraform layers",
		},
		[]string{"namespace", "name", "repository", "branch", "path", "action"},
	)

	layerRunDuration := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_layer_run_duration_seconds",
			Help: "Duration of the last run for Terraform layers in seconds",
		},
		[]string{"namespace", "name", "repository", "branch", "path", "action", "status"},
	)

	layerConditions := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_layer_condition",
			Help: "Condition status for Terraform layers (1=true, 0=false, -1=unknown)",
		},
		[]string{"namespace", "name", "repository", "branch", "path", "condition", "status", "reason"},
	)

	repositoryStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_repository_status",
			Help: "Status of Terraform repositories",
		},
		[]string{"namespace", "name", "url", "branch", "status"},
	)

	repositoriesTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_repositories_total",
			Help: "Total number of Terraform repositories",
		},
	)

	runStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_run_status",
			Help: "Status of Terraform runs",
		},
		[]string{"namespace", "name", "layer_name", "action", "state"},
	)

	runDuration := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_run_duration_seconds",
			Help: "Duration of Terraform runs in seconds",
		},
		[]string{"namespace", "name", "layer_name", "action", "state"},
	)

	runRetries := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_run_retries",
			Help: "Number of retries for Terraform runs",
		},
		[]string{"namespace", "name", "layer_name", "action", "state"},
	)

	pullRequestStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_pullrequest_status",
			Help: "Status of Terraform pull requests",
		},
		[]string{"namespace", "name", "repository", "pr_id", "state"},
	)

	pullRequestsTotal := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_pullrequests_total",
			Help: "Total number of Terraform pull requests",
		},
	)

	// Register all metrics with the registry
	registry.MustRegister(
		layerStatusGauge,
		layerStateGauge,
		layerRunsTotal,
		layerLastRunTime,
		layerRunDuration,
		layerConditions,
		repositoryStatusGauge,
		repositoriesTotal,
		runStatusGauge,
		runDuration,
		runRetries,
		pullRequestStatusGauge,
		pullRequestsTotal,
	)

	return &MetricsCollector{
		client:                 client,
		registry:               registry,
		layerStatusGauge:       layerStatusGauge,
		layerStateGauge:        layerStateGauge,
		layerRunsTotal:         layerRunsTotal,
		layerLastRunTime:       layerLastRunTime,
		layerRunDuration:       layerRunDuration,
		layerConditions:        layerConditions,
		repositoryStatusGauge:  repositoryStatusGauge,
		repositoriesTotal:      repositoriesTotal,
		runStatusGauge:         runStatusGauge,
		runDuration:            runDuration,
		runRetries:             runRetries,
		pullRequestStatusGauge: pullRequestStatusGauge,
		pullRequestsTotal:      pullRequestsTotal,
	}
}

// GetRegistry returns the Prometheus registry for this collector
func (mc *MetricsCollector) GetRegistry() *prometheus.Registry {
	return mc.registry
}

// CollectMetrics collects all metrics from Kubernetes resources
func (mc *MetricsCollector) CollectMetrics(ctx context.Context) error {
	// Reset all metrics to avoid stale data
	mc.resetMetrics()

	// Collect layer metrics
	if err := mc.collectLayerMetrics(ctx); err != nil {
		log.Errorf("Failed to collect layer metrics: %v", err)
		return err
	}

	// Collect repository metrics
	if err := mc.collectRepositoryMetrics(ctx); err != nil {
		log.Errorf("Failed to collect repository metrics: %v", err)
		return err
	}

	// Collect run metrics
	if err := mc.collectRunMetrics(ctx); err != nil {
		log.Errorf("Failed to collect run metrics: %v", err)
		return err
	}

	// Collect pull request metrics
	if err := mc.collectPullRequestMetrics(ctx); err != nil {
		log.Errorf("Failed to collect pull request metrics: %v", err)
		return err
	}

	return nil
}

func (mc *MetricsCollector) resetMetrics() {
	mc.layerStatusGauge.Reset()
	mc.layerStateGauge.Reset()
	mc.layerLastRunTime.Reset()
	mc.layerRunDuration.Reset()
	mc.layerConditions.Reset()
	mc.repositoryStatusGauge.Reset()
	mc.runStatusGauge.Reset()
	mc.runDuration.Reset()
	mc.runRetries.Reset()
	mc.pullRequestStatusGauge.Reset()
}

func (mc *MetricsCollector) collectLayerMetrics(ctx context.Context) error {
	var layers configv1alpha1.TerraformLayerList
	if err := mc.client.List(ctx, &layers); err != nil {
		return err
	}

	for _, layer := range layers.Items {
		namespace := layer.Namespace
		name := layer.Name
		repository := layer.Spec.Repository.Name
		branch := layer.Spec.Branch
		path := layer.Spec.Path

		// Layer status metrics (based on UI logic)
		status := mc.getLayerUIStatus(layer)
		statusValue := mc.statusToValue(status)

		mc.layerStatusGauge.WithLabelValues(namespace, name, repository, branch, path, status).Set(statusValue)

		// Layer state metrics
		state := layer.Status.State
		if state == "" {
			state = "unknown"
		}
		mc.layerStateGauge.WithLabelValues(namespace, name, repository, branch, path, state).Set(1)

		// Layer condition metrics
		for _, condition := range layer.Status.Conditions {
			conditionValue := mc.conditionToValue(condition.Status)
			mc.layerConditions.WithLabelValues(
				namespace, name, repository, branch, path,
				condition.Type, string(condition.Status), condition.Reason,
			).Set(conditionValue)
		}

		// Last run metrics
		if layer.Status.LastRun.Name != "" {
			runTime := layer.Status.LastRun.Date.Unix()
			action := layer.Status.LastRun.Action
			if action == "" {
				action = "unknown"
			}

			mc.layerLastRunTime.WithLabelValues(namespace, name, repository, branch, path, action).Set(float64(runTime))

			// Count total runs (approximation based on latest runs)
			mc.layerRunsTotal.WithLabelValues(namespace, name, repository, branch, path, action).Add(float64(len(layer.Status.LatestRuns)))
		}
	}

	return nil
}

func (mc *MetricsCollector) collectRepositoryMetrics(ctx context.Context) error {
	var repositories configv1alpha1.TerraformRepositoryList
	if err := mc.client.List(ctx, &repositories); err != nil {
		return err
	}

	mc.repositoriesTotal.Set(float64(len(repositories.Items)))

	for _, repo := range repositories.Items {
		namespace := repo.Namespace
		name := repo.Name
		url := repo.Spec.Repository.Url
		branch := "main" // Default branch since it's not stored in the repository spec

		// Repository status based on conditions
		status := "unknown"
		if len(repo.Status.Conditions) > 0 {
			// Check if any condition indicates an error
			hasError := false
			for _, condition := range repo.Status.Conditions {
				if condition.Status == metav1.ConditionFalse && condition.Type != "Idle" {
					hasError = true
					break
				}
			}
			if hasError {
				status = "error"
			} else {
				status = "success"
			}
		}

		statusValue := mc.statusToValue(status)
		mc.repositoryStatusGauge.WithLabelValues(namespace, name, url, branch, status).Set(statusValue)
	}

	return nil
}

func (mc *MetricsCollector) collectRunMetrics(ctx context.Context) error {
	var runs configv1alpha1.TerraformRunList
	if err := mc.client.List(ctx, &runs); err != nil {
		return err
	}

	for _, run := range runs.Items {
		namespace := run.Namespace
		name := run.Name
		layerName := run.Spec.Layer.Name
		action := run.Spec.Action
		state := run.Status.State

		if state == "" {
			state = "unknown"
		}

		mc.runStatusGauge.WithLabelValues(namespace, name, layerName, action, state).Set(1)

		// Run retries
		mc.runRetries.WithLabelValues(namespace, name, layerName, action, state).Set(float64(run.Status.Retries))

		// Run duration (if we can calculate it from creation time and last update)
		if run.CreationTimestamp.Time.IsZero() == false {
			var duration float64
			if len(run.Status.Conditions) > 0 {
				// Use the last condition update time as end time for finished runs
				lastCondition := run.Status.Conditions[len(run.Status.Conditions)-1]
				if state == "Succeeded" || state == "Failed" {
					duration = lastCondition.LastTransitionTime.Sub(run.CreationTimestamp.Time).Seconds()
				}
			}
			if duration > 0 {
				mc.runDuration.WithLabelValues(namespace, name, layerName, action, state).Set(duration)
			}
		}
	}

	return nil
}

func (mc *MetricsCollector) collectPullRequestMetrics(ctx context.Context) error {
	var pullRequests configv1alpha1.TerraformPullRequestList
	if err := mc.client.List(ctx, &pullRequests); err != nil {
		return err
	}

	mc.pullRequestsTotal.Set(float64(len(pullRequests.Items)))

	for _, pr := range pullRequests.Items {
		namespace := pr.Namespace
		name := pr.Name
		repository := pr.Spec.Repository.Name
		prID := pr.Spec.ID
		state := pr.Status.State

		if state == "" {
			state = "unknown"
		}

		mc.pullRequestStatusGauge.WithLabelValues(namespace, name, repository, prID, state).Set(1)
	}

	return nil
}

// getLayerUIStatus returns the UI status based on the layer state and conditions
// This mirrors the logic in internal/server/api/layers.go:getLayerState
func (mc *MetricsCollector) getLayerUIStatus(layer configv1alpha1.TerraformLayer) string {
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

	// Check for error conditions based on annotations
	if layer.Annotations[annotations.LastPlanSum] == "" {
		state = "error"
	}
	if layer.Annotations[annotations.LastApplySum] != "" && layer.Annotations[annotations.LastApplySum] == "" {
		state = "error"
	}

	// Check if layer is running
	for _, condition := range layer.Status.Conditions {
		if condition.Type == "IsRunning" && condition.Status == metav1.ConditionTrue {
			state = "running"
			break
		}
	}

	return state
}

func (mc *MetricsCollector) statusToValue(status string) float64 {
	switch status {
	case "disabled":
		return 0
	case "success":
		return 1
	case "warning":
		return 2
	case "error":
		return 3
	case "running":
		return 4
	default:
		return -1
	}
}

func (mc *MetricsCollector) conditionToValue(status metav1.ConditionStatus) float64 {
	switch status {
	case metav1.ConditionTrue:
		return 1
	case metav1.ConditionFalse:
		return 0
	case metav1.ConditionUnknown:
		return -1
	default:
		return -1
	}
}

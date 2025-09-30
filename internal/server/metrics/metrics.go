package metrics

import (
	"context"
	"sync"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Default refresh interval for background collection
	DefaultRefreshInterval = 30 * time.Second
	// Maximum number of runs to include in metrics (to control cardinality)
	MaxRunsInMetrics = 100
	// Cache TTL for metrics
	MetricsCacheTTL = 60 * time.Second
)

// MetricsCollector provides a scalable metrics collection implementation
type MetricsCollector struct {
	client          client.Client
	registry        *prometheus.Registry
	refreshInterval time.Duration
	maxRuns         int

	// Background collection
	stopCh         chan struct{}
	mutex          sync.RWMutex
	lastCollection time.Time
	cacheTTL       time.Duration

	// Optimized metrics with controlled cardinality
	layerStatusGauge *prometheus.GaugeVec
	layerStateGauge  *prometheus.GaugeVec
	layerRunsTotal   *prometheus.CounterVec
	layerLastRunTime *prometheus.GaugeVec

	// Summary metrics for better scalability
	layersByStatus    *prometheus.GaugeVec
	layersByNamespace *prometheus.GaugeVec
	runsByAction      *prometheus.GaugeVec
	runsByStatus      *prometheus.GaugeVec

	// Repository metrics
	repositoryStatusGauge *prometheus.GaugeVec

	// High-level aggregate metrics
	totalLayers       prometheus.Gauge
	totalRepositories prometheus.Gauge
	totalRuns         prometheus.Gauge
	totalPullRequests prometheus.Gauge

	// Performance metrics
	collectionDuration prometheus.Histogram
	apiCallsTotal      *prometheus.CounterVec

	// Error metrics
	collectionErrors *prometheus.CounterVec
}

type CollectorConfig struct {
	RefreshInterval time.Duration
	MaxRuns         int
	CacheTTL        time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(client client.Client, config *CollectorConfig) *MetricsCollector {
	if config == nil {
		config = &CollectorConfig{
			RefreshInterval: DefaultRefreshInterval,
			MaxRuns:         MaxRunsInMetrics,
		}
	}

	registry := prometheus.NewRegistry()

	// Individual layer metrics with layer and repository names
	layerStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_layer_status",
			Help: "Status of individual Terraform layers (0=disabled, 1=success, 2=warning, 3=error, 4=running)",
		},
		[]string{"namespace", "layer_name", "repository_name", "status"},
	)

	layerStateGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_layer_state",
			Help: "State of individual Terraform layers (1=active for the given state)",
		},
		[]string{"namespace", "layer_name", "repository_name", "state"},
	)

	layerRunsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "burrito_layer_runs_total",
			Help: "Total number of runs by namespace and action",
		},
		[]string{"namespace", "action", "status"},
	)

	layerLastRunTime := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_layer_last_run_timestamp",
			Help: "Timestamp of most recent run by namespace",
		},
		[]string{"namespace", "action"},
	)

	// Summary metrics
	layersByStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_layers_status_total",
			Help: "Total layers by status across all namespaces",
		},
		[]string{"status"},
	)

	layersByNamespace := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_layers_namespace_total",
			Help: "Total layers per namespace",
		},
		[]string{"namespace"},
	)

	runsByAction := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_runs_by_action_total",
			Help: "Total runs by action type",
		},
		[]string{"action"},
	)

	runsByStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_runs_by_status_total",
			Help: "Total runs by status",
		},
		[]string{"status"},
	)

	// Repository metrics
	repositoryStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "burrito_terraform_repository_status",
			Help: "Status of individual Terraform repositories (0=disabled, 1=success, 2=warning, 3=error)",
		},
		[]string{"namespace", "repository_name", "url", "status"},
	)

	// High-level aggregates
	totalLayers := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "burrito_layers_total",
			Help: "Total number of Terraform layers",
		},
	)

	totalRepositories := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "burrito_repositories_total",
			Help: "Total number of Terraform repositories",
		},
	)

	totalRuns := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "burrito_runs_total",
			Help: "Total number of Terraform runs",
		},
	)

	totalPullRequests := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "burrito_pullrequests_total",
			Help: "Total number of Terraform pull requests",
		},
	)

	// Performance metrics
	collectionDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "burrito_metrics_collection_duration_seconds",
			Help:    "Time spent collecting metrics",
			Buckets: prometheus.DefBuckets,
		},
	)

	apiCallsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "burrito_metrics_api_calls_total",
			Help: "Total API calls made during metrics collection",
		},
		[]string{"resource", "status"},
	)

	collectionErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "burrito_metrics_collection_errors_total",
			Help: "Total errors during metrics collection",
		},
		[]string{"resource", "error_type"},
	)

	// Register all metrics
	registry.MustRegister(
		layerStatusGauge, layerStateGauge, layerRunsTotal, layerLastRunTime,
		layersByStatus, layersByNamespace, runsByAction, runsByStatus,
		repositoryStatusGauge, totalLayers, totalRepositories, totalRuns, totalPullRequests,
		collectionDuration, apiCallsTotal, collectionErrors,
	)

	collector := &MetricsCollector{
		client:          client,
		registry:        registry,
		refreshInterval: config.RefreshInterval,
		maxRuns:         config.MaxRuns,
		cacheTTL:        MetricsCacheTTL,
		stopCh:          make(chan struct{}),

		layerStatusGauge:      layerStatusGauge,
		layerStateGauge:       layerStateGauge,
		layerRunsTotal:        layerRunsTotal,
		layerLastRunTime:      layerLastRunTime,
		layersByStatus:        layersByStatus,
		layersByNamespace:     layersByNamespace,
		runsByAction:          runsByAction,
		runsByStatus:          runsByStatus,
		repositoryStatusGauge: repositoryStatusGauge,
		totalLayers:           totalLayers,
		totalRepositories:     totalRepositories,
		totalRuns:             totalRuns,
		totalPullRequests:     totalPullRequests,
		collectionDuration:    collectionDuration,
		apiCallsTotal:         apiCallsTotal,
		collectionErrors:      collectionErrors,
	}

	return collector
}

// Start begins background metrics collection
func (smc *MetricsCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(smc.refreshInterval)
	defer ticker.Stop()

	// Initial collection
	smc.collectMetrics(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-smc.stopCh:
			return
		case <-ticker.C:
			smc.collectMetrics(ctx)
		}
	}
}

// Stop stops background collection
func (smc *MetricsCollector) Stop() {
	close(smc.stopCh)
}

// GetRegistry returns the Prometheus registry
func (smc *MetricsCollector) GetRegistry() *prometheus.Registry {
	return smc.registry
}

// GetMetrics returns cached metrics if still valid, otherwise triggers collection
func (smc *MetricsCollector) GetMetrics(ctx context.Context) error {
	smc.mutex.RLock()
	cacheValid := time.Since(smc.lastCollection) < smc.cacheTTL
	smc.mutex.RUnlock()

	if !cacheValid {
		return smc.collectMetrics(ctx)
	}

	return nil
}

func (smc *MetricsCollector) collectMetrics(ctx context.Context) error {
	start := time.Now()
	defer func() {
		smc.collectionDuration.Observe(time.Since(start).Seconds())
	}()

	smc.mutex.Lock()
	defer smc.mutex.Unlock()

	// Reset metrics
	smc.resetMetrics()

	// Collect all resource types with error handling
	if err := smc.collectLayerMetricsOptimized(ctx); err != nil {
		smc.collectionErrors.WithLabelValues("layers", "api_error").Inc()
		log.Errorf("Failed to collect layer metrics: %v", err)
	}

	if err := smc.collectRepositoryMetrics(ctx); err != nil {
		smc.collectionErrors.WithLabelValues("repositories", "api_error").Inc()
		log.Errorf("Failed to collect repository metrics: %v", err)
	}

	if err := smc.collectRunMetricsOptimized(ctx); err != nil {
		smc.collectionErrors.WithLabelValues("runs", "api_error").Inc()
		log.Errorf("Failed to collect run metrics: %v", err)
	}

	if err := smc.collectPullRequestMetricsOptimized(ctx); err != nil {
		smc.collectionErrors.WithLabelValues("pullrequests", "api_error").Inc()
		log.Errorf("Failed to collect pull request metrics: %v", err)
	}

	smc.lastCollection = time.Now()
	return nil
}

func (smc *MetricsCollector) resetMetrics() {
	smc.layerStatusGauge.Reset()
	smc.layerStateGauge.Reset()
	smc.layersByStatus.Reset()
	smc.layersByNamespace.Reset()
	smc.runsByAction.Reset()
	smc.runsByStatus.Reset()
	smc.repositoryStatusGauge.Reset()
}

func (smc *MetricsCollector) collectLayerMetricsOptimized(ctx context.Context) error {
	var layers configv1alpha1.TerraformLayerList

	// Use field selectors for optimization if needed
	listOptions := []client.ListOption{}

	smc.apiCallsTotal.WithLabelValues("layers", "started").Inc()
	if err := smc.client.List(ctx, &layers, listOptions...); err != nil {
		smc.apiCallsTotal.WithLabelValues("layers", "error").Inc()
		return err
	}
	smc.apiCallsTotal.WithLabelValues("layers", "success").Inc()

	// Aggregate counters
	statusCounts := make(map[string]map[string]int) // namespace -> status -> count
	stateCounts := make(map[string]map[string]int)  // namespace -> state -> count
	namespaceCounts := make(map[string]int)         // namespace -> total count
	globalStatusCounts := make(map[string]int)      // status -> total count

	var mostRecentRunTime int64

	for _, layer := range layers.Items {
		namespace := layer.Namespace
		layerName := layer.Name
		repositoryName := layer.Spec.Repository.Name

		// Initialize maps if needed
		if statusCounts[namespace] == nil {
			statusCounts[namespace] = make(map[string]int)
			stateCounts[namespace] = make(map[string]int)
		}

		// Get status and state
		status := smc.getLayerUIStatus(layer)
		state := layer.Status.State
		if state == "" {
			state = "unknown"
		}

		// Set individual layer metrics
		statusValue := smc.statusToValue(status)
		smc.layerStatusGauge.WithLabelValues(namespace, layerName, repositoryName, status).Set(statusValue)
		smc.layerStateGauge.WithLabelValues(namespace, layerName, repositoryName, state).Set(1)

		// Increment aggregate counters
		statusCounts[namespace][status]++
		stateCounts[namespace][state]++
		namespaceCounts[namespace]++
		globalStatusCounts[status]++

		// Track most recent run
		if layer.Status.LastRun.Name != "" && !layer.Status.LastRun.Date.Time.IsZero() {
			runTime := layer.Status.LastRun.Date.Unix()
			if runTime > mostRecentRunTime {
				mostRecentRunTime = runTime
				smc.layerLastRunTime.WithLabelValues(namespace, layer.Status.LastRun.Action).Set(float64(runTime))
			}
		}
	}

	// Set aggregated metrics
	smc.totalLayers.Set(float64(len(layers.Items)))

	for namespace, counts := range statusCounts {
		for status, count := range counts {
			smc.layerStatusGauge.WithLabelValues(namespace, status).Set(float64(count))
		}
	}

	for namespace, counts := range stateCounts {
		for state, count := range counts {
			smc.layerStateGauge.WithLabelValues(namespace, state).Set(float64(count))
		}
	}

	for namespace, count := range namespaceCounts {
		smc.layersByNamespace.WithLabelValues(namespace).Set(float64(count))
	}

	for status, count := range globalStatusCounts {
		smc.layersByStatus.WithLabelValues(status).Set(float64(count))
	}

	return nil
}

func (smc *MetricsCollector) collectRepositoryMetrics(ctx context.Context) error {
	var repositories configv1alpha1.TerraformRepositoryList

	smc.apiCallsTotal.WithLabelValues("repositories", "started").Inc()
	if err := smc.client.List(ctx, &repositories); err != nil {
		smc.apiCallsTotal.WithLabelValues("repositories", "error").Inc()
		return err
	}
	smc.apiCallsTotal.WithLabelValues("repositories", "success").Inc()

	smc.totalRepositories.Set(float64(len(repositories.Items)))

	for _, repo := range repositories.Items {
		namespace := repo.Namespace
		name := repo.Name
		url := repo.Spec.Repository.Url

		// Repository status based on conditions
		status := "success" // default
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
			}
		}

		statusValue := smc.statusToValue(status)
		smc.repositoryStatusGauge.WithLabelValues(namespace, name, url, status).Set(statusValue)
	}

	return nil
}

func (smc *MetricsCollector) collectRunMetricsOptimized(ctx context.Context) error {
	var runs configv1alpha1.TerraformRunList

	// Limit to recent runs to control cardinality
	listOptions := []client.ListOption{}
	if smc.maxRuns > 0 {
		listOptions = append(listOptions, client.Limit(int64(smc.maxRuns)))
	}

	smc.apiCallsTotal.WithLabelValues("runs", "started").Inc()
	if err := smc.client.List(ctx, &runs, listOptions...); err != nil {
		smc.apiCallsTotal.WithLabelValues("runs", "error").Inc()
		return err
	}
	smc.apiCallsTotal.WithLabelValues("runs", "success").Inc()

	smc.totalRuns.Set(float64(len(runs.Items)))

	// Aggregate run metrics
	actionCounts := make(map[string]int)
	statusCounts := make(map[string]int)
	runTotals := make(map[string]map[string]map[string]int) // namespace -> action -> status -> count

	for _, run := range runs.Items {
		action := run.Spec.Action
		state := run.Status.State
		if state == "" {
			state = "unknown"
		}
		namespace := run.Namespace

		actionCounts[action]++
		statusCounts[state]++

		// Initialize nested maps
		if runTotals[namespace] == nil {
			runTotals[namespace] = make(map[string]map[string]int)
		}
		if runTotals[namespace][action] == nil {
			runTotals[namespace][action] = make(map[string]int)
		}
		runTotals[namespace][action][state]++
	}

	// Set metrics
	for action, count := range actionCounts {
		smc.runsByAction.WithLabelValues(action).Set(float64(count))
	}

	for status, count := range statusCounts {
		smc.runsByStatus.WithLabelValues(status).Set(float64(count))
	}

	for namespace, actions := range runTotals {
		for action, states := range actions {
			for state, count := range states {
				smc.layerRunsTotal.WithLabelValues(namespace, action, state).Add(float64(count))
			}
		}
	}

	return nil
}

func (smc *MetricsCollector) collectPullRequestMetricsOptimized(ctx context.Context) error {
	var pullRequests configv1alpha1.TerraformPullRequestList

	smc.apiCallsTotal.WithLabelValues("pullrequests", "started").Inc()
	if err := smc.client.List(ctx, &pullRequests); err != nil {
		smc.apiCallsTotal.WithLabelValues("pullrequests", "error").Inc()
		return err
	}
	smc.apiCallsTotal.WithLabelValues("pullrequests", "success").Inc()

	smc.totalPullRequests.Set(float64(len(pullRequests.Items)))

	return nil
}

// getLayerUIStatus returns the UI status (same logic as original)
func (smc *MetricsCollector) getLayerUIStatus(layer configv1alpha1.TerraformLayer) string {
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

func (smc *MetricsCollector) statusToValue(status string) float64 {
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

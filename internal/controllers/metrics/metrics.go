package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type BurritoMetrics struct {
	LayerStatusGauge *prometheus.GaugeVec
	LayerStateGauge  *prometheus.GaugeVec

	LayersByStatus    *prometheus.GaugeVec
	LayersByNamespace *prometheus.GaugeVec

	RepositoryStatusGauge *prometheus.GaugeVec

	RunsByAction *prometheus.GaugeVec
	RunsByStatus *prometheus.GaugeVec

	RunsCreatedTotal   *prometheus.CounterVec
	RunsCompletedTotal *prometheus.CounterVec
	RunsFailedTotal    *prometheus.CounterVec

	TotalLayers       prometheus.Gauge
	TotalRepositories prometheus.Gauge
	TotalRuns         prometheus.Gauge
	TotalPullRequests prometheus.Gauge

	ReconcileDuration *prometheus.HistogramVec
	ReconcileTotal    *prometheus.CounterVec
	ReconcileErrors   *prometheus.CounterVec
}

var (
	Metrics *BurritoMetrics
)

// InitMetrics initializes and registers all Burrito metrics with controller-runtime's default registry
func InitMetrics() *BurritoMetrics {
	if Metrics != nil {
		return Metrics
	}

	m := &BurritoMetrics{
		LayerStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "burrito_terraform_layer_status",
				Help: "Status of individual Terraform layers (1=layer exists with this status, status is identified by label)",
			},
			[]string{"namespace", "layer_name", "repository_name", "status"},
		), LayerStateGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "burrito_terraform_layer_state",
				Help: "State of individual Terraform layers (1=active for the given state)",
			},
			[]string{"namespace", "layer_name", "repository_name", "state"},
		),

		LayersByStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "burrito_layers_status_total",
				Help: "Total layers by status across all namespaces",
			},
			[]string{"status"},
		),

		LayersByNamespace: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "burrito_layers_namespace_total",
				Help: "Total layers per namespace",
			},
			[]string{"namespace"},
		),

		RepositoryStatusGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "burrito_terraform_repository_status",
				Help: "Status of individual Terraform repositories (1=repository exists with this status, status is identified by label)",
			},
			[]string{"namespace", "repository_name", "url", "status"},
		),

		RunsByAction: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "burrito_runs_by_action",
				Help: "Current number of runs by action type",
			},
			[]string{"action"},
		),

		RunsByStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "burrito_runs_by_status",
				Help: "Current number of runs by status",
			},
			[]string{"status"},
		),

		RunsCreatedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "burrito_runs_created_total",
				Help: "Total number of runs created (cumulative)",
			},
			[]string{"namespace", "action"},
		),

		RunsCompletedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "burrito_runs_completed_total",
				Help: "Total number of runs completed successfully (cumulative)",
			},
			[]string{"namespace", "action"},
		),

		RunsFailedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "burrito_runs_failed_total",
				Help: "Total number of runs that failed (cumulative)",
			},
			[]string{"namespace", "action"},
		),

		// High-level aggregates
		TotalLayers: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "burrito_layers_total",
				Help: "Total number of Terraform layers",
			},
		),

		TotalRepositories: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "burrito_repositories_total",
				Help: "Total number of Terraform repositories",
			},
		),

		TotalRuns: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "burrito_runs_total",
				Help: "Total number of Terraform runs",
			},
		),

		TotalPullRequests: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "burrito_pullrequests_total",
				Help: "Total number of Terraform pull requests",
			},
		),

		// Reconciliation performance metrics
		ReconcileDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "burrito_controller_reconcile_duration_seconds",
				Help:    "Time spent in controller reconciliation",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"controller"},
		),

		ReconcileTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "burrito_controller_reconcile_total",
				Help: "Total number of reconciliations per controller",
			},
			[]string{"controller", "result"},
		),

		ReconcileErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "burrito_controller_reconcile_errors_total",
				Help: "Total reconciliation errors per controller",
			},
			[]string{"controller", "error_type"},
		),
	}

	// Register all metrics with controller-runtime's default registry
	metrics.Registry.MustRegister(
		m.LayerStatusGauge,
		m.LayerStateGauge,
		m.LayersByStatus,
		m.LayersByNamespace,
		m.RepositoryStatusGauge,
		m.RunsByAction,
		m.RunsByStatus,
		m.RunsCreatedTotal,
		m.RunsCompletedTotal,
		m.RunsFailedTotal,
		m.TotalLayers,
		m.TotalRepositories,
		m.TotalRuns,
		m.TotalPullRequests,
		m.ReconcileDuration,
		m.ReconcileTotal,
		m.ReconcileErrors,
	)

	Metrics = m
	return m
}

func GetMetrics() *BurritoMetrics {
	if Metrics == nil {
		return InitMetrics()
	}
	return Metrics
}

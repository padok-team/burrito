package metrics

import (
	"context"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// AggregationInterval is how often we recalculate aggregate metrics
	AggregationInterval = 30 * time.Second
)

// MetricsAggregator periodically calculates aggregate metrics
type MetricsAggregator struct {
	client   client.Client
	stopCh   chan struct{}
	interval time.Duration
}

func NewMetricsAggregator(client client.Client, interval time.Duration) *MetricsAggregator {
	if interval == 0 {
		interval = AggregationInterval
	}

	return &MetricsAggregator{
		client:   client,
		stopCh:   make(chan struct{}),
		interval: interval,
	}
}

func (ma *MetricsAggregator) Start(ctx context.Context) {
	ticker := time.NewTicker(ma.interval)
	defer ticker.Stop()

	ma.aggregate(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ma.stopCh:
			return
		case <-ticker.C:
			ma.aggregate(ctx)
		}
	}
}

func (ma *MetricsAggregator) Stop() {
	close(ma.stopCh)
}

func (ma *MetricsAggregator) aggregate(ctx context.Context) {
	m := GetMetrics()
	if m == nil {
		return
	}

	if err := ma.aggregateLayers(ctx, m); err != nil {
		log.Errorf("Failed to aggregate layer metrics: %v", err)
	}

	if err := ma.aggregateRepositories(ctx, m); err != nil {
		log.Errorf("Failed to aggregate repository metrics: %v", err)
	}

	if err := ma.aggregateRuns(ctx, m); err != nil {
		log.Errorf("Failed to aggregate run metrics: %v", err)
	}

	if err := ma.aggregatePullRequests(ctx, m); err != nil {
		log.Errorf("Failed to aggregate pull request metrics: %v", err)
	}
}

func (ma *MetricsAggregator) aggregateLayers(ctx context.Context, m *BurritoMetrics) error {
	var layers configv1alpha1.TerraformLayerList
	if err := ma.client.List(ctx, &layers); err != nil {
		return err
	}

	m.LayersByStatus.Reset()
	m.LayersByNamespace.Reset()

	statusCounts := make(map[string]int)
	namespaceCounts := make(map[string]int)

	for _, layer := range layers.Items {
		status := GetLayerUIStatus(layer)
		namespace := layer.Namespace

		statusCounts[status]++
		namespaceCounts[namespace]++
	}

	m.TotalLayers.Set(float64(len(layers.Items)))

	for status, count := range statusCounts {
		m.LayersByStatus.WithLabelValues(status).Set(float64(count))
	}

	for namespace, count := range namespaceCounts {
		m.LayersByNamespace.WithLabelValues(namespace).Set(float64(count))
	}

	return nil
}

func (ma *MetricsAggregator) aggregateRepositories(ctx context.Context, m *BurritoMetrics) error {
	var repositories configv1alpha1.TerraformRepositoryList
	if err := ma.client.List(ctx, &repositories); err != nil {
		return err
	}

	m.TotalRepositories.Set(float64(len(repositories.Items)))
	return nil
}

func (ma *MetricsAggregator) aggregateRuns(ctx context.Context, m *BurritoMetrics) error {
	var runs configv1alpha1.TerraformRunList
	if err := ma.client.List(ctx, &runs); err != nil {
		return err
	}

	m.RunsByAction.Reset()
	m.RunsByStatus.Reset()

	actionCounts := make(map[string]int)
	statusCounts := make(map[string]int)

	for _, run := range runs.Items {
		action := run.Spec.Action
		state := run.Status.State
		if state == "" {
			state = "unknown"
		}

		actionCounts[action]++
		statusCounts[state]++
	}

	m.TotalRuns.Set(float64(len(runs.Items)))

	for action, count := range actionCounts {
		m.RunsByAction.WithLabelValues(action).Set(float64(count))
	}

	for status, count := range statusCounts {
		m.RunsByStatus.WithLabelValues(status).Set(float64(count))
	}

	return nil
}

func (ma *MetricsAggregator) aggregatePullRequests(ctx context.Context, m *BurritoMetrics) error {
	var pullRequests configv1alpha1.TerraformPullRequestList
	if err := ma.client.List(ctx, &pullRequests); err != nil {
		return err
	}

	m.TotalPullRequests.Set(float64(len(pullRequests.Items)))
	return nil
}

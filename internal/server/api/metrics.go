package api

import (
	"context"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/padok-team/burrito/internal/server/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// Global metrics collector instance (singleton pattern for efficiency)
var (
	metricsCollector     *metrics.MetricsCollector
	metricsCollectorOnce sync.Once
	metricsCollectorMux  sync.RWMutex
)

// MetricsHandler handles the /metrics endpoint with optimized collection
func (a *API) MetricsHandler(c echo.Context) error {
	collector := a.getOrCreateMetricsCollector()

	// Use cached metrics or trigger collection if cache expired
	ctx := context.Background()
	if err := collector.GetMetrics(ctx); err != nil {
		log.Errorf("Failed to get metrics: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to collect metrics")
	}

	// Serve metrics from the collector's registry
	handler := promhttp.HandlerFor(collector.GetRegistry(), promhttp.HandlerOpts{
		ErrorLog:      log.StandardLogger(),
		ErrorHandling: promhttp.ContinueOnError,
	})
	handler.ServeHTTP(c.Response().Writer, c.Request())

	return nil
}

// getOrCreateMetricsCollector returns the singleton metrics collector instance
func (a *API) getOrCreateMetricsCollector() *metrics.MetricsCollector {
	metricsCollectorMux.RLock()
	if metricsCollector != nil {
		defer metricsCollectorMux.RUnlock()
		return metricsCollector
	}
	metricsCollectorMux.RUnlock()

	metricsCollectorOnce.Do(func() {
		metricsCollectorMux.Lock()
		defer metricsCollectorMux.Unlock()

		config := &metrics.CollectorConfig{
			RefreshInterval: metrics.DefaultRefreshInterval,
			MaxRuns:         metrics.MaxRunsInMetrics,
		}

		metricsCollector = metrics.NewMetricsCollector(a.Client, config)

		// Start background collection
		go func() {
			ctx := context.Background()
			metricsCollector.Start(ctx)
		}()

		log.Info("Started metrics collector with background refresh")
	})

	return metricsCollector
}

// StopMetricsCollector stops the background metrics collection
func StopMetricsCollector() {
	metricsCollectorMux.RLock()
	defer metricsCollectorMux.RUnlock()

	if metricsCollector != nil {
		metricsCollector.Stop()
		log.Info("Stopped metrics collector")
	}
}

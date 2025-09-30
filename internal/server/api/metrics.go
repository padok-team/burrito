package api

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/padok-team/burrito/internal/server/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

// MetricsHandler handles the /metrics endpoint for Prometheus scraping
func (a *API) MetricsHandler(c echo.Context) error {
	// Create a new metrics collector with the client
	collector := metrics.NewMetricsCollector(a.Client)

	// Collect metrics from Kubernetes resources
	ctx := context.Background()
	if err := collector.CollectMetrics(ctx); err != nil {
		log.Errorf("Failed to collect metrics: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to collect metrics")
	}

	// Use Prometheus HTTP handler with the collector's registry to serve metrics
	handler := promhttp.HandlerFor(collector.GetRegistry(), promhttp.HandlerOpts{})
	handler.ServeHTTP(c.Response().Writer, c.Request())

	return nil
}

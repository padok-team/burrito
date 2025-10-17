package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

func TestMetricsHandler(t *testing.T) {
	// Create a fake Kubernetes client with test data
	scheme := runtime.NewScheme()
	_ = configv1alpha1.AddToScheme(scheme)

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Create API instance
	api := &API{
		Client: client,
	}

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := api.MetricsHandler(c)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check that response contains Prometheus metrics
	responseBody := rec.Body.String()
	assert.Contains(t, responseBody, "# HELP")
	assert.Contains(t, responseBody, "burrito_")

	// Check for specific metrics (metrics that are always present)
	expectedMetrics := []string{
		"burrito_layers_total",
		"burrito_repositories_total",
		"burrito_runs_total",
		"burrito_metrics_collection_duration_seconds",
	}

	for _, metric := range expectedMetrics {
		assert.Contains(t, responseBody, metric, "Expected metric %s not found in response", metric)
	}
}

func TestMetricsHandlerContentType(t *testing.T) {
	// Create a fake Kubernetes client
	scheme := runtime.NewScheme()
	_ = configv1alpha1.AddToScheme(scheme)

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	// Create API instance
	api := &API{
		Client: client,
	}

	// Create test request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute handler
	err := api.MetricsHandler(c)

	// Assertions
	assert.NoError(t, err)

	// Check content type is correct for Prometheus
	contentType := rec.Header().Get("Content-Type")
	assert.True(t, strings.Contains(contentType, "text/plain") || strings.Contains(contentType, "text"),
		"Expected text content type for Prometheus metrics, got: %s", contentType)
}

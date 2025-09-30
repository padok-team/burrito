package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

func TestMetricsHandler(t *testing.T) {
	// Create a fake Kubernetes client with test data
	scheme := runtime.NewScheme()
	_ = configv1alpha1.AddToScheme(scheme)

	// Create some test layers
	testLayer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-layer",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "test-repo",
				Namespace: "default",
			},
			Branch: "main",
			Path:   "/terraform",
		},
		Status: configv1alpha1.TerraformLayerStatus{
			State: "Idle",
			Conditions: []metav1.Condition{
				{
					Type:   "IsRunning",
					Status: metav1.ConditionFalse,
					Reason: "NotRunning",
				},
			},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(testLayer).
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
	assert.Contains(t, responseBody, "burrito_terraform")

	// Check for specific metrics
	expectedMetrics := []string{
		"burrito_terraform_layer_status",
		"burrito_terraform_layer_state",
		"burrito_terraform_repositories_total",
		"burrito_terraform_pullrequests_total",
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

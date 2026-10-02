package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestGetLayerState(t *testing.T) {
	baseAnnotations := map[string]string{
		annotations.LastPlanSum: "some-sum",
	}
	someCondition := []metav1.Condition{{
		Type:   "some-condition",
		Status: metav1.ConditionTrue,
		Reason: "some-reason",
	}}

	tests := []struct {
		name     string
		layer    configv1alpha1.TerraformLayer
		expected string
	}{
		{
			name: "no conditions is disabled",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{Annotations: baseAnnotations},
				Status:     configv1alpha1.TerraformLayerStatus{},
			},
			expected: "disabled",
		},
		{
			name: "max retries reached",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{Annotations: baseAnnotations},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: someCondition,
					State:      "MaxRetriesReached",
				},
			},
			expected: "retriesExhausted",
		},
		{
			name: "max retries reached without plan sum stays retries exhausted",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: someCondition,
					State:      "MaxRetriesReached",
				},
			},
			expected: "retriesExhausted",
		},
		{
			name: "missing plan sum is an error",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: someCondition,
					State:      "Idle",
				},
			},
			expected: "error",
		},
		{
			name: "default state is success",
			layer: configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{Annotations: baseAnnotations},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: someCondition,
					State:      "Idle",
				},
			},
			expected: "success",
		},
	}

	a := &API{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := a.getLayerState(tt.layer); got != tt.expected {
				t.Errorf("getLayerState() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// stubDatastore overrides the datastore calls used by the server API handlers.
type stubDatastore struct {
	datastore.Client
	content []byte
	err     error
}

func (s *stubDatastore) GetPlan(namespace, layer, run, attempt, format string) ([]byte, error) {
	return s.content, s.err
}

func (s *stubDatastore) GetStateGraph(namespace, layer string) ([]byte, error) {
	return s.content, s.err
}

func newTestContext(names, values []string) (echo.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest(http.MethodGet, "/", nil), rec)
	c.SetParamNames(names...)
	c.SetParamValues(values...)
	return c, rec
}

func newTestAPI(t *testing.T, funcs *interceptor.Funcs, objects ...client.Object) *API {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("could not build scheme: %s", err)
	}
	builder := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...)
	if funcs != nil {
		builder = builder.WithInterceptorFuncs(*funcs)
	}
	return &API{Client: builder.Build()}
}

func testObjects() (*configv1alpha1.TerraformRepository, *configv1alpha1.TerraformLayer, *configv1alpha1.TerraformRun) {
	repo := &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"},
	}
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{Name: "layer", Namespace: "default", Annotations: map[string]string{annotations.LastPlanSum: "sum"}},
		Spec: configv1alpha1.TerraformLayerSpec{
			Path:       "terraform/",
			Branch:     "main",
			Repository: configv1alpha1.TerraformLayerRepository{Name: "repo", Namespace: "default"},
		},
		Status: configv1alpha1.TerraformLayerStatus{
			LastRun: configv1alpha1.TerraformLayerRun{Name: "run"},
			LatestRuns: []configv1alpha1.TerraformLayerRun{
				{Name: "run", Commit: "abc", Author: "me", Message: "msg", Action: "plan"},
			},
		},
	}
	run := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{Name: "run", Namespace: "default"},
		Spec:       configv1alpha1.TerraformRunSpec{Action: "plan"},
		Status:     configv1alpha1.TerraformRunStatus{State: "Succeeded", Commit: "abc", Author: "me", Message: "msg"},
	}
	return repo, layer, run
}

func TestLayersHandler(t *testing.T) {
	repo, layer, run := testObjects()
	a := newTestAPI(t, nil, repo, layer, run)
	c, rec := newTestContext(nil, nil)
	if err := a.LayersHandler(c); err != nil {
		t.Fatalf("LayersHandler() error = %s", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("LayersHandler() status = %d, want %d", rec.Code, http.StatusOK)
	}
	resp := layersResponse{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("could not decode response: %s", err)
	}
	if len(resp.Results) != 1 || resp.Results[0].LastRun.Author != "me" || resp.Results[0].IsRunning {
		t.Errorf("LayersHandler() results = %+v", resp.Results)
	}
}

func TestLayersHandlerListErrors(t *testing.T) {
	for _, kind := range []string{"TerraformLayerList", "TerraformRunList", "TerraformRepositoryList"} {
		t.Run(kind, func(t *testing.T) {
			a := newTestAPI(t, &interceptor.Funcs{
				List: func(ctx context.Context, cl client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
					if fmt.Sprintf("%T", list) == "*v1alpha1."+kind {
						return errors.New("boom")
					}
					return cl.List(ctx, list, opts...)
				},
			})
			c, rec := newTestContext(nil, nil)
			if err := a.LayersHandler(c); err != nil {
				t.Fatalf("LayersHandler() error = %s", err)
			}
			if rec.Code != http.StatusInternalServerError {
				t.Errorf("LayersHandler() status = %d, want %d", rec.Code, http.StatusInternalServerError)
			}
		})
	}
}

func TestLayerHandler(t *testing.T) {
	repo, layer, run := testObjects()
	missingRun := layer.DeepCopy()
	missingRun.Name = "missing-run"
	missingRun.Status.LastRun.Name = "nope"
	missingRepo := layer.DeepCopy()
	missingRepo.Name = "missing-repo"
	missingRepo.Spec.Repository.Name = "nope"
	getErr := &interceptor.Funcs{
		Get: func(ctx context.Context, cl client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if key.Name == "repo" {
				return errors.New("boom")
			}
			return cl.Get(ctx, key, obj, opts...)
		},
	}

	tests := []struct {
		name     string
		funcs    *interceptor.Funcs
		layer    string
		expected int
	}{
		{name: "found", layer: "layer", expected: http.StatusOK},
		{name: "last run missing", layer: "missing-run", expected: http.StatusOK},
		{name: "layer not found", layer: "unknown", expected: http.StatusNotFound},
		{name: "repository not found", layer: "missing-repo", expected: http.StatusNotFound},
		{name: "repository get error", funcs: getErr, layer: "layer", expected: http.StatusInternalServerError},
		{name: "layer get error", funcs: getErr, layer: "repo", expected: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestAPI(t, tt.funcs, repo, layer, run, missingRun, missingRepo)
			c, rec := newTestContext([]string{"namespace", "layer"}, []string{"default", tt.layer})
			if err := a.LayerHandler(c); err != nil {
				t.Fatalf("LayerHandler() error = %s", err)
			}
			if rec.Code != tt.expected {
				t.Errorf("LayerHandler() status = %d, want %d", rec.Code, tt.expected)
			}
		})
	}
}

func TestGetStateGraphHandler(t *testing.T) {
	tests := []struct {
		name     string
		values   []string
		store    *stubDatastore
		expected int
	}{
		{name: "ok", values: []string{"default", "layer"}, store: &stubDatastore{content: []byte(`{}`)}, expected: http.StatusOK},
		{name: "missing params", values: []string{"default", ""}, store: &stubDatastore{}, expected: http.StatusBadRequest},
		{name: "storage error", values: []string{"default", "layer"}, store: &stubDatastore{err: errors.New("boom")}, expected: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &API{Datastore: tt.store}
			c, rec := newTestContext([]string{"namespace", "layer"}, tt.values)
			if err := a.GetStateGraphHandler(c); err != nil {
				t.Fatalf("GetStateGraphHandler() error = %s", err)
			}
			if rec.Code != tt.expected {
				t.Errorf("GetStateGraphHandler() status = %d, want %d", rec.Code, tt.expected)
			}
		})
	}
}

func TestGetPlanHandler(t *testing.T) {
	ok := []string{"default", "layer", "run", "0"}
	tests := []struct {
		name     string
		values   []string
		store    *stubDatastore
		expected int
	}{
		{name: "ok", values: ok, store: &stubDatastore{content: []byte(`{}`)}, expected: http.StatusOK},
		{name: "missing params", values: []string{"default", "layer", "run", ""}, store: &stubDatastore{}, expected: http.StatusBadRequest},
		{name: "not found", values: ok, store: &stubDatastore{err: &storageerrors.StorageError{Err: errors.New("nil"), Nil: true}}, expected: http.StatusNotFound},
		{name: "storage error", values: ok, store: &stubDatastore{err: errors.New("boom")}, expected: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &API{Datastore: tt.store}
			c, rec := newTestContext([]string{"namespace", "layer", "run", "attempt"}, tt.values)
			if err := a.GetPlanHandler(c); err != nil {
				t.Fatalf("GetPlanHandler() error = %s", err)
			}
			if rec.Code != tt.expected {
				t.Errorf("GetPlanHandler() status = %d, want %d", rec.Code, tt.expected)
			}
		})
	}
}

// nolint
package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/server/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestServerAPI(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Server API Suite")
}

func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = configv1alpha1.AddToScheme(scheme)
	return scheme
}

func getContext(method string, path string) *echo.Context {
	e := echo.New()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

var _ = Describe("Server API", func() {
	repository := &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "repo",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformRepositorySpec{
			Repository: configv1alpha1.TerraformRepositoryRepository{
				Url: "https://github.com/padok-team/burrito.git",
			},
		},
	}
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-layer",
			Namespace: "default",
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Path:   "terraform/",
			Branch: "main",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      repository.Name,
				Namespace: repository.Namespace,
			},
		},
	}
	run := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-run",
			Namespace: "default",
		},
		Status: configv1alpha1.TerraformRunStatus{
			Attempts: []configv1alpha1.Attempt{
				{PodName: "pod-1", Number: 0},
			},
		},
	}

	newAPI := func() *api.API {
		cl := fake.NewClientBuilder().
			WithScheme(newTestScheme()).
			WithObjects(repository, layer.DeepCopy(), run).
			Build()
		return &api.API{
			Client:    cl,
			Datastore: datastore.NewMockClient(),
		}
	}

	Describe("LayersHandler", func() {
		It("should return 200 OK", func() {
			a := newAPI()
			context := getContext(http.MethodGet, "/api/layers")
			err := a.LayersHandler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})
	})

	Describe("SyncLayerHandler", func() {
		It("should return 200 OK", func() {
			a := newAPI()
			context := getContext(http.MethodPost, "/api/layers/default/test-layer/sync")
			context.SetPathValues(echo.PathValues{
				{Name: "namespace", Value: "default"},
				{Name: "layer", Value: "test-layer"},
			})
			err := a.SyncLayerHandler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})
	})

	Describe("RepositoriesHandler", func() {
		It("should return 200 OK", func() {
			a := newAPI()
			context := getContext(http.MethodGet, "/api/repositories")
			err := a.RepositoriesHandler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})
	})

	Describe("GetLogsHandler", func() {
		It("should return 200 OK", func() {
			a := newAPI()
			context := getContext(http.MethodGet, "/api/logs/default/test-layer/test-run/0")
			context.SetPathValues(echo.PathValues{
				{Name: "namespace", Value: "default"},
				{Name: "layer", Value: "test-layer"},
				{Name: "run", Value: "test-run"},
				{Name: "attempt", Value: "0"},
			})
			err := a.GetLogsHandler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})
	})

	Describe("GetAttemptsHandler", func() {
		It("should return 200 OK", func() {
			a := newAPI()
			context := getContext(http.MethodGet, "/api/run/default/test-run/attempts")
			context.SetPathValues(echo.PathValues{
				{Name: "namespace", Value: "default"},
				{Name: "run", Value: "test-run"},
			})
			err := a.GetAttemptsHandler(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(context.Response().(*echo.Response).Status).To(Equal(http.StatusOK))
		})
	})
})

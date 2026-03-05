package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/server/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Sync API", func() {
	var e *echo.Echo

	BeforeEach(func() {
		e = echo.New()
	})

	Describe("ApplyLayerHandler", func() {
		It("should trigger apply on an existing layer", func() {
			layer := &configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "my-layer",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:   "modules/test",
					Branch: "main",
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				WithObjects(layer).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodPost, "/api/layers/default/my-layer/apply", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setRouteParams(c, []string{"namespace", "layer"}, []string{"default", "my-layer"})

			err := a.ApplyLayerHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var body map[string]string
			Expect(json.Unmarshal(rec.Body.Bytes(), &body)).NotTo(HaveOccurred())
			Expect(body["status"]).To(Equal("Layer apply triggered"))
		})

		It("should return error when layer does not exist", func() {
			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodPost, "/api/layers/default/nonexistent/apply", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setRouteParams(c, []string{"namespace", "layer"}, []string{"default", "nonexistent"})

			err := a.ApplyLayerHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("should return conflict when layer is managed by TerraformPullRequest", func() {
			layer := &configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pr-layer",
					Namespace: "default",
					Labels: map[string]string{
						"burrito/managed-by": "terraform-pullrequest",
					},
					Annotations: map[string]string{},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:   "modules/test",
					Branch: "main",
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				WithObjects(layer).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodPost, "/api/layers/default/pr-layer/apply", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setRouteParams(c, []string{"namespace", "layer"}, []string{"default", "pr-layer"})

			err := a.ApplyLayerHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusConflict))

			var body map[string]string
			Expect(json.Unmarshal(rec.Body.Bytes(), &body)).NotTo(HaveOccurred())
			Expect(body["error"]).To(ContainSubstring("Manual apply is not allowed"))
		})

		It("should not block when managed-by label exists but is empty", func() {
			layer := &configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-label-layer",
					Namespace: "default",
					Labels: map[string]string{
						"burrito/managed-by": "",
					},
					Annotations: map[string]string{},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:   "modules/test",
					Branch: "main",
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				WithObjects(layer).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodPost, "/api/layers/default/empty-label-layer/apply", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setRouteParams(c, []string{"namespace", "layer"}, []string{"default", "empty-label-layer"})

			err := a.ApplyLayerHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	Describe("SyncLayerHandler", func() {
		It("should trigger sync on an existing layer", func() {
			layer := &configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "sync-layer",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:   "modules/test",
					Branch: "main",
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				WithObjects(layer).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodPost, "/api/layers/default/sync-layer/sync", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setRouteParams(c, []string{"namespace", "layer"}, []string{"default", "sync-layer"})

			err := a.SyncLayerHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var body map[string]string
			Expect(json.Unmarshal(rec.Body.Bytes(), &body)).NotTo(HaveOccurred())
			Expect(body["status"]).To(Equal("Layer sync triggered"))
		})

		It("should return conflict when sync is already pending", func() {
			layer := &configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "already-syncing",
					Namespace: "default",
					Annotations: map[string]string{
						annotations.SyncNow: "true",
					},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:   "modules/test",
					Branch: "main",
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				WithObjects(layer).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodPost, "/api/layers/default/already-syncing/sync", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setRouteParams(c, []string{"namespace", "layer"}, []string{"default", "already-syncing"})

			err := a.SyncLayerHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusConflict))
		})

		It("should return error when layer does not exist", func() {
			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodPost, "/api/layers/default/nonexistent/sync", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setRouteParams(c, []string{"namespace", "layer"}, []string{"default", "nonexistent"})

			err := a.SyncLayerHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})

})

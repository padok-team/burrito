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
	"github.com/padok-team/burrito/internal/server/utils"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Layers API", func() {
	var e *echo.Echo

	BeforeEach(func() {
		e = echo.New()
	})

	Describe("LayersHandler", func() {
		It("should return layers with hasValidPlan, manualSyncStatus, and autoApply computed correctly", func() {
			// Layer with a valid plan + apply-now annotation + autoApply
			layerWithPlanAndApply := &configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "layer-with-plan",
					Namespace: "default",
					Annotations: map[string]string{
						annotations.LastPlanSum: "AuP6pMNxWsbSZKnxZvxD842wy0qaF9JCX8HW1nFeL1I=",
						annotations.ApplyNow:    "true",
					},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:   "modules/vpc",
					Branch: "main",
					Repository: configv1alpha1.TerraformLayerRepository{
						Name:      "my-repo",
						Namespace: "default",
					},
				},
			}

			// Layer without plan sum + sync-now annotation + autoApply false
			layerWithoutPlan := &configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "layer-no-plan",
					Namespace: "default",
					Annotations: map[string]string{
						annotations.SyncNow: "true",
					},
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Path:   "modules/rds",
					Branch: "main",
					Repository: configv1alpha1.TerraformLayerRepository{
						Name:      "my-repo",
						Namespace: "default",
					},
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: boolPtr(false),
					},
				},
			}

			repo := &configv1alpha1.TerraformRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "my-repo",
					Namespace: "default",
				},
				Spec: configv1alpha1.TerraformRepositorySpec{
					RemediationStrategy: configv1alpha1.RemediationStrategy{
						AutoApply: boolPtr(true),
					},
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(newScheme()).
				WithObjects(layerWithPlanAndApply, layerWithoutPlan, repo).
				Build()

			a := &api.API{Client: fakeClient}

			req := httptest.NewRequest(http.MethodGet, "/api/layers", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := a.LayersHandler(c)
			Expect(err).NotTo(HaveOccurred())
			Expect(rec.Code).To(Equal(http.StatusOK))

			var resp struct {
				Results []struct {
					Name             string                 `json:"name"`
					HasValidPlan     bool                   `json:"hasValidPlan"`
					ManualSyncStatus utils.ManualSyncStatus `json:"manualSyncStatus"`
					AutoApply        bool                   `json:"autoApply"`
				} `json:"results"`
			}
			Expect(json.Unmarshal(rec.Body.Bytes(), &resp)).NotTo(HaveOccurred())
			Expect(resp.Results).To(HaveLen(2))

			// Find each layer in results (order not guaranteed)
			byName := map[string]struct {
				HasValidPlan     bool
				ManualSyncStatus utils.ManualSyncStatus
				AutoApply        bool
			}{}
			for _, r := range resp.Results {
				byName[r.Name] = struct {
					HasValidPlan     bool
					ManualSyncStatus utils.ManualSyncStatus
					AutoApply        bool
				}{r.HasValidPlan, r.ManualSyncStatus, r.AutoApply}
			}

			// layer-with-plan: has plan sum → hasValidPlan=true, apply-now → manualSyncStatus=annotated, repo autoApply=true
			Expect(byName["layer-with-plan"].HasValidPlan).To(BeTrue())
			Expect(byName["layer-with-plan"].ManualSyncStatus).To(Equal(utils.ManualSyncAnnotated))
			Expect(byName["layer-with-plan"].AutoApply).To(BeTrue())

			// layer-no-plan: no plan sum → hasValidPlan=false, sync-now → manualSyncStatus=annotated, repo autoApply=true (inherited)
			Expect(byName["layer-no-plan"].HasValidPlan).To(BeFalse())
			Expect(byName["layer-no-plan"].ManualSyncStatus).To(Equal(utils.ManualSyncAnnotated))
			Expect(byName["layer-no-plan"].AutoApply).To(BeFalse())
		})
	})
})

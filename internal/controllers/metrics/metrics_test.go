package metrics_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/metrics"
)

func TestMetrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Metrics Suite")
}

var _ = BeforeSuite(func() {
	// Reset global metrics for testing
	metrics.Metrics = nil
})

var _ = Describe("Metrics", func() {
	Describe("InitMetrics", func() {
		It("should initialize all metrics", func() {
			metrics.Metrics = nil
			m := metrics.InitMetrics()

			Expect(m).NotTo(BeNil())
			Expect(m.LayerStatusGauge).NotTo(BeNil())
			Expect(m.RepositoryStatusGauge).NotTo(BeNil())
			Expect(m.TotalLayers).NotTo(BeNil())
			Expect(m.TotalRepositories).NotTo(BeNil())
			Expect(m.TotalRuns).NotTo(BeNil())
			Expect(m.TotalPullRequests).NotTo(BeNil())
		})

		It("should return the same instance when called again", func() {
			m1 := metrics.InitMetrics()
			m2 := metrics.GetMetrics()
			Expect(m1).To(Equal(m2))
		})
	})

	Describe("GetLayerUIStatus", func() {
		Context("when layer has no conditions", func() {
			It("should return disabled", func() {
				layer := configv1alpha1.TerraformLayer{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"runner.terraform.padok.cloud/plan-sum": "abc123",
						},
					},
					Status: configv1alpha1.TerraformLayerStatus{
						Conditions: []metav1.Condition{},
					},
				}
				Expect(metrics.GetLayerStatus(layer)).To(Equal("disabled"))
			})
		})

		Context("when layer is in ApplyNeeded state with no changes", func() {
			It("should return success", func() {
				layer := configv1alpha1.TerraformLayer{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"runner.terraform.padok.cloud/plan-sum": "abc123",
						},
					},
					Status: configv1alpha1.TerraformLayerStatus{
						Conditions: []metav1.Condition{
							{Type: "Ready", Status: metav1.ConditionTrue},
						},
						State:      "ApplyNeeded",
						LastResult: "Plan: 0 to create, 0 to update, 0 to delete",
					},
				}
				Expect(metrics.GetLayerStatus(layer)).To(Equal("success"))
			})
		})

		Context("when layer is in ApplyNeeded state with changes", func() {
			It("should return warning", func() {
				layer := configv1alpha1.TerraformLayer{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"runner.terraform.padok.cloud/plan-sum": "abc123",
						},
					},
					Status: configv1alpha1.TerraformLayerStatus{
						Conditions: []metav1.Condition{
							{Type: "Ready", Status: metav1.ConditionTrue},
						},
						State:      "ApplyNeeded",
						LastResult: "Plan: 1 to create, 0 to update, 0 to delete",
					},
				}
				Expect(metrics.GetLayerStatus(layer)).To(Equal("warning"))
			})
		})

		Context("when layer is in PlanNeeded state", func() {
			It("should return warning", func() {
				layer := configv1alpha1.TerraformLayer{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"runner.terraform.padok.cloud/plan-sum": "abc123",
						},
					},
					Status: configv1alpha1.TerraformLayerStatus{
						Conditions: []metav1.Condition{
							{Type: "Ready", Status: metav1.ConditionTrue},
						},
						State: "PlanNeeded",
					},
				}
				Expect(metrics.GetLayerStatus(layer)).To(Equal("warning"))
			})
		})

		Context("when layer has IsRunning condition", func() {
			It("should return running", func() {
				layer := configv1alpha1.TerraformLayer{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"runner.terraform.padok.cloud/plan-sum": "abc123",
						},
					},
					Status: configv1alpha1.TerraformLayerStatus{
						Conditions: []metav1.Condition{
							{Type: "IsRunning", Status: metav1.ConditionTrue},
						},
						State: "PlanNeeded",
					},
				}
				Expect(metrics.GetLayerStatus(layer)).To(Equal("running"))
			})
		})
	})

	Describe("GetRepositoryStatus", func() {
		Context("when repository has no conditions", func() {
			It("should return Synced", func() {
				repo := configv1alpha1.TerraformRepository{
					Status: configv1alpha1.TerraformRepositoryStatus{
						Conditions: []metav1.Condition{},
					},
				}
				Expect(metrics.GetRepositoryStatus(repo)).To(Equal("Synced"))
			})
		})

		Context("when repository state is Synced", func() {
			It("should return Synced", func() {
				repo := configv1alpha1.TerraformRepository{
					Status: configv1alpha1.TerraformRepositoryStatus{
						State: "Synced",
						Conditions: []metav1.Condition{
							{Type: "Ready", Status: metav1.ConditionTrue},
						},
					},
				}
				Expect(metrics.GetRepositoryStatus(repo)).To(Equal("Synced"))
			})
		})

		Context("when repository state is SyncNeeded", func() {
			It("should return SyncNeeded", func() {
				repo := configv1alpha1.TerraformRepository{
					Status: configv1alpha1.TerraformRepositoryStatus{
						State: "SyncNeeded",
						Conditions: []metav1.Condition{
							{Type: "Ready", Status: metav1.ConditionFalse},
						},
					},
				}
				Expect(metrics.GetRepositoryStatus(repo)).To(Equal("SyncNeeded"))
			})
		})
	})

	Describe("UpdateLayerMetrics", func() {
		BeforeEach(func() {
			if metrics.Metrics == nil {
				metrics.InitMetrics()
			}
		})

		It("should not panic when updating layer metrics", func() {
			layer := configv1alpha1.TerraformLayer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-layer",
					Namespace: "test-namespace",
				},
				Spec: configv1alpha1.TerraformLayerSpec{
					Repository: configv1alpha1.TerraformLayerRepository{
						Name: "test-repo",
					},
				},
				Status: configv1alpha1.TerraformLayerStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
					},
					State: "Idle",
				},
			}

			Expect(func() {
				metrics.UpdateLayerMetrics(layer)
			}).NotTo(Panic())
			Expect(metrics.Metrics).NotTo(BeNil())
		})
	})

	Describe("UpdateRepositoryMetrics", func() {
		BeforeEach(func() {
			if metrics.Metrics == nil {
				metrics.InitMetrics()
			}
		})

		It("should not panic when updating repository metrics", func() {
			repo := configv1alpha1.TerraformRepository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-repo",
					Namespace: "test-namespace",
				},
				Spec: configv1alpha1.TerraformRepositorySpec{
					Repository: configv1alpha1.TerraformRepositoryRepository{
						Url: "https://github.com/test/repo",
					},
				},
				Status: configv1alpha1.TerraformRepositoryStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
					},
				},
			}

			Expect(func() {
				metrics.UpdateRepositoryMetrics(repo)
			}).NotTo(Panic())
			Expect(metrics.Metrics).NotTo(BeNil())
		})
	})

	Describe("MetricsRegistration", func() {
		var m *metrics.BurritoMetrics
		var registry *prometheus.Registry

		BeforeEach(func() {
			if metrics.Metrics != nil {
				m = metrics.Metrics
			} else {
				m = metrics.InitMetrics()
			}
			registry = prometheus.NewRegistry()
		})

		It("should have all metrics initialized", func() {
			Expect(m.LayerStatusGauge).NotTo(BeNil())
			Expect(m.LayersByStatus).NotTo(BeNil())
			Expect(m.LayersByNamespace).NotTo(BeNil())
			Expect(m.RepositoryStatusGauge).NotTo(BeNil())
			Expect(m.RunsByAction).NotTo(BeNil())
			Expect(m.RunsByStatus).NotTo(BeNil())
			Expect(m.TotalLayers).NotTo(BeNil())
			Expect(m.TotalRepositories).NotTo(BeNil())
			Expect(m.TotalRuns).NotTo(BeNil())
			Expect(m.TotalPullRequests).NotTo(BeNil())
			Expect(m.ReconcileDuration).NotTo(BeNil())
			Expect(m.ReconcileTotal).NotTo(BeNil())
			Expect(m.ReconcileErrors).NotTo(BeNil())
		})

		It("should gather metrics without error", func() {
			_, err := registry.Gather()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

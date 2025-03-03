package syncwindow_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/utils/syncwindow"
)

const testTime = "Sun May  8 11:21:53 UTC 2023"

func TestSyncWindow(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SyncWindow Suite")
}

type MockClock struct{}

func (m *MockClock) Now() time.Time {
	t, _ := time.Parse(time.UnixDate, testTime)
	return t
}

var _ = Describe("SyncWindow", func() {
	Describe("When checking if sync is blocked", func() {
		Context("With no sync windows", func() {
			It("Should not block sync", func() {
				windows := []configv1alpha1.SyncWindow{}
				blocked, reason := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeFalse())
				Expect(reason).To(BeEmpty())
			})
		})

		Context("With deny windows", func() {
			It("Should block sync when in deny window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "* * * * *", // Every minute
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				blocked, reason := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeTrue())
				Expect(reason).To(Equal(syncwindow.BlockReasonInsideDenyWindow))
			})

			It("Should not block sync when layer doesn't match pattern", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"prod-*"},
					},
				}
				blocked, reason := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeFalse())
				Expect(reason).To(BeEmpty())
			})
		})

		Context("With allow windows", func() {
			It("Should allow sync during allow window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindAllow,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				blocked, reason := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeFalse())
				Expect(reason).To(BeEmpty())
			})

			It("Should block sync outside allow window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindAllow,
						Schedule: "0 0 31 2 *", // Never occurs (Feb 31)
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				blocked, reason := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeTrue())
				Expect(reason).To(Equal(syncwindow.BlockReasonOutsideAllowWindow))
			})
		})

		Context("With mixed windows", func() {
			It("Should block sync when in deny window, even if in allow window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindAllow,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				blocked, reason := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeTrue())
				Expect(reason).To(Equal(syncwindow.BlockReasonInsideDenyWindow))
			})

			It("Should allow sync when in allow window and not in deny window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindAllow,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "0 0 31 2 *", // Never occurs
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				blocked, reason := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeFalse())
				Expect(reason).To(BeEmpty())
			})
		})

		Context("With layer patterns", func() {
			It("Should match exact layer names", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"test-layer"},
					},
				}
				blocked, _ := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeTrue())
				blocked, _ = syncwindow.IsSyncBlocked(windows, "other-layer")
				Expect(blocked).To(BeFalse())
			})

			It("Should match wildcard patterns", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				blocked, _ := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeTrue())
				blocked, _ = syncwindow.IsSyncBlocked(windows, "test-other")
				Expect(blocked).To(BeTrue())
				blocked, _ = syncwindow.IsSyncBlocked(windows, "prod-layer")
				Expect(blocked).To(BeFalse())
			})

			It("Should match all layers with *", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				blocked, _ := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeTrue())
				blocked, _ = syncwindow.IsSyncBlocked(windows, "prod-layer")
				Expect(blocked).To(BeTrue())
				blocked, _ = syncwindow.IsSyncBlocked(windows, "any-layer")
				Expect(blocked).To(BeTrue())
			})
		})

		Context("With invalid configurations", func() {
			It("Should handle invalid schedule formats", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "invalid-cron",
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				blocked, _ := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeFalse())
			})

			It("Should handle invalid duration formats", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     configv1alpha1.SyncWindowKindDeny,
						Schedule: "* * * * *",
						Duration: "invalid-duration",
						Layers:   []string{"*"},
					},
				}
				blocked, _ := syncwindow.IsSyncBlocked(windows, "test-layer")
				Expect(blocked).To(BeFalse())
			})
		})
	})
})

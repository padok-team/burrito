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
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeFalse())
			})
		})

		Context("With deny windows", func() {
			It("Should block sync when in deny window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "deny",
						Schedule: "* * * * *", // Every minute
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeTrue())
			})

			It("Should not block sync when layer doesn't match pattern", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "deny",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"prod-*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeFalse())
			})
		})

		Context("With allow windows", func() {
			It("Should allow sync during allow window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "allow",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeFalse())
			})

			It("Should block sync outside allow window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "allow",
						Schedule: "0 0 31 2 *", // Never occurs (Feb 31)
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeTrue())
			})
		})

		Context("With mixed windows", func() {
			It("Should block sync when in deny window, even if in allow window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "allow",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
					{
						Kind:     "deny",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeTrue())
			})

			It("Should allow sync when in allow window and not in deny window", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "allow",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
					{
						Kind:     "deny",
						Schedule: "0 0 31 2 *", // Never occurs
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeFalse())
			})
		})

		Context("With layer patterns", func() {
			It("Should match exact layer names", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "deny",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"test-layer"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeTrue())
				Expect(syncwindow.IsSyncBlocked(windows, "other-layer")).To(BeFalse())
			})

			It("Should match wildcard patterns", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "deny",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"test-*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeTrue())
				Expect(syncwindow.IsSyncBlocked(windows, "test-other")).To(BeTrue())
				Expect(syncwindow.IsSyncBlocked(windows, "prod-layer")).To(BeFalse())
			})

			It("Should match all layers with *", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "deny",
						Schedule: "* * * * *",
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeTrue())
				Expect(syncwindow.IsSyncBlocked(windows, "prod-layer")).To(BeTrue())
				Expect(syncwindow.IsSyncBlocked(windows, "any-layer")).To(BeTrue())
			})
		})

		Context("With invalid configurations", func() {
			It("Should handle invalid schedule formats", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "deny",
						Schedule: "invalid-cron",
						Duration: "1h",
						Layers:   []string{"*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeFalse())
			})

			It("Should handle invalid duration formats", func() {
				windows := []configv1alpha1.SyncWindow{
					{
						Kind:     "deny",
						Schedule: "* * * * *",
						Duration: "invalid-duration",
						Layers:   []string{"*"},
					},
				}
				Expect(syncwindow.IsSyncBlocked(windows, "test-layer")).To(BeFalse())
			})
		})
	})
})

package storage_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage"
	"github.com/padok-team/burrito/internal/datastore/storage/mock"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var Storage *storage.Storage

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Storage Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	s := storage.Storage{
		Backend: mock.New(),
		Config:  config.TestConfig(),
	}
	Storage = &s
})

var _ = Describe("Storage", func() {
	Describe("Keys", func() {
		Describe("When skip leading slash in key is disabled", func() {
			It("should return the key with a leading slash", func() {
				key := Storage.ComputePlanKey("namespace", "layer", "run", "attempt", "format")
				Expect(key).Should(HavePrefix("/"))
			})
		})
		Describe("When skip leading slash in key is enabled", func() {
			It("should return the key without a leading slash", func() {
				testConfig := config.TestConfig()
				testConfig.Datastore.SkipLeadingSlashInKey = true
				storage := storage.Storage{
					Backend: mock.New(),
					Config:  testConfig,
				}
				key := storage.ComputePlanKey("namespace", "layer", "run", "attempt", "format")
				Expect(key).ShouldNot(HavePrefix("/"))
			})
		})
	})
})

// nolint
package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/api"
	"github.com/padok-team/burrito/internal/datastore/storage"
	"github.com/padok-team/burrito/internal/datastore/storage/mock"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var API *api.API
var e *echo.Echo

func TestDatastoreAPI(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Datastore API Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	// run tests without encryption by default
	s := storage.Storage{
		Backend: mock.New(),
		Config: config.Config{
			Datastore: config.DatastoreConfig{
				Storage: config.StorageConfig{
					Mock: true,
					Encryption: config.EncryptionConfig{
						Enabled: false,
					},
				},
			},
		},
		EncryptionManager: &storage.EncryptionManager{},
	}
	API = &api.API{}
	API.Storage = s
	API.Storage.PutLogs("default", "test1", "test1", "0", []byte("test1"))
	API.Storage.PutPlan("default", "test1", "test1", "0", "json", []byte("test1"))
	API.Storage.PutPlan("default", "test1", "test1", "0", "bin", []byte("test1"))
	API.Storage.PutPlan("default", "test1", "test1", "0", "short", []byte("test1"))
	API.Storage.PutPlan("default", "test1", "test1", "0", "pretty", []byte("test1"))
	API.Storage.PutGitBundle("default", "test1", "main", "abc123", []byte("test-bundle"))

	e = echo.New()
})

func getContext(method string, path string, params map[string]string, body []byte) echo.Context {
	buf := bytes.NewBuffer(body)
	req := httptest.NewRequest(method, path, buf)
	rec := httptest.NewRecorder()
	context := e.NewContext(req, rec)
	for k, v := range params {
		context.QueryParams().Add(k, v)
	}
	return context
}

var _ = Describe("Datastore API", func() {
	Describe("Read Operations", func() {
		Describe("Logs", func() {
			Describe("When attempt is present and log is present in storage", func() {
				It("should return the log with a 200 OK", func() {
					context := getContext(http.MethodGet, "/logs", map[string]string{
						"namespace": "default",
						"layer":     "test1",
						"run":       "test1",
						"attempt":   "0",
					}, nil)
					err := API.GetLogsHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusOK))
				})
			})
			Describe("When attempt is not present and log is present in storage", func() {
				It("should return the log with a 200 OK", func() {
					context := getContext(http.MethodGet, "/logs", map[string]string{
						"namespace": "default",
						"layer":     "test1",
						"run":       "test1",
					}, nil)
					err := API.GetLogsHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusOK))
				})
			})
			Describe("Log does not exist", func() {
				Describe("When attempt is present", func() {
					It("should return 404 Not Found", func() {
						context := getContext(http.MethodGet, "/logs", map[string]string{
							"namespace": "notfound",
							"layer":     "notfound",
							"run":       "notfound",
							"attempt":   "0",
						}, nil)
						err := API.GetLogsHandler(context)
						Expect(err).NotTo(HaveOccurred())
						Expect(context.Response().Status).To(Equal(http.StatusNotFound))
					})
				})
				Describe("When attempt is not present", func() {
					It("should return 404 Not Found", func() {
						context := getContext(http.MethodGet, "/logs", map[string]string{
							"namespace": "notfound",
							"layer":     "notfound",
							"run":       "notfound",
						}, nil)
						err := API.GetLogsHandler(context)
						Expect(err).NotTo(HaveOccurred())
						Expect(context.Response().Status).To(Equal(http.StatusNotFound))
					})
				})
			})
		})
		Describe("Plans", func() {
			Describe("Plan exists", func() {
				Describe("Format is not present", func() {
					It("should return the plan with a 200 OK if attempt is present", func() {
						context := getContext(http.MethodGet, "/plans", map[string]string{
							"namespace": "default",
							"layer":     "test1",
							"run":       "test1",
							"attempt":   "0",
						}, nil)
						err := API.GetPlanHandler(context)
						Expect(err).NotTo(HaveOccurred())
						Expect(context.Response().Status).To(Equal(http.StatusOK))
					})
					It("should return the plan with a 200 OK if attempt is not present", func() {
						context := getContext(http.MethodGet, "/plans", map[string]string{
							"namespace": "default",
							"layer":     "test1",
							"run":       "test1",
						}, nil)
						err := API.GetPlanHandler(context)
						Expect(err).NotTo(HaveOccurred())
						Expect(context.Response().Status).To(Equal(http.StatusOK))
					})
				})
				Describe("Format is present", func() {
					It("should return the plan with a 200 OK if attempt is present", func() {
						context := getContext(http.MethodGet, "/plans", map[string]string{
							"namespace": "default",
							"layer":     "test1",
							"run":       "test1",
							"attempt":   "0",
							"format":    "json",
						}, nil)
						err := API.GetPlanHandler(context)
						Expect(err).NotTo(HaveOccurred())
						Expect(context.Response().Status).To(Equal(http.StatusOK))
					})
					It("should return the plan with a 200 OK if attempt is not present", func() {
						context := getContext(http.MethodGet, "/plans", map[string]string{
							"namespace": "default",
							"layer":     "test1",
							"run":       "test1",
							"format":    "json",
						}, nil)
						err := API.GetPlanHandler(context)
						Expect(err).NotTo(HaveOccurred())
						Expect(context.Response().Status).To(Equal(http.StatusOK))
					})
				})
			})
			Describe("Plan does not exist", func() {
				It("should return 404 Not found if attempt is present", func() {
					context := getContext(http.MethodGet, "/plans", map[string]string{
						"namespace": "notfound",
						"layer":     "notfound",
						"run":       "notfound",
						"attempt":   "0",
					}, nil)
					err := API.GetPlanHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusNotFound))
				})
				It("should return 404 Not found if attempt is not present", func() {
					context := getContext(http.MethodGet, "/plans", map[string]string{
						"namespace": "notfound",
						"layer":     "notfound",
						"run":       "notfound",
					}, nil)
					err := API.GetPlanHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusNotFound))
				})
			})
		})
		Describe("Revisions", func() {
			Describe("Store Revision", func() {
				It("should return 200 OK when storing a revision", func() {
					body := []byte(`test-bundle`)
					context := getContext(http.MethodPut, "/revisions", map[string]string{
						"namespace": "default",
						"name":      "test1",
						"ref":       "main",
						"revision":  "def456",
					}, body)
					err := API.PutGitBundleHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusOK))
				})

				It("should return 400 Bad Request when missing parameters", func() {
					body := []byte(`test-bundle`)
					context := getContext(http.MethodPut, "/revisions", map[string]string{
						"namespace": "default",
						"name":      "test1",
						// missing ref and revision
					}, body)
					err := API.PutGitBundleHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusBadRequest))
				})
			})
		})
		Describe("Write", func() {
			Describe("Logs", func() {
				It("should return 200 OK", func() {
					body := []byte(`test1`)
					context := getContext(http.MethodPut, "/logs", map[string]string{
						"namespace": "default",
						"layer":     "test1",
						"run":       "test1",
						"attempt":   "0",
					}, body)
					err := API.PutLogsHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusOK))
				})
			})
			Describe("Plans", func() {
				It("should return 200 OK", func() {
					body := []byte(`test1`)
					context := getContext(http.MethodPut, "/plans", map[string]string{
						"namespace": "default",
						"layer":     "test1",
						"run":       "test1",
						"attempt":   "0",
						"format":    "json",
					}, body)
					err := API.PutPlanHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusOK))
				})
			})
		})
		Describe("Write with Encryption", func() {
			Describe("Plans", func() {
				It("should return 200 OK with encryption enabled", func() {
					// Set up encryption key environment variable
					encryptionKey := "test-encryption-key-for-api-testing-123"
					err := os.Setenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY", encryptionKey)
					Expect(err).NotTo(HaveOccurred())
					defer os.Unsetenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")

					// Create storage with encryption enabled
					config := config.Config{
						Datastore: config.DatastoreConfig{
							Storage: config.StorageConfig{
								Mock: true,
								Encryption: config.EncryptionConfig{
									Enabled: true,
								},
							},
						},
					}

					encryptedStorage := storage.New(config)
					encryptedAPI := &api.API{}
					encryptedAPI.Storage = encryptedStorage

					// Test data
					body := []byte(`{"format_version":"1.1","terraform_version":"1.0.0","planned_values":{}}`)
					context := getContext(http.MethodPut, "/plans", map[string]string{
						"namespace": "encrypted-test",
						"layer":     "test-layer",
						"run":       "test-run",
						"attempt":   "0",
						"format":    "json",
					}, body)

					// Store plan with encryption
					err = encryptedAPI.PutPlanHandler(context)
					Expect(err).NotTo(HaveOccurred())
					Expect(context.Response().Status).To(Equal(http.StatusOK))

					// Verify that data was stored encrypted by checking the raw backend
					// The encrypted data should be different from the original
					storedData, err := encryptedStorage.Backend.Get("layers/encrypted-test/test-layer/test-run/0/plan.json")
					Expect(err).NotTo(HaveOccurred())
					Expect(storedData).NotTo(Equal(body), "stored data should be encrypted and different from original")

					// Verify that the storage layer can decrypt it correctly
					retrievedData, err := encryptedStorage.GetPlan("encrypted-test", "test-layer", "test-run", "0", "json")
					Expect(err).NotTo(HaveOccurred())
					Expect(retrievedData).To(Equal(body), "decrypted data should match original")
				})
			})
		})
	})
})

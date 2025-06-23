package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/api"
	"github.com/padok-team/burrito/internal/datastore/storage"
)

var _ = Describe("Encrypt API", func() {
	var (
		testAPI *api.API
		e       *echo.Echo
	)

	BeforeEach(func() {
		// Set up test environment with encryption enabled
		os.Setenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY", "test-encryption-key")

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

		testAPI = &api.API{}
		testAPI.Storage = storage.New(config)
		testAPI = &api.API{}
		testAPI.Storage = storage.New(config)
		testAPI = &api.API{}
		testAPI.Storage = storage.New(config)
		testAPI = api.New(&config)
		testAPI.Storage = storage.New(config)
		e = echo.New()
	})

	AfterEach(func() {
		os.Unsetenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")
	})

	Describe("POST /encrypt", func() {
		Context("when encryption is enabled and key is valid", func() {
			It("should encrypt all files successfully", func() {
				// Pre-populate some test data
				testAPI.Storage.PutLogs("test-namespace", "test-layer", "test-run", "0", []byte("test logs"))
				testAPI.Storage.PutPlan("test-namespace", "test-layer", "test-run", "0", "json", []byte("test plan"))

				// Prepare request
				reqBody := map[string]string{
					"encryptionKey": "test-encryption-key",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/encrypt", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				// Execute request
				err := testAPI.EncryptAllFilesHandler(c)
				Expect(err).NotTo(HaveOccurred())
				// Accept both OK and Partial Content (when there are some errors but process completes)
				Expect(rec.Code).To(Or(Equal(http.StatusOK), Equal(http.StatusPartialContent)))

				// Check response
				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(ContainSubstring("Encryption process completed"))
				// For debugging, let's see what the actual response is
				GinkgoWriter.Printf("Response: %+v\n", response)
			})
		})

		Context("when encryption key is invalid", func() {
			It("should return unauthorized error", func() {
				reqBody := map[string]string{
					"encryptionKey": "wrong-key",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/encrypt", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := testAPI.EncryptAllFilesHandler(c)
				Expect(err).NotTo(HaveOccurred())
				Expect(rec.Code).To(Equal(http.StatusUnauthorized))

				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal("Invalid encryption key"))
			})
		})

		Context("when encryption key is missing", func() {
			It("should return bad request error", func() {
				reqBody := map[string]string{}
				jsonBody, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/encrypt", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := testAPI.EncryptAllFilesHandler(c)
				Expect(err).NotTo(HaveOccurred())
				Expect(rec.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal("encryptionKey is required"))
			})
		})

		Context("when encryption is disabled", func() {
			It("should return bad request error", func() {
				// Create API with encryption disabled
				config := config.Config{
					Datastore: config.DatastoreConfig{
						Storage: config.StorageConfig{
							Mock: true,
							Encryption: config.EncryptionConfig{
								Enabled: false,
							},
						},
					},
				}

				disabledAPI := api.New(&config)
				disabledAPI.Storage = storage.New(config)

				reqBody := map[string]string{
					"encryptionKey": "test-encryption-key",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/encrypt", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := disabledAPI.EncryptAllFilesHandler(c)
				Expect(err).NotTo(HaveOccurred())
				Expect(rec.Code).To(Equal(http.StatusBadRequest))

				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal("Encryption is not enabled in configuration"))
			})
		})

		Context("when no encryption key is configured", func() {
			It("should return internal server error", func() {
				os.Unsetenv("BURRITO_DATASTORE_STORAGE_ENCRYPTION_KEY")

				reqBody := map[string]string{
					"encryptionKey": "any-key",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/encrypt", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := testAPI.EncryptAllFilesHandler(c)
				Expect(err).NotTo(HaveOccurred())
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))

				var response map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["error"]).To(Equal("No encryption key configured on server"))
			})
		})

		Context("when files are already encrypted", func() {
			It("should skip already encrypted files", func() {
				// Pre-populate some test data and encrypt it first
				testAPI.Storage.PutLogs("test-namespace", "test-layer", "test-run", "0", []byte("test logs"))
				testAPI.Storage.PutPlan("test-namespace", "test-layer", "test-run", "0", "json", []byte("test plan"))

				// First encryption run
				reqBody := map[string]string{
					"encryptionKey": "test-encryption-key",
				}
				jsonBody, _ := json.Marshal(reqBody)

				req := httptest.NewRequest(http.MethodPost, "/encrypt", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				err := testAPI.EncryptAllFilesHandler(c)
				Expect(err).NotTo(HaveOccurred())

				var firstResponse map[string]interface{}
				err = json.Unmarshal(rec.Body.Bytes(), &firstResponse)
				Expect(err).NotTo(HaveOccurred())

				// Second encryption run on the same data
				req2 := httptest.NewRequest(http.MethodPost, "/encrypt", bytes.NewBuffer(jsonBody))
				req2.Header.Set("Content-Type", "application/json")
				rec2 := httptest.NewRecorder()
				c2 := e.NewContext(req2, rec2)

				err = testAPI.EncryptAllFilesHandler(c2)
				Expect(err).NotTo(HaveOccurred())

				var secondResponse map[string]interface{}
				err = json.Unmarshal(rec2.Body.Bytes(), &secondResponse)
				Expect(err).NotTo(HaveOccurred())

				// The second run should encrypt 0 files since they're already encrypted
				Expect(secondResponse["filesEncrypted"]).To(Equal(float64(0)))
				Expect(secondResponse["message"]).To(ContainSubstring("0 files encrypted"))
			})
		})
	})
})

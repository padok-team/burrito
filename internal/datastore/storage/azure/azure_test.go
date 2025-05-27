// nolint
// Package azure_test contains integration tests for Azure storage backend.
//
// These tests require an Azurite container running on localhost:10000.
// To skip these tests when Azurite is not available, set SKIP_AZURITE_TESTS environment variable.
//
// To run Azurite locally:
// docker run -p 10000:10000 mcr.microsoft.com/azure-storage/azurite azurite-blob --loose --blobHost 0.0.0.0
//
// These tests verify the functionality of Azure storage operations including:
// - Get: retrieving blobs with handling for missing keys
// - Set: storing blobs with data integrity verification
// - Delete: removing blobs with validation and handling missing keys
// - Check: verifying existence of blobs and retrieving metadata
// - List: retrieving all blob keys with a given prefix
package azure_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage/azure"
	storageErrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Constants for Azurite connection
const (
	azuriteConnString = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"
	containerName     = "test-container"
	testKey           = "test-get-key"
	testPrefix        = "prefix-test-"
)

var (
	azureClient  *azblob.Client
	azureBackend *azure.Azure
	testData     = []byte("Hello from Azurite test")
)

func TestAzureStorage(t *testing.T) {
	// Skip if SKIP_AZURITE_TESTS is set
	if os.Getenv("SKIP_AZURITE_TESTS") != "" {
		t.Skip("Skipping Azurite tests")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Azure Storage Suite")
}

var _ = BeforeSuite(func() {
	// Setup logging
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	// Create Azurite client
	var err error
	azureClient, err = azblob.NewClientFromConnectionString(azuriteConnString, nil)
	if err != nil {
		Skip("Failed to create Azurite client: " + err.Error())
	}

	// Verify Azurite is running
	_, err = azureClient.ServiceClient().GetProperties(context.Background(), nil)
	if err != nil {
		Skip("Azurite is not available: " + err.Error())
	}

	// Create test container
	ctx := context.Background()
	_, err = azureClient.CreateContainer(ctx, containerName, nil)
	if err != nil && !isContainerAlreadyExists(err) {
		Fail("Failed to create container: " + err.Error())
	}

	// Create Azure storage backend
	azureBackend = azure.New(config.AzureConfig{
		StorageAccount: "devstoreaccount1",
		Container:      containerName,
	}, azureClient)

	// Setup test data
	err = azureBackend.Set(testKey, testData, 0)
	Expect(err).NotTo(HaveOccurred(), "Failed to set up test data")

	// Setup additional test data for List operation
	for i := 1; i <= 3; i++ {
		prefixKey := testPrefix + string(rune('0'+i))
		err = azureBackend.Set(prefixKey, []byte("List test data "+string(rune('0'+i))), 0)
		Expect(err).NotTo(HaveOccurred(), "Failed to set up prefix test data")
	}

	// Setup layer test files in the required structure
	layerTestFiles := map[string]string{
		"/layers/ns/layer/run/0/run.log":    "Run log content for run 0",
		"/layers/ns/layer/run/1/short.diff": "Short diff content for run 1",
		"/layers/ns/layer/run/1/plan.bin":   "Plan binary content for run 1",
		"/layers/ns/layer/run/1/run.log":    "Run log content for run 1",
	}

	for filePath, content := range layerTestFiles {
		err = azureBackend.Set(filePath, []byte(content), 0)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Failed to set up layer test file: %s", filePath))
	}
})

var _ = AfterSuite(func() {
	if azureBackend != nil && azureClient != nil {
		// Clean up test data
		_ = azureBackend.Delete(testKey)
		_ = azureBackend.Delete("non-existent-key")

		// Clean up prefix test data
		for i := 1; i <= 3; i++ {
			_ = azureBackend.Delete(testPrefix + string(rune('0'+i)))
		}

		// Clean up layer test files
		layerTestFilePaths := []string{
			"/layers/ns/layer/run/0/run.log",
			"/layers/ns/layer/run/1/short.diff",
			"/layers/ns/layer/run/1/plan.bin",
			"/layers/ns/layer/run/1/run.log",
		}

		for _, filePath := range layerTestFilePaths {
			_ = azureBackend.Delete(filePath)
		}
	}
})

var _ = Describe("Azure Storage", func() {
	Describe("Get Operation", func() {
		Context("When the key exists", func() {
			It("should retrieve the correct data", func() {
				data, err := azureBackend.Get(testKey)
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(Equal(testData))
			})
		})

		Context("When the key does not exist", func() {
			BeforeEach(func() {
				// Ensure the key doesn't exist
				_ = azureBackend.Delete("non-existent-key")
			})

			It("should return a StorageError with Nil=true", func() {
				data, err := azureBackend.Get("non-existent-key")

				// Verify error is not nil
				Expect(err).To(HaveOccurred())

				// Verify it's the expected error type
				storageErr, ok := err.(*storageErrors.StorageError)
				Expect(ok).To(BeTrue(), "Error should be a StorageError")
				Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true")

				// Data should be nil
				Expect(data).To(BeNil())
			})
		})
	})

	Describe("Set Operation", func() {
		var dynamicTestKey string

		BeforeEach(func() {
			// Generate a unique key for each test
			dynamicTestKey = "dynamic-test-key-" + strings.Replace(GinkgoT().Name(), " ", "-", -1)
		})

		AfterEach(func() {
			// Clean up after each test
			_ = azureBackend.Delete(dynamicTestKey)
		})

		It("should store data that can be retrieved later", func() {
			// Set some test data
			testValue := []byte("Dynamic test data")
			err := azureBackend.Set(dynamicTestKey, testValue, 0)
			Expect(err).NotTo(HaveOccurred(), "Set operation should not fail")

			// Retrieve and verify the data
			retrievedData, err := azureBackend.Get(dynamicTestKey)
			Expect(err).NotTo(HaveOccurred(), "Get operation should not fail")
			Expect(retrievedData).To(Equal(testValue), "Retrieved data should match what was set")
		})
	})

	Describe("Delete Operation", func() {
		var deleteTestKey string

		BeforeEach(func() {
			// Create a test key and set some data
			deleteTestKey = "delete-test-key-" + strings.Replace(GinkgoT().Name(), " ", "-", -1)
			err := azureBackend.Set(deleteTestKey, []byte("Test data to delete"), 0)
			Expect(err).NotTo(HaveOccurred(), "Set operation during setup should not fail")

			// Verify the key exists
			_, err = azureBackend.Get(deleteTestKey)
			Expect(err).NotTo(HaveOccurred(), "Get operation during setup should not fail")
		})

		It("should delete existing keys", func() {
			// Delete the key
			err := azureBackend.Delete(deleteTestKey)
			Expect(err).NotTo(HaveOccurred(), "Delete operation should not fail")

			// Try to get the deleted key - should fail with StorageError.Nil = true
			_, err = azureBackend.Get(deleteTestKey)
			Expect(err).To(HaveOccurred(), "Get after delete should fail")

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true after deletion")
		})

		It("should handle deleting non-existent keys gracefully", func() {
			// Delete a key that doesn't exist
			err := azureBackend.Delete("non-existent-delete-key")

			// Should return StorageError with Nil=true but not error out
			Expect(err).To(HaveOccurred())
			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")
		})
	})

	Describe("Check Operation", func() {
		Context("When the key exists", func() {
			It("should return ContentMD5 without error", func() {
				md5, err := azureBackend.Check(testKey)
				Expect(err).NotTo(HaveOccurred(), "Check operation should not fail for existing key")
				Expect(md5).NotTo(BeNil(), "ContentMD5 should not be nil")
				Expect(len(md5)).To(BeNumerically(">", 0), "ContentMD5 should not be empty")
			})
		})

		Context("When the key does not exist", func() {
			It("should return error with StorageError.Nil=true", func() {
				md5, err := azureBackend.Check("non-existent-check-key")

				// Should have error
				Expect(err).To(HaveOccurred())

				// Should be StorageError with Nil=true
				storageErr, ok := err.(*storageErrors.StorageError)
				Expect(ok).To(BeTrue(), "Error should be a StorageError")
				Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")

				// MD5 should be empty
				Expect(md5).To(HaveLen(0), "ContentMD5 should be empty for non-existent key")
			})
		})
	})

	Describe("List Operation", func() {
		Context("When keys with the prefix exist", func() {
			It("should return all keys with the given prefix", func() {
				keys, err := azureBackend.List(testPrefix)

				// Should not error
				Expect(err).NotTo(HaveOccurred(), "List operation should not fail")

				// Should find all 3 keys with the prefix
				Expect(keys).To(HaveLen(3), "Should find all 3 test keys with the prefix")

				// All keys should start with the prefix
				for _, key := range keys {
					Expect(strings.HasPrefix(key, testPrefix)).To(BeTrue(), "All keys should start with the prefix")
				}

				// Should include all the expected keys
				expectedKeys := []string{
					testPrefix + "1",
					testPrefix + "2",
					testPrefix + "3",
				}
				for _, expectedKey := range expectedKeys {
					Expect(keys).To(ContainElement(expectedKey), "Results should contain "+expectedKey)
				}
			})
		})

		Context("When no keys with the prefix exist", func() {
			It("should return an empty list without error", func() {
				nonExistentPrefix := "non-existent-prefix-"
				keys, err := azureBackend.List(nonExistentPrefix)

				// Should not error
				Expect(err).NotTo(HaveOccurred(), "List operation should not fail")

				// Should return empty list
				Expect(keys).To(HaveLen(0), "Should return empty list for prefix with no matches")
			})
		})
	})

	Describe("Layer Files Structure", func() {
		It("should be able to access all layer files", func() {
			// Test files that should exist in the Azure storage
			expectedLayerFiles := []string{
				"/layers/ns/layer/run/0/run.log",
				"/layers/ns/layer/run/1/short.diff",
				"/layers/ns/layer/run/1/plan.bin",
				"/layers/ns/layer/run/1/run.log",
			}

			// Verify each file exists and has content
			for _, filePath := range expectedLayerFiles {
				data, err := azureBackend.Get(filePath)
				Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Should be able to get %s", filePath))
				Expect(data).NotTo(BeEmpty(), fmt.Sprintf("Content of %s should not be empty", filePath))
			}
		})

		It("should return correct content for each layer file", func() {
			// Map of files and their expected contents
			expectedContents := map[string]string{
				"/layers/ns/layer/run/0/run.log":    "Run log content for run 0",
				"/layers/ns/layer/run/1/short.diff": "Short diff content for run 1",
				"/layers/ns/layer/run/1/plan.bin":   "Plan binary content for run 1",
				"/layers/ns/layer/run/1/run.log":    "Run log content for run 1",
			}

			for filePath, expectedContent := range expectedContents {
				data, err := azureBackend.Get(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(expectedContent))
			}
		})

		It("should list all files in a layer run directory", func() {
			// List all files in the run/1/ directory
			keys, err := azureBackend.List("/layers/ns/layer/run/1/")
			Expect(err).NotTo(HaveOccurred())

			// Expect 3 files in run/1/
			Expect(keys).To(HaveLen(3))

			// Check for specific files
			expectedFiles := []string{
				"/layers/ns/layer/run/1/short.diff",
				"/layers/ns/layer/run/1/plan.bin",
				"/layers/ns/layer/run/1/run.log",
			}

			for _, expectedFile := range expectedFiles {
				Expect(keys).To(ContainElement(expectedFile))
			}
		})
	})
})

// Helper function to check if error is "container already exists"
func isContainerAlreadyExists(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "ContainerAlreadyExists") ||
		strings.Contains(errMsg, "The specified container already exists") ||
		strings.Contains(errMsg, "409")
}

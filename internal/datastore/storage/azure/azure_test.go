package azure_test

// we'll use Azurite for local testing

import (
	"context"
	"crypto/md5"
	"encoding/hex"
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

const (
	storageAccountName = "devstoreaccount1"
	containerName      = "test-container"
)

var (
	azureClient              *azblob.Client
	azureBackend             *azure.Azure
	azuriteConnString               = "DefaultEndpointsProtocol=http;AccountName=" + storageAccountName + ";AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/" + storageAccountName
	firstLayerFile           string = "/layers/ns/layer/run/0/run.log"
	firstLayerFileContent    string = "Run log content for run 0"
	firstLayerFileContentMD5 string
	expectedLayerTestFiles   = map[string]string{
		firstLayerFile:                      firstLayerFileContent,
		"/layers/ns/layer/run/1/short.diff": "Short diff content for run 1",
		"/layers/ns/layer/run/1/plan.bin":   "Plan binary content for run 1",
		"/layers/ns/layer/run/1/run.log":    "Run log content for run 1",
	}
)

func TestAzureStorage(t *testing.T) {
	// Skip if SKIP_AZURITE_TESTS is set
	if os.Getenv("SKIP_AZURITE_TESTS") != "" {
		t.Skip("Skipping Azurite tests")
	}

	RegisterFailHandler(Fail)
	RunSpecs(t, "Azure Storage Suite")
}

func isContainerPresent(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "ContainerAlreadyExists") ||
		strings.Contains(errMsg, "The specified container already exists") ||
		strings.Contains(errMsg, "409")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	var err error
	azureClient, err = azblob.NewClientFromConnectionString(azuriteConnString, nil)
	if err != nil {
		Skip("Failed to create Azurite client: " + err.Error())
	}

	_, err = azureClient.ServiceClient().GetProperties(context.Background(), nil)
	if err != nil {
		Fail("Azurite is not available: " + err.Error())
	}

	ctx := context.Background()
	_, err = azureClient.CreateContainer(ctx, containerName, nil)
	if err != nil && !isContainerPresent(err) {
		Fail("Failed to create container: " + err.Error())
	}

	azureBackend = azure.New(config.AzureConfig{
		StorageAccount: storageAccountName,
		Container:      containerName,
	}, azureClient)

	// Setup test data
	for filePath, content := range expectedLayerTestFiles {
		err = azureBackend.Set(filePath, []byte(content), 0)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Failed to set up layer test file: %s", filePath))
	}
	firstLayerFileContentMD5Bytes := md5.Sum([]byte(firstLayerFileContent))
	firstLayerFileContentMD5 = hex.EncodeToString(firstLayerFileContentMD5Bytes[:])
})

var _ = AfterSuite(func() {
	if azureBackend != nil && azureClient != nil {
		for _, filePath := range expectedLayerTestFiles {
			_ = azureBackend.Delete(filePath)
		}
	}
})

var _ = Describe("Azure Storage", func() {
	Describe("Get Operation", func() {
		It("should return correct content for each layer file", func() {
			for filePath, expectedContent := range expectedLayerTestFiles {
				data, err := azureBackend.Get(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(expectedContent))
			}
		})
		It("should return a StorageError with Nil=true", func() {
			data, err := azureBackend.Get("non-existent-key")

			Expect(err).To(HaveOccurred())

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true")

			Expect(data).To(BeNil())
		})
	})

	Describe("Set Operation", func() {
		var dynamicTestKey string

		BeforeEach(func() {
			dynamicTestKey = "dynamic-test-key-" + strings.Replace(GinkgoT().Name(), " ", "-", -1)
		})

		AfterEach(func() {
			_ = azureBackend.Delete(dynamicTestKey)
		})

		It("should store data that can be retrieved later", func() {
			testValue := []byte("Dynamic test data")
			err := azureBackend.Set(dynamicTestKey, testValue, 0)
			Expect(err).NotTo(HaveOccurred(), "Set operation should not fail")

			retrievedData, err := azureBackend.Get(dynamicTestKey)
			Expect(err).NotTo(HaveOccurred(), "Get operation should not fail")
			Expect(retrievedData).To(Equal(testValue), "Retrieved data should match what was set")
		})
	})

	Describe("List Operation", func() {
		It("should list content (non-recursive) in a layer run directory", func() {
			keys, err := azureBackend.List("/layers/ns/layer/run/")
			Expect(err).NotTo(HaveOccurred())

			// Expect 2 folders in run/
			Expect(keys).To(HaveLen(2))
			expectedFiles := []string{
				"/layers/ns/layer/run/0/",
				"/layers/ns/layer/run/1/",
			}

			for _, expectedFile := range expectedFiles {
				Expect(keys).To(ContainElement(expectedFile))
			}
		})

		It("should return error for non-existent prefix", func() {
			nonExistentPrefix := "/layers/non-existent-namespace/non-existent-layer/"
			keys, err := azureBackend.List(nonExistentPrefix)

			Expect(err).To(HaveOccurred(), "List operation should fail for non-existent prefix")

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true")

			Expect(keys).To(BeNil(), "Keys should be nil when prefix doesn't exist")
		})
	})

	Describe("Delete Operation", func() {
		var deleteTestKey string

		BeforeEach(func() {
			deleteTestKey = "delete-test-key-" + strings.Replace(GinkgoT().Name(), " ", "-", -1)
			err := azureBackend.Set(deleteTestKey, []byte("Test data to delete"), 0)
			Expect(err).NotTo(HaveOccurred(), "Set operation during setup should not fail")

			_, err = azureBackend.Get(deleteTestKey)
			Expect(err).NotTo(HaveOccurred(), "Get operation during setup should not fail")
		})

		It("should delete existing keys", func() {
			err := azureBackend.Delete(deleteTestKey)
			Expect(err).NotTo(HaveOccurred(), "Delete operation should not fail")

			_, err = azureBackend.Get(deleteTestKey)
			Expect(err).To(HaveOccurred(), "Get after delete should fail")

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true after deletion")
		})

		It("should handle deleting non-existent keys gracefully", func() {
			err := azureBackend.Delete("non-existent-delete-key")

			Expect(err).To(HaveOccurred())
			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")
		})
	})

	Describe("Check Operation", func() {
		Context("When the key exists", func() {
			It("should return ContentMD5 without error", func() {
				md5, err := azureBackend.Check(firstLayerFile)
				Expect(err).NotTo(HaveOccurred(), "Check operation should not fail for existing key")
				Expect(md5).NotTo(BeNil(), "ContentMD5 should not be nil")
				Expect(len(md5)).To(BeNumerically(">", 0), "ContentMD5 should not be empty")
				md5string := fmt.Sprintf("%x", md5)
				Expect(md5string).To(Equal("08e8de8c53789b20e50b15da9fb290ad"), "ContentMD5 should match expected value for test file")
			})
		})

		Context("When the key does not exist", func() {
			It("should return error with StorageError.Nil=true", func() {
				md5, err := azureBackend.Check("non-existent-check-key")

				Expect(err).To(HaveOccurred())

				storageErr, ok := err.(*storageErrors.StorageError)
				Expect(ok).To(BeTrue(), "Error should be a StorageError")
				Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")

				Expect(md5).To(HaveLen(0), "ContentMD5 should be empty for non-existent key")
			})
		})
	})
})

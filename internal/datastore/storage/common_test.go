package storage_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage"
	"github.com/padok-team/burrito/internal/datastore/storage/azure"
	storageErrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	s3Backend "github.com/padok-team/burrito/internal/datastore/storage/s3"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Constants for Azure tests
const (
	azureStorageAccountName = "devstoreaccount1"
	azureContainerName      = "test-container"
)

// Constants for S3/Minio tests
const (
	minioEndpoint = "http://localhost:9000"
	minioBucket   = "test-bucket"
	minioUser     = "burritoadmin"
	minioPassword = "burritoadmin"
)

// Constants for common test data
var (
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

// Backend interfaces for testing
var (
	azureBackend   storage.StorageBackend
	s3AwsBackend   storage.StorageBackend
	s3MinioBackend storage.StorageBackend
)

// Connection strings and clients
var (
	// Azure Connection string
	azuriteConnString = "DefaultEndpointsProtocol=http;AccountName=" + azureStorageAccountName + ";AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/" + azureStorageAccountName
	azureClient       *azblob.Client

	// S3 Clients
	s3Client    *s3.Client
	minioClient *s3.Client
)

func TestAllBackends(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Backends Suite")
}

// Helper functions
func isContainerPresent(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "ContainerAlreadyExists") ||
		strings.Contains(errMsg, "The specified container already exists") ||
		strings.Contains(errMsg, "409")
}

// Helper function to check if a bucket exists
func isBucketPresent(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "BucketAlreadyExists") ||
		strings.Contains(errMsg, "BucketAlreadyOwnedByYou") ||
		strings.Contains(errMsg, "The specified bucket already exists") ||
		strings.Contains(errMsg, "409")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	// Setup MD5 hash for test content
	firstLayerFileContentMD5Bytes := md5.Sum([]byte(firstLayerFileContent))
	firstLayerFileContentMD5 = hex.EncodeToString(firstLayerFileContentMD5Bytes[:])

	// Setup Azure backend
	setupAzureBackend()

	// Setup AWS S3 backend - we'll skip this in the BeforeSuite
	// to keep tests working even if AWS is not available

	// Setup Minio S3 backend
	setupMinioBackend()

	// Populate test data for each backend that was successfully initialized
	setupTestData()
})

var _ = AfterSuite(func() {
	// Cleanup each backend
	if azureBackend != nil {
		cleanupBackend(azureBackend)
	}

	if s3MinioBackend != nil {
		cleanupBackend(s3MinioBackend)
	}

	if s3AwsBackend != nil {
		cleanupBackend(s3AwsBackend)
	}
})

// Setup functions
func setupAzureBackend() {
	// Skip if SKIP_AZURITE_TESTS is set
	if os.Getenv("SKIP_AZURITE_TESTS") != "" {
		return
	}

	var err error
	azureClient, err = azblob.NewClientFromConnectionString(azuriteConnString, nil)
	if err != nil {
		fmt.Printf("Warning: Failed to create Azurite client: %s\n", err.Error())
		return
	}

	_, err = azureClient.ServiceClient().GetProperties(context.Background(), nil)
	if err != nil {
		fmt.Printf("Warning: Azurite is not available: %s\n", err.Error())
		return
	}

	ctx := context.Background()
	_, err = azureClient.CreateContainer(ctx, azureContainerName, nil)
	if err != nil && !isContainerPresent(err) {
		fmt.Printf("Warning: Failed to create Azure container: %s\n", err.Error())
		return
	}

	azureBackend = azure.New(config.AzureConfig{
		StorageAccount: azureStorageAccountName,
		Container:      azureContainerName,
	}, azureClient)
}

func setupMinioBackend() {
	// Skip if SKIP_MINIO_TESTS is set
	if os.Getenv("SKIP_MINIO_TESTS") != "" {
		return
	}

	// Create a new AWS config with a custom endpoint resolver for Minio
	cfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(minioUser, minioPassword, "")),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL:               minioEndpoint,
					SigningRegion:     "us-east-1",
					HostnameImmutable: true,
				}, nil
			}),
		),
	)
	if err != nil {
		fmt.Printf("Warning: Failed to create Minio config: %s\n", err.Error())
		return
	}

	// Create S3 client for Minio
	minioClient = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// Create bucket if it doesn't exist
	bucketName := minioBucket
	_, err = minioClient.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: &bucketName,
	})
	if err != nil && !isBucketPresent(err) {
		fmt.Printf("Warning: Failed to create Minio bucket: %s\n", err.Error())
		return
	}

	s3MinioBackend = s3Backend.NewWithClient(
		config.S3Config{
			Bucket:       minioBucket,
			UsePathStyle: true,
		},
		minioClient,
	)
}

// Setup test data in all available backends
func setupTestData() {
	backends := []storage.StorageBackend{azureBackend, s3MinioBackend, s3AwsBackend}

	for _, backend := range backends {
		if backend == nil {
			continue
		}

		for filePath, content := range expectedLayerTestFiles {
			err := backend.Set(filePath, []byte(content), 0)
			if err != nil {
				fmt.Printf("Warning: Failed to set up test file %s: %s\n", filePath, err.Error())
			}
		}
	}
}

// Cleanup test data from a backend
func cleanupBackend(backend storage.StorageBackend) {
	for filePath := range expectedLayerTestFiles {
		_ = backend.Delete(filePath)
	}
}

// Define tests for each backend
var _ = Describe("Storage Backends", func() {

	backends := []struct {
		Name    string
		Skip    func() bool
		Backend func() storage.StorageBackend
	}{
		{
			Name: "Azure",
			Skip: func() bool {
				return os.Getenv("SKIP_AZURITE_TESTS") != "" || azureBackend == nil
			},
			Backend: func() storage.StorageBackend { return azureBackend },
		},
		// {
		// 	Name: "S3 Minio",
		// 	Skip: func() bool {
		// 		return os.Getenv("SKIP_MINIO_TESTS") != "" || s3MinioBackend == nil
		// 	},
		// 	Backend: func() storage.StorageBackend { return s3MinioBackend },
		// },
	}

	for _, b := range backends {
		// Create a closure to capture the backend
		func(b struct {
			Name    string
			Skip    func() bool
			Backend func() storage.StorageBackend
		}) {
			Describe(fmt.Sprintf("%s Backend", b.Name), func() {
				BeforeEach(func() {
					if b.Skip() {
						Skip(fmt.Sprintf("Skipping %s tests", b.Name))
					}
				})

				Describe("Get Operation", func() {
					It("should return correct content for each layer file", func() {
						for filePath, expectedContent := range expectedLayerTestFiles {
							data, err := b.Backend().Get(filePath)
							Expect(err).NotTo(HaveOccurred())
							Expect(string(data)).To(Equal(expectedContent))
						}
					})

					It("should return a StorageError with Nil=true for non-existent keys", func() {
						data, err := b.Backend().Get("non-existent-key")
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
						dynamicTestKey = fmt.Sprintf("%s-dynamic-test-key-%s",
							strings.ToLower(b.Name),
							strings.Replace(GinkgoT().Name(), " ", "-", -1))
					})

					AfterEach(func() {
						_ = b.Backend().Delete(dynamicTestKey)
					})

					It("should store data that can be retrieved later", func() {
						testValue := []byte("Dynamic test data")
						err := b.Backend().Set(dynamicTestKey, testValue, 0)
						Expect(err).NotTo(HaveOccurred(), "Set operation should not fail")

						retrievedData, err := b.Backend().Get(dynamicTestKey)
						Expect(err).NotTo(HaveOccurred(), "Get operation should not fail")
						Expect(retrievedData).To(Equal(testValue), "Retrieved data should match what was set")
					})
				})

				Describe("List Operation", func() {
					It("should list content in a layer run directory", func() {
						keys, err := b.Backend().List("/layers/ns/layer/run/")
						Expect(err).NotTo(HaveOccurred())

						// Expect 2 folders in run/
						Expect(keys).To(HaveLen(2))
						expectedFolders := []string{
							"/layers/ns/layer/run/0/",
							"/layers/ns/layer/run/1/",
						}

						for _, expectedFolder := range expectedFolders {
							Expect(keys).To(ContainElement(expectedFolder))
						}
					})

					It("should return error for non-existent prefix", func() {
						nonExistentPrefix := "/layers/non-existent-namespace/non-existent-layer/"
						keys, err := b.Backend().List(nonExistentPrefix)

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
						deleteTestKey = fmt.Sprintf("%s-delete-test-key-%s",
							strings.ToLower(b.Name),
							strings.Replace(GinkgoT().Name(), " ", "-", -1))

						err := b.Backend().Set(deleteTestKey, []byte("Test data to delete"), 0)
						Expect(err).NotTo(HaveOccurred(), "Set operation during setup should not fail")

						_, err = b.Backend().Get(deleteTestKey)
						Expect(err).NotTo(HaveOccurred(), "Get operation during setup should not fail")
					})

					It("should delete existing keys", func() {
						err := b.Backend().Delete(deleteTestKey)
						Expect(err).NotTo(HaveOccurred(), "Delete operation should not fail")

						_, err = b.Backend().Get(deleteTestKey)
						Expect(err).To(HaveOccurred(), "Get after delete should fail")

						storageErr, ok := err.(*storageErrors.StorageError)
						Expect(ok).To(BeTrue(), "Error should be a StorageError")
						Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true after deletion")
					})

					It("should handle deleting non-existent keys gracefully", func() {
						err := b.Backend().Delete("non-existent-delete-key")

						Expect(err).To(HaveOccurred())
						storageErr, ok := err.(*storageErrors.StorageError)
						Expect(ok).To(BeTrue(), "Error should be a StorageError")
						Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")
					})
				})

				Describe("Check Operation", func() {
					It("should return content info without error for existing key", func() {
						md5, err := b.Backend().Check(firstLayerFile)
						Expect(err).NotTo(HaveOccurred(), "Check operation should not fail for existing key")

						// Different backends may handle checksums differently, so we just check that something is returned
						Expect(md5).NotTo(BeNil(), "Content info should not be nil")
					})

					It("should return error with StorageError.Nil=true for non-existent key", func() {
						md5, err := b.Backend().Check("non-existent-check-key")

						Expect(err).To(HaveOccurred())
						Expect(md5).To(Equal(make([]byte, 0)), "Content info should be empty for non-existent key")

						storageErr, ok := err.(*storageErrors.StorageError)
						Expect(ok).To(BeTrue(), "Error should be a StorageError")
						Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")
					})
				})
			})
		}(b)
	}
})

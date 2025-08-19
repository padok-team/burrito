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
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/datastore/storage"
	"github.com/padok-team/burrito/internal/datastore/storage/azure"
	storageErrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/datastore/storage/gcs"
	"github.com/padok-team/burrito/internal/datastore/storage/mock"
	"github.com/padok-team/burrito/internal/datastore/storage/s3"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	storageAccountName = "devstoreaccount1"
	containerName      = "test-container"
	testBucketName     = "test-bucket"
)

var (
	// Azure settings
	azureClient       = &azblob.Client{}
	azureBackend      *azure.Azure
	azuriteConnString = "DefaultEndpointsProtocol=http;AccountName=" + storageAccountName + ";AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/" + storageAccountName

	// S3 settings
	s3Backend *s3.S3

	// GCS settings
	gcsBackend *gcs.GCS

	// Mock backend
	mockBackend *mock.Mock

	// Common test data
	firstLayerFile           string = "/layers/ns/layer/run/0/run.log"
	firstLayerFileContent    string = "Run log content for run 0"
	firstLayerFileContentMD5 string
	expectedLayerTestFiles   = map[string]string{
		firstLayerFile:                      firstLayerFileContent,
		"/layers/ns/layer/run/1/short.diff": "Short diff content for run 1",
		"/layers/ns/layer/run/1/plan.bin":   "Plan binary content for run 1",
		"/layers/ns/layer/run/1/run.log":    "Run log content for run 1",
	}

	// Test backends
	backends map[string]storage.StorageBackend
)

func TestStorageBackends(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Backends Suite")
}

func isAzureContainerPresent(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "ContainerAlreadyExists") ||
		strings.Contains(errMsg, "The specified container already exists") ||
		strings.Contains(errMsg, "409")
}

func isS3BucketPresent(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "BucketAlreadyExists") ||
		strings.Contains(errMsg, "BucketAlreadyOwnedByYou") ||
		strings.Contains(errMsg, "409")
}

func isGCSBucketPresent(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "Try another name. Bucket names must be globally unique") ||
		strings.Contains(errMsg, "Conflict") ||
		strings.Contains(errMsg, "409")
}

func setupS3Bucket(backendType string) {
	os.Setenv("AWS_ACCESS_KEY_ID", "burritoadmin")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "burritoadmin")
	os.Setenv("AWS_REGION", "eu-west-3")

	// Use the "new" function from the s3backend
	s3Config := config.S3Config{
		Bucket:       "test-bucket",
		UsePathStyle: true,
	}
	s3Backend = s3.New(s3Config)

	// Create bucket if it doesn't exist
	ctx := context.Background()
	region := os.Getenv("AWS_REGION")
	createBucketInput := &awss3.CreateBucketInput{
		Bucket: aws.String(s3Config.Bucket),
		// Set public-read-write ACL for the bucket
		ACL: types.BucketCannedACLPublicReadWrite,
		// Set the region for the bucket
		CreateBucketConfiguration: &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		},
	}
	_, err := s3Backend.Client.CreateBucket(ctx, createBucketInput)

	// Add full access ACL for burritoadmin
	bucketCreatedOrExists := err == nil || isS3BucketPresent(err)
	if bucketCreatedOrExists {
		// Set bucket policy for full access
		bucketPolicy := `{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {"AWS": "*"},
						"Action": "s3:*",
						"Resource": [
							"arn:aws:s3:::test-bucket",
							"arn:aws:s3:::test-bucket/*"
						]
					}
				]
			}`

		putPolicyInput := &awss3.PutBucketPolicyInput{
			Bucket: aws.String(s3Config.Bucket),
			Policy: aws.String(bucketPolicy),
		}

		_, policyErr := s3Backend.Client.PutBucketPolicy(ctx, putPolicyInput)
		if policyErr != nil {
			fmt.Printf("%s - Warning: Failed to set bucket policy: %v\n", backendType, policyErr)
		}

		backends[backendType] = s3Backend
	} else {
		fmt.Printf("%s - Failed to create S3 bucket: %v\n", backendType, err)
	}
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	backends = make(map[string]storage.StorageBackend)

	mockBackend = mock.New()
	backends["mock"] = mockBackend

	if os.Getenv("SKIP_AZURITE_TESTS") == "" {
		var err error
		azureClient, err = azblob.NewClientFromConnectionString(azuriteConnString, nil)
		if err == nil {
			_, err = azureClient.ServiceClient().GetProperties(context.Background(), nil)
			if err == nil {
				ctx := context.Background()
				_, err = azureClient.CreateContainer(ctx, containerName, nil)
				if err == nil || isAzureContainerPresent(err) {
					azureBackend = azure.New(config.AzureConfig{
						StorageAccount: storageAccountName,
						Container:      containerName,
					}, azureClient)
					backends["azure"] = azureBackend
				}
			}
		}
	}

	if os.Getenv("SKIP_MINIO_TESTS") == "" {
		os.Setenv("AWS_ENDPOINT_URL_S3", "http://localhost:9000")
		setupS3Bucket("minio")
	}

	if os.Getenv("SKIP_AWS_TESTS") == "" {
		os.Setenv("AWS_ENDPOINT_URL_S3", "http://localhost:4566")
		setupS3Bucket("aws")
	}

	if os.Getenv("SKIP_GCS_TESTS") == "" {

		os.Setenv("STORAGE_EMULATOR_HOST", "http://localhost:8000")
		gcsBackend = gcs.New(config.GCSConfig{
			Bucket: testBucketName,
		})

		ctx := context.Background()
		err := gcsBackend.Client.Bucket(testBucketName).Create(ctx, "projectID", nil)
		bucketCreatedOrExists := false

		if err != nil {
			errMsg := err.Error()
			if isGCSBucketPresent(err) || strings.Contains(errMsg, "409") {
				bucketCreatedOrExists = true
			} else {
				fmt.Printf("Failed to create GCS bucket %s: %v\n", testBucketName, err)
			}
		} else {
			bucketCreatedOrExists = true
		}

		if bucketCreatedOrExists {
			// Verify the bucket exists
			bkt := gcsBackend.Client.Bucket(testBucketName)
			_, err := bkt.Attrs(ctx)
			if err != nil {
				fmt.Printf("Error verifying bucket exists: %v\n", err)
			} else {
				backends["gcs"] = gcsBackend
			}
		}
	}

	// Setup test data for each backend
	for name, backend := range backends {
		for filePath, content := range expectedLayerTestFiles {
			err := backend.Set(filePath, []byte(content), 0)
			fmt.Printf("Setting up layer test file for %s backend: %s\n", name, filePath)
			Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Failed to set up layer test file for %s backend: %s", name, filePath))
		}
	}

	// Calculate MD5 of test content
	firstLayerFileContentMD5Bytes := md5.Sum([]byte(firstLayerFileContent))
	firstLayerFileContentMD5 = hex.EncodeToString(firstLayerFileContentMD5Bytes[:])
})

var _ = AfterSuite(func() {
	// // Clean up test data
	for _, backend := range backends {
		for filePath := range expectedLayerTestFiles {
			_ = backend.Delete(filePath)
		}
	}
})

var _ = Describe("Storage Backends", func() {
	// Using DescribeTable to test multiple backends
	DescribeTable("List Operation",
		func(backendName string) {
			backend, ok := backends[backendName]
			if !ok {
				Skip(fmt.Sprintf("Backend %s is not available", backendName))
			}

			By("should list content (non-recursive) in a layer run directory")
			keys, err := backend.List("/layers/ns/layer/run")
			Expect(err).NotTo(HaveOccurred())

			// Expect 2 folders in run/
			Expect(keys).To(HaveLen(2))
			expectedFolders := []string{
				"/layers/ns/layer/run/0",
				"/layers/ns/layer/run/1",
			}

			for _, expectedFolder := range expectedFolders {
				Expect(keys).To(ContainElement(expectedFolder))
			}

			By("should list content (non-recursive) in a layer run/ directory")
			keys, err = backend.List("/layers/ns/layer/run/")
			Expect(err).NotTo(HaveOccurred())

			// Expect 2 folders in run/
			Expect(keys).To(HaveLen(2))
			expectedFolders = []string{
				"/layers/ns/layer/run/0",
				"/layers/ns/layer/run/1",
			}

			for _, expectedFolder := range expectedFolders {
				Expect(keys).To(ContainElement(expectedFolder))
			}

			By("should list content (non-recursive) in a layer attempt directory")
			keys, err = backend.List("/layers/ns/layer/run/1/")
			fmt.Printf("Keys in run/1/: %v\n", keys)
			Expect(err).NotTo(HaveOccurred())

			// Expect 3 items in run/1/
			Expect(keys).To(HaveLen(3))
			expectedFolders = []string{
				"/layers/ns/layer/run/1/plan.bin",
				"/layers/ns/layer/run/1/run.log",
				"/layers/ns/layer/run/1/short.diff",
			}

			for _, expectedFolder := range expectedFolders {
				Expect(keys).To(ContainElement(expectedFolder))
			}

			By("should return error for non-existent prefix")
			nonExistentPrefix := "/layers/non-existent-namespace/non-existent-layer/"
			keys, err = backend.List(nonExistentPrefix)

			Expect(err).To(HaveOccurred(), "List operation should fail for non-existent prefix")

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true")

			Expect(keys).To(BeNil(), "Keys should be nil when prefix doesn't exist")
		},
		Entry("Mock backend", "mock"),
		Entry("Azure backend", "azure"),
		Entry("S3 backend - AWS", "aws"),
		Entry("S3 backend - Minio", "minio"),
		Entry("GCS backend", "gcs"),
	)

	DescribeTable("Get Operation",
		func(backendName string) {
			backend, ok := backends[backendName]
			if !ok {
				Skip(fmt.Sprintf("Backend %s is not available", backendName))
			}

			By("should return correct content for each layer file")
			for filePath, expectedContent := range expectedLayerTestFiles {
				data, err := backend.Get(filePath)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(expectedContent))
			}

			By("should return a StorageError with Nil=true on inexistent keys")
			data, err := backend.Get("non-existent-key")

			Expect(err).To(HaveOccurred())

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true")

			Expect(data).To(BeNil())
		},
		Entry("Mock backend", "mock"),
		Entry("Azure backend", "azure"),
		Entry("S3 backend - AWS", "aws"),
		Entry("S3 backend - Minio", "minio"),
		Entry("GCS backend", "gcs"),
	)

	DescribeTable("Set Operation",
		func(backendName string) {
			backend, ok := backends[backendName]
			if !ok {
				Skip(fmt.Sprintf("Backend %s is not available", backendName))
			}

			dynamicTestKey := fmt.Sprintf("dynamic-test-key-%s-%s", backendName, strings.Replace(GinkgoT().Name(), " ", "-", -1))
			defer backend.Delete(dynamicTestKey) //nolint:errcheck

			testValue := []byte("Dynamic test data")

			By("should store data")
			err := backend.Set(dynamicTestKey, testValue, 0)
			Expect(err).NotTo(HaveOccurred(), "Set operation should not fail")

			By("should get previously set data")
			retrievedData, err := backend.Get(dynamicTestKey)
			Expect(err).NotTo(HaveOccurred(), "Get operation should not fail")
			Expect(retrievedData).To(Equal(testValue), "Retrieved data should match what was set")
		},
		Entry("Mock backend", "mock"),
		Entry("Azure backend", "azure"),
		Entry("S3 backend - AWS", "aws"),
		Entry("S3 backend - Minio", "minio"),
		Entry("GCS backend", "gcs"),
	)

	DescribeTable("Delete Operation",
		func(backendName string) {
			backend, ok := backends[backendName]
			if !ok {
				Skip(fmt.Sprintf("Backend %s is not available", backendName))
			}

			deleteTestKey := fmt.Sprintf("delete-test-key-%s-%s", backendName, strings.Replace(GinkgoT().Name(), " ", "-", -1))

			err := backend.Set(deleteTestKey, []byte("Test data to delete"), 0)
			Expect(err).NotTo(HaveOccurred(), "Set operation during setup should not fail")

			_, err = backend.Get(deleteTestKey)
			Expect(err).NotTo(HaveOccurred(), "Get operation during setup should not fail")

			By("should delete existing keys")
			err = backend.Delete(deleteTestKey)
			Expect(err).NotTo(HaveOccurred(), "Delete operation should not fail")

			_, err = backend.Get(deleteTestKey)
			Expect(err).To(HaveOccurred(), "Get after delete should fail")

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true after deletion")

			By("should handle deleting non-existent keys")
			err = backend.Delete("non-existent-delete-key")

			Expect(err).To(HaveOccurred())
			storageErr, ok = err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")
		},
		Entry("Mock backend", "mock"),
		Entry("Azure backend", "azure"),
		Entry("S3 backend - AWS", "aws"),
		Entry("S3 backend - Minio", "minio"),
		Entry("GCS backend", "gcs"),
	)

	DescribeTable("Check Operation",
		func(backendName string) {
			backend, ok := backends[backendName]
			if !ok {
				Skip(fmt.Sprintf("Backend %s is not available", backendName))
			}

			By("Testing check operation for existing keys")
			md5, err := backend.Check(firstLayerFile)
			Expect(err).NotTo(HaveOccurred(), "Check operation should not fail for existing key")

			if backendName != "mock" { // Mock might return different MD5
				Expect(md5).NotTo(BeNil(), "ContentMD5 should not be nil")
				Expect(len(md5)).To(BeNumerically(">", 0), "ContentMD5 should not be empty")
			}

			By("Testing check operation for non-existent keys")
			md5, err = backend.Check("non-existent-check-key")

			Expect(err).To(HaveOccurred())

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true for non-existent key")

			Expect(md5).To(HaveLen(0), "ContentMD5 should be empty for non-existent key")
		},
		Entry("Mock backend", "mock"),
		Entry("Azure backend", "azure"),
		Entry("S3 backend - AWS", "aws"),
		Entry("S3 backend - Minio", "minio"),
		Entry("GCS backend", "gcs"),
	)

	DescribeTable("ListRecursive Operation",
		func(backendName string) {
			backend, ok := backends[backendName]
			if !ok {
				Skip(fmt.Sprintf("Backend %s is not available", backendName))
			}

			By("should recursively list all files under a prefix")
			keys, err := backend.ListRecursive("/layers/ns/layer/run")
			Expect(err).NotTo(HaveOccurred())

			// Expect all 4 files to be listed recursively
			Expect(keys).To(HaveLen(4))
			expectedFiles := []string{
				"/layers/ns/layer/run/0/run.log",
				"/layers/ns/layer/run/1/plan.bin",
				"/layers/ns/layer/run/1/run.log",
				"/layers/ns/layer/run/1/short.diff",
			}

			for _, expectedFile := range expectedFiles {
				Expect(keys).To(ContainElement(expectedFile))
			}

			By("should recursively list all files under a prefix with trailing slash")
			keys, err = backend.ListRecursive("/layers/ns/layer/run/")
			Expect(err).NotTo(HaveOccurred())

			// Expect all 4 files to be listed recursively
			Expect(keys).To(HaveLen(4))
			for _, expectedFile := range expectedFiles {
				Expect(keys).To(ContainElement(expectedFile))
			}

			By("should recursively list files in a specific run directory")
			keys, err = backend.ListRecursive("/layers/ns/layer/run/1/")
			Expect(err).NotTo(HaveOccurred())

			// Expect 3 files in run/1/
			Expect(keys).To(HaveLen(3))
			expectedFiles = []string{
				"/layers/ns/layer/run/1/plan.bin",
				"/layers/ns/layer/run/1/run.log",
				"/layers/ns/layer/run/1/short.diff",
			}

			for _, expectedFile := range expectedFiles {
				Expect(keys).To(ContainElement(expectedFile))
			}

			By("should return error for non-existent prefix")
			nonExistentPrefix := "/layers/non-existent-namespace/non-existent-layer/"
			keys, err = backend.ListRecursive(nonExistentPrefix)

			Expect(err).To(HaveOccurred(), "ListRecursive operation should fail for non-existent prefix")

			storageErr, ok := err.(*storageErrors.StorageError)
			Expect(ok).To(BeTrue(), "Error should be a StorageError")
			Expect(storageErr.Nil).To(BeTrue(), "StorageError.Nil should be true")

			Expect(keys).To(BeNil(), "Keys should be nil when prefix doesn't exist")
		},
		Entry("Mock backend", "mock"),
		Entry("Azure backend", "azure"),
		Entry("S3 backend - AWS", "aws"),
		Entry("S3 backend - Minio", "minio"),
		Entry("GCS backend", "gcs"),
	)
})

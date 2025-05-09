package client

import (
	"fmt"
	"os"

	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

const (
	TestRepoNamespace = "default"
	TestRepoName      = "burrito"
	TestRef           = "main"
	TestRevision      = "TEST_REVISION"
)

func isBundleTestValues(namespace, name, ref, revision string) bool {
	return namespace == TestRepoNamespace && name == TestRepoName && ref == TestRef && revision == TestRevision
}

type MockClient struct {
	// Store latest revisions in memory for testing
	revisions map[string]string
	// Store bundles in memory for testing
	bundles map[string][]byte
}

func NewMockClient() *MockClient {
	return &MockClient{
		revisions: make(map[string]string),
		bundles:   make(map[string][]byte),
	}
}

func (c *MockClient) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	return nil, nil
}

func (c *MockClient) PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error {
	return nil
}

func (c *MockClient) GetLogs(namespace string, layer string, run string, attempt string) ([]string, error) {
	return nil, nil
}

func (c *MockClient) PutLogs(namespace string, layer string, run string, attempt string, content []byte) error {
	return nil
}

func (c *MockClient) GetAttempts(namespace string, layer string, run string) (int, error) {
	return 0, nil
}

func (c *MockClient) PutGitBundle(namespace, name, ref, revision string, bundle []byte) error {
	// Not used in tests yet
	if isBundleTestValues(namespace, name, ref, revision) {
		return nil
	}

	revKey := fmt.Sprintf("%s/%s/%s", namespace, name, ref)
	c.revisions[revKey] = revision

	bundleKey := fmt.Sprintf("%s/%s/%s/%s", namespace, name, ref, revision)
	c.bundles[bundleKey] = bundle

	return nil
}

func (c *MockClient) CheckGitBundle(namespace, name, ref, revision string) (bool, error) {
	// Used by TerraformRun Controller tests
	if isBundleTestValues(namespace, name, ref, revision) {
		return true, nil
	}

	revKey := fmt.Sprintf("%s/%s/%s", namespace, name, ref)
	if rev, ok := c.revisions[revKey]; ok {
		if rev == revision {
			return true, nil
		}
	}
	return false, nil
}

func (c *MockClient) GetGitBundle(namespace, name, ref, revision string) ([]byte, error) {
	// Used by Runner tests
	if isBundleTestValues(namespace, name, ref, revision) {
		bundle, err := os.ReadFile("testdata/burrito-examples.bundle")
		if err != nil {
			return nil, err
		}
		return bundle, nil
	}

	bundleKey := fmt.Sprintf("%s/%s/%s/%s", namespace, name, ref, revision)
	if bundle, ok := c.bundles[bundleKey]; ok {
		return bundle, nil
	}

	return nil, &storageerrors.StorageError{
		Err: fmt.Errorf("bundle not found"),
		Nil: true,
	}
}

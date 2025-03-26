package client

import (
	"fmt"
)

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
	revKey := fmt.Sprintf("%s/%s/%s", namespace, name, ref)
	c.revisions[revKey] = revision

	bundleKey := fmt.Sprintf("%s/%s/%s/%s", namespace, name, ref, revision)
	c.bundles[bundleKey] = bundle

	return nil
}

func (c *MockClient) CheckGitBundle(namespace, name, ref, revision string) (bool, error) {
	return false, nil
}

func (c *MockClient) GetGitBundle(namespace, name, ref, revision string) ([]byte, error) {
	return nil, nil
}

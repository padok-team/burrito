package client

type MockClient struct {
}

func NewMockClient() *MockClient {
	return &MockClient{}
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

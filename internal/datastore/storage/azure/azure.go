package azure

import "github.com/padok-team/burrito/internal/burrito/config"

// Implements Storage interface using Azure Blob Storage

type Azure struct {
	// Azure Blob Storage client
	Client interface{}
}

// New creates a new Azure Blob Storage client
func New(config config.AzureConfig) *Azure {
	return &Azure{}
}

func (a *Azure) Get(string) ([]byte, error) {
	return nil, nil
}

func (a *Azure) Set(string, []byte, int) error {
	return nil
}

func (a *Azure) Delete(string) error {
	return nil
}

func (a *Azure) List(string) ([]string, error) {
	return nil, nil
}

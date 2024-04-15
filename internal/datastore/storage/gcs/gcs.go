package gcs

import "github.com/padok-team/burrito/internal/burrito/config"

// Implements Storage interface using Google Cloud Storage

type GCS struct {
	// GCS Blob Storage client
	Client interface{}
}

// New creates a new Google Cloud Storage client
func New(config config.GCSConfig) *GCS {
	return &GCS{}
}

func (a *GCS) Get(string) ([]byte, error) {
	return nil, nil
}

func (a *GCS) Set(string, []byte, int) error {
	return nil
}

func (a *GCS) Delete(string) error {
	return nil
}

func (a *GCS) List(string) ([]string, error) {
	return nil, nil
}

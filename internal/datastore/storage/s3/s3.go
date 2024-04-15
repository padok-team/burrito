package s3

import "github.com/padok-team/burrito/internal/burrito/config"

// Implements Storage interface using Google Cloud Storage

type S3 struct {
	// GCS Blob Storage client
	Client interface{}
}

// New creates a new Google Cloud Storage client
func New(config config.S3Config) *S3 {
	return &S3{}
}

func (a *S3) Get(string) ([]byte, error) {
	return nil, nil
}

func (a *S3) Set(string, []byte, int) error {
	return nil
}

func (a *S3) Delete(string) error {
	return nil
}

func (a *S3) List(string) ([]string, error) {
	return nil, nil
}

package s3

import (
	"bytes"
	"io"

	"github.com/aws/aws-sdk-go/aws/session"
	storage "github.com/aws/aws-sdk-go/service/s3"
	"github.com/padok-team/burrito/internal/burrito/config"
)

// Implements Storage interface using Google Cloud Storage

type S3 struct {
	// GCS Blob Storage client
	Client *storage.S3
	Config config.S3Config
}

// New creates a new Google Cloud Storage client
func New(config config.S3Config) *S3 {
	session := session.Must(session.NewSession())
	client := storage.New(session)
	return &S3{
		Config: config,
		Client: client,
	}
}

func (a *S3) Get(key string) ([]byte, error) {
	input := &storage.GetObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &key,
	}

	result, err := a.Client.GetObject(input)
	if err != nil {
		return nil, err
	}

	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a *S3) Set(key string, data []byte, ttl int) error {
	input := &storage.PutObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &key,
		Body:   bytes.NewReader(data),
	}

	_, err := a.Client.PutObject(input)
	if err != nil {
		return err
	}

	return nil
}

func (a *S3) Delete(key string) error {
	input := &storage.DeleteObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &key,
	}

	_, err := a.Client.DeleteObject(input)
	if err != nil {
		return err
	}

	return nil
}

func (a *S3) List(prefix string) ([]string, error) {
	input := &storage.ListObjectsInput{
		Bucket: &a.Config.Bucket,
		Prefix: &prefix,
	}

	result, err := a.Client.ListObjects(input)
	if err != nil {
		return nil, err
	}

	keys := make([]string, len(result.Contents))
	for i, obj := range result.Contents {
		keys[i] = *obj.Key
	}

	return keys, nil
}

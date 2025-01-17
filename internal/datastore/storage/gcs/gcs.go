package gcs

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/padok-team/burrito/internal/burrito/config"
	errors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"google.golang.org/api/iterator"
)

// Implements Storage interface using Google Cloud Storage
type GCS struct {
	// GCS Blob Storage client
	Client *storage.Client
	Config config.GCSConfig
}

// New creates a new Google Cloud Storage client
func New(config config.GCSConfig) *GCS {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		panic(err)
	}
	return &GCS{
		Config: config,
		Client: client,
	}
}

func (a *GCS) Get(key string) ([]byte, error) {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	obj := bucket.Object(key)
	reader, err := obj.NewReader(ctx)
	if err == storage.ErrObjectNotExist {
		return make([]byte, 0), &errors.StorageError{
			Err: err,
			Nil: true,
		}
	}
	if err != nil {
		return make([]byte, 0), &errors.StorageError{
			Err: err,
			Nil: false,
		}
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return make([]byte, 0), &errors.StorageError{
			Err: err,
			Nil: false,
		}
	}

	return data, nil
}

func (a *GCS) Set(key string, data []byte, ttl int) error {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	obj := bucket.Object(key)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	_, err := writer.Write(data)
	if err != nil {
		return &errors.StorageError{
			Err: err,
			Nil: false,
		}
	}

	return nil
}

func (a *GCS) Check(key string) ([]byte, error) {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	obj := bucket.Object(key)
	metadata, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return make([]byte, 0), &errors.StorageError{
			Err: err,
			Nil: true,
		}
	}
	if err != nil {
		return make([]byte, 0), &errors.StorageError{
			Err: err,
			Nil: false,
		}
	}
	return metadata.MD5, nil
}

func (a *GCS) Delete(key string) error {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	obj := bucket.Object(key)
	err := obj.Delete(ctx)
	if err == storage.ErrObjectNotExist {
		return &errors.StorageError{
			Err: err,
			Nil: true,
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (a *GCS) List(prefix string) ([]string, error) {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	it := bucket.Objects(ctx, &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	})

	var objects []string
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		objects = append(objects, objAttrs.Prefix)
	}

	return objects, nil
}

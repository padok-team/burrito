package gcs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/padok-team/burrito/internal/burrito/config"
	storageErrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/datastore/storage/utils"
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
	client, err := storage.NewClient(context.Background(), storage.WithJSONReads())
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
	storageKey := strings.TrimPrefix(key, "/")
	obj := bucket.Object(storageKey)
	reader, err := obj.NewReader(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return nil, &storageErrors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}
	if err != nil {
		return make([]byte, 0), &storageErrors.StorageError{
			Err: fmt.Errorf("error reading object %s: %w", key, err),
			Nil: false,
		}
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		return make([]byte, 0), &storageErrors.StorageError{
			Err: err,
			Nil: false,
		}
	}

	return data, nil
}

func (a *GCS) Set(key string, data []byte, ttl int) error {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	storageKey := strings.TrimPrefix(key, "/")
	obj := bucket.Object(storageKey)
	writer := obj.NewWriter(ctx)
	defer writer.Close()

	_, err := writer.Write(data)
	if err != nil {
		return &storageErrors.StorageError{
			Err: fmt.Errorf("error setting object %s: %w", storageKey, err),
			Nil: false,
		}
	}

	return nil
}

func (a *GCS) Check(key string) ([]byte, error) {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	storageKey := strings.TrimPrefix(key, "/")
	obj := bucket.Object(storageKey)
	metadata, err := obj.Attrs(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return make([]byte, 0), &storageErrors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}
	if err != nil {
		return make([]byte, 0), &storageErrors.StorageError{
			Err: fmt.Errorf("error checking object %s: %w", key, err),
			Nil: false,
		}
	}
	return metadata.MD5, nil
}

func (a *GCS) Delete(key string) error {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	storageKey := strings.TrimPrefix(key, "/")
	obj := bucket.Object(storageKey)
	err := obj.Delete(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return &storageErrors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}
	if err != nil {
		return &storageErrors.StorageError{
			Err: fmt.Errorf("error deleting object %s: %w", key, err),
			Nil: false,
		}
	}

	return nil
}

func (a *GCS) List(prefix string) ([]string, error) {
	ctx := context.Background()
	bucket := a.Client.Bucket(a.Config.Bucket)
	listPrefix := utils.SanitizePrefix(prefix)

	it := bucket.Objects(ctx, &storage.Query{
		Prefix:    listPrefix,
		Delimiter: "/",
	})

	var objects []string
	foundItems := false

	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listing objects with prefix %s: %w", listPrefix, err)
		}
		if objAttrs.Prefix != "" {
			objects = append(objects, "/"+strings.TrimSuffix(objAttrs.Prefix, "/"))
			foundItems = true
		}

		if objAttrs.Name != "" {
			objects = append(objects, "/"+objAttrs.Name)
			foundItems = true
		}
	}

	// If no items were found, return a StorageError with Nil=true
	if !foundItems {
		return nil, &storageErrors.StorageError{
			Err: fmt.Errorf("prefix %s not found", prefix),
			Nil: true,
		}
	}

	return objects, nil
}

package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	sdk "github.com/aws/aws-sdk-go-v2/config"
	storage "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go"
	"github.com/padok-team/burrito/internal/burrito/config"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/datastore/storage/utils"
)

// Implements Storage interface using AWS S3
type S3 struct {
	Client *storage.Client
	Config config.S3Config
}

// New creates a new AWS S3 client
func New(config config.S3Config) *S3 {
	sdkConfig, err := sdk.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	client := storage.NewFromConfig(sdkConfig, func(o *storage.Options) {
		o.UsePathStyle = config.UsePathStyle
	})
	return &S3{
		Config: config,
		Client: client,
	}
}

func (a *S3) Get(key string) ([]byte, error) {
	trimmedKey := strings.TrimPrefix(key, "/")

	input := &storage.GetObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &trimmedKey,
	}

	result, err := a.Client.GetObject(context.TODO(), input)
	if err != nil {
		var noKey *types.NoSuchKey
		if errors.As(err, &noKey) {
			return nil, &storageerrors.StorageError{
				Err: fmt.Errorf("object %s not found", key),
				Nil: true,
			}
		}
		return nil, fmt.Errorf("error getting object %s: %w", key, err)
	}

	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (a *S3) Check(key string) ([]byte, error) {
	trimmedKey := strings.TrimPrefix(key, "/")

	input := &storage.HeadObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &trimmedKey,
	}

	result, err := a.Client.HeadObject(context.TODO(), input)
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				return make([]byte, 0), &storageerrors.StorageError{
					Err: fmt.Errorf("object %s not found", key),
					Nil: true,
				}
			default:
				break
			}
		}
		return make([]byte, 0), fmt.Errorf("error checking object %s: %w", key, err)
	}

	// S3 returns a checksum only if the object was uploaded with one
	if result.ChecksumSHA256 != nil {
		return []byte(*result.ChecksumSHA256), nil
	}

	// Fall back to ETag (MD5) if ChecksumSHA256 is not available
	// This is common in Minio and some S3 implementations
	if result.ETag != nil {
		// ETag is usually returned with quotes, so we need to remove them
		etag := strings.Trim(*result.ETag, "\"")
		return []byte(etag), nil
	}

	// If no checksum is available, return an empty byte array
	return make([]byte, 0), nil
}

func (a *S3) Set(key string, data []byte, ttl int) error {
	trimmedKey := strings.TrimPrefix(key, "/")

	input := &storage.PutObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &trimmedKey,
		Body:   bytes.NewReader(data),
	}
	_, err := a.Client.PutObject(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

func (a *S3) Delete(key string) error {
	trimmedKey := strings.TrimPrefix(key, "/")

	// First check if the file exists
	checkInput := &storage.HeadObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &trimmedKey,
	}

	_, checkErr := a.Client.HeadObject(context.TODO(), checkInput)
	if checkErr != nil {
		var apiError smithy.APIError
		if errors.As(checkErr, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				return &storageerrors.StorageError{
					Err: fmt.Errorf("object %s not found", key),
					Nil: true,
				}
			default:
				return fmt.Errorf("error checking object %s: %w", key, checkErr)
			}
		}
		return fmt.Errorf("error checking object %s: %w", key, checkErr)
	}

	// File exists, proceed with deletion
	input := &storage.DeleteObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &trimmedKey,
	}

	_, err := a.Client.DeleteObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("error deleting object %s: %w", key, err)
	}

	return nil
}

func (a *S3) List(prefix string) ([]string, error) {
	listPrefix := utils.SanitizePrefix(prefix)

	input := &storage.ListObjectsV2Input{
		Bucket:    &a.Config.Bucket,
		Prefix:    aws.String(listPrefix),
		Delimiter: aws.String("/"),
	}
	result, err := a.Client.ListObjectsV2(context.TODO(), input)

	if err != nil {
		return nil, fmt.Errorf("error listing objects with prefix %s: %w", prefix, err)
	}

	// If there are no CommonPrefixes and no Contents, the prefix doesn't exist
	if len(result.CommonPrefixes) == 0 && len(result.Contents) == 0 {
		return nil, &storageerrors.StorageError{
			Err: fmt.Errorf("prefix %s not found", prefix),
			Nil: true,
		}
	}

	var keys []string

	// Add directories
	for _, obj := range result.CommonPrefixes {
		keys = append(keys, "/"+strings.TrimSuffix(*obj.Prefix, "/"))
	}

	// Add files
	for _, obj := range result.Contents {
		keys = append(keys, "/"+*obj.Key)
	}

	return keys, nil
}

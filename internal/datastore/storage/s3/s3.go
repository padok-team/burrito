package s3

import (
	"bytes"
	"context"
	"fmt"
	"io"

	sdk "github.com/aws/aws-sdk-go-v2/config"
	storage "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/padok-team/burrito/internal/burrito/config"
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
	input := &storage.GetObjectInput{
		Bucket: &a.Config.Bucket,
		Key:    &key,
	}

	result, err := a.Client.GetObject(context.TODO(), input)
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

	_, err := a.Client.PutObject(context.TODO(), input)
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

	_, err := a.Client.DeleteObject(context.TODO(), input)
	if err != nil {
		return err
	}

	return nil
}

func (a *S3) List(prefix string) ([]string, error) {
	input := &storage.ListObjectsV2Input{
		Bucket:    &a.Config.Bucket,
		Prefix:    aws.String(fmt.Sprintf("%s/", prefix)),
		Delimiter: aws.String("/"),
	}
	result, err := a.Client.ListObjectsV2(context.TODO(), input)

	if err != nil {
		return nil, err
	}

	keys := make([]string, len(result.CommonPrefixes))
	for i, obj := range result.CommonPrefixes {
		keys[i] = *obj.Prefix
	}

	return keys, nil
}

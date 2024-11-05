package azure

import (
	"context"

	identity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	storage "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	container "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	errors "github.com/padok-team/burrito/internal/datastore/storage/error"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/padok-team/burrito/internal/burrito/config"
)

// Implements Storage interface using Azure Blob Storage

type Azure struct {
	// Azure Blob Storage client
	Client *storage.Client
	Config *config.AzureConfig
}

// New creates a new Azure Blob Storage client
func New(config *config.AzureConfig) *Azure {
	credential, err := identity.NewDefaultAzureCredential(nil)
	if err != nil {
		panic(err)
	}
	client, err := storage.NewClient("https://"+config.StorageAccount+".blob.core.windows.net", credential, nil)
	if err != nil {
		panic(err)
	}
	return &Azure{
		Client: client,
		Config: config,
	}
}

func (a *Azure) Get(key string) ([]byte, error) {
	content := make([]byte, 0)
	_, err := a.Client.DownloadBuffer(context.Background(), a.Config.Container, key, content, &blob.DownloadBufferOptions{})
	if bloberror.HasCode(err, bloberror.BlobNotFound) {
		return nil, &errors.StorageError{
			Err: err,
			Nil: true,
		}
	}
	if err != nil {
		return nil, &errors.StorageError{
			Err: err,
			Nil: false,
		}
	}
	return content, nil
}

func (a *Azure) Set(key string, value []byte, ttl int) error {
	_, err := a.Client.UploadBuffer(context.Background(), a.Config.Container, key, value, nil)
	if err != nil {
		return &errors.StorageError{
			Err: err,
		}
	}
	return nil
}

func (a *Azure) Delete(key string) error {
	_, err := a.Client.DeleteBlob(context.Background(), a.Config.Container, key, nil)
	if err != nil {
		return &errors.StorageError{
			Err: err,
		}
	}
	return nil
}

func (a *Azure) List(prefix string) ([]string, error) {
	keys := []string{}
	marker := ""
	pager := a.Client.NewListBlobsFlatPager(a.Config.Container, &container.ListBlobsFlatOptions{
		Prefix: &prefix,
		Marker: &marker,
	})
	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, err
		}

		for _, blob := range resp.Segment.BlobItems {
			keys = append(keys, *blob.Name)
		}
	}
	return keys, nil
}

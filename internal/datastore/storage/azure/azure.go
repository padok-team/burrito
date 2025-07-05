package azure

import (
	"context"
	"fmt"
	"strings"

	identity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	storage "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	container "github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	errors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/datastore/storage/utils"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/padok-team/burrito/internal/burrito/config"
)

// Implements Storage interface using Azure Blob Storage

type Azure struct {
	// Azure Blob Storage client
	Client          *storage.Client
	Config          config.AzureConfig
	ContainerClient *container.Client
}

// New creates a new Azure Blob Storage client
// If client is nil, a new one will be created using the provided config
// This function can also be used for testing with Azurite by providing a pre-configured client
func New(config config.AzureConfig, client *storage.Client) *Azure {
	// If no client is provided, create one
	if client == nil {
		credential, err := identity.NewDefaultAzureCredential(nil)
		if err != nil {
			panic(err)
		}
		newClient, err := storage.NewClient("https://"+config.StorageAccount+".blob.core.windows.net", credential, nil)
		if err != nil {
			panic(err)
		}
		client = newClient
	}

	containerClient := client.ServiceClient().NewContainerClient(config.Container)

	return &Azure{
		Client:          client,
		Config:          config,
		ContainerClient: containerClient,
	}
}

func (a *Azure) Get(key string) ([]byte, error) {
	// First, check if the blob exists and get its size
	blobClient := a.Client.ServiceClient().NewContainerClient(a.Config.Container).NewBlobClient(key)
	props, err := blobClient.GetProperties(context.Background(), nil)

	// Handle case where blob doesn't exist
	if bloberror.HasCode(err, bloberror.BlobNotFound) {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}

	// Handle other errors
	if err != nil {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("error getting object %s: %w", key, err),
			Nil: false,
		}
	}

	// Get content length and prepare buffer with appropriate size
	contentLength := int(*props.ContentLength)
	content := make([]byte, contentLength)

	// Download the blob into the pre-allocated buffer
	_, err = a.Client.DownloadBuffer(context.Background(), a.Config.Container, key, content, &blob.DownloadBufferOptions{})

	if err != nil {
		return nil, &errors.StorageError{
			Err: err,
			Nil: false,
		}
	}

	return content, nil
}

func (a *Azure) Check(key string) ([]byte, error) {
	resp, err := a.Client.ServiceClient().NewContainerClient(a.Config.Container).NewBlobClient(key).GetProperties(context.Background(), nil)
	if bloberror.HasCode(err, bloberror.BlobNotFound) {
		return make([]byte, 0), &errors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}
	if err != nil {
		return make([]byte, 0), &errors.StorageError{
			Err: fmt.Errorf("error checking object %s: %w", key, err),
			Nil: false,
		}
	}
	return resp.ContentMD5, nil
}

func (a *Azure) Set(key string, value []byte, ttl int) error {
	_, err := a.Client.UploadBuffer(context.Background(), a.Config.Container, key, value, nil)
	if err != nil {
		return &errors.StorageError{
			Err: fmt.Errorf("error setting object %s: %w", key, err),
			Nil: false,
		}
	}
	return nil
}

func (a *Azure) Delete(key string) error {
	_, err := a.Client.DeleteBlob(context.Background(), a.Config.Container, key, nil)
	if bloberror.HasCode(err, bloberror.BlobNotFound) {
		return &errors.StorageError{
			Err: fmt.Errorf("object %s not found", key),
			Nil: true,
		}
	}
	if err != nil {
		return &errors.StorageError{
			Err: fmt.Errorf("error deleting object %s: %w", key, err),
			Nil: false,
		}
	}
	return nil
}

func (a *Azure) List(prefix string) ([]string, error) {
	keys := []string{}
	listPrefix := fmt.Sprintf("/%s", utils.SanitizePrefix(prefix))

	pager := a.ContainerClient.NewListBlobsHierarchyPager("/", &container.ListBlobsHierarchyOptions{
		Prefix: &listPrefix,
	})

	// Variable to track if any items were found
	foundItems := false

	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, &errors.StorageError{
				Err: fmt.Errorf("error listing objects with prefix %s: %w", prefix, err),
				Nil: false,
			}
		}

		// If we have blob items or prefixes, mark that we found items
		if len(resp.Segment.BlobItems) > 0 || len(resp.Segment.BlobPrefixes) > 0 {
			foundItems = true
		}

		for _, blob := range resp.Segment.BlobItems {
			keys = append(keys, *blob.Name)
		}

		for _, prefix := range resp.Segment.BlobPrefixes {
			keys = append(keys, strings.TrimSuffix(*prefix.Name, "/"))
		}
	}

	// If no items were found, return a StorageError with Nil=true
	if !foundItems {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("prefix %s not found", prefix),
			Nil: true,
		}
	}

	return keys, nil
}

// ListRecursive recursively lists all files under a prefix
func (a *Azure) ListRecursive(prefix string) ([]string, error) {
	keys := []string{}
	listPrefix := fmt.Sprintf("/%s", utils.SanitizePrefix(prefix))

	pager := a.ContainerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Prefix: &listPrefix,
	})

	// Variable to track if any items were found
	foundItems := false

	for pager.More() {
		resp, err := pager.NextPage(context.TODO())
		if err != nil {
			return nil, &errors.StorageError{
				Err: fmt.Errorf("error listing objects with prefix %s: %w", prefix, err),
				Nil: false,
			}
		}

		// If we have blob items, mark that we found items
		if len(resp.Segment.BlobItems) > 0 {
			foundItems = true
		}

		for _, blob := range resp.Segment.BlobItems {
			keys = append(keys, *blob.Name)
		}
	}

	// If no items were found, return a StorageError with Nil=true
	if !foundItems {
		return nil, &errors.StorageError{
			Err: fmt.Errorf("prefix %s not found", prefix),
			Nil: true,
		}
	}

	return keys, nil
}

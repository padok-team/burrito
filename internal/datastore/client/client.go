package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/padok-team/burrito/internal/datastore/api"
	"github.com/padok-team/burrito/internal/datastore/storage"
)

const (
	DefaultURL  = "https://datastore.burrito-system"
	DefaultPath = "<>" //TODO: Add default sa token path
)

type Client interface {
	GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error)
	PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error
}

type DefaultClient struct {
	URL  string
	Path string
}

func NewDefaultClient() *DefaultClient {
	return &DefaultClient{
		URL:  DefaultURL,
		Path: DefaultPath,
	}
}

func (c *DefaultClient) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	queryParams := url.Values{
		"namespace": {namespace},
		"layer":     {layer},
		"run":       {run},
		"attempt":   {attempt},
		"format":    {format},
	}
	url := "https://" + c.URL + "/plan?" + queryParams.Encode()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, &storage.StorageError{
			Err: fmt.Errorf("no plan for this attempt"),
			Nil: true,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not get plan, there's an issue with the storage backend")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *DefaultClient) PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error {
	queryParams := url.Values{
		"namespace": {namespace},
		"layer":     {layer},
		"run":       {run},
		"attempt":   {attempt},
		"format":    {format},
	}
	url := "https://" + c.URL + "/plan?" + queryParams.Encode()
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}
	requestBody := api.PutLogsRequest{
		Content: string(content),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	stringReader := strings.NewReader(string(body))
	stringReadCloser := io.NopCloser(stringReader)
	req.Body = stringReadCloser
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not put plan, there's an issue with the storage backend")
	}
	return nil
}

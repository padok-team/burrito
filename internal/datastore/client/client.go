package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/padok-team/burrito/internal/datastore/api"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

const (
	DefaultURL  = "https://datastore.burrito-system"
	DefaultPath = "/var/run/secrets/token/burrito"
)

type Client interface {
	GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error)
	PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error
	GetLogs(namespace string, layer string, run string, attempt string) ([]string, error)
	PutLogs(namespace string, layer string, run string, attempt string, content []byte) error
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
		return nil, &storageerrors.StorageError{
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
	jresp := api.GetPlanResponse{}
	err = json.Unmarshal(b, &jresp)
	if err != nil {
		return nil, err
	}
	return jresp.Plan, nil
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

func (c *DefaultClient) GetLogs(namespace string, layer string, run string, attempt string) ([]string, error) {
	queryParams := url.Values{
		"namespace": {namespace},
		"layer":     {layer},
		"run":       {run},
		"attempt":   {attempt},
	}
	url := "https://" + c.URL + "/logs?" + queryParams.Encode()
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, &storageerrors.StorageError{
			Err: fmt.Errorf("no logs for this attempt"),
			Nil: true,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("could not get logs, there's an issue with the storage backend")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	jresp := api.GetLogsResponse{}
	err = json.Unmarshal(b, &jresp)
	if err != nil {
		return nil, err
	}
	return jresp.Results, nil
}

func (c *DefaultClient) PutLogs(namespace string, layer string, run string, attempt string, content []byte) error {
	queryParams := url.Values{
		"namespace": {namespace},
		"layer":     {layer},
		"run":       {run},
		"attempt":   {attempt},
	}
	url := "https://" + c.URL + "/logs?" + queryParams.Encode()
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
		return fmt.Errorf("could not put logs, there's an issue with the storage backend")
	}
	return nil
}

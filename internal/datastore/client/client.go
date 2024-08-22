package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/padok-team/burrito/internal/datastore/api"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultHostname  = "burrito-datastore.burrito-system"
	DefaultTokenPath = "/var/run/secrets/token/burrito"
)

type Client interface {
	GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error)
	PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error
	GetLogs(namespace string, layer string, run string, attempt string) ([]string, error)
	PutLogs(namespace string, layer string, run string, attempt string, content []byte) error
}

type DefaultClient struct {
	Hostname  string
	tokenPath string
	scheme    string
	client    *http.Client
}

func NewDefaultClient(useTLS bool) *DefaultClient {
	scheme := "http"
	if useTLS {
		log.Info("using TLS for datastore")
		scheme = "https"
	}
	return &DefaultClient{
		Hostname:  DefaultHostname,
		tokenPath: DefaultTokenPath,
		scheme:    scheme,
		client:    &http.Client{},
	}
}

func (c *DefaultClient) buildRequest(path string, queryParams url.Values, method string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s://%s%s?%s", c.scheme, c.Hostname, path, queryParams.Encode())
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	token, err := os.ReadFile(c.tokenPath)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(token))
	return req, nil
}

func (c *DefaultClient) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	req, err := c.buildRequest("/api/plans", url.Values{
		"namespace": {namespace},
		"layer":     {layer},
		"run":       {run},
		"attempt":   {attempt},
		"format":    {format},
	}, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
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
	return b, nil
}

func (c *DefaultClient) PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error {
	req, err := c.buildRequest(
		"/api/plans",
		url.Values{
			"namespace": {namespace},
			"layer":     {layer},
			"run":       {run},
			"attempt":   {attempt},
			"format":    {format},
		},
		http.MethodPut,
		bytes.NewBuffer(content),
	)
	req.Header.Set("Content-Type", "application/octet-stream")
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		message, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("could not put plan, there's an issue reading the response from datastore: %s", err)
		}
		return fmt.Errorf("could not put plan, there's an issue with the storage backend: %s", string(message))
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
	req, err := c.buildRequest(
		"/api/logs",
		queryParams,
		http.MethodGet,
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
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
	req, err := c.buildRequest(
		"/api/logs",
		queryParams,
		http.MethodPut,
		bytes.NewBuffer(content),
	)
	req.Header.Set("Content-Type", "application/octet-stream")
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("could not put logs, there's an issue with the storage backend")
	}
	return nil
}

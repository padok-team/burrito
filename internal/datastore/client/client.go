package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/padok-team/burrito/internal/datastore/api"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

const (
	DefaultHostname = "datastore.burrito-system"
	DefaultPath     = "/var/run/secrets/token/burrito"
)

type Client interface {
	GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error)
	PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error
	GetLogs(namespace string, layer string, run string, attempt string) ([]string, error)
	PutLogs(namespace string, layer string, run string, attempt string, content []byte) error
	GetAttempts(namespace string, layer string, run string) (int, error)
}

type DefaultClient struct {
	Hostname string
	Path     string
	Scheme   string
	client   *http.Client
}

func NewDefaultClient() *DefaultClient {
	return &DefaultClient{
		Hostname: DefaultHostname,
		Path:     DefaultPath,
		Scheme:   "http",
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (c *DefaultClient) buildRequest(path string, queryParams url.Values, method string, body io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s://%s%s?%s", c.Scheme, c.Hostname, path, queryParams.Encode())
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	token, err := os.ReadFile(c.Path)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(token))
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *DefaultClient) GetPlan(namespace string, layer string, run string, attempt string, format string) ([]byte, error) {
	req, err := c.buildRequest("/plan", url.Values{
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
	jresp := api.GetPlanResponse{}
	err = json.Unmarshal(b, &jresp)
	if err != nil {
		return nil, err
	}
	return jresp.Plan, nil
}

func (c *DefaultClient) PutPlan(namespace string, layer string, run string, attempt string, format string, content []byte) error {
	requestBody := api.PutLogsRequest{
		Content: string(content),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	req, err := c.buildRequest(
		"/plan",
		url.Values{
			"namespace": {namespace},
			"layer":     {layer},
			"run":       {run},
			"attempt":   {attempt},
			"format":    {format},
		},
		http.MethodPut,
		strings.NewReader(string(body)),
	)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
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
	req, err := c.buildRequest(
		"/logs",
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
	requestBody := api.PutLogsRequest{
		Content: string(content),
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	queryParams := url.Values{
		"namespace": {namespace},
		"layer":     {layer},
		"run":       {run},
		"attempt":   {attempt},
	}
	req, err := c.buildRequest(
		"/logs",
		queryParams,
		http.MethodPut,
		strings.NewReader(string(body)),
	)
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

func (c *DefaultClient) GetAttempts(namespace string, layer string, run string) (int, error) {
	req, err := c.buildRequest(
		"/attempts",
		url.Values{
			"namespace": {namespace},
			"layer":     {layer},
			"run":       {run},
		},
		http.MethodGet,
		nil,
	)
	if err != nil {
		return 0, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("could not get attempts, there's an issue with the storage backend")
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	jresp := api.GetAttemptsResponse{}
	err = json.Unmarshal(b, &jresp)
	if err != nil {
		return 0, err
	}
	return jresp.Attempts, nil
}

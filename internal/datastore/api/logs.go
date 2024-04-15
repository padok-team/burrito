package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/padok-team/burrito/internal/datastore/storage"
)

type GetLogsResponse struct {
	Results []string `json:"results"`
}

type PutLogsRequest struct {
	Content string `json:"content"`
}

func getArgs(c echo.Context) (string, string, string, string, error) {
	namespace := c.QueryParam("namespace")
	layer := c.QueryParam("layer")
	run := c.QueryParam("run")
	attempt := c.QueryParam("attempt")
	if namespace == "" || layer == "" || run == "" {
		return "", "", "", "", fmt.Errorf("missing query parameters")
	}
	return namespace, layer, run, attempt, nil
}

func (a *API) GetLogsHandler(c echo.Context) error {
	var err error
	var content []byte
	namespace, layer, run, attempt, err := getArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	response := GetLogsResponse{}
	if attempt == "" {
		content, err = a.Storage.GetLatestLogs(namespace, layer, run)
	} else {
		content, err = a.Storage.GetLogs(namespace, layer, run, attempt)
	}
	if storage.NotFound(err) {
		return c.String(http.StatusNotFound, "No logs for this attempt")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get logs, there's an issue with the storage backend")
	}
	for _, line := range strings.Split(string(content), "\n") {
		response.Results = append(response.Results, line)
	}
	return c.JSON(http.StatusOK, &response)
}

func (a *API) PutLogsHandler(c echo.Context) error {
	var err error
	var content []byte
	namespace, layer, run, attempt, err := getArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	request := PutLogsRequest{}
	if err := c.Bind(&request); err != nil {
		return c.String(http.StatusBadRequest, "could not read request body")
	}
	content = []byte(request.Content)
	err = a.Storage.PutLogs(namespace, layer, run, attempt, content)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not put logs, there's an issue with the storage backend")
	}
	return c.NoContent(http.StatusOK)
}

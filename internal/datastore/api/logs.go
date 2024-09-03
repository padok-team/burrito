package api

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

type GetLogsResponse struct {
	Results []string `json:"results"`
}

func getLogsArgs(c echo.Context) (string, string, string, string, error) {
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
	namespace, layer, run, attempt, err := getLogsArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	response := GetLogsResponse{}
	if attempt == "" {
		content, err = a.Storage.GetLatestLogs(namespace, layer, run)
	} else {
		content, err = a.Storage.GetLogs(namespace, layer, run, attempt)
	}
	if storageerrors.NotFound(err) {
		return c.String(http.StatusNotFound, "No logs for this attempt")
	}
	if err != nil {
		c.Logger().Errorf("Could not get logs, there's an issue with the storage backend : %s", err)
		return c.String(http.StatusInternalServerError, "could not get logs, there's an issue with the storage backend")
	}
	response.Results = append(response.Results, strings.Split(string(content), "\n")...)
	return c.JSON(http.StatusOK, &response)
}

func (a *API) PutLogsHandler(c echo.Context) error {
	var err error
	namespace, layer, run, attempt, err := getLogsArgs(c)
	if attempt == "" {
		return c.String(http.StatusBadRequest, "missing query parameters")
	}
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	content, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not read request body: "+err.Error())
	}
	err = a.Storage.PutLogs(namespace, layer, run, attempt, content)
	if err != nil {
		c.Logger().Errorf("Could not put logs, there's an issue with the storage backend : %s", err)
		return c.String(http.StatusInternalServerError, "could not put logs, there's an issue with the storage backend")
	}
	return c.NoContent(http.StatusOK)
}

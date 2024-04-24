package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type Attempt struct {
	AttemptNumber string `param:"attempt"`
	Namespace     string `param:"namespace"`
	Layer         string `param:"layer"`
	Run           string `param:"run"`
}

type GetLogsResponse struct {
	Results []string `json:"results"`
}

func getLogsArgs(c echo.Context) (string, string, string, string, error) {
	attempt := Attempt{}
	if err := c.Bind(attempt); err != nil {
		return "", "", "", "", fmt.Errorf("missing query parameters")
	}
	return attempt.Namespace, attempt.Layer, attempt.Run, attempt.AttemptNumber, nil
}

// logs/${namespace}/${layer}/${runId}/${attemptId}
func (a *API) GetLogsHandler(c echo.Context) error {
	namespace, layer, run, attempt, err := getLogsArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	response := GetLogsResponse{}
	content, err := a.Datastore.GetLogs(namespace, layer, run, attempt)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get logs, there's an issue with the storage backend")
	}
	response.Results = content
	return c.JSON(http.StatusOK, &response)
}

func (a *API) DownloadLogsHandler(c echo.Context) error {
	namespace, layer, run, attempt, err := getLogsArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	content, err := a.Datastore.GetLogs(namespace, layer, run, attempt)
	file := strings.Join(content, "\n")
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get logs, there's an issue with the storage backend")
	}
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s_%s_%s.log", namespace, layer, run, attempt))
	c.Response().Header().Set("Content-Type", "application/octet-stream")
	return c.Blob(http.StatusOK, "application/octet-stream", []byte(file))
}

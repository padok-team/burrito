package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

type GetPlanResponse struct {
	Plan   []byte `json:"plan"`
	Format string `json:"format"`
}

type PutPlanRequest struct {
	Plan []byte `json:"content"`
}

func getPlanArgs(c echo.Context) (string, string, string, string, string, error) {
	namespace := c.QueryParam("namespace")
	layer := c.QueryParam("layer")
	run := c.QueryParam("run")
	attempt := c.QueryParam("attempt")
	format := c.QueryParam("format")
	if namespace == "" || layer == "" || run == "" {
		return "", "", "", "", "", fmt.Errorf("missing query parameters")
	}
	if format == "" {
		format = "json"
	}
	return namespace, layer, run, attempt, format, nil
}

func (a *API) GetPlanHandler(c echo.Context) error {
	var err error
	var content []byte
	namespace, layer, run, attempt, format, err := getPlanArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	response := GetPlanResponse{}
	if attempt == "" {
		content, err = a.Storage.GetLatestPlan(namespace, layer, run, format)
	} else {
		content, err = a.Storage.GetPlan(namespace, layer, run, attempt, format)
	}
	if storageerrors.NotFound(err) {
		return c.String(http.StatusNotFound, "No logs for this attempt")
	}
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get logs, there's an issue with the storage backend")
	}
	response.Plan = content
	response.Format = format
	return c.JSON(http.StatusOK, &response)
}

func (a *API) PutPlanHandler(c echo.Context) error {
	var err error
	var content []byte
	namespace, layer, run, attempt, format, err := getPlanArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if attempt == "" || format == "" {
		return c.String(http.StatusBadRequest, "missing query parameters")
	}
	request := PutPlanRequest{}
	if err := c.Bind(&request); err != nil {
		return c.String(http.StatusBadRequest, "could not read request body")
	}
	content = []byte(request.Plan)
	err = a.Storage.PutPlan(namespace, layer, run, attempt, format, content)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not put logs, there's an issue with the storage backend: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

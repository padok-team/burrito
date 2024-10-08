package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

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
	if attempt == "" {
		content, err = a.Storage.GetLatestPlan(namespace, layer, run, format)
	} else {
		content, err = a.Storage.GetPlan(namespace, layer, run, attempt, format)
	}
	if storageerrors.NotFound(err) {
		return c.String(http.StatusNotFound, "No plan for this attempt")
	}
	if err != nil {
		c.Logger().Errorf("Could not get plan, there's an issue with the storage backend : %s", err)
		return c.String(http.StatusInternalServerError, "could not get plan, there's an issue with the storage backend")
	}
	return c.Blob(http.StatusOK, "application/octet-stream", content)
}

func (a *API) PutPlanHandler(c echo.Context) error {
	var err error
	namespace, layer, run, attempt, format, err := getPlanArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	if attempt == "" || format == "" {
		return c.String(http.StatusBadRequest, "missing query parameters")
	}
	content, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not read request body: "+err.Error())
	}
	err = a.Storage.PutPlan(namespace, layer, run, attempt, format, content)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not put plan, there's an issue with the storage backend: "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
)

func getPlanArgs(c echo.Context) (string, string, string, string, error) {
	namespace := c.Param("namespace")
	layer := c.Param("layer")
	run := c.Param("run")
	attempt := c.Param("attempt")

	if namespace == "" || layer == "" || run == "" || attempt == "" {
		return "", "", "", "", fmt.Errorf("missing path parameters")
	}

	return namespace, layer, run, attempt, nil
}

// plans/${namespace}/${layer}/${run}/${attempt}?format=<format>
func (a *API) GetPlanHandler(c echo.Context) error {
	namespace, layer, run, attempt, err := getPlanArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	content, err := a.Datastore.GetPlan(namespace, layer, run, attempt, "json")
	if storageerrors.NotFound(err) {
		return c.String(http.StatusNotFound, "plan not found for this attempt")
	}
	if err != nil {
		c.Logger().Errorf("failed to retrieve plan %s/%s run %s attempt %s: %v", namespace, layer, run, attempt, err)
		return c.String(http.StatusInternalServerError, "could not get plan, there's an issue with the storage backend")
	}

	return c.JSONBlob(http.StatusOK, content)
}

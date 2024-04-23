package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type GetAttemptsResponse struct {
	Count int `json:"count"`
}

type Run struct {
	Namespace string `param:"namespace"`
	Layer     string `param:"layer"`
	Run       string `param:"run"`
}

func getRunAttemptArgs(c echo.Context) (string, string, string, error) {
	run := Run{}
	if err := c.Bind(run); err != nil {
		return "", "", "", fmt.Errorf("missing query parameters")
	}
	return run.Namespace, run.Layer, run.Run, nil
}

func (a *API) GetAttemptsHandler(c echo.Context) error {
	namespace, layer, run, err := getRunAttemptArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	attempts, err := a.Datastore.GetAttempts(namespace, layer, run)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get run attempt, there's an issue with the storage backend")
	}
	response := GetAttemptsResponse{Count: attempts}
	return c.JSON(http.StatusOK, &response)
}

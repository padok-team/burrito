package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type GetAttemptsResponse struct {
	Count int `json:"count"`
}

func getRunAttemptArgs(c echo.Context) (string, string, string, error) {
	namespace := c.Param("namespace")
	layer := c.Param("layer")
	run := c.Param("run")
	if namespace == "" || layer == "" || run == "" {
		return "", "", "", fmt.Errorf("missing query parameters")
	}
	return namespace, layer, run, nil
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

package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type GetAttemptsResponse struct {
	Attempts int `json:"count"`
}

func getAttemptsArgs(c echo.Context) (string, string, string, error) {
	namespace := c.QueryParam("namespace")
	layer := c.QueryParam("layer")
	run := c.QueryParam("run")
	if namespace == "" || layer == "" || run == "" {
		return "", "", "", fmt.Errorf("missing query parameters")
	}
	return namespace, layer, run, nil
}

func (a *API) GetAttemptsHandler(c echo.Context) error {
	namespace, layer, run, err := getAttemptsArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	attempts, err := a.Storage.GetAttempts(namespace, layer, run)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get attempts, there's an issue with the storage backend")
	}
	response := GetAttemptsResponse{
		Attempts: attempts,
	}
	return c.JSON(http.StatusOK, &response)
}

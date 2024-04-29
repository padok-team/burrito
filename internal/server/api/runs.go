package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type GetAttemptsResponse struct {
	Count int `json:"count"`
}

func getRunAttemptArgs(c echo.Context) (string, string, error) {
	namespace := c.Param("namespace")
	run := c.Param("run")
	if namespace == "" || run == "" {
		return "", "", fmt.Errorf("missing query parameters")
	}
	return namespace, run, nil
}

func (a *API) GetAttemptsHandler(c echo.Context) error {
	namespace, run, err := getRunAttemptArgs(c)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	runObject := &configv1alpha1.TerraformRun{}
	err = a.Client.Get(context.Background(), types.NamespacedName{Name: run, Namespace: namespace}, runObject)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get run attempt, there's an issue with the cluster: "+err.Error())
	}
	response := GetAttemptsResponse{Count: len(runObject.Status.Attempts)}
	return c.JSON(http.StatusOK, &response)
}

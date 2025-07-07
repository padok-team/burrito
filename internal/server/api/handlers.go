package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	terraformv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (a *API) SyncLayerHandler(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("layer")
	layer := &terraformv1alpha1.TerraformLayer{}
	err := a.Client.Get(c.Request().Context(), client.ObjectKey{Namespace: namespace, Name: name}, layer)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not get terraform layer")
	}
	// Add sync annotation to trigger manual sync
	err = annotations.Add(c.Request().Context(), a.Client, layer, map[string]string{
		annotations.SyncNow: "true",
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not add sync annotation")
	}
	return c.String(http.StatusOK, "OK")
}

func (a *API) ApplyLayerHandler(c echo.Context) error {
	namespace := c.Param("namespace")
	name := c.Param("layer")
	layer := &terraformv1alpha1.TerraformLayer{}
	err := a.Client.Get(c.Request().Context(), client.ObjectKey{Namespace: namespace, Name: name}, layer)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not get terraform layer")
	}

	// Check if layer is managed by TerraformPullRequest controller
	if managedBy, exists := layer.Labels["burrito/managed-by"]; exists && managedBy != "" {
		return c.String(http.StatusForbidden, "Manual apply is not allowed on layers managed by TerraformPullRequest controller")
	}

	// Add apply annotation to trigger manual apply
	err = annotations.Add(c.Request().Context(), a.Client, layer, map[string]string{
		annotations.ApplyNow: "true",
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Could not add apply annotation")
	}
	return c.String(http.StatusOK, "OK")
}

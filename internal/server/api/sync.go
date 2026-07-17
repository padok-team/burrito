package api

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/server/utils"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (a *API) SyncLayerHandler(c echo.Context) error {
	layer := &configv1alpha1.TerraformLayer{}
	err := a.Client.Get(context.Background(), client.ObjectKey{
		Namespace: c.Param("namespace"),
		Name:      c.Param("layer"),
	}, layer)
	if err != nil {
		log.Errorf("could not get terraform layer: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while getting the layer"})
	}
	syncStatus := utils.GetManualSyncStatus(*layer)
	if syncStatus == utils.ManualSyncAnnotated || syncStatus == utils.ManualSyncPending {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Layer sync already triggered"})
	}

	err = annotations.Add(context.Background(), a.Client, layer, map[string]string{
		annotations.SyncNow: "true",
	})
	if err != nil {
		log.Errorf("could not update terraform layer annotations: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while updating the layer annotations"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "Layer sync triggered"})
}

func (a *API) ApplyLayerHandler(c echo.Context) error {
	layer := &configv1alpha1.TerraformLayer{}
	err := a.Client.Get(context.Background(), client.ObjectKey{
		Namespace: c.Param("namespace"),
		Name:      c.Param("layer"),
	}, layer)
	if err != nil {
		log.Errorf("could not get terraform layer: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while getting the layer"})
	}
	// Check if layer is managed by TerraformPullRequest controller
	if managedBy, exists := layer.Labels["burrito/managed-by"]; exists && managedBy != "" {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Manual apply is not allowed on layers managed by TerraformPullRequest controller"})
	}
	// Add apply annotation to trigger manual apply
	err = annotations.Add(context.Background(), a.Client, layer, map[string]string{
		annotations.ApplyNow: "true",
	})
	if err != nil {
		log.Errorf("could not update terraform layer annotations: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while updating the layer annotations"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "Layer apply triggered"})
}

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	running, err := a.isLayerRunning(*layer)
	if err != nil {
		log.Errorf("could not check if layer is running: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while checking if the layer is running"})
	}
	if running {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Layer is already running"})
	}

	err = a.createPlanRun(*layer)
	if err != nil {
		log.Errorf("could not create plan run: %s", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "An error occurred while creating the plan run"})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "Layer sync triggered"})
}

func (a *API) fetchAssociatedRuns(layer configv1alpha1.TerraformLayer) ([]configv1alpha1.TerraformRun, error) {
	runs := &configv1alpha1.TerraformRunList{}
	err := a.Client.List(context.Background(), runs, client.InNamespace(layer.Namespace), client.MatchingLabels{"burrito/managed-by": layer.Name})
	if err != nil {
		log.Errorf("could not list terraform runs: %s", err)
		return nil, err
	}
	return runs.Items, nil
}

func (a *API) isLayerRunning(layer configv1alpha1.TerraformLayer) (bool, error) {
	runs, err := a.fetchAssociatedRuns(layer)
	if err != nil {
		return false, err
	}
	for _, run := range runs {
		if run.Status.State != "Failed" && run.Status.State != "Succeeded" {
			return true, nil
		}
	}
	return false, nil
}

func getPlanRun(layer *configv1alpha1.TerraformLayer) configv1alpha1.TerraformRun {
	artifact := configv1alpha1.Artifact{}
	return configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-manual-", layer.Name, "plan"),
			Namespace:    layer.Namespace,
			Labels: map[string]string{
				"burrito/managed-by": layer.Name,
				"burrito/manual-run": "true",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: layer.GetAPIVersion(),
					Kind:       layer.GetKind(),
					Name:       layer.Name,
					UID:        layer.UID,
				},
			},
		},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: string("plan"),
			Layer: configv1alpha1.TerraformRunLayer{
				Name:      layer.Name,
				Namespace: layer.Namespace,
			},
			Artifact: artifact,
		},
	}
}

func (a *API) createPlanRun(l configv1alpha1.TerraformLayer) error {
	run := getPlanRun(&l)
	err := a.Client.Create(context.Background(), &run)
	if err != nil {
		return err
	}
	return nil
}

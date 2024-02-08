package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type layer struct {
	UID        string `json:"uid"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Path       string `json:"path"`
	State      string `json:"state"`
	RunCount   int    `json:"runCount"`
	LastResult string `json:"lastResult"`
	IsRunning  bool   `json:"isRunning"`
	IsPR       bool   `json:"isPR"`
}

type layersResponse struct {
	Results []layer `json:"results"`
}

func (a *API) LayersHandler(c echo.Context) error {
	layers := &configv1alpha1.TerraformLayerList{}
	err := a.Client.List(context.Background(), layers)
	if err != nil {
		log.Errorf("could not list terraform layers: %s", err)
		return c.String(http.StatusInternalServerError, "could not list terraform layers")
	}
	if err != nil {
		log.Errorf("could not list terraform layers: %s", err)
		return c.String(http.StatusInternalServerError, "could not list terraform layers")
	}
	results := []layer{}
	for _, l := range layers.Items {
		layerRunning, runCount := a.getLayerRunInfo(l)
		results = append(results, layer{
			UID:        string(l.UID),
			Name:       l.Name,
			Namespace:  l.Namespace,
			Repository: fmt.Sprintf("%s/%s", l.Spec.Repository.Namespace, l.Spec.Repository.Name),
			Branch:     l.Spec.Branch,
			Path:       l.Spec.Path,
			State:      a.getLayerState(l),
			RunCount:   runCount,
			LastResult: l.Status.LastResult,
			IsRunning:  layerRunning,
			IsPR:       a.isLayerPR(l),
		})
	}
	return c.JSON(http.StatusOK, &layersResponse{
		Results: results,
	},
	)
}

func (a *API) getLayerState(layer configv1alpha1.TerraformLayer) string {
	state := "success"
	switch {
	case len(layer.Status.Conditions) == 0:
		state = "disabled"
	case layer.Status.State == "ApplyNeeded":
		if layer.Status.LastResult == "Plan: 0 to create, 0 to update, 0 to delete" {
			state = "success"
		} else {
			state = "warning"
		}
	case layer.Status.State == "PlanNeeded":
		state = "warning"
	}
	if layer.Annotations[annotations.LastPlanSum] == "" {
		state = "error"
	}
	if layer.Annotations[annotations.LastApplySum] != "" && layer.Annotations[annotations.LastApplySum] == "" {
		state = "error"
	}
	return state
}

func (a *API) getLayerRunInfo(layer configv1alpha1.TerraformLayer) (layerRunning bool, runCount int) {
	runs := &configv1alpha1.TerraformRunList{}
	requirement, _ := labels.NewRequirement("burrito/managed-by", selection.Equals, []string{layer.Name})
	selector := labels.NewSelector().Add(*requirement)
	err := a.Client.List(context.Background(), runs, client.MatchingLabelsSelector{Selector: selector})
	runCount = len(runs.Items)
	if err != nil {
		log.Errorf("could not list terraform runs, returning false: %s", err)
		layerRunning = false
		return
	}
	for _, r := range runs.Items {
		if r.Status.State == "Running" {
			layerRunning = true
			return
		}
	}
	layerRunning = false
	return
}

func (a *API) isLayerPR(layer configv1alpha1.TerraformLayer) bool {
	if len(layer.OwnerReferences) == 0 {
		return false
	}
	return layer.OwnerReferences[0].Kind == "TerraformPullRequest"
}

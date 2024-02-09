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
	LastRunAt  string `json:"lastRunAt"`
	LastResult string `json:"lastResult"`
	IsRunning  bool   `json:"isRunning"`
	IsPR       bool   `json:"isPR"`
	LatestRuns []run  `json:"latestRuns"`
}

type run struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Action     string `json:"action"`
	Status     string `json:"status"`
	LastRun    string `json:"lastRun"`
	Retries    int    `json:"retries"`
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
		layerRunning, runCount, latestRuns, lastRunAt := a.getLayerRunInfo(l)
		results = append(results, layer{
			UID:        string(l.UID),
			Name:       l.Name,
			Namespace:  l.Namespace,
			Repository: fmt.Sprintf("%s/%s", l.Spec.Repository.Namespace, l.Spec.Repository.Name),
			Branch:     l.Spec.Branch,
			Path:       l.Spec.Path,
			State:      a.getLayerState(l),
			RunCount:   runCount,
			LastRunAt:  lastRunAt,
			LastResult: l.Status.LastResult,
			IsRunning:  layerRunning,
			IsPR:       a.isLayerPR(l),
			LatestRuns: latestRuns,
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

func (a *API) getLayerRunInfo(layer configv1alpha1.TerraformLayer) (layerRunning bool, runCount int, latestRuns []run, lastRunAt string) {
	runs := &configv1alpha1.TerraformRunList{}
	requirement, _ := labels.NewRequirement("burrito/managed-by", selection.Equals, []string{layer.Name})
	selector := labels.NewSelector().Add(*requirement)
	err := a.Client.List(context.Background(), runs, client.MatchingLabelsSelector{Selector: selector})
	runCount = len(runs.Items)
	layerRunning = false
	if err != nil {
		log.Errorf("could not list terraform runs, returning false: %s", err)
		return
	}
	lastRunAt = runs.Items[len(runs.Items)-1].Status.LastRun
	for i := len(runs.Items)-1 ; i >= 0 ; i-- {
		r := runs.Items[i]
		if r.Status.State == "Running" {
			layerRunning = true
			if (len(runs.Items)-1-i) >= 5 {
				return
			}
		}
		if (len(runs.Items)-1-i) < 5 {
			latestRuns = append(latestRuns, run{
				ID:         string(r.UID),
				Name:       r.Name,
				Namespace:  r.Namespace,
				Action:     r.Spec.Action,
				Status:     r.Status.State,
				LastRun:    r.Status.LastRun,
				Retries:    r.Status.Retries,
			})
		}
	}
	return
}

func (a *API) isLayerPR(layer configv1alpha1.TerraformLayer) bool {
	if len(layer.OwnerReferences) == 0 {
		return false
	}
	return layer.OwnerReferences[0].Kind == "TerraformPullRequest"
}

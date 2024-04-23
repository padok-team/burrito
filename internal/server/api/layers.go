package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
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
	Name   string `json:"id"`
	Commit string `json:"commit"`
	Date   string `json:"date"`
	Action string `json:"action"`
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
		run, err := a.getLatestRun(l)
		if err != nil {
			log.Errorf("could not get latest run for layer %s: %s", l.Name, err)
		}
		results = append(results, layer{
			UID:        string(l.UID),
			Name:       l.Name,
			Namespace:  l.Namespace,
			Repository: fmt.Sprintf("%s/%s", l.Spec.Repository.Namespace, l.Spec.Repository.Name),
			Branch:     l.Spec.Branch,
			Path:       l.Spec.Path,
			State:      a.getLayerState(l),
			RunCount:   len(l.Status.LatestRuns),
			LastRunAt:  l.Status.LastRun.Date.Format(time.RFC3339),
			LastResult: l.Status.LastResult,
			IsRunning:  run.Status.State != "Succeeded" && run.Status.State != "Failed",
			IsPR:       a.isLayerPR(l),
			LatestRuns: transformLatestRuns(l.Status.LatestRuns),
		})
	}
	return c.JSON(http.StatusOK, &layersResponse{
		Results: results,
	},
	)
}

func transformLatestRuns(runs []configv1alpha1.TerraformLayerRun) []run {
	results := []run{}
	for _, r := range runs {
		results = append(results, run{
			Name:   r.Name,
			Commit: r.Commit,
			Date:   r.Date.Format(time.RFC3339),
			Action: r.Action,
		})
	}
	return results
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

func (a *API) getLayerRunState(layer configv1alpha1.TerraformLayer) string {
	if len(layer.Status.LatestRuns) == 0 {
		return "disabled"
	}
	return layer.Status.LatestRuns[0].Action
}

func (a *API) getLatestRun(layer configv1alpha1.TerraformLayer) (configv1alpha1.TerraformRun, error) {
	run := &configv1alpha1.TerraformRun{}
	err := a.Client.Get(context.Background(), types.NamespacedName{
		Namespace: layer.Namespace,
		Name:      layer.Status.LastRun.Name,
	}, run)
	if err != nil {
		log.Errorf("could not get latest run for layer %s: %s", layer.Name, err)
	}
	return *run, err
}

func (a *API) isLayerPR(layer configv1alpha1.TerraformLayer) bool {
	if len(layer.OwnerReferences) == 0 {
		return false
	}
	return layer.OwnerReferences[0].Kind == "TerraformPullRequest"
}

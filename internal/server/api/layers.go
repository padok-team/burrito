package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/server/utils"
	log "github.com/sirupsen/logrus"
)

type layer struct {
	UID              string                 `json:"uid"`
	Name             string                 `json:"name"`
	Namespace        string                 `json:"namespace"`
	Repository       string                 `json:"repository"`
	Branch           string                 `json:"branch"`
	Path             string                 `json:"path"`
	State            string                 `json:"state"`
	RunCount         int                    `json:"runCount"`
	LastRun          Run                    `json:"lastRun"`
	LastRunAt        string                 `json:"lastRunAt"`
	LastResult       string                 `json:"lastResult"`
	IsRunning        bool                   `json:"isRunning"`
	IsPR             bool                   `json:"isPR"`
	LatestRuns       []Run                  `json:"latestRuns"`
	ManualSyncStatus utils.ManualSyncStatus `json:"manualSyncStatus"`
}

type Run struct {
	Name   string `json:"id"`
	Commit string `json:"commit"`
	Date   string `json:"date"`
	Action string `json:"action"`
}

type layersResponse struct {
	Results []layer `json:"results"`
}

func (a *API) getLayersAndRuns() ([]configv1alpha1.TerraformLayer, map[string]configv1alpha1.TerraformRun, error) {
	layers := &configv1alpha1.TerraformLayerList{}
	err := a.Client.List(context.Background(), layers)
	if err != nil {
		log.Errorf("could not list TerraformLayers: %s", err)
		return nil, nil, err
	}
	runs := &configv1alpha1.TerraformRunList{}
	indexedRuns := map[string]configv1alpha1.TerraformRun{}
	err = a.Client.List(context.Background(), runs)
	if err != nil {
		log.Errorf("could not list TerraformRuns: %s", err)
		return nil, nil, err
	}
	for _, run := range runs.Items {
		indexedRuns[fmt.Sprintf("%s/%s", run.Namespace, run.Name)] = run
	}
	return layers.Items, indexedRuns, err
}

func (a *API) LayersHandler(c echo.Context) error {
	layers, runs, err := a.getLayersAndRuns()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("could not list terraform layers or runs: %s", err))
	}
	results := []layer{}
	for _, l := range layers {
		if err != nil {
			log.Errorf("could not get latest run for layer %s: %s", l.Name, err)
		}
		run, ok := runs[fmt.Sprintf("%s/%s", l.Namespace, l.Status.LastRun.Name)]
		runAPI := Run{}
		running := false
		if ok {
			runAPI = Run{
				Name:   run.Name,
				Commit: "",
				Date:   run.CreationTimestamp.Format(time.RFC3339),
				Action: run.Spec.Action,
			}
			running = runStillRunning(run)
		}
		results = append(results, layer{
			UID:              string(l.UID),
			Name:             l.Name,
			Namespace:        l.Namespace,
			Repository:       fmt.Sprintf("%s/%s", l.Spec.Repository.Namespace, l.Spec.Repository.Name),
			Branch:           l.Spec.Branch,
			Path:             l.Spec.Path,
			State:            a.getLayerState(l),
			RunCount:         len(l.Status.LatestRuns),
			LastRun:          runAPI,
			LastRunAt:        l.Status.LastRun.Date.Format(time.RFC3339),
			LastResult:       l.Status.LastResult,
			IsRunning:        running,
			IsPR:             a.isLayerPR(l),
			LatestRuns:       transformLatestRuns(l.Status.LatestRuns),
			ManualSyncStatus: utils.GetManualSyncStatus(l),
		})
	}
	return c.JSON(http.StatusOK, &layersResponse{
		Results: results,
	},
	)
}

func runStillRunning(run configv1alpha1.TerraformRun) bool {
	if run.Status.State != "Failed" && run.Status.State != "Succeeded" {
		return true
	}
	return false
}

func transformLatestRuns(runs []configv1alpha1.TerraformLayerRun) []Run {
	results := []Run{}
	for _, r := range runs {
		results = append(results, Run{
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

func (a *API) isLayerPR(layer configv1alpha1.TerraformLayer) bool {
	if len(layer.OwnerReferences) == 0 {
		return false
	}
	return layer.OwnerReferences[0].Kind == "TerraformPullRequest"
}

type LayerStatusCounts struct {
	Total       int `json:"total"`
	Ok          int `json:"ok"`
	OutOfSync   int `json:"outOfSync"`
	Error       int `json:"error"`
	Disabled    int `json:"disabled"`
	ApplyNeeded int `json:"applyNeeded"`
	PlanNeeded  int `json:"planNeeded"`
	Running     int `json:"running"`
}

func (a *API) LayersStatusHandler(c echo.Context) error {
	layers, runs, err := a.getLayersAndRuns()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("could not list terraform layers: %s", err))
	}

	counts := LayerStatusCounts{}
	for _, layer := range layers {
		if a.isLayerPR(layer) {
			continue // Skip PR layers for status counts
		}

		counts.Total++

		// Check if layer is currently running
		run, ok := runs[fmt.Sprintf("%s/%s", layer.Namespace, layer.Status.LastRun.Name)]
		isRunning := false
		if ok {
			isRunning = runStillRunning(run)
		}

		if isRunning {
			counts.Running++
		}

		// Get the raw state from the layer
		rawState := layer.Status.State
		layerState := a.getLayerState(layer)

		switch layerState {
		case "success":
			counts.Ok++
		case "warning":
			counts.OutOfSync++
			// Differentiate between ApplyNeeded and PlanNeeded for status bar
			if rawState == "ApplyNeeded" {
				counts.ApplyNeeded++
			} else if rawState == "PlanNeeded" {
				counts.PlanNeeded++
			}
		case "error":
			counts.Error++
		case "disabled":
			counts.Disabled++
		}
	}

	return c.JSON(http.StatusOK, counts)
}

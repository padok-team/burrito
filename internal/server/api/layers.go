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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
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
	AutoApply        bool                   `json:"autoApply"`
	OpenTofu         bool                   `json:"openTofu"`
	Terraform        bool                   `json:"terraform"`
	Terragrunt       bool                   `json:"terragrunt"`
}

type Run struct {
	Name    string `json:"id"`
	Commit  string `json:"commit"`
	Author  string `json:"author"`
	Message string `json:"message"`
	Date    string `json:"date"`
	Action  string `json:"action"`
}

type layersResponse struct {
	Results []layer `json:"results"`
}

func (a *API) getLayersRunsRepositories() ([]configv1alpha1.TerraformLayer, map[string]configv1alpha1.TerraformRun, map[string]configv1alpha1.TerraformRepository, error) {
	layers := &configv1alpha1.TerraformLayerList{}
	err := a.Client.List(context.Background(), layers)
	if err != nil {
		log.Errorf("could not list TerraformLayers: %s", err)
		return nil, nil, nil, err
	}
	runs := &configv1alpha1.TerraformRunList{}
	indexedRuns := map[string]configv1alpha1.TerraformRun{}
	err = a.Client.List(context.Background(), runs)
	if err != nil {
		log.Errorf("could not list TerraformRuns: %s", err)
		return nil, nil, nil, err
	}
	for _, run := range runs.Items {
		indexedRuns[fmt.Sprintf("%s/%s", run.Namespace, run.Name)] = run
	}
	repositories := &configv1alpha1.TerraformRepositoryList{}
	indexedRepositories := map[string]configv1alpha1.TerraformRepository{}
	err = a.Client.List(context.Background(), repositories)
	if err != nil {
		log.Errorf("could not list TerraformRepositories: %s", err)
		return nil, nil, nil, err
	}
	for _, repo := range repositories.Items {
		indexedRepositories[fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)] = repo
	}
	return layers.Items, indexedRuns, indexedRepositories, err
}

func (a *API) getLayer(namespace, name string) (*configv1alpha1.TerraformLayer, error) {
	layer := &configv1alpha1.TerraformLayer{}
	err := a.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, layer)
	if err != nil {
		log.Errorf("could not get TerraformLayer %s/%s: %s", namespace, name, err)
		return nil, err
	}
	return layer, nil
}

func (a *API) getRepository(namespace, name string) (*configv1alpha1.TerraformRepository, error) {
	repo := &configv1alpha1.TerraformRepository{}
	err := a.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, repo)
	if err != nil {
		log.Errorf("could not get TerraformRepository %s/%s: %s", namespace, name, err)
		return nil, err
	}
	return repo, nil
}

func (a *API) getRun(namespace, name string) (configv1alpha1.TerraformRun, bool) {
	run := configv1alpha1.TerraformRun{}
	err := a.Client.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, &run)
	if err != nil {
		log.Errorf("could not get TerraformRun %s/%s: %s", namespace, name, err)
		return run, false
	}
	return run, true
}

func (a *API) LayersHandler(c echo.Context) error {
	layers, runs, repositories, err := a.getLayersRunsRepositories()
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
				Name:    run.Name,
				Commit:  run.Status.Commit,
				Author:  run.Status.Author,
				Message: run.Status.Message,
				Date:    run.CreationTimestamp.Format(time.RFC3339),
				Action:  run.Spec.Action,
			}
			running = runStillRunning(run)
		}
		r, ok := repositories[fmt.Sprintf("%s/%s", l.Spec.Repository.Namespace, l.Spec.Repository.Name)]
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
			AutoApply:        configv1alpha1.GetAutoApplyEnabled(&r, &l),
			OpenTofu:         configv1alpha1.GetOpenTofuEnabled(&r, &l),
			Terraform:        configv1alpha1.GetTerraformEnabled(&r, &l),
			Terragrunt:       configv1alpha1.GetTerragruntEnabled(&r, &l),
		})
	}
	return c.JSON(http.StatusOK, &layersResponse{
		Results: results,
	},
	)
}

func (a *API) LayerHandler(c echo.Context) error {
	namespace := c.Param("namespace")
	layerName := c.Param("layer")
	layerObj, err := a.getLayer(namespace, layerName)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.String(http.StatusNotFound, fmt.Sprintf("terraform layer %s/%s not found", namespace, layerName))
		}
		log.Errorf("could not get TerraformLayer %s/%s: %s", namespace, layerName, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("could not get terraform layer %s/%s: %s", namespace, layerName, err))
	}
	run, ok := configv1alpha1.TerraformRun{}, false
	if layerObj.Status.LastRun.Name != "" {
		run, ok = a.getRun(namespace, layerObj.Status.LastRun.Name)
		if !ok {
			log.Warnf("could not find last run %s for layer %s/%s", layerObj.Status.LastRun.Name, namespace, layerName)
		}
	}
	runAPI := Run{}
	running := false
	if ok {
		runAPI = Run{
			Name:    run.Name,
			Commit:  run.Status.Commit,
			Author:  run.Status.Author,
			Message: run.Status.Message,
			Date:    run.CreationTimestamp.Format(time.RFC3339),
			Action:  run.Spec.Action,
		}
		running = runStillRunning(run)
	}
	repositoryObj, err := a.getRepository(layerObj.Spec.Repository.Namespace, layerObj.Spec.Repository.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			return c.String(http.StatusNotFound, fmt.Sprintf("terraform repository %s/%s not found", layerObj.Spec.Repository.Namespace, layerObj.Spec.Repository.Name))
		}
		log.Errorf("could not get TerraformRepository %s/%s: %s", layerObj.Spec.Repository.Namespace, layerObj.Spec.Repository.Name, err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("could not get terraform repository %s/%s: %s", layerObj.Spec.Repository.Namespace, layerObj.Spec.Repository.Name, err))
	}
	result := layer{
		UID:              string(layerObj.UID),
		Name:             layerObj.Name,
		Namespace:        layerObj.Namespace,
		Repository:       fmt.Sprintf("%s/%s", layerObj.Spec.Repository.Namespace, layerObj.Spec.Repository.Name),
		Branch:           layerObj.Spec.Branch,
		Path:             layerObj.Spec.Path,
		State:            a.getLayerState(*layerObj),
		RunCount:         len(layerObj.Status.LatestRuns),
		LastRun:          runAPI,
		LastRunAt:        layerObj.Status.LastRun.Date.Format(time.RFC3339),
		LastResult:       layerObj.Status.LastResult,
		IsRunning:        running,
		IsPR:             a.isLayerPR(*layerObj),
		LatestRuns:       transformLatestRuns(layerObj.Status.LatestRuns),
		ManualSyncStatus: utils.GetManualSyncStatus(*layerObj),
		AutoApply:        configv1alpha1.GetAutoApplyEnabled(repositoryObj, layerObj),
		OpenTofu:         configv1alpha1.GetOpenTofuEnabled(repositoryObj, layerObj),
		Terraform:        configv1alpha1.GetTerraformEnabled(repositoryObj, layerObj),
		Terragrunt:       configv1alpha1.GetTerragruntEnabled(repositoryObj, layerObj),
	}
	return c.JSON(http.StatusOK, result)
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
			Name:    r.Name,
			Commit:  r.Commit,
			Author:  r.Author,
			Message: r.Message,
			Date:    r.Date.Format(time.RFC3339),
			Action:  r.Action,
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

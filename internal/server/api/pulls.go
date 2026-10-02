package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type pullRequest struct {
	UID                  string  `json:"uid"`
	Name                 string  `json:"name"`
	Namespace            string  `json:"namespace"`
	ID                   string  `json:"id"`
	Repository           string  `json:"repository"`
	Branch               string  `json:"branch"`
	Base                 string  `json:"base"`
	State                string  `json:"state"`
	LastDiscoveredCommit string  `json:"lastDiscoveredCommit"`
	LastCommentedCommit  string  `json:"lastCommentedCommit"`
	Layers               []layer `json:"layers"`
}

type pullRequestsResponse struct {
	Results []pullRequest `json:"results"`
}

func (a *API) PullRequestsHandler(c echo.Context) error {
	prs := &configv1alpha1.TerraformPullRequestList{}
	if err := a.Client.List(context.Background(), prs); err != nil {
		log.Errorf("could not list TerraformPullRequests: %s", err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("could not list terraform pull requests: %s", err))
	}
	layers, runs, err := a.getLayersAndRuns()
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("could not list terraform layers or runs: %s", err))
	}

	// Ephemeral layers are owned by their pull request.
	layersByPR := map[string][]layer{}
	for _, l := range layers {
		for _, ref := range l.OwnerReferences {
			if ref.Kind == "TerraformPullRequest" {
				key := fmt.Sprintf("%s/%s", l.Namespace, ref.Name)
				layersByPR[key] = append(layersByPR[key], a.buildLayer(l, runs))
				break
			}
		}
	}

	results := []pullRequest{}
	for _, pr := range prs.Items {
		prLayers := layersByPR[fmt.Sprintf("%s/%s", pr.Namespace, pr.Name)]
		if prLayers == nil {
			prLayers = []layer{}
		}
		results = append(results, pullRequest{
			UID:                  string(pr.UID),
			Name:                 pr.Name,
			Namespace:            pr.Namespace,
			ID:                   pr.Spec.ID,
			Repository:           fmt.Sprintf("%s/%s", pr.Spec.Repository.Namespace, pr.Spec.Repository.Name),
			Branch:               pr.Spec.Branch,
			Base:                 pr.Spec.Base,
			State:                pr.Status.State,
			LastDiscoveredCommit: pr.Status.LastDiscoveredCommit,
			LastCommentedCommit:  pr.Status.LastCommentedCommit,
			Layers:               prLayers,
		})
	}
	return c.JSON(http.StatusOK, &pullRequestsResponse{Results: results})
}

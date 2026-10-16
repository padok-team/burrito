package api

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	urlutils "github.com/padok-team/burrito/internal/utils/url"
	log "github.com/sirupsen/logrus"
)

type pullRequest struct {
	UID                  string  `json:"uid"`
	Name                 string  `json:"name"`
	Namespace            string  `json:"namespace"`
	ID                   string  `json:"id"`
	Repository           string  `json:"repository"`
	URL                  string  `json:"url"`
	Branch               string  `json:"branch"`
	Base                 string  `json:"base"`
	State                string  `json:"state"`
	LastDiscoveredCommit string  `json:"lastDiscoveredCommit"`
	LastCommentedCommit  string  `json:"lastCommentedCommit"`
	Title                string  `json:"title"`
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
	repositories := &configv1alpha1.TerraformRepositoryList{}
	if err := a.Client.List(context.Background(), repositories); err != nil {
		log.Errorf("could not list TerraformRepositories: %s", err)
		return c.String(http.StatusInternalServerError, fmt.Sprintf("could not list terraform repositories: %s", err))
	}
	repositoryURLs := map[string]string{}
	for _, r := range repositories.Items {
		repositoryURLs[fmt.Sprintf("%s/%s", r.Namespace, r.Name)] = r.Spec.Repository.Url
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
		repository := fmt.Sprintf("%s/%s", pr.Spec.Repository.Namespace, pr.Spec.Repository.Name)
		results = append(results, pullRequest{
			UID:                  string(pr.UID),
			Name:                 pr.Name,
			Namespace:            pr.Namespace,
			ID:                   pr.Spec.ID,
			Repository:           repository,
			URL:                  pullRequestURL(repositoryURLs[repository], pr.Spec.ID),
			Branch:               pr.Spec.Branch,
			Base:                 pr.Spec.Base,
			State:                pr.Status.State,
			LastDiscoveredCommit: pr.Status.LastDiscoveredCommit,
			LastCommentedCommit:  pr.Status.LastCommentedCommit,
			Title:                pr.Status.Title,
			Layers:               prLayers,
		})
	}
	return c.JSON(http.StatusOK, &pullRequestsResponse{Results: results})
}

// pullRequestURL builds the web URL of a pull request from the repository URL.
// The server has no access to the provider credentials, so GitLab is detected
// from the host name; any other host is assumed to be GitHub-like.
func pullRequestURL(repositoryURL, id string) string {
	if repositoryURL == "" || id == "" {
		return ""
	}
	base := urlutils.NormalizeUrl(repositoryURL)
	parsed, err := url.Parse(base)
	if err != nil || parsed.Host == "" {
		return ""
	}
	if strings.Contains(strings.ToLower(parsed.Host), "gitlab") {
		return fmt.Sprintf("%s/-/merge_requests/%s", base, id)
	}
	return fmt.Sprintf("%s/pull/%s", base, id)
}

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type layer struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Path       string `json:"path"`
	LastResult string `json:"last_result"`
	IsRunning  bool   `json:"is_running"`
	IsPR       bool   `json:"is_pr"`
}

type layersResponse struct {
	Results []layer `json:"results"`
}

func (a *API) LayersHandler(c echo.Context) error {
	layers := &configv1alpha1.TerraformLayerList{}
	err := a.Client.List(context.Background(), layers)
	if err != nil {
		log.Errorf("could not list terraform repositories: %s", err)
		return c.String(http.StatusInternalServerError, "could not list terraform repositories")
	}
	results := []layer{}
	for _, l := range layers.Items {
		results = append(results, layer{
			Name:       l.Name,
			Namespace:  l.Namespace,
			Repository: fmt.Sprintf("%s/%s", l.Spec.Repository.Namespace, l.Spec.Repository.Name),
			Branch:     l.Spec.Branch,
			Path:       l.Spec.Path,
			LastResult: l.Status.LastResult,
			IsRunning:  a.isLayerRunning(l),
			IsPR:       a.isLayerPR(l),
		})
	}
	return c.JSON(http.StatusOK, &layersResponse{
		Results: results,
	},
	)
}

func (a *API) isLayerRunning(layer configv1alpha1.TerraformLayer) bool {
	runs := &configv1alpha1.TerraformRunList{}
	requirement, _ := labels.NewRequirement("burrito/managed-by", selection.Equals, []string{layer.Name})
	selector := labels.NewSelector().Add(*requirement)
	err := a.Client.List(context.Background(), runs, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		log.Errorf("could not list terraform runs, returning false: %s", err)
		return false
	}
	err = a.Client.List(context.Background(), runs)
	if err != nil {
		log.Errorf("could not list terraform runs, returning false: %s", err)
		return false
	}
	for _, r := range runs.Items {
		if r.Status.State == "Running" {
			return true
		}
	}
	return false
}

func (a *API) isLayerPR(layer configv1alpha1.TerraformLayer) bool {
	return layer.OwnerReferences[0].Kind == "TerraformPullRequest"
}

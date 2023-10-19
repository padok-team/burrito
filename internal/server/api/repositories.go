package api

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type repository struct {
	Name string `json:"name"`
}

type repositoriesResponse struct {
	Results []repository `json:"results"`
}

func (a *API) RepositoriesHandler(c echo.Context) error {
	repositories := &configv1alpha1.TerraformRepositoryList{}
	err := a.Client.List(context.Background(), repositories)
	if err != nil {
		log.Errorf("could not list terraform repositories: %s", err)
		return c.String(http.StatusInternalServerError, "could not list terraform repositories")
	}
	results := []repository{}
	for _, r := range repositories.Items {
		results = append(results, repository{
			Name: r.Name,
		})
	}
	return c.JSON(http.StatusOK, &repositoriesResponse{
		Results: results,
	},
	)
}

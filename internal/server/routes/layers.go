package routes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func hashLayerId(l configv1alpha1.TerraformLayer) string {
	toHash := l.Name + l.Namespace + l.APIVersion + l.Kind
	hash := sha256.Sum256([]byte(toHash))
	return hex.EncodeToString(hash[:])
}

func getRepoForLayer(l *configv1alpha1.TerraformLayer, c client.Client) (*configv1alpha1.TerraformRepository, error) {
	repo := &configv1alpha1.TerraformRepository{}
	err := c.Get(context.Background(), client.ObjectKey{
		Namespace: l.Spec.Repository.Namespace,
		Name:      l.Spec.Repository.Name,
	}, repo)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (lc *LayerClient) GetAllLayersHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Infof("request received on /layers")

		layers := &configv1alpha1.TerraformLayerList{}
		err := lc.client.List(context.Background(), layers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		layerResponse := []Layer{}
		for _, layer := range layers.Items {
			// Get the repo associated to the layer
			repo, err := getRepoForLayer(&layer, lc.client)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Construct the object
			layerResponse = append(layerResponse, Layer{
				Id:        hashLayerId(layer),
				Name:      layer.Name,
				Namespace: layer.Namespace,
				RepoUrl:   repo.Spec.Repository.Url,
				Branch:    layer.Spec.Branch,
				Path:      layer.Spec.Path,
				Status:    layer.Status.State,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		jsonResponse, err := json.Marshal(layerResponse)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(jsonResponse)
	}
}

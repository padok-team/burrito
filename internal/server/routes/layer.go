package routes

import (
	"context"
	"encoding/json"
	"net/http"

	tfjson "github.com/hashicorp/terraform-json"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/storage"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getResourcesFromPlan(s storage.Storage, l *configv1alpha1.TerraformLayer) ([]Resource, error) {
	planKey := storage.GenerateKey(storage.LastPlannedArtifactJson, l)
	planBytes, err := s.Get(planKey)
	if err != nil {
		log.Errorf("could not get plan artifact %s: %s", planKey, err)
		return nil, err
	}
	plan := &tfjson.Plan{}
	err = json.Unmarshal(planBytes, plan)
	if err != nil {
		log.Errorf("error parsing terraform json plan: %s", err)
		return nil, err
	}

	resources := []Resource{}

	for _, r := range plan.ResourceChanges {
		resources = append(resources, Resource{
			Address: r.Address,
			Type:    r.Type,
			Status:  string(r.Change.Actions[0]),
		})
	}

	// TODO: avoid duplicates between ResourceChanges and ResourceDrift
	for _, r := range plan.ResourceDrift {
		resources = append(resources, Resource{
			Address: r.Address,
			Type:    r.Type,
			Status:  string(r.Change.Actions[0]),
		})
	}

	return resources, nil
}

// Call with /layer?name=XXXXX&namespace=YYYYY
func (lc *LayerClient) GetSpecificLayerHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		layerName, present := query["name"] //name=["XXXXXX"]
		if !present || len(layerName) == 0 {
			log.Errorf("layer name query param not present")
			http.Error(w, "name query parameter not present", http.StatusBadRequest)
			return
		}
		layerNamespace, present := query["namespace"] //namespace=["XXXXXX"]
		if !present || len(layerNamespace) == 0 {
			log.Errorf("layer NS query param not present")
			http.Error(w, "namespace query parameter not present", http.StatusBadRequest)
			return
		}

		log.Infof("getting layer %s/%s", layerNamespace[0], layerName[0])

		layer := &configv1alpha1.TerraformLayer{}
		err := lc.client.Get(context.Background(), client.ObjectKey{
			Namespace: layerNamespace[0],
			Name:      layerName[0],
		}, layer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		repo, err := getRepoForLayer(layer, lc.client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resources, err := getResourcesFromPlan(lc.storage, layer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		layerResponse := Layer{
			Id:                 hashLayerId(*layer),
			Name:               layer.Name,
			Namespace:          layer.Namespace,
			RepoUrl:            repo.Spec.Repository.Url,
			Branch:             layer.Spec.Branch,
			Path:               layer.Spec.Path,
			Status:             layer.Status.State,
			LastPlanCommit:     layer.Annotations[annotations.LastPlanCommit],
			LastApplyCommit:    layer.Annotations[annotations.LastApplyCommit],
			LastRelevantCommit: layer.Annotations[annotations.LastRelevantCommit],
			Resources:          resources,
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

func (lc *LayerClient) ForceApplyHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		layerName, present := query["name"] //name=["XXXXXX"]
		if !present || len(layerName) == 0 {
			log.Errorf("layer name query param not present")
			http.Error(w, "name query parameter not present", http.StatusBadRequest)
			return
		}
		layerNamespace, present := query["namespace"] //namespace=["XXXXXX"]
		if !present || len(layerNamespace) == 0 {
			log.Errorf("layer NS query param not present")
			http.Error(w, "namespace query parameter not present", http.StatusBadRequest)
			return
		}

		log.Infof("force apply layer %s/%s", layerNamespace[0], layerName[0])

		layer := &configv1alpha1.TerraformLayer{}
		err := lc.client.Get(context.Background(), client.ObjectKey{
			Namespace: layerNamespace[0],
			Name:      layerName[0],
		}, layer)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ann := map[string]string{}
		ann[annotations.ForceApply] = "1"
		err = annotations.Add(context.Background(), lc.client, layer, ann)
		if err != nil {
			log.Errorf("could not update terraform layer annotations to force apply: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Else, it worked !
	}
}

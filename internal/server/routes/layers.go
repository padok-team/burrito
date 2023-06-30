package routes

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func hashLayerId(l configv1alpha1.TerraformLayer) string {
	toHash := l.Name + l.Namespace + l.APIVersion + l.Kind
	hash := sha256.Sum256([]byte(toHash))
	return hex.EncodeToString(hash[:])
}

type Layer struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	RepoUrl string `json:"repoUrl"`
	Branch  string `json:"branch"`
	Path    string `json:"path"`
	Status  string `json:"status"`
}

func GetAllLayers(w http.ResponseWriter, r *http.Request) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	layers := &configv1alpha1.TerraformLayerList{}
	err = cl.List(context.Background(), layers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	layerResponse := []Layer{}
	for _, layer := range layers.Items {
		// Get the repo associated to the layer
		repo := &configv1alpha1.TerraformRepository{}
		err = cl.Get(context.Background(), client.ObjectKey{
			Namespace: layer.Spec.Repository.Namespace,
			Name:      layer.Spec.Repository.Name,
		}, repo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Construct the object
		layerResponse = append(layerResponse, Layer{
			Id:      hashLayerId(layer),
			Name:    layer.Name,
			RepoUrl: repo.Spec.Repository.Url,
			Branch:  layer.Spec.Branch,
			Path:    layer.Spec.Path,
			Status:  layer.Status.State,
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

package terraformrepository

import (
	"context"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

// Returns the list of layers that are managed by this repository
func (r *Reconciler) retrieveManagedLayers(ctx context.Context, repository *configv1alpha1.TerraformRepository) ([]configv1alpha1.TerraformLayer, error) {
	// get all layers that depends on the repository (layer.spec.repository.name == repository.name)
	layers := &configv1alpha1.TerraformLayerList{}
	if err := r.List(ctx, layers); err != nil {
		return nil, err
	}
	managedLayers := []configv1alpha1.TerraformLayer{}
	for _, layer := range layers.Items {
		if layer.Spec.Repository.Name == repository.Name {
			managedLayers = append(managedLayers, layer)
		}
	}
	return managedLayers, nil
}

// Returns a list of all refs (branches and tags) among a list of layers from the same repository (duplicated allowed)
func retrieveAllLayerRefs(layers []configv1alpha1.TerraformLayer) []string {
	refs := []string{}
	for _, layer := range layers {
		refs = append(refs, layer.Spec.Branch)
	}
	return refs
}

// Returns a list of all layers referencing a specific ref
func retrieveLayersForRef(ref string, layers []configv1alpha1.TerraformLayer) []configv1alpha1.TerraformLayer {
	result := []configv1alpha1.TerraformLayer{}
	for _, layer := range layers {
		if layer.Spec.Branch == ref {
			result = append(result, layer)
		}
	}
	return result
}

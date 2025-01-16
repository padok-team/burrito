package terraformrepository

import (
	"context"
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	gitCommon "github.com/padok-team/burrito/internal/utils/gitprovider/common"
)

// getRemoteRevision gets the latest revision (commit sha) for a given ref from the remote repository
func (r *Reconciler) getRemoteRevision(repository *configv1alpha1.TerraformRepository, ref string) (string, error) {
	// Get the appropriate provider for the repository
	provider, exists := r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)]
	if !exists {
		return "", fmt.Errorf("provider not found for repository %s/%s", repository.Namespace, repository.Name)
	}
	rev, err := provider.GetLatestRevisionForRef(repository, ref)
	if err != nil {
		return "", fmt.Errorf("failed to get latest revision for ref %s: %v", ref, err)
	}
	return rev, nil
}

// getRevisionBundle gets the git bundle for a given revision from the remote repository
func (r *Reconciler) getRevisionBundle(repository *configv1alpha1.TerraformRepository, ref string, revision string) ([]byte, error) {
	provider, exists := r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)]
	if !exists {
		return nil, fmt.Errorf("provider not found for repository %s/%s", repository.Namespace, repository.Name)
	}
	auth, err := provider.GetGitAuth()
	if err != nil {
		return nil, fmt.Errorf("failed to get git auth for repository %s/%s: %v", repository.Namespace, repository.Name, err)
	}
	bundle, err := gitCommon.GetGitBundle(repository, ref, revision, auth)
	if err != nil {
		return nil, fmt.Errorf("failed to get revision bundle for ref %s: %v", ref, err)
	}
	return bundle, nil
}

// retrieveLayerBranches returns the list of refs (branches and tags) that are managed by burrito for a specific repository
func (r *Reconciler) retrieveLayerBranches(ctx context.Context, repository *configv1alpha1.TerraformRepository) ([]string, error) {
	// get all layers that depends on the repository (layer.spec.repository.name == repository.name)
	layers := &configv1alpha1.TerraformLayerList{}
	if err := r.List(ctx, layers); err != nil {
		return nil, err
	}
	refs := []string{}
	for _, layer := range layers.Items {
		if layer.Spec.Repository.Name == repository.Name {
			refs = append(refs, layer.Spec.Branch)
		}
	}
	return refs, nil
}

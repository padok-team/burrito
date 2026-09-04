package terraformrepository

import (
	"context"

	"github.com/google/go-cmp/cmp"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/padok-team/burrito/internal/repository"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
)

// syncPullRequests keeps TerraformPullRequest resources for repositoryObj in sync with the
// remote open pull/merge requests: it creates or updates the desired ones and deletes the
// ones that are no longer open remotely.
func (r *Reconciler) syncPullRequests(ctx context.Context, repositoryObj *configv1alpha1.TerraformRepository) error {
	provider, err := r.getAPIProvider(repositoryObj)
	if err != nil {
		return err
	}

	remotePullRequests, err := provider.ListPullRequests(repositoryObj)
	if err != nil {
		return err
	}

	existingPullRequests := &configv1alpha1.TerraformPullRequestList{}
	if err := r.Client.List(ctx, existingPullRequests, client.InNamespace(repositoryObj.Namespace)); err != nil {
		return err
	}

	desiredPullRequests := map[string]configv1alpha1.TerraformPullRequest{}
	for _, pullRequest := range remotePullRequests {
		desiredPullRequests[pullRequest.Name] = pullRequest
		if err := r.applyDesiredPullRequest(ctx, &pullRequest); err != nil {
			return err
		}
	}

	for i := range existingPullRequests.Items {
		current := &existingPullRequests.Items[i]
		if !sameRepository(current, repositoryObj) {
			continue
		}
		if _, ok := desiredPullRequests[current.Name]; ok {
			continue
		}
		if err := r.deleteRemotePullRequest(ctx, current); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) getAPIProvider(repositoryObj *configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
	if r.APIProviderFactory != nil {
		return r.APIProviderFactory(repositoryObj)
	}
	return repository.GetAPIProviderFromRepository(r.Credentials, repositoryObj)
}

func (r *Reconciler) applyDesiredPullRequest(ctx context.Context, desired *configv1alpha1.TerraformPullRequest) error {
	current := &configv1alpha1.TerraformPullRequest{}
	err := r.Client.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if apierrors.IsNotFound(err) {
		err = r.Client.Create(ctx, desired.DeepCopy())
		if apierrors.IsAlreadyExists(err) {
			current = &configv1alpha1.TerraformPullRequest{}
			err = r.Client.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
			if err != nil {
				return err
			}
			current.Spec = desired.Spec
			mergePollingAnnotations(current, desired)
			return r.Client.Update(ctx, current)
		}
		return err
	}
	if err != nil {
		return err
	}
	if cmp.Diff(current.Spec, desired.Spec) == "" && pollingAnnotationsEqual(current, desired) {
		return nil
	}
	current.Spec = desired.Spec
	mergePollingAnnotations(current, desired)
	return r.Client.Update(ctx, current)
}

func pollingAnnotationsEqual(current *configv1alpha1.TerraformPullRequest, desired *configv1alpha1.TerraformPullRequest) bool {
	for key, desiredValue := range desired.Annotations {
		if current.Annotations[key] != desiredValue {
			return false
		}
	}
	return true
}

func mergePollingAnnotations(current *configv1alpha1.TerraformPullRequest, desired *configv1alpha1.TerraformPullRequest) {
	// Polling owns the remote commit annotation, but other annotations can come
	// from webhooks or users and must survive a periodic remote sync.
	if current.Annotations == nil {
		current.Annotations = map[string]string{}
	}
	for key, value := range desired.Annotations {
		current.Annotations[key] = value
	}
}

func (r *Reconciler) deleteRemotePullRequest(ctx context.Context, pullRequest *configv1alpha1.TerraformPullRequest) error {
	err := r.Client.Delete(ctx, pullRequest)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func sameRepository(pullRequest *configv1alpha1.TerraformPullRequest, repositoryObj *configv1alpha1.TerraformRepository) bool {
	return pullRequest.Spec.Repository.Name == repositoryObj.Name && pullRequest.Spec.Repository.Namespace == repositoryObj.Namespace
}

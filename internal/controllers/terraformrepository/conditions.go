package terraformrepository

import (
	"context"
	"fmt"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsLastSyncTooOld checks if the last sync was too long ago
func (r *Reconciler) IsLastSyncTooOld(repository *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsLastSyncTooOld",
		ObservedGeneration: repository.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	// Get last sync date from annotation
	lastSyncStr, exists := repository.Annotations[annotations.LastSyncDate]

	// If no sync has ever happened, we need one
	if !exists {
		condition.Reason = "NoSyncYet"
		condition.Message = "Repository has never been synced"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	lastSync, err := time.Parse(time.UnixDate, lastSyncStr)
	if err != nil {
		condition.Reason = "InvalidSyncDate"
		condition.Message = fmt.Sprintf("Invalid last sync date format: %v", err)
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	nextSyncTime := lastSync.Add(r.Config.Controller.Timers.RepositorySync)
	now := time.Now()

	if nextSyncTime.Before(now) {
		condition.Reason = "SyncTooOld"
		condition.Message = fmt.Sprintf("Last sync was more than %s ago", r.Config.Controller.Timers.RepositorySync)
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	condition.Reason = "SyncRecent"
	condition.Message = fmt.Sprintf("Last sync was less than %s ago", r.Config.Controller.Timers.RepositorySync)
	condition.Status = metav1.ConditionFalse
	return condition, false
}

// HasLastSyncFailed checks if the last sync failed
// A sync can fail if at least one of the refs managed by burrito could not be synced with the datastore
func (r *Reconciler) HasLastSyncFailed(repository *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasLastSyncFailed",
		ObservedGeneration: repository.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	lastSyncStatus, exists := repository.Annotations[annotations.LastSyncStatus]
	if !exists {
		condition.Reason = "NoSyncYet"
		condition.Message = "Repository has never been synced"
		condition.Status = metav1.ConditionTrue
		return condition, false
	}

	if lastSyncStatus == annotations.SyncStatusFailed {
		condition.Reason = "SyncFailed"
		condition.Message = "Last sync failed"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	condition.Reason = "SyncSucceeded"
	condition.Message = "Last sync succeeded"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

// AreRemoteRevisionsDifferent checks if any remote revision is different from its stored version
// Returns true if at least one ref has a different latest revision in datastore than in remote repository
func (r *Reconciler) AreRemoteRevisionsDifferent(ctx context.Context, repository *configv1alpha1.TerraformRepository, refs map[string]bool) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "AreRemoteRevisionsDifferent",
		ObservedGeneration: repository.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	for ref := range refs {
		// Get latest revision from remote repository
		remoteRevision, err := r.getRemoteRevision(ctx, repository, ref)
		if err != nil {
			condition.Reason = "ErrorGettingRemoteRevision"
			condition.Message = fmt.Sprintf("Failed to get remote revision for ref %s: %v", ref, err)
			condition.Status = metav1.ConditionUnknown
			return condition, false
		}

		// Get latest revision from datastore
		storedRevision, err := r.Datastore.GetLatestRevision(repository.Namespace, repository.Name, ref)
		if err != nil {
			condition.Reason = "ErrorGettingStoredRevision"
			condition.Message = fmt.Sprintf("Failed to get stored revision for ref %s: %v", ref, err)
			condition.Status = metav1.ConditionUnknown
			return condition, false
		}

		if remoteRevision != storedRevision {
			condition.Reason = "RevisionsDiffer"
			condition.Message = fmt.Sprintf("Remote revision is different from stored revision for ref %s", ref)
			condition.Status = metav1.ConditionTrue
			return condition, true
		}
	}

	condition.Reason = "RevisionsMatch"
	condition.Message = "All remote revisions match their stored versions"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

// getRemoteRevision gets the latest revision for a given ref from the remote repository
func (r *Reconciler) getRemoteRevision(ctx context.Context, repository *configv1alpha1.TerraformRepository, ref string) (string, error) {
	// Get the appropriate provider for the repository
	// provider, exists := r.Providers[repository.Spec.Repository.Provider]
	// if !exists {
	// 	return "", fmt.Errorf("provider %s not found", repository.Spec.Repository.Provider)
	// }

	// TODO: Implement provider-specific logic to get the latest revision
	// This might involve:
	// 1. Using the provider's API to get the latest commit for the ref
	// 2. Cloning the repository and getting the ref's HEAD
	// For now, return an error
	return "", fmt.Errorf("getRemoteRevision not implemented")
}

// getRevisionBundle gets the git bundle for a given revision from the remote repository
func (r *Reconciler) getRevisionBundle(ctx context.Context, repository *configv1alpha1.TerraformRepository, ref string, revision string) ([]byte, error) {
	// TODO: Implement provider-specific logic to get the git bundle
	return nil, fmt.Errorf("getRevisionBundle not implemented")
}

package terraformrepository

import (
	"context"
	"fmt"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Add newly found branches in the repository's branch state object
func mergeBranchesWithBranchState(found []string, branchStates []configv1alpha1.BranchState) []configv1alpha1.BranchState {
	for _, branch := range found {
		if _, ok := configv1alpha1.GetBranchState(branch, branchStates); !ok {
			branchStates = append(branchStates, configv1alpha1.BranchState{
				Name: branch,
			})
		}
	}
	return branchStates
}

// IsLastSyncTooOld checks if the last sync was too long ago for at least one of the branches tracked by the repository
func (r *Reconciler) IsLastSyncTooOld(repo *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsLastSyncTooOld",
		ObservedGeneration: repo.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	layerBranches, err := r.retrieveLayerBranches(context.Background(), repo)
	if err != nil {
		condition.Reason = "ErrorListingLayers"
		condition.Message = err.Error()
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	if len(layerBranches) == 0 {
		condition.Reason = "NoBranches"
		condition.Message = "No branches managed by this repository, no layers found"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}

	branchStates := repo.Status.Branches
	branchStates = mergeBranchesWithBranchState(layerBranches, branchStates)

	for _, branch := range branchStates {
		// If no sync has ever happened for this branch, we need one
		if branch.LastSyncDate == "" {
			condition.Reason = "NoSyncYet"
			condition.Message = fmt.Sprintf("Repository has never been synced for branch %s", branch.Name)
			condition.Status = metav1.ConditionTrue
			return condition, true
		}

		lastSync, err := time.Parse(time.UnixDate, branch.LastSyncDate)
		if err != nil {
			condition.Reason = "InvalidSyncDate"
			condition.Message = fmt.Sprintf("Invalid last sync date format for branch %s: %v", branch.Name, err)
			condition.Status = metav1.ConditionTrue
			return condition, true
		}

		nextSyncTime := lastSync.Add(r.Config.Controller.Timers.RepositorySync)
		now := time.Now()

		if nextSyncTime.Before(now) {
			condition.Reason = "SyncTooOld"
			condition.Message = fmt.Sprintf("Last sync for %s was more than %s ago", branch.Name, r.Config.Controller.Timers.RepositorySync)
			condition.Status = metav1.ConditionTrue
			return condition, true
		}
	}

	condition.Reason = "SyncRecent"
	condition.Message = fmt.Sprintf("Last sync for all branches was less than %s ago", r.Config.Controller.Timers.RepositorySync)
	condition.Status = metav1.ConditionFalse
	return condition, false
}

// HasLastSyncFailed checks if the last sync failed
// A sync can fail if at least one of the refs managed by burrito could not be synced with the datastore
func (r *Reconciler) HasLastSyncFailed(repo *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasLastSyncFailed",
		ObservedGeneration: repo.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	layerBranches, err := r.retrieveLayerBranches(context.Background(), repo)
	if err != nil {
		condition.Reason = "ErrorListingLayers"
		condition.Message = err.Error()
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	if len(layerBranches) == 0 {
		condition.Reason = "NoBranches"
		condition.Message = "No branches managed by this repository, no layers found"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}

	branchStates := repo.Status.Branches
	branchStates = mergeBranchesWithBranchState(layerBranches, branchStates)

	for _, branch := range branchStates {
		if branch.LastSyncStatus == "" {
			condition.Reason = "NoSyncYet"
			condition.Message = fmt.Sprintf("Repository has never been synced on branch %s yet", branch.Name)
			condition.Status = metav1.ConditionTrue
			return condition, false
		}

		if branch.LastSyncStatus == SyncStatusFailed {
			condition.Reason = "SyncFailed"
			condition.Message = fmt.Sprintf("Last sync failed for branch %s", branch.Name)
			condition.Status = metav1.ConditionTrue
			return condition, true
		}
	}

	condition.Reason = "SyncSucceeded"
	condition.Message = "Last sync succeeded for all branches"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

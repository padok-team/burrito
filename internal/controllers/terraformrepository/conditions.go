package terraformrepository

import (
	"fmt"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsLastSyncTooOld checks if the last sync was too long ago
func (r *Reconciler) IsLastSyncTooOld(repo *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsLastSyncTooOld",
		ObservedGeneration: repo.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	// If no sync has ever happened, we need one
	if repo.Status.LastSyncDate == "" {
		condition.Reason = "NoSyncYet"
		condition.Message = "Repository has never been synced"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	lastSync, err := time.Parse(time.UnixDate, repo.Status.LastSyncDate)
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
func (r *Reconciler) HasLastSyncFailed(repo *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasLastSyncFailed",
		ObservedGeneration: repo.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}

	if repo.Status.LastSyncStatus == "" {
		condition.Reason = "NoSyncYet"
		condition.Message = "Repository has never been synced"
		condition.Status = metav1.ConditionTrue
		return condition, false
	}

	if repo.Status.LastSyncStatus == SyncStatusFailed {
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

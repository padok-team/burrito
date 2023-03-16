package terraformlayer

import (
	"fmt"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) IsPlanArtifactUpToDate(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsPlanArtifactUpToDate",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	value, ok := t.Annotations[annotations.LastPlanDate]
	if !ok {
		condition.Reason = "NoPlanHasRunYet"
		condition.Message = "No plan has run on this layer yet"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	lastPlanDate, err := time.Parse(time.UnixDate, value)
	if err != nil {
		condition.Reason = "ParseError"
		condition.Message = "Could not parse time from annotation"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	nextPlanDate := lastPlanDate.Add(r.Config.Controller.Timers.DriftDetection)
	now := time.Now()
	if nextPlanDate.After(now) {
		condition.Reason = "PlanIsRecent"
		condition.Message = fmt.Sprintf("The plan has been made less than %s ago.", r.Config.Controller.Timers.DriftDetection)
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "PlanIsTooOld"
	condition.Message = fmt.Sprintf("The plan has been made more than %s ago.", r.Config.Controller.Timers.DriftDetection)
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) IsLastRelevantCommitPlanned(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsLastRelevantCommitPlanned",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	lastPlannedCommit, ok := t.Annotations[annotations.LastPlanCommit]
	if !ok {
		condition.Reason = "NoPlanHasRunYet"
		condition.Message = "No plan has run on this layer yet"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	lastBranchCommit, ok := t.Annotations[annotations.LastBranchCommit]
	if !ok {
		condition.Reason = "NoCommitReceived"
		condition.Message = "No commit has been received from webhook"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	lastRelevantCommit, ok := t.Annotations[annotations.LastRelevantCommit]
	if !ok {
		condition.Reason = "NoCommitReceived"
		condition.Message = "No commit has been received from webhook"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if lastBranchCommit != lastRelevantCommit {
		condition.Reason = "CommitAlreadyHadnled"
		condition.Message = "The last relevant commit should already have been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if lastPlannedCommit == lastBranchCommit {
		condition.Reason = "LastRelevantCommitPlanned"
		condition.Message = "The last relevant commit has already been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "LastRelevantCommitNotPlanned"
	condition.Message = "The last received relevant commit has not been planned yet"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) IsApplyUpToDate(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsApplyUpToDate",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	planHash, ok := t.Annotations[annotations.LastPlanSum]
	if !ok {
		condition.Reason = "NoPlanHasRunYet"
		condition.Message = "No plan has run on this layer yet"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	applyHash, ok := t.Annotations[annotations.LastApplySum]
	if !ok {
		condition.Reason = "NoApplyHasRan"
		condition.Message = "Apply has not ran yet but a plan is available, launching apply"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if applyHash != planHash {
		condition.Reason = "NewPlanAvailable"
		condition.Message = "Apply will run."
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Reason = "ApplyUpToDate"
	condition.Message = "Last planned artifact is the same as the last applied one"
	condition.Status = metav1.ConditionTrue
	return condition, true
}

func (r *Reconciler) HasFailed(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasFailed",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	result, ok := t.Annotations[annotations.Failure]
	if !ok {
		condition.Reason = "NoRunYet"
		condition.Message = "Terraform has not ran yet"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if string(result) == "0" {
		condition.Reason = "RunExitedGracefully"
		condition.Message = "Last run exited gracefully"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Status = metav1.ConditionTrue
	condition.Reason = "TerraformRunFailure"
	condition.Message = "Terraform has failed, look at the runner logs"
	return condition, true
}

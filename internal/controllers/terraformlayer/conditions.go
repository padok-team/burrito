package terraformlayer

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
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
	planHash, ok := t.Annotations[annotations.LastPlanSum]
	if !ok || planHash == "" {
		condition.Reason = "LastPlanFailed"
		condition.Message = "Last plan run has failed"
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
	now := r.Clock.Now()
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
	if lastPlannedCommit == lastBranchCommit || lastPlannedCommit == lastRelevantCommit {
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
	if applyHash == "" {
		condition.Reason = "LastApplyFailed"
		condition.Message = "Last apply run has failed."
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

func (r *Reconciler) IsInFailureGracePeriod(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsInFailureGracePeriod",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if failure, ok := t.Annotations[annotations.Failure]; !ok || failure == "0" {
		condition.Reason = "NoFailureYet"
		condition.Message = "No failure has been detected yet"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	lastFailureDate, err := GetLastActionTime(r, t)
	if err != nil {
		condition.Reason = "CouldNotGetLastActionTime"
		condition.Message = "Could not get last action time from layer annotations"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}

	nextFailure := lastFailureDate.Add(GetLayerExponentialBackOffTime(r.Config.Controller.Timers.FailureGracePeriod, t))
	now := r.Clock.Now()
	if nextFailure.After(now) {
		condition.Reason = "InFailureGracePeriod"
		condition.Message = fmt.Sprintf("The failure grace period is still active (until %s).", nextFailure)
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "FailureGracePeriodOver"
	condition.Message = fmt.Sprintf("The failure grace period is over (since %s).", now.Sub(nextFailure))
	condition.Status = metav1.ConditionFalse
	return condition, false
}

// func (r *Reconciler) HasFailed(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
// 	condition := metav1.Condition{
// 		Type:               "HasFailed",
// 		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
// 		Status:             metav1.ConditionUnknown,
// 		LastTransitionTime: metav1.NewTime(time.Now()),
// 	}
// 	result, ok := t.Annotations[annotations.Failure]
// 	if !ok {
// 		condition.Reason = "NoRunYet"
// 		condition.Message = "Terraform has not ran yet"
// 		condition.Status = metav1.ConditionFalse
// 		return condition, false
// 	}
// 	if string(result) == "0" {
// 		condition.Reason = "RunExitedGracefully"
// 		condition.Message = "Last run exited gracefully"
// 		condition.Status = metav1.ConditionFalse
// 		return condition, false
// 	}
// 	condition.Status = metav1.ConditionTrue
// 	condition.Reason = "TerraformRunFailure"
// 	condition.Message = "Terraform has failed, look at the runner logs"
// 	return condition, true
// }

func GetLastActionTime(r *Reconciler, layer *configv1alpha1.TerraformLayer) (time.Time, error) {
	var lastActionTime time.Time
	lastPlanTimeAnnotation, ok := layer.Annotations[annotations.LastPlanDate]
	if !ok {
		return r.Clock.Now(), errors.New("never ran a plan on this layer")
	}
	lastActionTime, err := time.Parse(time.UnixDate, lastPlanTimeAnnotation)
	if err != nil {
		return r.Clock.Now(), err
	}

	lastApplyTimeAnnotation, ok := layer.Annotations[annotations.LastApplyDate]
	if !ok {
		return lastActionTime, nil
	}
	lastApplyTime, err := time.Parse(time.UnixDate, lastApplyTimeAnnotation)
	if err != nil {
		return r.Clock.Now(), err
	}

	if lastApplyTime.After(lastActionTime) {
		lastActionTime = lastApplyTime
	}
	return lastActionTime, nil
}

func LayerFilesHaveChanged(layer configv1alpha1.TerraformLayer, changedFiles []string) bool {
	if len(changedFiles) == 0 {
		return true
	}

	// At last one changed file must be under refresh path
	for _, f := range changedFiles {
		f = ensureAbsPath(f)
		if strings.Contains(f, layer.Spec.Path) {
			return true
		}
		// Check if the file is under an additionnal trigger path
		if val, ok := layer.Annotations[annotations.AdditionnalTriggerPaths]; ok {
			for _, p := range strings.Split(val, ",") {
				p = ensureAbsPath(p)
				// Handle relative parent paths (like "../")
				p = filepath.Clean(filepath.Join(layer.Spec.Path, p))
				if strings.Contains(f, p) {
					return true
				}
			}
		}
	}

	return false
}

func ensureAbsPath(input string) string {
	if !filepath.IsAbs(input) {
		return string(filepath.Separator) + input
	}
	return input
}

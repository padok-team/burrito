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
	lastPlanDate, _ := time.Parse(time.UnixDate, value)
	delta, _ := time.ParseDuration(r.Config.Controller.Timers.DriftDetection)
	nextPlanDate := lastPlanDate.Add(delta)
	if nextPlanDate.After(lastPlanDate) {
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

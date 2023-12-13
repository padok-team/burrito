package terraformlayer

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Reconciler) HasBeenInitialized(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasBeenPlannedOnLastRelevantCommit",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	value, ok := t.Annotations[annotations.LastBranchCommit]
	if !ok || value == "" {
		condition.Reason = "NoAnnotation"
		condition.Message = "No annotation present on layer"
		return condition, false
	}
	condition.Reason = "HasBeenInitialized"
	condition.Message = "Layer has been initialized"
	condition.Status = metav1.ConditionTrue
	return condition, true
}

func (r *Reconciler) IsRunning(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsRunning",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if t.Status.CurrentRun == "" {
		condition.Reason = "NoCurrentRun"
		condition.Message = "No TerraformRun is currently running"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Reason = "IsRunning"
	condition.Message = "A TerraformRun is currently running"
	condition.Status = metav1.ConditionTrue
	return condition, true
}

func (r *Reconciler) HasBeenPlannedOnLastRelevantCommit(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasBeenPlannedOnLastRelevantCommit",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if t.Status.Plan.Commit == "" {
		condition.Reason = "NeverPlanned"
		condition.Message = "Layer has never been planned"
		return condition, false
	}
	if t.Annotations[annotations.LastBranchCommit] == t.Status.Plan.Commit {
		condition.Reason = "PlannedOnLastBranchCommit"
		condition.Message = "The last branch commit has been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	value, ok := t.Annotations[annotations.LastRelevantCommit]
	if ok {
		if value == t.Status.Plan.Commit {
			condition.Reason = "PlannedOnLastRelevantCommit"
			condition.Message = "The last relevant commit has been planned"
			condition.Status = metav1.ConditionTrue
			return condition, true
		}
		condition.Reason = "NotPlannedOnLastRelevantCommit"
		condition.Message = "The last relevant commit has not been planned"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Reason = "NoRelevantCommit"
	condition.Message = "No relevant commit for layer, passing to drift detection to determine if it needs planning"
	condition.Status = metav1.ConditionUnknown
	return condition, true
}

func (r *Reconciler) ShouldBeCheckedForDrift(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "ShouldBeCheckedForDrift",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if t.Status.Plan.Commit == "" {
		condition.Reason = "NeverPlanned"
		condition.Message = "Layer has never been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	lastPlanDate, err := time.Parse(time.UnixDate, t.Status.Plan.Date)
	if err != nil {
		condition.Reason = "ParseError"
		condition.Message = "Could not parse time from status"
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

func (r *Reconciler) ShouldBeApplied(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "ShouldBeApplied",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if t.Status.Plan.Commit == "" {
		condition.Reason = "NeverPlanned"
		condition.Message = "Layer has never been planned"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if t.Status.Apply.Commit == "" {
		condition.Reason = "NeverApplied"
		condition.Message = "Layer has never been applied"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if t.Status.Apply.Commit == t.Status.Plan.Commit {
		condition.Reason = "AlreadyApplied"
		condition.Message = "Layer has already been applied"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Reason = "ShouldBeApplied"
	condition.Message = "Layer should be applied"
	condition.Status = metav1.ConditionTrue
	return condition, true
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

package terraformlayer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	terraformrun "github.com/padok-team/burrito/internal/controllers/terraformrun"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type lastRunRetryInfo struct {
	reachedLimit bool
	action       string
}

func (r *Reconciler) IsRunning(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsRunning",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if t.Status.LastRun.Name == "" {
		condition.Reason = "NoRunHasRunYet"
		condition.Message = "No run has run on this layer yet"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	run := configv1alpha1.TerraformRun{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: t.Namespace,
		Name:      t.Status.LastRun.Name,
	}, &run)
	if errors.IsNotFound(err) {
		condition.Reason = "RunNotFound"
		condition.Message = "The Last Run could not be fetched, it might have been manually deleted, considering layer is not running"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if err != nil {
		condition.Reason = "RunRetrievalError"
		condition.Message = "An error happened while fetching the last run, this is likely a bug, considering layer is running"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if run.Status.State != "Succeeded" && run.Status.State != "Failed" {
		condition.Reason = "RunStillRunning"
		condition.Message = "The last run is still running"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "RunFinished"
	condition.Message = "The last run has finished"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) IsLastPlanTooOld(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsLastPlanTooOld",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	value, ok := t.Annotations[annotations.LastPlanDate]
	if !ok {
		condition.Reason = "NoPlanHasRunYet"
		condition.Message = "No plan has run on this layer yet"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	lastPlanDate, err := time.Parse(time.UnixDate, value)
	if err != nil {
		condition.Reason = "ParseError"
		condition.Message = "Burrito could not parse the time from the annotation, this is likely a bug, considering plan is recent to lock the behavior"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	nextPlanDate := lastPlanDate.Add(r.Config.Controller.Timers.DriftDetection)
	now := r.Clock.Now()
	if nextPlanDate.After(now) {
		condition.Reason = "PlanIsRecent"
		condition.Message = fmt.Sprintf("The plan has been made less than %s ago.", r.Config.Controller.Timers.DriftDetection)
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Reason = "PlanIsTooOld"
	condition.Message = fmt.Sprintf("The plan has been made more than %s ago.", r.Config.Controller.Timers.DriftDetection)
	condition.Status = metav1.ConditionTrue
	return condition, true
}

func (r *Reconciler) HasLastPlanFailed(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasLastPlanFailed",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	value, ok := t.Annotations[annotations.LastPlanSum]
	if !ok {
		condition.Reason = "NoPlanHasRunYet"
		condition.Message = "No plan has run on this layer yet"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if value == "" {
		condition.Reason = "NoPlanSum"
		condition.Message = "The last plan has no sum, considering plan failed"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "LastPlanHasSucceeded"
	condition.Message = "The last plan has succeeded"
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
		condition.Reason = "NoRelevantCommitReceived"
		condition.Message = "No relevant commit has been received from webhook, letting drift detection take a decision"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if lastPlannedCommit == lastBranchCommit || lastPlannedCommit == lastRelevantCommit {
		condition.Reason = "LastRelevantCommitPlanned"
		condition.Message = "The last relevant commit has already been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	lastBranchCommitDate, ok := t.Annotations[annotations.LastBranchCommitDate]
	if !ok {
		condition.Reason = "NoDatePresent"
		condition.Message = "The last received branch commit does not have a date, can't take a decision"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	lastPlanDate, err := time.Parse(time.UnixDate, t.Annotations[annotations.LastPlanDate])
	if err != nil {
		condition.Reason = "ParseError"
		condition.Message = "Could not parse time from annotation, this is likely a bug, considering layer has been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	lastBranchCommitDateParsed, err := time.Parse(time.UnixDate, lastBranchCommitDate)
	if err != nil {
		condition.Reason = "ParseError"
		condition.Message = "Could not parse time from annotation, this is likely a bug, considering layer has been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if lastPlanDate.After(lastBranchCommitDateParsed) {
		condition.Reason = "LastPlanIsMoreRecentThanLastBranchCommit"
		condition.Message = "The last plan is more recent than the last branch commit, considering layer has been planned"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "LastRelevantCommitNotPlanned"
	condition.Message = "The last relevant commit has not been planned yet"
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
	if !ok || planHash == "" {
		condition.Reason = "NoPlanHasRunYet"
		condition.Message = "No plan has run on this layer yet"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	applyHash, ok := t.Annotations[annotations.LastApplySum]
	if !ok {
		condition.Reason = "NoApplyHasRun"
		condition.Message = "Apply has not run yet but a plan is available, launching apply"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if applyHash == "" {
		applyDate, err := time.Parse(time.UnixDate, t.Annotations[annotations.LastApplyDate])
		if err != nil {
			condition.Reason = "ParseError"
			condition.Message = "Could not parse time from annotation, this is likely a bug, considering apply is up to date"
			condition.Status = metav1.ConditionTrue
			return condition, true
		}
		planDate, err := time.Parse(time.UnixDate, t.Annotations[annotations.LastPlanDate])
		if err != nil {
			condition.Reason = "ParseError"
			condition.Message = "Could not parse time from annotation, this is likely a bug, considering apply is up to date"
			condition.Status = metav1.ConditionTrue
			return condition, true
		}
		if planDate.After(applyDate) {
			condition.Reason = "NewPlanAvailable"
			condition.Message = "A new plan is available, apply will run."
			condition.Status = metav1.ConditionFalse
			return condition, false
		}
		condition.Reason = "ApplyUpToDate"
		condition.Message = "Apply has failed, waiting for another plan to run"
		condition.Status = metav1.ConditionTrue
		return condition, true
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

func (r *Reconciler) IsSyncScheduled(t *configv1alpha1.TerraformLayer) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsSyncScheduled",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	// check if annotations.SyncNow is present
	if _, ok := t.Annotations[annotations.SyncNow]; ok {
		condition.Reason = "SyncScheduled"
		condition.Message = "A sync has been manually scheduled"
		condition.Status = metav1.ConditionTrue
		// Remove the annotation to avoid running the sync again
		err := annotations.Remove(context.Background(), r.Client, t, annotations.SyncNow)
		if err != nil {
			log.Errorf("Failed to remove annotation %s from layer %s: %s", annotations.SyncNow, t.Name, err)
		}
		return condition, true
	}
	condition.Reason = "NoSyncScheduled"
	condition.Message = "No sync has been manually scheduled"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) HasLastRunReachedRetryLimit(layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (metav1.Condition, lastRunRetryInfo) {
	condition := metav1.Condition{
		Type:               "HasLastRunReachedRetryLimit",
		ObservedGeneration: layer.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if layer.Status.LastRun.Name == "" {
		condition.Reason = "NoRunYet"
		condition.Message = "No run has been created for this layer yet"
		condition.Status = metav1.ConditionFalse
		return condition, lastRunRetryInfo{}
	}
	run := configv1alpha1.TerraformRun{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: layer.Namespace,
		Name:      layer.Status.LastRun.Name,
	}, &run)
	if err != nil {
		condition.Reason = "RunRetrievalError"
		condition.Message = "Could not fetch the last run"
		condition.Status = metav1.ConditionFalse
		return condition, lastRunRetryInfo{}
	}
	if run.Status.State != "Failed" {
		condition.Reason = "LastRunNotFailed"
		condition.Message = "The last run has not failed"
		condition.Status = metav1.ConditionFalse
		return condition, lastRunRetryInfo{action: run.Spec.Action}
	}
	maxRetries := terraformrun.GetMaxRetries(r.Config.Controller.TerraformMaxRetries, repo, layer)
	if run.Status.Retries < maxRetries {
		condition.Reason = "RetryLimitNotReached"
		condition.Message = "The last run has not reached the retry limit"
		condition.Status = metav1.ConditionFalse
		return condition, lastRunRetryInfo{action: run.Spec.Action}
	}
	currentRevision := layer.Annotations[annotations.LastRelevantCommit]
	if currentRevision != "" && run.Spec.Layer.Revision != currentRevision {
		condition.Reason = "NewRevisionAvailable"
		condition.Message = "The last run reached retry limit but a new revision is available"
		condition.Status = metav1.ConditionFalse
		return condition, lastRunRetryInfo{action: run.Spec.Action}
	}
	condition.Reason = "HasReachedRetryLimit"
	condition.Message = fmt.Sprintf("The last %s run has reached the retry limit (%d)", run.Spec.Action, maxRetries)
	condition.Status = metav1.ConditionTrue
	return condition, lastRunRetryInfo{reachedLimit: true, action: run.Spec.Action}
}

func LayerFilesHaveChanged(layer configv1alpha1.TerraformLayer, changedFiles []string) bool {
	if len(changedFiles) == 0 {
		return true
	}

	patterns := additionalTriggerPaths(layer)

	// At least one changed file must be under refresh path
	for _, f := range changedFiles {
		f = ensureAbsPath(f)
		if strings.Contains(f, layer.Spec.Path) {
			return true
		}
		// Check if the file is under an additional trigger path
		for _, p := range patterns {
			if matchesTriggerPath(layer.Spec.Path, p, f) {
				return true
			}
		}
	}

	return false
}

// additionalTriggerPaths returns the trigger path patterns to check in addition to
// the layer's path. spec.additionalTriggerPaths takes precedence over the deprecated
// annotation: when the spec field is set, the annotation is ignored entirely.
func additionalTriggerPaths(layer configv1alpha1.TerraformLayer) []string {
	if len(layer.Spec.AdditionalTriggerPaths) > 0 {
		return layer.Spec.AdditionalTriggerPaths
	}
	if val, ok := layer.Annotations[annotations.AdditionnalTriggerPaths]; ok {
		log.Warnf("DEPRECATED: layer %s/%s uses annotation %s, use spec.additionalTriggerPaths instead", layer.Namespace, layer.Name, annotations.AdditionnalTriggerPaths)
		return strings.Split(val, ",")
	}
	return nil
}

// matchesTriggerPath resolves pattern relative to layerPath (supporting "../" to escape
// it) and checks whether changedFile matches. Patterns containing glob characters are
// matched recursively (e.g. "../**/*.yaml"); plain patterns keep the historical substring match.
func matchesTriggerPath(layerPath string, pattern string, changedFile string) bool {
	pattern = ensureAbsPath(pattern)
	// Handle relative parent paths (like "../")
	pattern = filepath.Clean(filepath.Join(layerPath, pattern))

	if strings.ContainsAny(pattern, "*?[") {
		sep := string(filepath.Separator)
		matched, err := doublestar.Match(strings.TrimPrefix(pattern, sep), strings.TrimPrefix(changedFile, sep))
		return err == nil && matched
	}

	return strings.Contains(changedFile, pattern)
}

func ensureAbsPath(input string) string {
	if !filepath.IsAbs(input) {
		return string(filepath.Separator) + input
	}
	return input
}

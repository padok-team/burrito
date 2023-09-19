package terraformrun

import (
	"context"
	"fmt"
	"math"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *Reconciler) getPodPhase(name string, namespace string) corev1.PodPhase {
	pod := &corev1.Pod{}
	err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, pod)
	if err != nil {
		log.Errorf("conditions: could not get runner pod %s: %s", name, err)
		return corev1.PodUnknown
	}
	return pod.Status.Phase
}

func (r *Reconciler) HasStatus(t *configv1alpha1.TerraformRun) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasStatus",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	if t.Status.State != "" {
		condition.Reason = "HasStatus"
		condition.Message = "This run has a status"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "HasNoStatus"
	condition.Message = "This run has no status"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) HasReachedRetryLimit(
	run *configv1alpha1.TerraformRun,
	layer *configv1alpha1.TerraformLayer,
	repo *configv1alpha1.TerraformRepository,
) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasReachedRetryLimit",
		ObservedGeneration: run.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	// TODO: Fix this, layer with no remediation strategy will have 0 as default value
	maxRetries := int(math.Min(
		float64(repo.Spec.RemediationStrategy.OnError.MaxRetries),
		float64(layer.Spec.RemediationStrategy.OnError.MaxRetries),
	))
	if run.Status.Retries >= maxRetries {
		condition.Reason = "HasReachedRetryLimit"
		condition.Message = fmt.Sprintf("This run has reached the retry limit (%d)", maxRetries)
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "HasNotReachedRetryLimit"
	condition.Message = fmt.Sprintf("This run has not reached the retry limit (%d)", maxRetries)
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) HasSucceeded(t *configv1alpha1.TerraformRun) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasSucceeded",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	currentState := t.Status.State
	podPhase := r.getPodPhase(t.Status.RunnerPod, t.Namespace)
	if currentState == "Suceeded" || podPhase == corev1.PodSucceeded {
		condition.Reason = "HasSucceeded"
		condition.Message = "This run has succeeded"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "HasNotSucceeded"
	condition.Message = "This run has not succeeded"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) IsRunning(t *configv1alpha1.TerraformRun) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsRunning",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	currentState := t.Status.State
	runnerPod := t.Status.RunnerPod
	if (currentState == "Initial" || currentState == "Retrying" || currentState == "Running") && runnerPod != "" {
		podPhase := r.getPodPhase(runnerPod, t.Namespace)
		if podPhase == corev1.PodPending || podPhase == corev1.PodRunning {
			condition.Reason = "IsRunning"
			condition.Message = fmt.Sprintf("This run is currently running with pod %s", runnerPod)
			condition.Status = metav1.ConditionTrue
			return condition, true
		}
	}
	condition.Reason = "NotRunning"
	condition.Message = "This run is not running"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) IsInFailureGracePeriod(t *configv1alpha1.TerraformRun) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsInFailureGracePeriod",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	podPhase := r.getPodPhase(t.Status.RunnerPod, t.Namespace)
	if podPhase == corev1.PodFailed {
		lastFailureTime, err := getLastActionTime(r, t)
		if err != nil {
			condition.Reason = "CouldNotGetLastActionTime"
			condition.Message = "Could not get last action time from resource status"
			condition.Status = metav1.ConditionFalse
			return condition, false
		}
		nextFailure := lastFailureTime.Add(getRunExponentialBackOffTime(r.Config.Controller.Timers.FailureGracePeriod, t))
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
	condition.Reason = "NoFailureYet"
	condition.Message = "No failure has been detected yet"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func getLastActionTime(r *Reconciler, run *configv1alpha1.TerraformRun) (time.Time, error) {
	lastActionTime, err := time.Parse(time.UnixDate, run.Status.LastRun)
	if err != nil {
		return r.Clock.Now(), err
	}
	return lastActionTime, nil
}

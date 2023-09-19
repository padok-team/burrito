package terraformrun

import (
	"context"
	"fmt"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

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

// func (r *Reconciler) HasReachedRetryLimit(t *configv1alpha1.TerraformRun) (metav1.Condition, bool) {

// }

func (r *Reconciler) HasSucceeded(t *configv1alpha1.TerraformRun) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "HasSucceeded",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	isPodSuceeded := r.isRunnerPodInPhase(t.Status.RunnerPod, t.Namespace, corev1.PodSucceeded)
	if isPodSuceeded {
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
	if currentState == "" {
		condition.Reason = "HasNotRunYet"
		condition.Message = "This run has not run yet"
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if currentState == "Running" && runnerPod != "" {
		condition.Reason = "IsRunning"
		condition.Message = fmt.Sprintf("This run is currently running with pod %s", runnerPod)
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	if currentState == "Suceeded" || currentState == "Failed" {
		condition.Reason = "HasFinished"
		condition.Message = fmt.Sprintf("This run has finished with state %s", currentState)
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if currentState == "FailureGracePeriod" {
		condition.Reason = "FailureGracePeriod"
		condition.Message = "This run is in the failure grace period"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Reason = "UnknownState"
	condition.Message = "This run is in an unknown state"
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) isRunnerPodInPhase(podName string, namespace string, phase corev1.PodPhase) bool {
	pod := &corev1.Pod{}
	err := r.Client.Get(context.Background(), types.NamespacedName{
		Name:      podName,
		Namespace: namespace,
	}, pod)
	if err != nil {
		return false
	}
	return pod.Status.Phase == phase
}

func (r *Reconciler) IsInFailureGracePeriod(t *configv1alpha1.TerraformRun) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsInFailureGracePeriod",
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	currentState := t.Status.State
	if currentState == "" {
		condition.Reason = "HasNotRunYet"
		condition.Message = "This run has not run yet"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	podFailed := r.isRunnerPodInPhase(t.Status.RunnerPod, t.Namespace, corev1.PodFailed)
	if (currentState == "Running" && podFailed) || currentState == "FailureGracePeriod" {
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

	condition.Reason = "HasFinished"
	condition.Message = "This run has successfully finished"
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

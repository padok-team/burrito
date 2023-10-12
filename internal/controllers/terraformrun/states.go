package terraformrun

import (
	"context"
	"fmt"
	"strings"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/lock"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type RunInfo struct {
	Retries   int
	LastRun   string
	RunnerPod string
}

func getRunInfo(run *configv1alpha1.TerraformRun) RunInfo {
	return RunInfo{
		Retries:   run.Status.Retries,
		LastRun:   run.Status.LastRun,
		RunnerPod: run.Status.RunnerPod,
	}
}

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformRun, *configv1alpha1.TerraformLayer, *configv1alpha1.TerraformRepository) (ctrl.Result, RunInfo)

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, hasStatus := r.HasStatus(run)
	c2, hasReachedRetryLimit := r.HasReachedRetryLimit(run, layer, repo)
	c3, hasSucceeded := r.HasSucceeded(run)
	c4, isRunning := r.IsRunning(run)
	c5, isInFailureGracePeriod := r.IsInFailureGracePeriod(run)
	conditions := []metav1.Condition{c1, c2, c3, c4, c5}
	switch {
	case !hasStatus:
		log.Infof("run %s is in initial state", run.Name)
		return &Initial{}, conditions
	case hasSucceeded:
		log.Infof("run %s has succeeded", run.Name)
		return &Succeeded{}, conditions
	case isInFailureGracePeriod && !hasReachedRetryLimit && !isRunning:
		log.Infof("run %s is in failure grace period", run.Name)
		return &FailureGracePeriod{}, conditions
	case isInFailureGracePeriod && hasReachedRetryLimit && !isRunning:
		log.Infof("run %s has reached retry limit, marking run as failed", run.Name)
		return &Failed{}, conditions
	case !isRunning && !hasReachedRetryLimit:
		log.Infof("run %s has not reach retry limit, retrying...", run.Name)
		return &Retrying{}, conditions
	case isRunning:
		log.Infof("run %s is running", run.Name)
		return &Running{}, conditions
	default:
		log.Infof("run %s is failed", run.Name)
		return &Failed{}, conditions
	}
}

type Initial struct{}

func (s *Initial) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (ctrl.Result, RunInfo) {
		log := log.WithContext(ctx)
		err := lock.CreateLock(ctx, r.Client, layer, run)
		if err != nil {
			log.Errorf("could not set lock on run %s for layer %s, requeuing resource: %s", run.Name, layer.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, RunInfo{}
		}
		pod := r.getPod(run, layer, repo)
		err = r.Client.Create(ctx, &pod)
		if err != nil {
			log.Errorf("failed to create pod for run %s: %s", run.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, RunInfo{}
		}
		runInfo := RunInfo{
			Retries:   0,
			LastRun:   r.Clock.Now().Format(time.UnixDate),
			RunnerPod: pod.Name,
		}
		// Minimal time (1s) to transit from Initial state to Running state
		return ctrl.Result{RequeueAfter: time.Duration(1 * time.Second)}, runInfo
	}
}

type Running struct{}

func (s *Running) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (ctrl.Result, RunInfo) {
		// Wait and do nothing
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, getRunInfo(run)
	}
}

type FailureGracePeriod struct{}

func (s *FailureGracePeriod) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (ctrl.Result, RunInfo) {
		lastActionTime, ok := getLastActionTime(r, run)
		if ok != nil {
			log.Errorf("could not get lastActionTime on run %s,: %s", run.Name, ok)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, getRunInfo(run)
		}
		expTime := GetRunExponentialBackOffTime(r.Config.Controller.Timers.FailureGracePeriod, run)
		endIdleTime := lastActionTime.Add(expTime)
		now := r.Clock.Now()
		if endIdleTime.After(now) {
			log.Infof("the grace period is over for run %v, new retry", run.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, getRunInfo(run)
		}
		return ctrl.Result{RequeueAfter: now.Sub(endIdleTime)}, getRunInfo(run)
	}
}

type Retrying struct{}

func (s *Retrying) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (ctrl.Result, RunInfo) {
		log := log.WithContext(ctx)
		runInfo := getRunInfo(run)
		pod := r.getPod(run, layer, repo)
		err := r.Client.Create(ctx, &pod)
		if err != nil {
			log.Errorf("failed to create retry pod for run %s: %s", run.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, runInfo
		}
		runInfo = RunInfo{
			Retries:   runInfo.Retries + 1,
			LastRun:   r.Clock.Now().Format(time.UnixDate),
			RunnerPod: pod.Name,
		}
		// Minimal time (1s) to transit from Retrying state to Running state
		return ctrl.Result{RequeueAfter: time.Duration(1 * time.Second)}, runInfo
	}
}

type Succeeded struct{}

func (s *Succeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (ctrl.Result, RunInfo) {
		// Try to delete lock if it still exists
		log := log.WithContext(ctx)
		err := lock.DeleteLock(ctx, r.Client, layer, run)
		if err != nil {
			log.Errorf("could not delete lock for run %s: %s", run.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, getRunInfo(run)
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, getRunInfo(run)
	}
}

type Failed struct{}

func (s *Failed) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository) (ctrl.Result, RunInfo) {
		// Try to delete lock if it still exists
		log := log.WithContext(ctx)
		err := lock.DeleteLock(ctx, r.Client, layer, run)
		if err != nil {
			log.Errorf("could not delete lock for run %s: %s", run.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, getRunInfo(run)
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, getRunInfo(run)
	}
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

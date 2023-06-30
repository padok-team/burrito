package terraformlayer

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/lock"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformLayer, *configv1alpha1.TerraformRepository) ctrl.Result

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, layer *configv1alpha1.TerraformLayer) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, isPlanArtifactUpToDate := r.IsPlanArtifactUpToDate(layer)
	c2, isApplyUpToDate := r.IsApplyUpToDate(layer)
	c3, isLastRelevantCommitPlanned := r.IsLastRelevantCommitPlanned(layer)
	c4, isInFailureGracePeriod := r.IsInFailureGracePeriod(layer)
	conditions := []metav1.Condition{c1, c2, c3, c4}
	switch {
	case isInFailureGracePeriod:
		log.Infof("layer %s is in failure grace period", layer.Name)
		return &FailureGracePeriod{}, conditions
	case isPlanArtifactUpToDate && isApplyUpToDate && isLastRelevantCommitPlanned:
		log.Infof("layer %s is up to date, waiting for a new drift detection cycle", layer.Name)
		return &Idle{}, conditions
	case !isPlanArtifactUpToDate || !isLastRelevantCommitPlanned:
		log.Infof("layer %s needs to be planned, acquiring lock and creating a new runner", layer.Name)
		return &PlanNeeded{}, conditions
	case isPlanArtifactUpToDate && !isApplyUpToDate:
		log.Infof("layer %s needs to be applied, acquiring lock and creating a new runner", layer.Name)
		return &ApplyNeeded{}, conditions
	default:
		log.Infof("layer %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is an intended behavior.", layer.Name)
		return &Idle{}, conditions
	}
}

type FailureGracePeriod struct{}

func (s *FailureGracePeriod) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		lastActionTime, ok := GetLastActionTime(r, layer)
		if ok != nil {
			log.Errorf("could not get lastActionTime on layer %s,: %s", layer.Name, ok)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		expTime := GetLayerExponentialBackOffTime(r.Config.Controller.Timers.FailureGracePeriod, layer)
		endIdleTime := lastActionTime.Add(expTime)
		now := r.Clock.Now()
		if endIdleTime.After(now) {
			log.Infof("the grace period is over for layer %v, new retry", layer.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
		}
		return ctrl.Result{RequeueAfter: now.Sub(endIdleTime)}
	}
}

type Idle struct{}

func (s *Idle) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}
	}
}

type PlanNeeded struct{}

func (s *PlanNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.WithContext(ctx)
		err := lock.CreateLock(ctx, r.Client, layer)
		if err != nil {
			log.Errorf("could not set lock on layer %s, requeing resource: %s", layer.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		pod := r.getPod(layer, repository, "plan")
		err = r.Client.Create(ctx, &pod)
		if err != nil {
			log.Errorf("failed to create pod for Plan action on layer %s: %s", layer.Name, err)
			_ = lock.DeleteLock(ctx, r.Client, layer)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
	}
}

type ApplyNeeded struct{}

func (s *ApplyNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.WithContext(ctx)
		remediationStrategy := getRemediationStrategy(repository, layer)
		if remediationStrategy != configv1alpha1.AutoApplyRemediationStrategy && !isForcedApply(layer) {
			log.Infof("layer %s is in dry mode, no action taken", layer.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}
		}
		err := lock.CreateLock(ctx, r.Client, layer)
		if err != nil {
			log.Errorf("could not set lock on layer %s, requeing resource: %s", layer.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		pod := r.getPod(layer, repository, "apply")
		err = r.Client.Create(ctx, &pod)
		if err != nil {
			log.Errorf("failed to create pod for Apply action on layer %s: %s", layer.Name, err)
			_ = lock.DeleteLock(ctx, r.Client, layer)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
	}
}

func getRemediationStrategy(repo *configv1alpha1.TerraformRepository, layer *configv1alpha1.TerraformLayer) configv1alpha1.RemediationStrategy {
	result := configv1alpha1.DryRemediationStrategy
	if len(repo.Spec.RemediationStrategy) > 0 {
		result = repo.Spec.RemediationStrategy
	}
	if len(layer.Spec.RemediationStrategy) > 0 {
		result = layer.Spec.RemediationStrategy
	}
	return result
}

func isForcedApply(layer *configv1alpha1.TerraformLayer) bool {
	if val, ok := layer.Annotations[annotations.ForceApply]; ok {
		return val == "1"
	}
	return false
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

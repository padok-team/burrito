package terraformlayer

import (
	"context"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/lock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type State interface {
	getHandler() func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result
}

func (r *Reconciler) GetState(ctx context.Context, l *configv1alpha1.TerraformLayer) (State, []metav1.Condition) {
	log := log.FromContext(ctx)
	c1, isPlanArtifactUpToDate := r.IsPlanArtifactUpToDate(l)
	c2, isApplyUpToDate := r.IsApplyUpToDate(l)
	c3, isLastConcerningCommitPlanned := r.IsLastConcernginCommitPlanned(l)
	// c3, hasFailed := HasFailed(r)
	conditions := []metav1.Condition{c1, c2, c3}
	switch {
	case isPlanArtifactUpToDate && isApplyUpToDate:
		log.Info("Layer is up to date, waiting for a new drift detection cycle")
		return &IdleState{}, conditions
	case !isPlanArtifactUpToDate || !isLastConcerningCommitPlanned:
		log.Info("Layer needs to be planned, acquiring lock and creating a new runner")
		return &PlanNeededState{}, conditions
	case isPlanArtifactUpToDate && !isApplyUpToDate:
		log.Info("Layer needs to be applied, acquiring lock and creating a new runner")
		return &ApplyNeededState{}, conditions
	default:
		log.Info("Layer is in an unknown state, defaulting to idle. If this happens please file an issue, this is an intended behavior.")
		return &IdleState{}, conditions
	}
}

type IdleState struct{}

func (s *IdleState) getHandler() func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.FromContext(ctx)
		delta, err := time.ParseDuration(t.Config.Controller.Timers.DriftDetection)
		if err != nil {
			log.Error(err, "could not parse timer drift detection period")
			return ctrl.Result{}
		}
		return ctrl.Result{RequeueAfter: delta}
	}
}

type PlanNeededState struct{}

func (s *PlanNeededState) getHandler() func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.FromContext(ctx)
		deltaOnError, err := time.ParseDuration(t.Config.Controller.Timers.OnError)
		if err != nil {
			log.Error(err, "could not parse timer on error period")
			return ctrl.Result{}
		}
		err = lock.CreateLock(ctx, t.Client, r)
		if err != nil {
			log.Error(err, "Could not set lock on layer, requeing resource")
			return ctrl.Result{RequeueAfter: deltaOnError}
		}
		pod := t.getPod(r, repository, "plan")
		err = t.Client.Create(ctx, &pod)
		if err != nil {
			log.Error(err, "Failed to create pod for Plan action")
			_ = lock.DeleteLock(ctx, t.Client, r)
			return ctrl.Result{RequeueAfter: deltaOnError}
		}
		delta, err := time.ParseDuration(t.Config.Controller.Timers.WaitAction)
		if err != nil {
			log.Error(err, "could not parse timer wait action period")
			return ctrl.Result{}
		}
		return ctrl.Result{RequeueAfter: delta}
	}
}

type ApplyNeededState struct{}

func (s *ApplyNeededState) getHandler() func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.FromContext(ctx)
		deltaDriftDetection, err := time.ParseDuration(t.Config.Controller.Timers.DriftDetection)
		if err != nil {
			log.Error(err, "could not parse timer drift detection period")
			return ctrl.Result{}
		}
		remediationStrategy := getRemediationStrategy(repository, r)
		if remediationStrategy != configv1alpha1.AutoApplyRemediationStrategy {
			log.Info("layer is in dry mode, no action taken")
			return ctrl.Result{RequeueAfter: deltaDriftDetection}
		}
		deltaOnError, err := time.ParseDuration(t.Config.Controller.Timers.OnError)
		if err != nil {
			log.Error(err, "could not parse timer on error period")
			return ctrl.Result{}
		}
		err = lock.CreateLock(ctx, t.Client, r)
		if err != nil {
			log.Error(err, "Could not set lock on layer, requeing resource")
			return ctrl.Result{RequeueAfter: deltaOnError}
		}
		pod := t.getPod(r, repository, "apply")
		err = t.Client.Create(ctx, &pod)
		if err != nil {
			log.Error(err, "[TerraformApplyNeeded] Failed to create pod for Apply action")
			_ = lock.DeleteLock(ctx, t.Client, r)
			return ctrl.Result{RequeueAfter: deltaOnError}
		}
		delta, err := time.ParseDuration(t.Config.Controller.Timers.WaitAction)
		if err != nil {
			log.Error(err, "could not parse timer wait action period")
			return ctrl.Result{}
		}
		return ctrl.Result{RequeueAfter: delta}
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

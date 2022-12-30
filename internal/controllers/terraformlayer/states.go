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
	// c3, hasFailed := HasFailed(r)
	conditions := []metav1.Condition{c1, c2}
	switch {
	case isPlanArtifactUpToDate && isApplyUpToDate:
		log.Info("Layer is up to date, waiting for a new drift detection cycle")
		return &IdleState{}, conditions
	case isPlanArtifactUpToDate && !isApplyUpToDate:
		log.Info("Layer needs to be applied, acquiring lock and creating a new runner")
		return &ApplyNeededState{}, conditions
	case !isPlanArtifactUpToDate:
		log.Info("Layer needs to be planned, acquiring lock and creating a new runner")
		return &PlanNeededState{}, conditions
	default:
		log.Info("Layer is in an unknown state, defaulting to idle. If this happens please file an issue, this is an intended behavior.")
		return &IdleState{}, conditions
	}
}

type IdleState struct{}

func (s *IdleState) getHandler() func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		delta, _ := time.ParseDuration(t.Config.Controller.Timers.DriftDetection)
		return ctrl.Result{RequeueAfter: delta}
	}
}

type PlanNeededState struct{}

func (s *PlanNeededState) getHandler() func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.FromContext(ctx)
		err := lock.CreateLock(ctx, t.Client, r)
		if err != nil {
			log.Error(err, "Could not set lock on layer, requeing resource")
			delta, _ := time.ParseDuration(t.Config.Controller.Timers.OnError)
			return ctrl.Result{RequeueAfter: delta}
		}
		pod := t.getPod(r, repository, "plan")
		err = t.Client.Create(ctx, &pod)
		if err != nil {
			log.Error(err, "Failed to create pod for Plan action")
			_ = lock.DeleteLock(ctx, t.Client, r)
			delta, _ := time.ParseDuration(t.Config.Controller.Timers.OnError)
			return ctrl.Result{RequeueAfter: delta}
		}
		delta, _ := time.ParseDuration(t.Config.Controller.Timers.WaitAction)
		return ctrl.Result{RequeueAfter: delta}
	}
}

type ApplyNeededState struct{}

func (s *ApplyNeededState) getHandler() func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, t *Reconciler, r *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.FromContext(ctx)
		err := lock.CreateLock(ctx, t.Client, r)
		if err != nil {
			log.Error(err, "Could not set lock on layer, requeing resource")
			delta, _ := time.ParseDuration(t.Config.Controller.Timers.OnError)
			return ctrl.Result{RequeueAfter: delta}
		}
		pod := t.getPod(r, repository, "apply")
		err = t.Client.Create(ctx, &pod)
		if err != nil {
			log.Error(err, "[TerraformApplyNeeded] Failed to create pod for Apply action")
			_ = lock.DeleteLock(ctx, t.Client, r)
			delta, _ := time.ParseDuration(t.Config.Controller.Timers.OnError)
			return ctrl.Result{RequeueAfter: delta}
		}
		delta, _ := time.ParseDuration(t.Config.Controller.Timers.WaitAction)
		return ctrl.Result{RequeueAfter: delta}
	}
}
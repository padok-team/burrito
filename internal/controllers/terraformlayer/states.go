package terraformlayer

import (
	"context"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/lock"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State interface {
	getHandler() func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result
}

func (r *Reconciler) GetState(ctx context.Context, layer *configv1alpha1.TerraformLayer) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, isPlanArtifactUpToDate := r.IsPlanArtifactUpToDate(layer)
	c2, isApplyUpToDate := r.IsApplyUpToDate(layer)
	c3, isLastConcerningCommitPlanned := r.IsLastConcernginCommitPlanned(layer)
	// c3, hasFailed := HasFailed(r)
	conditions := []metav1.Condition{c1, c2, c3}
	switch {
	case isPlanArtifactUpToDate && isApplyUpToDate:
		log.Infof("layer %s is up to date, waiting for a new drift detection cycle", layer.Name)
		return &IdleState{}, conditions
	case !isPlanArtifactUpToDate || !isLastConcerningCommitPlanned:
		log.Infof("layer %s needs to be planned, acquiring lock and creating a new runner", layer.Name)
		return &PlanNeededState{}, conditions
	case isPlanArtifactUpToDate && !isApplyUpToDate:
		log.Infof("layer %s needs to be applied, acquiring lock and creating a new runner", layer.Name)
		return &ApplyNeededState{}, conditions
	default:
		log.Infof("layer %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is an intended behavior.", layer.Name)
		return &IdleState{}, conditions
	}
}

type IdleState struct{}

func (s *IdleState) getHandler() func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}
	}
}

type PlanNeededState struct{}

func (s *PlanNeededState) getHandler() func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
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

type ApplyNeededState struct{}

func (s *ApplyNeededState) getHandler() func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ctrl.Result {
		log := log.WithContext(ctx)
		remediationStrategy := getRemediationStrategy(repository, layer)
		if remediationStrategy != configv1alpha1.AutoApplyRemediationStrategy {
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

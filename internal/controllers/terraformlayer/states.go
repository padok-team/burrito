package terraformlayer

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformLayer, *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun)

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, layer *configv1alpha1.TerraformLayer) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, isPlanArtifactUpToDate := r.IsPlanArtifactUpToDate(layer)
	c2, isApplyUpToDate := r.IsApplyUpToDate(layer)
	c3, isLastRelevantCommitPlanned := r.IsLastRelevantCommitPlanned(layer)
	conditions := []metav1.Condition{c1, c2, c3}
	switch {
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

type Idle struct{}

func (s *Idle) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}, nil
	}
}

type PlanNeeded struct{}

func (s *PlanNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
		log := log.WithContext(ctx)
		run := r.getRun(layer, repository, "plan")
		err := r.Client.Create(ctx, &run)
		if err != nil {
			log.Errorf("failed to create TerraformRun for Plan action on layer %s: %s", layer.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, &run
	}
}

type ApplyNeeded struct{}

func (s *ApplyNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
		log := log.WithContext(ctx)
		remediationStrategy := configv1alpha1.GetRemediationStrategy(repository, layer)
		if !remediationStrategy.AutoApply {
			log.Infof("layer %s is in dry mode, no action taken", layer.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}, nil
		}
		run := r.getRun(layer, repository, "apply")
		err := r.Client.Create(ctx, &run)
		if err != nil {
			log.Errorf("failed to create TerraformRun for Apply action on layer %s: %s", layer.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, &run
	}
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

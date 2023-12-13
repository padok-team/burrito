package terraformlayer

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	Initial     string = "Initial"
	Running     string = "Running"
	Idle        string = "Idle"
	PlanNeeded  string = "PlanNeeded"
	ApplyNeeded string = "ApplyNeeded"
)

type State struct {
	handler func(context.Context, *Reconciler, *configv1alpha1.TerraformLayer, *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun)
	Status  configv1alpha1.TerraformLayerStatus
}

func initialHandler(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
	branchStatus, ok := repository.Status.BranchStatuses[layer.Spec.Branch]
	if !ok {
		log.Infof("repository %s has no commit referenced for this branch %s, doing nothing", repository.Name, layer.Spec.Branch)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
	}
	err := annotations.Add(ctx, r.Client, layer, map[string]string{
		annotations.LastBranchCommit: branchStatus.Commit,
	})
	if err != nil {
		log.Errorf("failed to add annotation %s to layer %s: %s", annotations.LastBranchCommit, layer.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
	}
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
}

func runningHandler(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
}

func idleHandler(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
}

func planNeededHandler(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
	run := r.getRun(layer, repository, "plan")
	err := r.Client.Create(ctx, &run)
	if err != nil {
		log.Errorf("failed to create TerraformRun for Plan action on layer %s: %s", layer.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
	}
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, &run
}

func applyNeededHandler(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
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

func (r *Reconciler) GetState(ctx context.Context, layer *configv1alpha1.TerraformLayer) State {
	log := log.WithContext(ctx)
	conditions := []metav1.Condition{}
	state := State{
		Status: configv1alpha1.TerraformLayerStatus{
			Conditions: conditions,
		},
	}
	c0, HasBeenInitialized := r.HasBeenInitialized(layer)
	conditions = append(conditions, c0)
	if !HasBeenInitialized {
		log.Infof("layer %s has not been initialized yet, initialize its annotations", layer.Name)
		state.handler = initialHandler
		state.Status.State = Initial
		return state
	}
	c1, isRunning := r.IsRunning(layer)
	conditions = append(conditions, c1)
	if isRunning {
		log.Infof("layer %s is running, waiting for it to finish", layer.Name)
		state.handler = runningHandler
		state.Status.State = Running
		return state
	}
	c2, HasBeenPlannedOnLastRelevantCommit := r.HasBeenPlannedOnLastRelevantCommit(layer)
	conditions = append(conditions, c2)
	if !HasBeenPlannedOnLastRelevantCommit {
		log.Infof("layer %s has not been planned on last relevant commit, planning", layer.Name)
		state.handler = planNeededHandler
		state.Status.State = PlanNeeded
		return state
	}
	c3, ShouldBeCheckedForDrift := r.ShouldBeCheckedForDrift(layer)
	conditions = append(conditions, c3)
	if ShouldBeCheckedForDrift {
		log.Infof("layer %s should be checked for drift, planning", layer.Name)
		state.handler = planNeededHandler
		state.Status.State = PlanNeeded
		return state
	}
	c4, ShouldBeApplied := r.ShouldBeApplied(layer)
	conditions = append(conditions, c4)
	if ShouldBeApplied {
		log.Infof("layer %s should be applied, applying", layer.Name)
		state.handler = applyNeededHandler
		state.Status.State = ApplyNeeded
		return state
	}
	state.handler = idleHandler
	state.Status.State = Idle
	return state
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

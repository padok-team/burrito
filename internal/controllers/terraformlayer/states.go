package terraformlayer

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/utils/syncwindow"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformLayer, *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun)

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, layer *configv1alpha1.TerraformLayer) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, IsRunning := r.IsRunning(layer)
	c2, IsLastPlanTooOld := r.IsLastPlanTooOld(layer)
	c3, IsLastRelevantCommitPlanned := r.IsLastRelevantCommitPlanned(layer)
	c4, HasLastPlanFailed := r.HasLastPlanFailed(layer)
	c5, IsApplyUpToDate := r.IsApplyUpToDate(layer)
	c6, IsSyncScheduled := r.IsSyncScheduled(layer)
	conditions := []metav1.Condition{c1, c2, c3, c4, c5, c6}
	switch {
	case IsRunning:
		log.Infof("layer %s is running, waiting for the run to finish", layer.Name)
		return &Idle{}, conditions
	case IsLastPlanTooOld || !IsLastRelevantCommitPlanned:
		log.Infof("layer %s has an outdated plan, creating a new run", layer.Name)
		return &PlanNeeded{}, conditions
	case IsSyncScheduled:
		log.Infof("layer %s has a sync scheduled, creating a new run", layer.Name)
		return &PlanNeeded{}, conditions
	case !IsApplyUpToDate && !HasLastPlanFailed:
		log.Infof("layer %s needs to be applied, creating a new run", layer.Name)
		return &ApplyNeeded{}, conditions
	default:
		log.Infof("layer %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is not an intended behavior.", layer.Name)
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
		// Check for sync windows that would block the apply action
		if isActionBlocked(r, layer, repository, syncwindow.PlanAction) {
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
		}
		run := r.getRun(layer, repository, "plan")
		err := r.Client.Create(ctx, &run)
		if err != nil {
			r.Recorder.Event(layer, corev1.EventTypeWarning, "Reconciliation", "Failed to create TerraformRun for Plan action")
			log.Errorf("failed to create TerraformRun for Plan action on layer %s: %s", layer.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
		}
		r.Recorder.Event(layer, corev1.EventTypeNormal, "Reconciliation", "Created TerraformRun for Plan action")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, &run
	}
}

type ApplyNeeded struct{}

func (s *ApplyNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
		log := log.WithContext(ctx)
		autoApply := configv1alpha1.GetAutoApplyEnabled(repository, layer)
		if !autoApply {
			log.Infof("autoApply is disabled for layer %s, no apply action taken", layer.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}, nil
		}
		// Check for sync windows that would block the apply action
		if isActionBlocked(r, layer, repository, syncwindow.ApplyAction) {
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
		}
		run := r.getRun(layer, repository, "apply")
		err := r.Client.Create(ctx, &run)
		if err != nil {
			r.Recorder.Event(layer, corev1.EventTypeWarning, "Reconciliation", "Failed to create TerraformRun for Apply action")
			log.Errorf("failed to create TerraformRun for Apply action on layer %s: %s", layer.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
		}
		r.Recorder.Event(layer, corev1.EventTypeNormal, "Reconciliation", "Created TerraformRun for Apply action")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, &run
	}
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

func isActionBlocked(r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, action syncwindow.Action) bool {
	defaultSyncWindows := r.Config.Controller.DefaultSyncWindows
	syncBlocked, reason := syncwindow.IsSyncBlocked(append(repository.Spec.SyncWindows, defaultSyncWindows...), action, layer.Name)
	if syncBlocked {
		if reason == syncwindow.BlockReasonInsideDenyWindow {
			log.Infof("layer %s is in a deny window, no %s action taken", layer.Name, string(action))
			r.Recorder.Eventf(layer, corev1.EventTypeNormal, "Reconciliation", "Layer is in a deny window, no %s action taken", string(action))
		} else if reason == syncwindow.BlockReasonOutsideAllowWindow {
			log.Infof("layer %s is outside an allow window, no %s action taken", layer.Name, string(action))
			r.Recorder.Eventf(layer, corev1.EventTypeNormal, "Reconciliation", "Layer is outside an allow window, no %s action taken", string(action))
		}
		return true
	}
	return false
}

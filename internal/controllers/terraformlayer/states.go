package terraformlayer

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
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

func (r *Reconciler) GetState(ctx context.Context, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (State, []metav1.Condition) {
	log := log.WithContext(ctx)

	planApprovalRequired := configv1alpha1.GetPlanApprovalRequired(repository, layer)

	c1, IsRunning := r.IsRunning(layer)
	c2, IsLastPlanTooOld := r.IsLastPlanTooOld(layer)
	c3, IsLastRelevantCommitPlanned := r.IsLastRelevantCommitPlanned(layer)
	c4, HasLastPlanFailed := r.HasLastPlanFailed(layer)
	c5, IsApplyUpToDate := r.IsApplyUpToDate(layer)
	c6, IsSyncScheduled := r.IsSyncScheduled(layer)
	c7, IsPlanApproved := r.IsPlanApproved(layer)
	conditions := []metav1.Condition{c1, c2, c3, c4, c5, c6, c7}

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
		if planApprovalRequired {
			if IsPlanApproved {
				// Approval required and granted, so apply plan
				log.Infof("layer %s last plan is approved, creating a new run", layer.Name)
				return &ApplyNeeded{}, conditions
			}

			// Plan not approved, so move to `PlanApprovalNeeded` state
			log.Infof("layer %s needs to be approved before application", layer.Name)
			return &PlanApprovalNeeded{}, conditions
		} else {
			log.Infof("layer %s needs to be applied, creating a new run", layer.Name)
			return &ApplyNeeded{}, conditions
		}
	case IsApplyUpToDate:
		log.Infof("layer %s has applied the latest plan", layer.Name)
		return &Idle{}, conditions
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
		revision, ok := layer.Annotations[annotations.LastRelevantCommit]
		if !ok {
			r.Recorder.Event(layer, corev1.EventTypeWarning, "Reconciliation", "Layer has no last relevant commit annotation, Plan run not created")
			log.Errorf("layer %s has no last relevant commit annotation, run not created", layer.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
		}
		run := r.getRun(layer, revision, PlanAction)
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

type PlanApprovalNeeded struct{}

func (s *PlanApprovalNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
		log := log.WithContext(ctx)

		// Check for sync windows that would block the apply action
		if isActionBlocked(r, layer, repository, syncwindow.PlanAction) {
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
		}

		log.Infof("waiting for plan approval on layer %s", layer.Name)
		r.Recorder.Event(layer, corev1.EventTypeNormal, "Reconciliation", "Waiting for plan approval")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
	}
}

type ApplyNeeded struct{}

func (s *ApplyNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (ctrl.Result, *configv1alpha1.TerraformRun) {
		log := log.WithContext(ctx)

		autoApply := configv1alpha1.GetAutoApplyEnabled(repository, layer)
		planApprovalRequired := configv1alpha1.GetPlanApprovalRequired(repository, layer)

		if !autoApply && !planApprovalRequired {
			log.Infof("autoApply and planApprovalRequired is disabled for layer %s, no apply action taken", layer.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}, nil
		}

		// Check for sync windows that would block the apply action
		if isActionBlocked(r, layer, repository, syncwindow.ApplyAction) {
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
		}

		revision, ok := layer.Annotations[annotations.LastRelevantCommit]
		if !ok {
			r.Recorder.Event(layer, corev1.EventTypeWarning, "Reconciliation", "Layer has no last relevant commit annotation, Apply run not created")
			log.Errorf("layer %s has no last relevant commit annotation, run not created", layer.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, nil
		}
		run := r.getRun(layer, revision, ApplyAction)
		if err := r.Client.Create(ctx, &run); err != nil {
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
		switch reason {
		case syncwindow.BlockReasonInsideDenyWindow:
			log.Infof("layer %s is in a deny window, no %s action taken", layer.Name, string(action))
			r.Recorder.Eventf(layer, corev1.EventTypeNormal, "Reconciliation", "Layer is in a deny window, no %s action taken", string(action))
		case syncwindow.BlockReasonOutsideAllowWindow:
			log.Infof("layer %s is outside an allow window, no %s action taken", layer.Name, string(action))
			r.Recorder.Eventf(layer, corev1.EventTypeNormal, "Reconciliation", "Layer is outside an allow window, no %s action taken", string(action))
		}
		return true
	}
	return false
}

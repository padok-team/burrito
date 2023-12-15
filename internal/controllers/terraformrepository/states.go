package terraformrepository

import (
	"context"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	UpToDate     string = "UpToDate"
	UpdateNeeded string = "UpdateNeeded"
)

type State struct {
	handler func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository) ctrl.Result
	Status  configv1alpha1.TerraformRepositoryStatus
}

func (r *Reconciler) GetState(ctx context.Context, repository *configv1alpha1.TerraformRepository) State {
	var state State
	log := log.WithContext(ctx)
	c1, isLastCloneTooOld := r.IsLastCloneTooOld(repository)
	c2, wasABranchRecentlyUpdated := r.WasABranchRecentlyUpdated(repository)
	conditions := []metav1.Condition{c1, c2}
	state = State{
		Status: configv1alpha1.TerraformRepositoryStatus{
			Conditions: conditions,
		},
	}
	switch {
	case wasABranchRecentlyUpdated || isLastCloneTooOld:
		log.Infof("repository %s needs to be updated", repository.Name)
		state.handler = updateHandler
		state.Status.State = UpdateNeeded
	default:
		log.Infof("repository %s is up to date", repository.Name)
		state.handler = uptodateHandler
		state.Status.State = UpToDate
	}
	return state
}

func uptodateHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository) ctrl.Result {
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

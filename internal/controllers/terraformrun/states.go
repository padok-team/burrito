package terraformrun

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformRun) ctrl.Result

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, run *configv1alpha1.TerraformRun) (State, []metav1.Condition) {
	return nil, []metav1.Condition{}
	// log := log.WithContext(ctx)
	// c1, isPlanArtifactUpToDate := r.IsPlanArtifactUpToDate(layer)
	// c2, isApplyUpToDate := r.IsApplyUpToDate(layer)
	// c3, isLastRelevantCommitPlanned := r.IsLastRelevantCommitPlanned(layer)
	// c4, isInFailureGracePeriod := r.IsInFailureGracePeriod(layer)
	// conditions := []metav1.Condition{c1, c2, c3, c4}
	// switch {
	// case isInFailureGracePeriod:
	// 	log.Infof("layer %s is in failure grace period", layer.Name)
	// 	return &FailureGracePeriod{}, conditions
	// case isPlanArtifactUpToDate && isApplyUpToDate && isLastRelevantCommitPlanned:
	// 	log.Infof("layer %s is up to date, waiting for a new drift detection cycle", layer.Name)
	// 	return &Idle{}, conditions
	// case !isPlanArtifactUpToDate || !isLastRelevantCommitPlanned:
	// 	log.Infof("layer %s needs to be planned, acquiring lock and creating a new runner", layer.Name)
	// 	return &PlanNeeded{}, conditions
	// case isPlanArtifactUpToDate && !isApplyUpToDate:
	// 	log.Infof("layer %s needs to be applied, acquiring lock and creating a new runner", layer.Name)
	// 	return &ApplyNeeded{}, conditions
	// default:
	// 	log.Infof("layer %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is an intended behavior.", layer.Name)
	// 	return &Idle{}, conditions
	// }
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

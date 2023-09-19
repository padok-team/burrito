package terraformrun

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformRun, *configv1alpha1.TerraformLayer) ctrl.Result

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, run *configv1alpha1.TerraformRun) (State, []metav1.Condition) {
	// log := log.WithContext(ctx)
	// c1, hasStatus := r.HasStatus(run)
	// c2, isRunning := r.IsRunning(run)
	// c3, isFinished := r.IsFinished(run)
	// c4, isInFailureGracePeriod := r.IsInFailureGracePeriod(run)
	// conditions := []metav1.Condition{c1, c2, c3}
	// switch {
	// case isInFailureGracePeriod:
	// 	log.Infof("run %s is in failure grace period", run.Name)
	// 	return &FailureGracePeriod{}, conditions
	// case !isInFailureGracePeriod && isRunning:
	// 	log.Infof("run %s is running", run.Name)
	// 	return &Running{}, conditions
	// case isFinished && c2.Reason == "Succeeded":
	// 	log.Infof("run %s is finished and has succeeded", run.Name)
	// 	return &Succeeded{}, conditions
	// case isFinished && c2.Reason == "Failed":
	// 	log.Infof("run %s is finished and has definitely failed", run.Name)
	// 	return &Failed{}, conditions
	// default:
	// 	log.Infof("layer %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is an intended behavior.", layer.Name)
	// 	return &Failed{}, conditions
	// }
	return nil, nil
}

type Initial struct{}

type Running struct{}

type FailureGracePeriod struct{}

func (s *FailureGracePeriod) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer) ctrl.Result {
		lastActionTime, ok := getLastActionTime(r, run)
		if ok != nil {
			log.Errorf("could not get lastActionTime on run %s,: %s", run.Name, ok)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		expTime := getRunExponentialBackOffTime(r.Config.Controller.Timers.FailureGracePeriod, run)
		endIdleTime := lastActionTime.Add(expTime)
		now := r.Clock.Now()
		if endIdleTime.After(now) {
			log.Infof("the grace period is over for run %v, new retry", run.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
		}
		return ctrl.Result{RequeueAfter: now.Sub(endIdleTime)}
	}
}

type Succeeded struct{}

type Failed struct{}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

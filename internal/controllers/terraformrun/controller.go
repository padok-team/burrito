/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package terraformrun

import (
	"context"
	"math"
	"time"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (c RealClock) Now() time.Time {
	return time.Now()
}

// RunReconcilier reconciles a TerraformRun object
type Reconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Config   *config.Config
	Recorder record.EventRecorder
	Clock
}

//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformruns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformruns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformruns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TerraformRun object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.WithContext(ctx)
	log.Infof("starting reconciliation for run %s/%s ...", req.Namespace, req.Name)
	run := &configv1alpha1.TerraformRun{}
	err := r.Client.Get(ctx, req.NamespacedName, run)
	if errors.IsNotFound(err) {
		log.Errorf("resource not found. Ignoring since object must be deleted: %s", err)
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Errorf("failed to get TerraformRun: %s", err)
		return ctrl.Result{}, err
	}
	if run.Status.State == "Succeeded" || run.Status.State == "Failed" {
		log.Infof("run %s is in a terminal state, ignoring...", run.Name)
		return ctrl.Result{}, nil
	}
	layer, err := r.getLinkedLayer(run)
	if err != nil {
		r.Recorder.Event(run, corev1.EventTypeWarning, "Reconciliation", "Could not get linked layer")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, err
	}
	repo, err := r.getLinkedRepo(run, layer)
	if err != nil {
		r.Recorder.Event(run, corev1.EventTypeWarning, "Reconciliation", "Could not get linked repository")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, err
	}
	state, conditions := r.GetState(ctx, run, layer, repo)
	result, runInfo := state.getHandler()(ctx, r, run, layer, repo)
	run.Status = configv1alpha1.TerraformRunStatus{
		Conditions: conditions,
		State:      getStateString(state),
		Retries:    runInfo.Retries,
		LastRun:    runInfo.LastRun,
		RunnerPod:  runInfo.RunnerPod,
	}
	err = r.Client.Status().Update(ctx, run)
	if err != nil {
		r.Recorder.Event(run, corev1.EventTypeWarning, "Reconciliation", "Could not update run status")
		log.Errorf("could not update run %s status: %s", run.Name, err)
	}
	log.Infof("finished reconciliation cycle for run %s/%s", run.Namespace, run.Name)
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Clock = RealClock{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformRun{}).
		WithEventFilter(ignorePredicate()).
		Complete(r)
}

func GetRunExponentialBackOffTime(DefaultRequeueAfter time.Duration, run *configv1alpha1.TerraformRun) time.Duration {
	var attempts = run.Status.Retries
	if attempts < 1 {
		return DefaultRequeueAfter
	}
	return getExponentialBackOffTime(DefaultRequeueAfter, attempts)
}

func getExponentialBackOffTime(DefaultRequeueAfter time.Duration, attempts int) time.Duration {
	var x float64 = float64(attempts)
	return time.Duration(int32(math.Exp(x))) * DefaultRequeueAfter
}

func ignorePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Update only if generation or annotations change, filter out anything else.
			// We only need to check generation or annotations change here, because it is only
			// updated on spec changes. On the other hand RevisionVersion
			// changes also on status changes. We want to omit reconciliation
			// for status updates.
			return (e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()) ||
				cmp.Diff(e.ObjectOld.GetAnnotations(), e.ObjectNew.GetAnnotations()) != ""
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

func (r *Reconciler) getLinkedLayer(run *configv1alpha1.TerraformRun) (*configv1alpha1.TerraformLayer, error) {
	layer := &configv1alpha1.TerraformLayer{}
	log.Infof("getting linked TerraformLayer to run %s", run.Name)
	err := r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: run.Spec.Layer.Namespace,
		Name:      run.Spec.Layer.Name,
	}, layer)
	if errors.IsNotFound(err) {
		log.Infof("TerraformLayer linked to run %s not found, ignoring run until it's modified: %s", run.Name, err)
		return nil, err
	}
	if err != nil {
		log.Errorf("failed to get TerraformLayer linked to run %s: %s", run.Name, err)
		return nil, err
	}
	return layer, nil
}

func (r *Reconciler) getLinkedRepo(run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer) (*configv1alpha1.TerraformRepository, error) {
	repo := &configv1alpha1.TerraformRepository{}
	log.Infof("getting linked TerraformRepository to run %s", run.Name)
	err := r.Client.Get(context.Background(), types.NamespacedName{
		Namespace: layer.Spec.Repository.Namespace,
		Name:      layer.Spec.Repository.Name,
	}, repo)
	if errors.IsNotFound(err) {
		log.Infof("TerraformRepository linked to run %s not found, ignoring run until it's modified: %s", run.Name, err)
		return nil, err
	}
	if err != nil {
		log.Errorf("failed to get TerraformRepository linked to run %s: %s", run.Name, err)
		return nil, err
	}
	return repo, nil
}

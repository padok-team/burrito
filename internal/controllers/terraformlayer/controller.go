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

package terraformlayer

import (
	"context"
	e "errors"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/lock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	log "github.com/sirupsen/logrus"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (c RealClock) Now() time.Time {
	return time.Now()
}

// Reconciler reconciles a TerraformLayer object
type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Config    *config.Config
	Recorder  record.EventRecorder
	Datastore datastore.Client
	Clock
}

//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformlayers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformlayers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformlayers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.WithContext(ctx)
	log.Infof("starting reconciliation for layer %s/%s ...", req.Namespace, req.Name)
	layer := &configv1alpha1.TerraformLayer{}
	err := r.Client.Get(ctx, req.NamespacedName, layer)
	if errors.IsNotFound(err) {
		log.Errorf("resource not found. Ignoring since object must be deleted: %s", err)
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Errorf("failed to get TerraformLayer: %s", err)
		return ctrl.Result{}, err
	}
	locked, err := lock.IsLayerLocked(ctx, r.Client, layer)
	if err != nil {
		log.Errorf("failed to get Lease Resource: %s", err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, err
	}
	if locked {
		log.Infof("TerraformLayer %s is locked, skipping reconciliation.", layer.Name)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
	}
	repository := &configv1alpha1.TerraformRepository{}
	log.Infof("getting Linked TerraformRepository to layer %s", layer.Name)
	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: layer.Spec.Repository.Namespace,
		Name:      layer.Spec.Repository.Name,
	}, repository)
	if errors.IsNotFound(err) {
		log.Infof("TerraformRepository linked to layer %s not found, ignoring layer until it's modified: %s", layer.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, err
	}
	if err != nil {
		log.Errorf("failed to get TerraformRepository linked to layer %s: %s", layer.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}, err
	}
	err = validateLayerConfig(layer, repository)
	if err != nil {
		r.Recorder.Event(layer, corev1.EventTypeWarning, "Reconciliation", err.Error())
		return ctrl.Result{}, err
	}
	state, conditions := r.GetState(ctx, layer, repository)
	lastResult := []byte("Layer has never been planned")
	if layer.Status.LastRun.Name != "" {
		lastResult, err = r.Datastore.GetPlan(layer.Namespace, layer.Name, layer.Status.LastRun.Name, "", "short")
		if err != nil {
			log.Errorf("failed to get plan for layer %s: %s", layer.Name, err)
			r.Recorder.Event(layer, corev1.EventTypeNormal, "Reconciliation", "Failed to get last Result")
			lastResult = []byte("Error getting last Result")
		}
	}
	result, run := state.getHandler()(ctx, r, layer, repository)
	lastRun := layer.Status.LastRun
	runHistory := layer.Status.LatestRuns
	if run != nil {
		lastRun = getRun(*run)
		runHistory = updateLatestRuns(runHistory, *run, *configv1alpha1.GetRunHistoryPolicy(repository, layer).KeepLastRuns)
	}
	layer.Status = configv1alpha1.TerraformLayerStatus{Conditions: conditions, State: getStateString(state), LastResult: string(lastResult), LastRun: lastRun, LatestRuns: runHistory}
	err = r.Client.Status().Update(ctx, layer)
	if err != nil {
		r.Recorder.Event(layer, corev1.EventTypeWarning, "Reconciliation", "Could not update layer status")
		log.Errorf("could not update layer %s status: %s", layer.Name, err)
	}
	err = r.cleanupRuns(ctx, layer, repository)
	if err != nil {
		log.Warningf("failed to cleanup runs for layer %s: %s", layer.Name, err)
	}
	log.Infof("finished reconciliation cycle for layer %s/%s", layer.Namespace, layer.Name)
	return result, nil
}

func (r *Reconciler) cleanupRuns(ctx context.Context, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) error {
	historyPolicy := configv1alpha1.GetRunHistoryPolicy(repository, layer)
	runs, err := r.getAllRuns(ctx, layer)
	if len(runs) < *historyPolicy.KeepLastRuns {
		log.Infof("no runs to delete for layer %s", layer.Name)
		return nil
	}
	if err != nil {
		return err
	}
	runsToKeep := map[string]bool{}
	for _, run := range layer.Status.LatestRuns {
		runsToKeep[run.Name] = true
	}
	toDelete := []*configv1alpha1.TerraformRun{}
	for _, run := range runs {
		if _, ok := runsToKeep[run.Name]; !ok {
			toDelete = append(toDelete, run)
		}
	}
	if len(toDelete) == 0 {
		log.Infof("no runs to delete for layer %s", layer.Name)
		return nil
	}
	err = deleteAll(ctx, r.Client, toDelete)
	if err != nil {
		return err
	}
	log.Infof("deleted %d runs for layer %s", len(toDelete), layer.Name)
	r.Recorder.Event(layer, corev1.EventTypeNormal, "Reconciliation", "Cleaned up old runs")
	return nil
}

func getRun(run configv1alpha1.TerraformRun) configv1alpha1.TerraformLayerRun {
	return configv1alpha1.TerraformLayerRun{
		Name:   run.Name,
		Commit: "",
		Date:   run.CreationTimestamp,
		Action: run.Spec.Action,
	}
}

func updateLatestRuns(runs []configv1alpha1.TerraformLayerRun, run configv1alpha1.TerraformRun, keep int) []configv1alpha1.TerraformLayerRun {
	oldestRun := &configv1alpha1.TerraformLayerRun{
		Date: metav1.NewTime(time.Now()),
	}
	var oldestRunIndex int
	newRun := getRun(run)
	for i, r := range runs {
		if r.Date.Before(&oldestRun.Date) {
			oldestRun = &r
			oldestRunIndex = i
		}
	}
	if oldestRun == nil || len(runs) < keep {
		return append(runs, newRun)
	}
	rs := append(runs[:oldestRunIndex], runs[oldestRunIndex+1:]...)
	return append(rs, newRun)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Clock = RealClock{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformLayer{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.Config.Controller.MaxConcurrentReconciles}).
		WithEventFilter(ignorePredicate()).
		Complete(r)
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

func validateLayerConfig(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) error {
	if !configv1alpha1.GetTerraformEnabled(repository, layer) && !configv1alpha1.GetOpenTofuEnabled(repository, layer) {
		return e.New("TerraformLayer configuration is invalid: Neither Terraform nor OpenTofu is enabled for this layer. Please enable one of them in the TerraformLayer or the TerraformRepository referenced by the layer.")
	}
	return nil
}

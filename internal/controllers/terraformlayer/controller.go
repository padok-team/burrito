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
	"time"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/lock"
	"github.com/padok-team/burrito/internal/storage"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Scheme  *runtime.Scheme
	Config  *config.Config
	Storage storage.Storage
	Clock
}

//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformlayers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformlayers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformlayers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TerraformLayer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.WithContext(ctx)
	log.Infof("starting reconciliation...")
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
		log.Infof("terraform layer %s is locked, skipping reconciliation.", layer.Name)
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
	err = r.cleanupRuns(ctx, layer, repository)
	if err != nil {
		log.Warningf("failed to cleanup runs for layer %s: %s", layer.Name, err)
	}
	state, conditions := r.GetState(ctx, layer)
	lastResult, err := r.Storage.Get(storage.GenerateKey(storage.LastPlanResult, layer))
	if err != nil {
		lastResult = []byte("Error getting last Result")
	}
	layer.Status = configv1alpha1.TerraformLayerStatus{Conditions: conditions, State: getStateString(state), LastResult: string(lastResult)}
	result := state.getHandler()(ctx, r, layer, repository)
	err = r.Client.Status().Update(ctx, layer)
	if err != nil {
		log.Errorf("could not update layer %s status: %s", layer.Name, err)
	}
	log.Infof("finished reconciliation cycle for layer %s", layer.Name)
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Clock = RealClock{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformLayer{}).
		WithEventFilter(ignorePredicate()).
		Complete(r)
}

func ignorePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

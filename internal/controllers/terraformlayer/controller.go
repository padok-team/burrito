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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

// Reconciler reconciles a TerraformLayer object
type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Config *config.Config
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
	log := log.FromContext(ctx)
	log.Info("Starting reconciliation")
	layer := &configv1alpha1.TerraformLayer{}
	err := r.Client.Get(ctx, req.NamespacedName, layer)
	if errors.IsNotFound(err) {
		log.Info("Resource not found. Ignoring since object must be deleted.")
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "Failed to get TerraformLayer")
		return ctrl.Result{}, err
	}
	deltaOnError, err := time.ParseDuration(r.Config.Controller.Timers.OnError)
	if err != nil {
		log.Error(err, "could not parse timer drift detection period")
		return ctrl.Result{}, err
	}
	locked, err := lock.IsLocked(ctx, r.Client, layer)
	if err != nil {
		log.Error(err, "Failed to get Lease Resource.")
		return ctrl.Result{RequeueAfter: deltaOnError}, err
	}
	deltaWaitAction, err := time.ParseDuration(r.Config.Controller.Timers.WaitAction)
	if err != nil {
		log.Error(err, "could not parse timer wait action period")
		return ctrl.Result{}, err
	}
	if locked {
		log.Info("Layer is locked, skipping reconciliation.")
		return ctrl.Result{RequeueAfter: deltaWaitAction}, nil
	}
	repository := &configv1alpha1.TerraformRepository{}
	log.Info("Getting Linked TerraformRepository")
	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: layer.Spec.Repository.Namespace,
		Name:      layer.Spec.Repository.Name,
	}, repository)
	if errors.IsNotFound(err) {
		log.Info("TerraformRepository not found, ignoring layer until it's modified.")
		return ctrl.Result{RequeueAfter: deltaOnError}, err
	}
	if err != nil {
		log.Error(err, "Failed to get TerraformRepository")
		return ctrl.Result{RequeueAfter: deltaOnError}, err
	}
	state, conditions := r.GetState(ctx, layer)
	layer.Status = configv1alpha1.TerraformLayerStatus{Conditions: conditions}
	result := state.getHandler()(ctx, r, layer, repository)
	err = r.Client.Status().Update(ctx, layer)
	if err != nil {
		log.Error(err, "Could not update resource status")
	}
	log.Info("Finished reconciliation cycle")
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformLayer{}).
		Complete(r)
}

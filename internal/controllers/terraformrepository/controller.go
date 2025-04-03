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

package terraformrepository

import (
	"context"

	"github.com/google/go-cmp/cmp"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/utils/gitprovider"
)

// RepositoryReconciler reconciles a TerraformRepository object
type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Recorder  record.EventRecorder
	Config    *config.Config
	Providers map[string]gitprovider.Provider
	Datastore datastore.Client
}

//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformrepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformrepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformrepositories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TerraformRepository object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.WithContext(ctx)
	log.Infof("starting reconciliation for repository %s/%s ...", req.Namespace, req.Name)

	// fetch the TerraformRepository instance
	repository := &configv1alpha1.TerraformRepository{}
	if err := r.Get(ctx, req.NamespacedName, repository); err != nil {
		log.Errorf("failed to get TerraformRepository: %s", err)
		// If the repository is not found, it might have been deleted
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get the current state and conditions
	state, conditions := r.GetState(ctx, repository)
	stateString := getStateString(state)

	// Update status conditions and state
	repository.Status.Conditions = conditions
	repository.Status.State = stateString

	// Execute the handler
	log.Infof("repository %s/%s is in state %s", repository.Namespace, repository.Name, stateString)
	result, branchStates := state.getHandler()(ctx, r, repository)
	repository.Status.Branches = branchStates
	if err := r.Status().Update(ctx, repository); err != nil {
		r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", "Could not update repository status")
		log.Errorf("failed to update repository status: %s", err)
	}
	log.Infof("finished reconciliation cycle for repository %s/%s", repository.Namespace, repository.Name)
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Providers = make(map[string]gitprovider.Provider)
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformRepository{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.Config.Controller.MaxConcurrentReconciles}).
		Watches(&configv1alpha1.TerraformLayer{}, handler.EnqueueRequestsFromMapFunc(
			func(ctx context.Context, obj client.Object) []reconcile.Request {
				log.Infof("repository controller has detected the following layer creation: %s/%s", obj.GetNamespace(), obj.GetName())
				layer := obj.(*configv1alpha1.TerraformLayer)
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{Namespace: layer.Spec.Repository.Namespace, Name: layer.Spec.Repository.Name}},
				}
			},
		)).
		WithEventFilter(ignorePredicate()).
		Complete(r)
}

func ignorePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates on TerraformLayer objects, we only watch their creation
			if _, ok := e.ObjectNew.(*configv1alpha1.TerraformLayer); ok {
				return false
			}
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

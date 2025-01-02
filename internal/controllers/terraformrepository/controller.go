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

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

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

	// Update status conditions
	repository.Status.Conditions = conditions
	if err := r.Status().Update(ctx, repository); err != nil {
		log.Errorf("failed to update repository status: %s", err)
		return ctrl.Result{}, err
	}

	// Get the handler for the current state
	handler := state.getHandler()

	// Execute the handler
	log.Infof("repository %s/%s is in state %s", repository.Namespace, repository.Name, getStateString(state))
	result, err := handler(ctx, r, repository)
	if err != nil {
		log.Errorf("error handling state %s: %s", getStateString(state), err)
		return ctrl.Result{}, err
	}

	// Update repository status with current state
	repository.Status.State = getStateString(state)
	if err := r.Status().Update(ctx, repository); err != nil {
		log.Errorf("failed to update repository status: %s", err)
		return ctrl.Result{}, err
	}

	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformRepository{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.Config.Controller.MaxConcurrentReconciles}).
		Complete(r)
}

// listManagedRefs returns the list of refs (branches and tags) that are managed by burrito for a specific repository
func (r *Reconciler) listManagedRefs(ctx context.Context, repository *configv1alpha1.TerraformRepository) (map[string]bool, error) {
	// get all layers that depends on the repository (layer.spec.repository.name == repository.name)
	layers := &configv1alpha1.TerraformLayerList{}
	if err := r.List(ctx, layers); err != nil {
		return nil, err
	}
	refs := map[string]bool{}
	for _, layer := range layers.Items {
		if layer.Spec.Repository.Name == repository.Name {
			refs[layer.Spec.Branch] = true
		}
	}
	return refs, nil
}

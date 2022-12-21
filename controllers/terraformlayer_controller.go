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

package controllers

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/padok-team/burrito/burrito/config"
	internal "github.com/padok-team/burrito/cache"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

// TerraformLayerReconciler reconciles a TerraformLayer object
type TerraformLayerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cache  internal.Cache
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
func (r *TerraformLayerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	layer := &configv1alpha1.TerraformLayer{}
	err := r.Client.Get(ctx, req.NamespacedName, layer)
	if errors.IsNotFound(err) {
		log.Log.Info("TerraformLayer resource not found. Ignoring since object must be deleted.")
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Log.Error(err, "Failed to get TerraformLayer.")
		return ctrl.Result{}, err
	}
	log.Log.Info("Reconciling TerraformLayer")
	repository := &configv1alpha1.TerraformRepository{}
	log.Log.Info("Getting Linked TerraformRepository")
	err = r.Client.Get(ctx, types.NamespacedName{
		Namespace: layer.Spec.Repository.Namespace,
		Name:      layer.Spec.Repository.Name,
	}, repository)
	if errors.IsNotFound(err) {
		log.Log.Info("TerraformRepository linked to TerraformLayer not found, ignoring layer until it's modified.")
		return ctrl.Result{RequeueAfter: time.Minute * 2}, err
	}
	if err != nil {
		log.Log.Error(err, "Failed to get TerraformRepository")
		return ctrl.Result{}, err
	}
	c := TerraformLayerConditions{Resource: layer, Cache: &r.Cache, Repository: repository}
	log.Log.Info("Evaluating Conditions")
	evalFunc, conditions := c.Evaluate()
	layer.Status = configv1alpha1.TerraformLayerStatus{Conditions: conditions}
	r.Client.Status().Update(ctx, layer)
	return evalFunc(ctx, r.Client), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TerraformLayerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Cache = internal.NewRedisCache(r.Config.Redis.URL, r.Config.Redis.Password, r.Config.Redis.Database)
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformLayer{}).
		Complete(r)
}

const (
	IsRunning              = "IsTerraformRunning"
	IsPlanArtifactUpToDate = "IsPlanArtifactUpToDate"
	IsApplyUpToDate        = "IsApplyUpToDate"
	HasFailed              = "HasTerraformFailed"
)

type TerraformLayerConditions struct {
	IsRunning              TerraformRunning
	IsPlanArtifactUpToDate TerraformPlanArtifactUpToDate
	IsApplyUpToDate        TerraformApplyUpToDate
	HasFailed              TerraformFailure
	Cache                  *internal.Cache
	Resource               *configv1alpha1.TerraformLayer
	Repository             *configv1alpha1.TerraformRepository
}

func (t *TerraformLayerConditions) Evaluate() (func(ctx context.Context, c client.Client) ctrl.Result, []metav1.Condition) {
	isTerraformRunning := t.IsRunning.Evaluate(*t.Cache, t.Resource)
	isPlanArtifactUpToDate := t.IsPlanArtifactUpToDate.Evaluate(*t.Cache, t.Resource)
	isApplyUpToDate := t.IsApplyUpToDate.Evaluate(*t.Cache, t.Resource)
	hasTerraformFailed := t.HasFailed.Evaluate(*t.Cache, t.Resource)
	conditions := []metav1.Condition{t.IsRunning.Condition, t.IsPlanArtifactUpToDate.Condition, t.IsApplyUpToDate.Condition, t.HasFailed.Condition}
	cache := *t.Cache
	switch {
	case isTerraformRunning:
		log.Log.Info("Terraform is already running on this layer, skipping")
		return func(ctx context.Context, c client.Client) ctrl.Result {
			return ctrl.Result{RequeueAfter: time.Minute * 2}
		}, conditions
	case !isTerraformRunning && isPlanArtifactUpToDate && isApplyUpToDate:
		log.Log.Info("Layer has not drifted")
		return func(ctx context.Context, c client.Client) ctrl.Result {
			return ctrl.Result{RequeueAfter: time.Minute * 20}
		}, conditions
	case !isTerraformRunning && isPlanArtifactUpToDate && !isApplyUpToDate && hasTerraformFailed:
		log.Log.Info("Layer needs to be applied but previous apply failed, launching a new runner")
		return func(ctx context.Context, c client.Client) ctrl.Result {
			pod := getPod(t.Resource, t.Repository, "apply")
			err := cache.Set(internal.GenerateKey(internal.Lock, t.Resource), []byte("1"), 0)
			if err != nil {
				log.Log.Error(err, "[TerraformApplyHasFailedPreviously] failed to create lock in cache")
				// TODO: time to requeue
				return ctrl.Result{}
			}
			err = c.Create(ctx, &pod)
			if err != nil {
				log.Log.Error(err, "[TerraformApplyHasFailedPreviously] Failed to create pod for Apply action")
				cache.Delete(internal.GenerateKey(internal.Lock, t.Resource))
			}
			//TODO: Implement Exponential backoff
			return ctrl.Result{RequeueAfter: time.Minute * 2}
		}, conditions
	case !isTerraformRunning && isPlanArtifactUpToDate && !isApplyUpToDate && !hasTerraformFailed:
		log.Log.Info("Layer needs to be applied, launching a new runner")
		return func(ctx context.Context, c client.Client) ctrl.Result {
			pod := getPod(t.Resource, t.Repository, "apply")
			err := cache.Set(internal.GenerateKey(internal.Lock, t.Resource), []byte("1"), 0)
			if err != nil {
				log.Log.Error(err, "[TerraformApplyNeeded] failed to create lock in cache")
				// TODO: time to requeue
				return ctrl.Result{}
			}
			err = c.Create(ctx, &pod)
			if err != nil {
				log.Log.Error(err, "[TerraformApplyNeeded] Failed to create pod for Apply action")
				cache.Delete(internal.GenerateKey(internal.Lock, t.Resource))
			}
			return ctrl.Result{RequeueAfter: time.Minute * 20}
		}, conditions
	case !isTerraformRunning && !isPlanArtifactUpToDate && hasTerraformFailed:
		log.Log.Info("Layer needs to be planned but previous plan failed, launching a new runner")
		return func(ctx context.Context, c client.Client) ctrl.Result {
			pod := getPod(t.Resource, t.Repository, "plan")
			err := cache.Set(internal.GenerateKey(internal.Lock, t.Resource), []byte("1"), 0)
			if err != nil {
				log.Log.Error(err, "[TerraformPlanHasFailedPreviously] failed to create lock in cache")
				// TODO: time to requeue
				return ctrl.Result{}
			}
			err = c.Create(ctx, &pod)
			if err != nil {
				log.Log.Error(err, "[TerraformPlanHasFailedPreviously] Failed to create pod for Plan action")
				cache.Delete(internal.GenerateKey(internal.Lock, t.Resource))
			}
			//TODO: Implement Exponential backoff
			return ctrl.Result{RequeueAfter: time.Minute * 2}
		}, conditions
	case !isTerraformRunning && !isPlanArtifactUpToDate && !hasTerraformFailed:
		log.Log.Info("Layer needs to be planned, launching a new runner")
		return func(ctx context.Context, c client.Client) ctrl.Result {
			pod := getPod(t.Resource, t.Repository, "plan")
			err := cache.Set(internal.GenerateKey(internal.Lock, t.Resource), []byte("1"), 0)
			if err != nil {
				log.Log.Error(err, "[TerraformPlanNeeded] failed to create lock in cache")
				// TODO: time to requeue
				return ctrl.Result{}
			}
			err = c.Create(ctx, &pod)
			if err != nil {
				log.Log.Error(err, "[TerraformPlanNeeded] Failed to create pod for Plan action")
				cache.Delete(internal.GenerateKey(internal.Lock, t.Resource))
			}
			return ctrl.Result{RequeueAfter: time.Minute * 20}
		}, conditions
	default:
		log.Log.Info("This controller is drunk")
		return func(ctx context.Context, c client.Client) ctrl.Result {
			//TODO: Add Log -> This should not have happened
			return ctrl.Result{}
		}, conditions
	}
}

type TerraformRunning struct {
	Condition metav1.Condition
}

func (c *TerraformRunning) Evaluate(cache internal.Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Condition = metav1.Condition{
		Type:               IsRunning,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
	}
	key := internal.GenerateKey(internal.Lock, t)
	_, err := cache.Get(key)
	if err != nil {
		c.Condition.Reason = "NoLockInCache"
		c.Condition.Message = "No lock has been found in Cache. Terraform is not running on this layer."
		c.Condition.Status = metav1.ConditionFalse
		return false
	}
	c.Condition.Reason = "LockInCache"
	c.Condition.Message = "Lock has been found in Cache. Terraform is already running on this layer."
	c.Condition.Status = metav1.ConditionTrue
	return true
}

type TerraformPlanArtifactUpToDate struct {
	Condition metav1.Condition
}

func (c *TerraformPlanArtifactUpToDate) Evaluate(cache internal.Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Condition = metav1.Condition{
		Type:               IsPlanArtifactUpToDate,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
	}
	key := internal.GenerateKey(internal.LastPlanDate, t)
	value, err := cache.Get(key)
	if err != nil {
		c.Condition.Reason = "NoTimestampInCache"
		c.Condition.Message = "The last plan date is not in cache."
		c.Condition.Status = metav1.ConditionFalse
		return false
	}
	unixTimestamp, _ := strconv.ParseInt(string(value), 10, 64)
	lastPlanDate := time.Unix(unixTimestamp, 0)
	nextPlanDate := lastPlanDate.Add(20 * time.Minute)
	now := time.Now()
	if nextPlanDate.After(now) {
		c.Condition.Reason = "PlanIsRecent"
		c.Condition.Message = "The plan has been made less than 20 minutes ago."
		c.Condition.Status = metav1.ConditionTrue
		return true
	}
	c.Condition.Reason = "PlanIsTooOld"
	c.Condition.Message = "The plan has been made more than 20 minutes ago."
	c.Condition.Status = metav1.ConditionFalse
	return false
}

type TerraformApplyUpToDate struct {
	Condition metav1.Condition
}

func (c *TerraformApplyUpToDate) Evaluate(cache internal.Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Condition = metav1.Condition{
		Type:               IsApplyUpToDate,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
	}
	key := internal.GenerateKey(internal.LastPlannedArtifact, t)
	planHash, err := cache.Get(key)
	if err != nil {
		c.Condition.Reason = "NoPlanYet"
		c.Condition.Message = "No plan has run yet, Layer might be new"
		c.Condition.Status = metav1.ConditionTrue
		return true
	}
	key = internal.GenerateKey(internal.LastAppliedArtifact, t)
	applyHash, err := cache.Get(key)
	if err != nil {
		c.Condition.Reason = "NoApplyHasRan"
		c.Condition.Message = "Apply has not ran yet but a plan is available, launching apply"
		c.Condition.Status = metav1.ConditionFalse
		return false
	}
	if bytes.Compare(planHash, applyHash) != 0 {
		c.Condition.Reason = "NewPlanAvailable"
		c.Condition.Message = "Apply will run."
		c.Condition.Status = metav1.ConditionFalse
		return false
	}
	c.Condition.Reason = "ApplyUpToDate"
	c.Condition.Message = "Last planned artifact is the same as the last applied one"
	c.Condition.Status = metav1.ConditionTrue
	return true
}

type TerraformFailure struct {
	Condition metav1.Condition
}

func (c *TerraformFailure) Evaluate(cache internal.Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Condition = metav1.Condition{
		Type:               HasFailed,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
	}
	key := internal.GenerateKey(internal.RunResult, t)
	result, err := cache.Get(key)
	if err != nil {
		c.Condition.Reason = "NoRunYet"
		c.Condition.Message = "Terraform has not ran yet"
		c.Condition.Status = metav1.ConditionFalse
		return false
	}
	if string(result) == "0" {
		c.Condition.Reason = "RunExitedGracefully"
		c.Condition.Message = "Last run exited gracefully"
		c.Condition.Status = metav1.ConditionFalse
		return false
	}
	c.Condition.Status = metav1.ConditionTrue
	key = internal.GenerateKey(internal.RunMessage, t)
	message, err := cache.Get(key)
	if err != nil {
		c.Condition.Reason = "UnexpectedRunnerFailure"
		c.Condition.Message = "Terraform runner might have crashed before updating Redis"
	}
	c.Condition.Reason = "TerraformRunFailure"
	c.Condition.Message = string(message)
	return true
}

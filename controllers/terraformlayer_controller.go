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
	"fmt"
	"hash/fnv"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

// TerraformLayerReconciler reconciles a TerraformLayer object
type TerraformLayerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cache  Cache
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
	err := r.Client.Get(context.TODO(), req.NamespacedName, layer)
	if errors.IsNotFound(err) {
		log.Log.Info("TerraformLayer resource not found. Ignoring since object must be deleted.")
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Log.Error(err, "Failed to get TerraformLayer.")
		return ctrl.Result{}, err
	}
	c := TerraformLayerConditions{Resource: layer, Cache: &r.Cache}
	evalFunc, conditions := c.Evaluate()
	layer.Status = configv1alpha1.TerraformLayerStatus{Conditions: conditions}
	r.Client.Status().Update(context.TODO(), layer)
	return evalFunc(), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TerraformLayerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Cache = newMemoryCache()
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformLayer{}).
		Complete(r)
}

const (
	RunningCondition = "IsTerraformRunning"
	PlanArtifact     = "IsPlanArtifactUpToDate"
	ApplyUpToDate    = "IsApplyUpToDate"
	TerraformFailure = "HasTerraformFailed"

	CachePrefixLock                = "lock-"
	CachePrefixLastPlanDate        = "lastPlandate"
	CachePrefixLastPlannedArtifact = "lastPlannedArtifact-"
	CachePrefixLastAppliedArtifact = "lastApplyArtifact-"
	CachePrefixRunResult           = "runResult-"
	CachePrefixRunMessage          = "runMessage-"
)

type TerraformLayerConditions struct {
	RunningCondition TerraformRunningCondition
	PlanArtifact     TerraformPlanArtifactCondition
	ApplyUpToDate    TerraformApplyUpToDateCondition
	TerraformFailure TerraformFailureCondition
	Cache            *Cache
	Resource         *configv1alpha1.TerraformLayer
}

func (t *TerraformLayerConditions) Evaluate() (func() ctrl.Result, []metav1.Condition) {
	isTerraformRunning := t.RunningCondition.Evaluate(*t.Cache, t.Resource)
	isPlanArtifactUpToDate := t.PlanArtifact.Evaluate(*t.Cache, t.Resource)
	isApplyUpToDate := t.ApplyUpToDate.Evaluate(*t.Cache, t.Resource)
	hasTerraformFailed := t.TerraformFailure.Evaluate(*t.Cache, t.Resource)
	conditions := []metav1.Condition{t.RunningCondition.Status, t.PlanArtifact.Status, t.ApplyUpToDate.Status, t.TerraformFailure.Status}
	switch {
	case isTerraformRunning:
		return func() ctrl.Result {
			return ctrl.Result{RequeueAfter: time.Minute * 2}
		}, conditions
	case !isTerraformRunning && isPlanArtifactUpToDate && isApplyUpToDate:
		return func() ctrl.Result {
			return ctrl.Result{RequeueAfter: time.Minute * 20}
		}, conditions
	case !isTerraformRunning && isPlanArtifactUpToDate && !isApplyUpToDate && hasTerraformFailed:
		return func() ctrl.Result {
			//TODO: Launch Apply
			//TODO: Implement Exponential backoff
			return ctrl.Result{}
		}, conditions
	case !isTerraformRunning && isPlanArtifactUpToDate && !isApplyUpToDate && !hasTerraformFailed:
		return func() ctrl.Result {
			//TODO: Launch Apply
			return ctrl.Result{RequeueAfter: time.Minute * 20}
		}, conditions
	case !isTerraformRunning && !isPlanArtifactUpToDate && hasTerraformFailed:
		return func() ctrl.Result {
			//TODO: Launch Plan
			//TODO: Implement Exponential backoff
			return ctrl.Result{}
		}, conditions
	case !isTerraformRunning && !isPlanArtifactUpToDate && !hasTerraformFailed:
		return func() ctrl.Result {
			//TODO: Launch Plan
			return ctrl.Result{RequeueAfter: time.Minute * 20}
		}, conditions
	default:
		return func() ctrl.Result {
			//TODO: Add Log -> This should not have happened
			return ctrl.Result{}
		}, conditions
	}
}

func computeHash(s ...string) string {
	beforeHash := ""
	strings.Join(s, beforeHash)
	h := fnv.New32a()
	h.Write([]byte(beforeHash))
	return fmt.Sprint(h.Sum32())
}

type TerraformCondition interface {
	Evaluate(cache Cache, t *configv1alpha1.TerraformLayer) bool
}

type TerraformRunningCondition struct {
	Status metav1.Condition
}

func (c *TerraformRunningCondition) Evaluate(cache Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Status = metav1.Condition{
		Type:               TerraformFailure,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
	}
	key := CachePrefixLock + computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path)
	_, err := cache.Get(key)
	if err != nil {
		c.Status.Reason = "NoLockInCache"
		c.Status.Message = "No lock has been found in Cache. Terraform is not running on this layer."
		c.Status.Status = metav1.ConditionFalse
		return false
	}
	c.Status.Reason = "LockInCache"
	c.Status.Message = "Lock has been found in Cache. Terraform is already running on this layer."
	c.Status.Status = metav1.ConditionTrue
	return true
}

type TerraformPlanArtifactCondition struct {
	Status metav1.Condition
}

func (c *TerraformPlanArtifactCondition) Evaluate(cache Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Status = metav1.Condition{
		Type:               TerraformFailure,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
	}
	key := CachePrefixLastPlanDate + computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path, t.Spec.Branch)
	value, err := cache.Get(key)
	if err != nil {
		c.Status.Reason = "NoTimestampInCache"
		c.Status.Message = "The last plan date is not in cache."
		c.Status.Status = metav1.ConditionFalse
		return false
	}
	unixTimestamp, _ := strconv.ParseInt(string(value), 10, 64)
	lastPlanDate := time.Unix(unixTimestamp, 0)
	nextPlanDate := lastPlanDate.Add(20 * time.Minute)
	now := time.Now()
	if nextPlanDate.After(now) {
		c.Status.Reason = "PlanIsRecent"
		c.Status.Message = "The plan has been made less than 20 minutes ago."
		c.Status.Status = metav1.ConditionTrue
		return true
	}
	c.Status.Reason = "PlanIsTooOld"
	c.Status.Message = "The plan has been made more than 20 minutes ago."
	c.Status.Status = metav1.ConditionFalse
	return false
}

type TerraformApplyUpToDateCondition struct {
	Status metav1.Condition
}

func (c *TerraformApplyUpToDateCondition) Evaluate(cache Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Status = metav1.Condition{
		Type:               ApplyUpToDate,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionTrue,
	}
	status := true
	key := CachePrefixLastPlannedArtifact + computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path, t.Spec.Branch)
	planHash, err := cache.Get(key)
	if err != nil {
		c.Status.Reason = "NoPlanYet"
		c.Status.Message = "No plan has run yet, Layer might be new"
		return status
	}
	key = CachePrefixLastAppliedArtifact + computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path)
	applyHash, err := cache.Get(key)
	if err != nil {
		c.Status.Reason = "NoApplyHasRan"
		c.Status.Message = "Apply has not ran yet but a plan is available, launching apply"
		status = false
	}
	if bytes.Compare(planHash, applyHash) != 0 {
		c.Status.Reason = "NewPlanAvailable"
		c.Status.Message = "Apply will run."
		status = false
	}
	return status
}

type TerraformFailureCondition struct {
	Status metav1.Condition
}

func (c *TerraformFailureCondition) Evaluate(cache Cache, t *configv1alpha1.TerraformLayer) bool {
	c.Status = metav1.Condition{
		Type:               TerraformFailure,
		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionFalse,
	}
	status := false
	key := CachePrefixRunResult + computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path, t.Spec.Branch)
	result, err := cache.Get(key)
	if err != nil {
		c.Status.Reason = "NoRunYet"
		c.Status.Message = "Terraform has not ran yet"
		return status
	}
	if string(result) == "1" {
		c.Status.Status = metav1.ConditionTrue
		status = true
	}
	key = CachePrefixRunMessage + computeHash(t.Spec.Repository.Name, t.Spec.Repository.Namespace, t.Spec.Path, t.Spec.Branch)
	message, err := cache.Get(key)
	if err != nil {
		c.Status.Reason = "UnexpectedRunnerFailure"
		c.Status.Message = "Terraform runner might have crashed before updating Redis"
	}
	c.Status.Message = string(message)
	return status
}

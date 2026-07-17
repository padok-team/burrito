package annotations

import (
	"context"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LastApplySum    string = "runner.terraform.padok.cloud/apply-sum"
	LastApplyDate   string = "runner.terraform.padok.cloud/apply-date"
	LastApplyCommit string = "runner.terraform.padok.cloud/apply-commit"
	// LastApplyRun    string = "runner.terraform.padok.cloud/apply-run"
	LastPlanCommit string = "runner.terraform.padok.cloud/plan-commit"
	LastPlanDate   string = "runner.terraform.padok.cloud/plan-date"
	LastPlanSum    string = "runner.terraform.padok.cloud/plan-sum"
	LastPlanRun    string = "runner.terraform.padok.cloud/plan-run"
	Lock           string = "runner.terraform.padok.cloud/lock"

	LastBranchCommit       string = "webhook.terraform.padok.cloud/branch-commit"
	LastBranchCommitDate   string = "webhook.terraform.padok.cloud/branch-commit-date"
	LastRelevantCommit     string = "webhook.terraform.padok.cloud/relevant-commit"
	LastRelevantCommitDate string = "webhook.terraform.padok.cloud/relevant-commit-date"
	SyncBranchNow          string = "webhook.terraform.padok.cloud/sync-"

	ForceApply              string = "notifications.terraform.padok.cloud/force-apply"
	AdditionnalTriggerPaths string = "config.terraform.padok.cloud/additionnal-trigger-paths"

	SyncNow        string = "api.terraform.padok.cloud/sync-now"
	ApplyNow       string = "api.terraform.padok.cloud/apply-now"
	AllowedTenants string = "credentials.terraform.padok.cloud/allowed-tenants"
)

func ComputeKeyForSyncBranchNow(branch string) string {
	return SyncBranchNow + strings.ReplaceAll(branch, "/", "--")
}

func Add(ctx context.Context, c client.Client, obj client.Object, annotations map[string]string) error {
	newObj := obj.DeepCopyObject().(client.Object)
	patch := client.MergeFrom(newObj)
	currentAnnotations := obj.GetAnnotations()
	if currentAnnotations == nil {
		currentAnnotations = make(map[string]string)
	}
	for k, v := range annotations {
		currentAnnotations[k] = v
	}

	obj.SetAnnotations(currentAnnotations)
	return c.Patch(ctx, obj, patch)
}

func Remove(ctx context.Context, c client.Client, obj client.Object, annotation string) error {
	newObj := obj.DeepCopyObject().(client.Object)
	patch := client.MergeFrom(newObj)
	annotations := obj.GetAnnotations()
	delete(annotations, annotation)
	obj.SetAnnotations(annotations)
	return c.Patch(ctx, obj, patch)
}

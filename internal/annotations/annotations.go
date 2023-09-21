package annotations

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LastApplySum    string = "runner.terraform.padok.cloud/apply-sum"
	LastApplyDate   string = "runner.terraform.padok.cloud/apply-date"
	LastApplyCommit string = "runner.terraform.padok.cloud/apply-commit"
	LastPlanCommit  string = "runner.terraform.padok.cloud/plan-commit"
	LastPlanDate    string = "runner.terraform.padok.cloud/plan-date"
	LastPlanSum     string = "runner.terraform.padok.cloud/plan-sum"
	Lock            string = "runner.terraform.padok.cloud/lock"

	LastBranchCommit   string = "webhook.terraform.padok.cloud/branch-commit"
	LastRelevantCommit string = "webhook.terraform.padok.cloud/relevant-commit"

	ForceApply string = "notifications.terraform.padok.cloud/force-apply"

	LastDiscoveredCommit string = "pullrequest.terraform.padok.cloud/last-discovered-commit"
	LastCommentedCommit  string = "pullrequest.terraform.padok.cloud/last-commented-commit"

	AdditionnalTriggerPaths string = "config.terraform.padok.cloud/additionnal-trigger-paths"
	PlanRunRetention        string = "config.terraform.padok.cloud/plan-run-retention"
	ApplyRunRetention       string = "config.terraform.padok.cloud/apply-run-retention"
)

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

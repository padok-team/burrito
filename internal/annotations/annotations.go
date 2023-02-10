package annotations

import (
	"context"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LastApplySum    string = "runner.terraform.padok.cloud/apply-sum"
	LastApplyDate   string = "runner.terraform.padok.cloud/apply-date"
	LastApplyCommit string = "runner.terraform.padok.cloud/apply-commit"
	LastPlanCommit  string = "runner.terraform.padok.cloud/plan-commit"
	LastPlanDate    string = "runner.terraform.padok.cloud/plan-date"
	LastPlanSum     string = "runner.terraform.padok.cloud/plan-sum"
	Failure         string = "runner.terraform.padok.cloud/failure"
	Lock            string = "runner.terraform.padok.cloud/lock"

	LastBranchCommit     string = "webhook.terraform.padok.cloud/branch-commit"
	LastConcerningCommit string = "webhook.terraform.padok.cloud/concerning-commit"

	ForceApply string = "notifications.terraform.padok.cloud/force-apply"
)

func Add(ctx context.Context, c client.Client, obj configv1alpha1.TerraformLayer, annotations map[string]string) error {
	patch := client.MergeFrom(obj.DeepCopy())
	currentAnnotations := obj.GetAnnotations()
	for k, v := range annotations {
		currentAnnotations[k] = v
	}
	obj.SetAnnotations(currentAnnotations)
	return c.Patch(ctx, &obj, patch)
}

func Remove(ctx context.Context, c client.Client, obj configv1alpha1.TerraformLayer, annotation string) error {
	patch := client.MergeFrom(obj.DeepCopy())
	annotations := obj.GetAnnotations()
	delete(annotations, annotation)
	obj.SetAnnotations(annotations)
	return c.Patch(ctx, &obj, patch)
}

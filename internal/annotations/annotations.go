package annotations

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LastBranchCommit   string = "repository.terraform.padok.cloud/branch-commit"
	LastRelevantCommit string = "repository.terraform.padok.cloud/relevant-commit"

	ForceApply              string = "notifications.terraform.padok.cloud/force-apply"
	AdditionnalTriggerPaths string = "config.terraform.padok.cloud/additionnal-trigger-paths"
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

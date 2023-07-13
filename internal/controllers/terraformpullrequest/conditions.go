package terraformpullrequest

import (
	"context"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *Reconciler) IsLastCommitDiscovered(pr *configv1alpha1.TerraformPullRequest) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsLastCommitDiscovered",
		ObservedGeneration: pr.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	lastDiscoveredCommit, ok := pr.Annotations[annotations.LastDiscoveredCommit]
	if !ok {
		condition.Reason = "NoCommitDiscovered"
		condition.Message = "Controller hasn't discovered any commit yet."
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	lastBranchCommit, ok := pr.Annotations[annotations.LastBranchCommit]
	if !ok {
		condition.Reason = "UnknownLastBranchCommit"
		condition.Message = "This should have happened"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if lastDiscoveredCommit == lastBranchCommit {
		condition.Reason = "LastCommitDiscovered"
		condition.Message = "The last commit has been discovered."
		condition.Status = metav1.ConditionTrue
		return condition, true
	}
	condition.Reason = "LastCommitNotDiscovered"
	condition.Message = "Last received commit is not the last discovered commit."
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) AreLayersStillPlanning(pr *configv1alpha1.TerraformPullRequest) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "AreLayersStillPlanning",
		ObservedGeneration: pr.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	layers, err := GetLinkedLayers(r.Client, pr)

	lastDiscoveredCommit, okDiscoveredCommit := pr.Annotations[annotations.LastDiscoveredCommit]
	prLastBranchCommit, okPRBranchCommit := pr.Annotations[annotations.LastBranchCommit]

	if !okPRBranchCommit {
		condition.Reason = "NoBranchCommitOnPR"
		condition.Message = "This should not have happened, report this as an issue"
		condition.Status = metav1.ConditionUnknown
		return condition, true
	}

	if !okDiscoveredCommit {
		condition.Reason = "NoCommitDiscovered"
		condition.Message = "Controller hasn't discovered any commit yet."
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	if lastDiscoveredCommit != prLastBranchCommit {
		condition.Reason = "StillNeedsDiscovery"
		condition.Message = "Controller hasn't discovered the latest commit yet."
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	if err != nil {
		condition.Reason = "ErrorListingLayers"
		condition.Message = err.Error()
		condition.Status = metav1.ConditionTrue
		return condition, true
	}

	for _, layer := range layers {
		lastRelevantCommit, okRelevantCommit := layer.Annotations[annotations.LastRelevantCommit]
		lastPlanCommit, okPlanCommit := layer.Annotations[annotations.LastPlanCommit]
		condition.Reason = "LayersStillPlanning"
		condition.Message = "Linked layers are still planning."
		condition.Status = metav1.ConditionTrue
		if !okPlanCommit {
			return condition, true
		}
		if !okRelevantCommit {
			condition.Reason = "NoRelevantCommitOnLayer"
			condition.Message = "This should not have happened, report this as an issue"
			condition.Status = metav1.ConditionUnknown
			return condition, true
		}
		if lastPlanCommit == lastRelevantCommit {
			continue
		}
		return condition, true
	}
	condition.Reason = "LayersNotPlanning"
	condition.Message = "Linked layers are not planning."
	condition.Status = metav1.ConditionFalse
	return condition, false
}

func (r *Reconciler) IsCommentUpToDate(pr *configv1alpha1.TerraformPullRequest) (metav1.Condition, bool) {
	condition := metav1.Condition{
		Type:               "IsCommentUpToDate",
		ObservedGeneration: pr.GetObjectMeta().GetGeneration(),
		Status:             metav1.ConditionUnknown,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	lastDiscoveredCommit, ok := pr.Annotations[annotations.LastDiscoveredCommit]
	if !ok {
		condition.Reason = "Unknown"
		condition.Message = "This should not have happened"
		condition.Status = metav1.ConditionUnknown
		return condition, true
	}
	lastCommentedCommit, ok := pr.Annotations[annotations.LastCommentedCommit]
	if !ok {
		condition.Reason = "NoCommentSent"
		condition.Message = "No comment has ever been sent"
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	if lastCommentedCommit != lastDiscoveredCommit {
		condition.Reason = "CommentOutdated"
		condition.Message = "The comment is outdated."
		condition.Status = metav1.ConditionFalse
		return condition, false
	}
	condition.Reason = "CommentUpToDate"
	condition.Message = "The comment is up to date."
	condition.Status = metav1.ConditionTrue
	return condition, true
}

func GetLinkedLayers(cl client.Client, pr *configv1alpha1.TerraformPullRequest) ([]configv1alpha1.TerraformLayer, error) {
	layers := configv1alpha1.TerraformLayerList{}
	requirement, err := labels.NewRequirement("burrito/managed-by", selection.Equals, []string{pr.Name})
	if err != nil {
		return nil, err
	}
	selector := labels.NewSelector().Add(*requirement)
	err = cl.List(context.TODO(), &layers, client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, err
	}
	return layers.Items, nil
}

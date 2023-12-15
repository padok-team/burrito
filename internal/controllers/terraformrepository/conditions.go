package terraformrepository

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

// func (r *Reconciler) IsUpToDate(t *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
// 	condition := metav1.Condition{
// 		Type:               "IsPlanArtifactUpToDate",
// 		ObservedGeneration: t.GetObjectMeta().GetGeneration(),
// 		Status:             metav1.ConditionUnknown,
// 		LastTransitionTime: metav1.NewTime(time.Now()),
// 	}

// 	return condition, false
// }

func (r *Reconciler) IsLastCloneTooOld(repository *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{}
	return condition, true
}

func (r *Reconciler) WasABranchRecentlyUpdated(repository *configv1alpha1.TerraformRepository) (metav1.Condition, bool) {
	condition := metav1.Condition{}
	return condition, true
}

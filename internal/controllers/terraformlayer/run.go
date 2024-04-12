package terraformlayer

import (
	"context"
	"fmt"
	"sort"
	"sync"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Action string

const (
	PlanAction  Action = "plan"
	ApplyAction Action = "apply"
)

func GetDefaultLabels(layer *configv1alpha1.TerraformLayer) map[string]string {
	return map[string]string{
		"burrito/managed-by": layer.Name,
	}
}

func (r *Reconciler) getRun(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, action Action) configv1alpha1.TerraformRun {
	return configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", layer.Name, action),
			Namespace:    layer.Namespace,
			Labels:       GetDefaultLabels(layer),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: layer.GetAPIVersion(),
					Kind:       layer.GetKind(),
					Name:       layer.Name,
					UID:        layer.UID,
				},
			},
		},
		Spec: configv1alpha1.TerraformRunSpec{
			Action: string(action),
			Layer: configv1alpha1.TerraformRunLayer{
				Name:      layer.Name,
				Namespace: layer.Namespace,
			},
		},
	}
}

func (r *Reconciler) getAllFinishedRuns(ctx context.Context, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) ([]*configv1alpha1.TerraformRun, error) {
	list := &configv1alpha1.TerraformRunList{}
	labelSelector := labels.NewSelector()
	for key, value := range GetDefaultLabels(layer) {
		requirement, err := labels.NewRequirement(key, selection.Equals, []string{value})
		if err != nil {
			return []*configv1alpha1.TerraformRun{}, err
		}
		labelSelector = labelSelector.Add(*requirement)
	}
	err := r.Client.List(
		ctx,
		list,
		client.MatchingLabelsSelector{Selector: labelSelector},
		&client.ListOptions{Namespace: layer.Namespace},
	)
	if err != nil {
		return []*configv1alpha1.TerraformRun{}, err
	}

	// Keep only runs with state Succeeded or Failed
	var runs []*configv1alpha1.TerraformRun
	for _, run := range list.Items {
		if run.Status.State == "Succeeded" || run.Status.State == "Failed" {
			runs = append(runs, &run)
		}
	}
	return runs, nil
}

func deleteAll(ctx context.Context, c client.Client, objs []*configv1alpha1.TerraformRun) error {
	var wg sync.WaitGroup
	errorCh := make(chan error, len(objs))

	deleteObject := func(obj *configv1alpha1.TerraformRun) {
		defer wg.Done()
		err := c.Delete(ctx, obj)
		if err != nil {
			errorCh <- fmt.Errorf("error deleting %s: %v", obj.Name, err)
		} else {
			log.Infof("deleted run %s", obj.Name)
		}
	}

	for _, obj := range objs {
		wg.Add(1)
		go deleteObject(obj)
	}

	go func() {
		wg.Wait()
		close(errorCh)
	}()

	var ret error = nil
	for err := range errorCh {
		if err != nil {
			ret = err
		}
	}

	return ret
}

func (r *Reconciler) cleanupRuns(ctx context.Context, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) error {
	historyPolicy := configv1alpha1.GetRunHistoryPolicy(repository, layer)

	runs, err := r.getAllFinishedRuns(ctx, layer, repository)
	if err != nil {
		return err
	}
	sortedRuns := sortAndSplitRunsByAction(runs)
	toDelete := []*configv1alpha1.TerraformRun{}
	if len(sortedRuns[string(PlanAction)]) <= *historyPolicy.KeepLastPlanRuns {
		log.Infof("no plan runs to delete for layer %s", layer.Name)
	} else {
		toDelete = append(toDelete, sortedRuns[string(PlanAction)][:len(sortedRuns[string(PlanAction)])-*historyPolicy.KeepLastPlanRuns]...)
	}
	if len(sortedRuns[string(ApplyAction)]) <= *historyPolicy.KeepLastApplyRuns {
		log.Infof("no apply runs to delete for layer %s", layer.Name)
	} else {
		toDelete = append(toDelete, sortedRuns[string(ApplyAction)][:len(sortedRuns[string(ApplyAction)])-*historyPolicy.KeepLastApplyRuns]...)
	}
	if len(toDelete) == 0 {
		log.Infof("no runs to delete for layer %s", layer.Name)
		return nil
	}
	err = deleteAll(ctx, r.Client, toDelete)
	if err != nil {
		return err
	}
	log.Infof("deleted %d runs for layer %s", len(toDelete), layer.Name)
	r.Recorder.Event(layer, corev1.EventTypeNormal, "Reconciliation", "Cleaned up old runs")
	return nil
}

func sortAndSplitRunsByAction(runs []*configv1alpha1.TerraformRun) map[string][]*configv1alpha1.TerraformRun {
	splittedRuns := map[string][]*configv1alpha1.TerraformRun{}
	for _, run := range runs {
		if _, ok := splittedRuns[run.Spec.Action]; !ok {
			splittedRuns[run.Spec.Action] = []*configv1alpha1.TerraformRun{}
		}
		splittedRuns[run.Spec.Action] = append(splittedRuns[run.Spec.Action], run)
	}
	for action := range splittedRuns {
		sort.Slice(splittedRuns[action], func(i, j int) bool {
			return splittedRuns[action][i].CreationTimestamp.Before(&splittedRuns[action][j].CreationTimestamp)
		})
	}
	return splittedRuns
}

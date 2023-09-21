package terraformlayer

import (
	"context"
	"fmt"
	"sync"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	log "github.com/sirupsen/logrus"
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

type runRetention struct {
	plan  time.Duration
	apply time.Duration
}

func getRetentions(layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) (runRetention, error) {
	var planRet time.Duration
	var applyRet time.Duration
	var err error

	if ann, ok := layer.Annotations[annotations.PlanRunRetention]; ok {
		planRet, err = time.ParseDuration(ann)
	} else if ann, ok := repository.Annotations[annotations.PlanRunRetention]; ok {
		planRet, err = time.ParseDuration(ann)
	} else {
		planRet = 0 * time.Second
	}

	if ann, ok := layer.Annotations[annotations.ApplyRunRetention]; ok {
		applyRet, err = time.ParseDuration(ann)
	} else if ann, ok := repository.Annotations[annotations.ApplyRunRetention]; ok {
		applyRet, err = time.ParseDuration(ann)
	} else {
		applyRet = 0 * time.Second
	}

	if err != nil {
		return runRetention{}, err
	}

	return runRetention{
		plan:  planRet,
		apply: applyRet,
	}, nil
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
	retentions, err := getRetentions(layer, repository)
	if err != nil {
		log.Errorf("could not get run retentions for layer %s: %s", layer.Name, err)
		return err
	}
	if retentions.plan == 0 && retentions.apply == 0 {
		// nothing to do
		log.Infof("no run retention set for layer %s, skipping cleanup", layer.Name)
		return nil
	}

	runs, err := r.getAllFinishedRuns(ctx, layer, repository)
	if err != nil {
		log.Errorf("could not get runs for layer %s: %s", layer.Name, err)
		return err
	}

	toDelete := []*configv1alpha1.TerraformRun{}
	for _, run := range runs {
		if run.Spec.Action == string(PlanAction) &&
			retentions.plan != 0 &&
			r.Clock.Now().Sub(run.CreationTimestamp.Time) > retentions.plan {
			toDelete = append(toDelete, run)
			continue
		}
		if run.Spec.Action == string(ApplyAction) &&
			retentions.apply != 0 &&
			r.Clock.Now().Sub(run.CreationTimestamp.Time) > retentions.apply {
			toDelete = append(toDelete, run)
			continue
		}
	}

	err = deleteAll(ctx, r.Client, toDelete)
	if err != nil {
		return err
	}

	return nil
}

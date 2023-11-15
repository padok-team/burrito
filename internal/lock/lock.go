package lock

import (
	"context"
	"fmt"
	"hash/fnv"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	coordination "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const lockPrefix string = "burrito-layer-lock"
const prLockPrefix string = "burrito-pr-layer-lock"

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func isPrLayer(layer *configv1alpha1.TerraformLayer) bool {
	if len(layer.GetOwnerReferences()) == 0 {
		return false
	}

	return layer.GetOwnerReferences()[0].Kind == "TerraformPullRequest"
}

func getLeaseName(layer *configv1alpha1.TerraformLayer) string {
	return fmt.Sprintf("%s-%d", lockPrefix, hash(layer.Spec.Repository.Name+layer.Spec.Repository.Namespace+layer.Spec.Path))
}

func getPullRequestLeaseName(layer *configv1alpha1.TerraformLayer) string {
	return fmt.Sprintf("%s-%d", prLockPrefix, hash(layer.Spec.Repository.Name+layer.Spec.Repository.Namespace+layer.Spec.Path))
}

func getLeaseLock(layer *configv1alpha1.TerraformLayer, run *configv1alpha1.TerraformRun) *coordination.Lease {
	identity := "burrito-controller"
	name := getLeaseName(layer)
	if isPrLayer(layer) {
		name = getPullRequestLeaseName(layer)
	}
	lease := &coordination.Lease{
		Spec: coordination.LeaseSpec{
			HolderIdentity: &identity,
		},
	}
	lease.SetName(name)
	lease.SetNamespace(layer.Namespace)
	lease.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: run.GetAPIVersion(),
			Kind:       run.GetKind(),
			Name:       run.Name,
			UID:        run.UID,
		},
	})
	return lease
}

func isPullRequestLockPresent(layer *configv1alpha1.TerraformLayer, c client.Client) (bool, error) {
	if !isPrLayer(layer) {
		return false, nil
	}
	err := c.Get(context.Background(), types.NamespacedName{
		Name:      getPullRequestLeaseName(layer),
		Namespace: layer.Namespace,
	}, &coordination.Lease{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

func isDefaultLockPresent(layer *configv1alpha1.TerraformLayer, c client.Client) (bool, error) {
	err := c.Get(context.Background(), types.NamespacedName{
		Name:      getLeaseName(layer),
		Namespace: layer.Namespace,
	}, &coordination.Lease{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	return true, nil
}

func IsLayerLocked(ctx context.Context, c client.Client, layer *configv1alpha1.TerraformLayer) (bool, error) {
	prLocked, err1 := isPullRequestLockPresent(layer, c)
	locked, err2 := isDefaultLockPresent(layer, c)
	if err1 != nil || err2 != nil {
		return true, fmt.Errorf("could not check lock status: %s, %s", err1, err2)
	}
	return locked || prLocked, nil
}

func CreateLock(ctx context.Context, c client.Client, layer *configv1alpha1.TerraformLayer, run *configv1alpha1.TerraformRun) error {
	leaseLock := getLeaseLock(layer, run)
	return c.Create(ctx, leaseLock)
}

func DeleteLock(ctx context.Context, c client.Client, layer *configv1alpha1.TerraformLayer, run *configv1alpha1.TerraformRun) error {
	leaseLock := getLeaseLock(layer, run)
	return c.Delete(ctx, leaseLock)
}

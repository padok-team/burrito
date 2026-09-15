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

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func getLeaseName(layer *configv1alpha1.TerraformLayer) string {
	return fmt.Sprintf("%s-%d", lockPrefix, hash(layer.Spec.Repository.Name+layer.Spec.Repository.Namespace+layer.Spec.Path))
}

func getLeaseLock(layer *configv1alpha1.TerraformLayer, run *configv1alpha1.TerraformRun) *coordination.Lease {
	identity := "burrito-controller"
	name := getLeaseName(layer)
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

func IsLayerLocked(ctx context.Context, c client.Client, layer *configv1alpha1.TerraformLayer) (bool, error) {
	lease, err := GetLock(ctx, c, layer)
	if err != nil {
		return false, err
	}
	return lease != nil, nil
}

// GetLock returns the Lease holding the lock on the given layer, or nil when
// the layer is not locked.
func GetLock(ctx context.Context, c client.Client, layer *configv1alpha1.TerraformLayer) (*coordination.Lease, error) {
	lease := &coordination.Lease{}
	err := c.Get(ctx, types.NamespacedName{
		Name:      getLeaseName(layer),
		Namespace: layer.Namespace,
	}, lease)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return lease, nil
}

func CreateLock(ctx context.Context, c client.Client, layer *configv1alpha1.TerraformLayer, run *configv1alpha1.TerraformRun) error {
	leaseLock := getLeaseLock(layer, run)
	return c.Create(ctx, leaseLock)
}

func DeleteLock(ctx context.Context, c client.Client, layer *configv1alpha1.TerraformLayer, run *configv1alpha1.TerraformRun) error {
	leaseLock := getLeaseLock(layer, run)
	err := c.Delete(ctx, leaseLock)
	if errors.IsNotFound(err) {
		// The lock may already have been released (e.g. as an orphaned lock
		// cleaned up by the layer controller): nothing left to do.
		return nil
	}
	return err
}

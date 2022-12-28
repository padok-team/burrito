package lock

import (
	"context"
	"fmt"
	"hash/fnv"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	coordination "k8s.io/api/coordination/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const lockPrefix string = "burrito-layer-lock"

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func getLeaseLock(obj *configv1alpha1.TerraformLayer) *coordination.Lease {
	identity := "burrito-controller"
	name := fmt.Sprintf("%s-%d", lockPrefix, hash(obj.Spec.Repository.Name+obj.Spec.Repository.Namespace+obj.Spec.Path))
	lease := &coordination.Lease{
		Spec: coordination.LeaseSpec{
			HolderIdentity: &identity,
		},
	}
	lease.SetName(name)
	lease.SetNamespace(obj.Namespace)
	return lease
}

func IsLocked(ctx context.Context, c client.Client, obj *configv1alpha1.TerraformLayer) (bool, error) {
	leaseLock := getLeaseLock(obj)
	err := c.Get(ctx, types.NamespacedName{
		Name:      leaseLock.Name,
		Namespace: leaseLock.Namespace,
	}, &coordination.Lease{})
	if errors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func CreateLock(ctx context.Context, c client.Client, obj *configv1alpha1.TerraformLayer) error {
	leaseLock := getLeaseLock(obj)
	return c.Create(ctx, leaseLock)
}

func DeleteLock(ctx context.Context, c client.Client, obj *configv1alpha1.TerraformLayer) error {
	leaseLock := getLeaseLock(obj)
	return c.Delete(ctx, leaseLock)
}

package terraformrepository

import (
	"context"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestReconcileRetriesStatusUpdateOnConflict(t *testing.T) {
	scheme := newTerraformRepositoryTestScheme(t)
	repository := testTerraformRepository("default", "repo")

	conflictReturned := false
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository).
		WithStatusSubresource(repository).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				if subResourceName == "status" && !conflictReturned {
					conflictReturned = true
					return errors.NewConflict(schema.GroupResource{Group: "config.terraform.padok.cloud", Resource: "terraformrepositories"}, obj.GetName(), nil)
				}
				return c.Status().Update(ctx, obj, opts...)
			},
		}).
		Build()

	reconciler := &Reconciler{
		Client:   cl,
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
		Clock:    RealClock{},
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: ktypes.NamespacedName{Name: "repo", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if !conflictReturned {
		t.Fatalf("expected the status update to hit a conflict at least once")
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.RepositorySync {
		t.Fatalf("expected repository sync requeue after a synced reconcile, got %s", result.RequeueAfter)
	}

	updated := &configv1alpha1.TerraformRepository{}
	if err := cl.Get(context.Background(), ktypes.NamespacedName{Name: "repo", Namespace: "default"}, updated); err != nil {
		t.Fatalf("failed to get repository: %v", err)
	}
	if updated.Status.State != "Synced" {
		t.Fatalf("expected status to be persisted despite the conflict retry, got state %q", updated.Status.State)
	}
}

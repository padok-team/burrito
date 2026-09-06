package terraformrepository

import (
	"context"
	"errors"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func TestReconcileReturnsErrorWhenStatusUpdateFails(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add TerraformRepository types to scheme: %v", err)
	}

	repository := &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{Name: "repo", Namespace: "default"},
	}
	expectedErr := errors.New("status update failed")
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository).
		WithStatusSubresource(repository).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				if subResourceName == "status" {
					return expectedErr
				}
				return client.SubResource(subResourceName).Update(ctx, obj, opts...)
			},
		}).
		Build()
	reconciler := &Reconciler{
		Client:   cl,
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: repository.Name, Namespace: repository.Namespace},
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected status update error, got %v", err)
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue, got %s", result.RequeueAfter)
	}
}

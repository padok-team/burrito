package terraformpullrequest

import (
	"context"
	"errors"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/event"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
)

func TestReconcileIgnoresMissingResources(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	reconciler := &Reconciler{
		Client:   fake.NewClientBuilder().WithScheme(scheme).Build(),
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: ktypes.NamespacedName{Name: "missing", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile returned error: %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Fatalf("expected no requeue for missing resources, got %s", result.RequeueAfter)
	}
}

func TestReconcileReturnsErrorWhenGetPullRequestFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	expectedErr := errors.New("get failed")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key ktypes.NamespacedName, obj client.Object, opts ...client.GetOption) error {
			return expectedErr
		},
	}).Build()
	reconciler := &Reconciler{Client: cl}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: ktypes.NamespacedName{Name: "pr", Namespace: "default"},
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected get error to be propagated, got %v", err)
	}
}

func TestReconcilePullRequestReturnsErrorWhenGetRepositoryFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	pr := terraformPullRequest("default", "repo-1", "1", "feature", "sha")
	expectedErr := errors.New("get repository failed")
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(pr).WithInterceptorFuncs(interceptor.Funcs{
		Get: func(ctx context.Context, c client.WithWatch, key ktypes.NamespacedName, obj client.Object, opts ...client.GetOption) error {
			if _, ok := obj.(*configv1alpha1.TerraformRepository); ok {
				return expectedErr
			}
			return c.Get(ctx, key, obj, opts...)
		},
	}).Build()
	reconciler := &Reconciler{Client: cl}

	_, err := reconciler.reconcilePullRequest(context.Background(), pr)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected get repository error to be propagated, got %v", err)
	}
}

func TestReconcilePullRequestReturnsErrorWhenStatusUpdateFails(t *testing.T) {
	scheme := newTerraformPullRequestTestScheme(t)
	repository := terraformRepository("default", "repo")
	pr := terraformPullRequest("default", "repo", "1", "feature", "sha")
	pr.Status.LastDiscoveredCommit = "sha"
	pr.Status.LastCommentedCommit = "sha"
	expectedErr := errors.New("status update failed")
	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(repository, pr).
		WithStatusSubresource(pr).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, client client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				if subResourceName == "status" {
					return expectedErr
				}
				return nil
			},
		}).
		Build()
	reconciler := &Reconciler{
		Client:   cl,
		Config:   config.TestConfig(),
		Recorder: record.NewFakeRecorder(10),
	}

	result, err := reconciler.reconcilePullRequest(context.Background(), pr)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected status update error, got %v", err)
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue, got %s", result.RequeueAfter)
	}
}

func TestIgnorePredicate(t *testing.T) {
	predicate := ignorePredicate()
	oldPR := &configv1alpha1.TerraformPullRequest{
		ObjectMeta: metav1.ObjectMeta{
			Generation: 1,
			Annotations: map[string]string{
				"key": "old",
			},
		},
	}
	newPR := oldPR.DeepCopy()

	if predicate.Update(event.UpdateEvent{ObjectOld: oldPR, ObjectNew: newPR}) {
		t.Fatalf("expected unchanged PR update to be ignored")
	}

	annotatedPR := oldPR.DeepCopy()
	annotatedPR.Annotations["key"] = "new"
	if !predicate.Update(event.UpdateEvent{ObjectOld: oldPR, ObjectNew: annotatedPR}) {
		t.Fatalf("expected annotation change to trigger reconciliation")
	}

	generatedPR := oldPR.DeepCopy()
	generatedPR.Generation = 2
	if !predicate.Update(event.UpdateEvent{ObjectOld: oldPR, ObjectNew: generatedPR}) {
		t.Fatalf("expected generation change to trigger reconciliation")
	}

	if predicate.Delete(event.DeleteEvent{Object: oldPR, DeleteStateUnknown: true}) {
		t.Fatalf("expected confirmed unknown delete state to be ignored")
	}
	if !predicate.Delete(event.DeleteEvent{Object: oldPR, DeleteStateUnknown: false}) {
		t.Fatalf("expected known delete state to trigger reconciliation")
	}
}

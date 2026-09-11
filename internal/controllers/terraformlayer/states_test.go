package terraformlayer

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type createErrorClient struct {
	client.Client
	err error
}

func (c createErrorClient) Create(context.Context, client.Object, ...client.CreateOption) error {
	return c.err
}

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

type planDatastoreClient struct {
	*datastore.MockClient
	plan []byte
}

func (c planDatastoreClient) GetPlan(string, string, string, string, string) ([]byte, error) {
	return c.plan, nil
}

func TestPlanNeededEventIncludesCreateError(t *testing.T) {
	createErr := errors.New("metadata.labels: Invalid value: must be no more than 63 characters")
	recorder := record.NewFakeRecorder(1)
	reconciler := &Reconciler{
		Client:   createErrorClient{err: createErr},
		Recorder: recorder,
		Config:   config.TestConfig(),
	}
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "infra-s3-buckets-logistics-toolings-dc1-nl-prd-defect-tracker-prd-nl1",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "abc123",
			},
		},
	}

	result, run := (&PlanNeeded{}).getHandler()(context.Background(), reconciler, layer, &configv1alpha1.TerraformRepository{})

	if run != nil {
		t.Fatalf("expected no run when creation fails")
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue, got %s", result.RequeueAfter)
	}
	assertEventContains(t, recorder, "Failed to create TerraformRun for Plan action: "+createErr.Error())
}

func TestApplyNeededEventIncludesCreateError(t *testing.T) {
	createErr := errors.New("metadata.labels: Invalid value: must be no more than 63 characters")
	recorder := record.NewFakeRecorder(1)
	reconciler := &Reconciler{
		Client:   createErrorClient{err: createErr},
		Recorder: recorder,
		Config:   config.TestConfig(),
	}
	autoApply := true
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "infra-s3-buckets-logistics-toolings-dc1-nl-prd-defect-tracker-prd-nl1",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "abc123",
				annotations.LastPlanRun:      "plan-run/0",
			},
		},
	}
	repository := &configv1alpha1.TerraformRepository{
		Spec: configv1alpha1.TerraformRepositorySpec{
			RemediationStrategy: configv1alpha1.RemediationStrategy{
				AutoApply: &autoApply,
			},
		},
	}

	result, run := (&ApplyNeeded{}).getHandler()(context.Background(), reconciler, layer, repository)

	if run != nil {
		t.Fatalf("expected no run when creation fails")
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue, got %s", result.RequeueAfter)
	}
	assertEventContains(t, recorder, "Failed to create TerraformRun for Apply action: "+createErr.Error())
}

func TestApplyNeededBlocksDestructivePlanWhenNonDestructiveApplyEnabled(t *testing.T) {
	recorder := record.NewFakeRecorder(1)
	reconciler := &Reconciler{
		Client:   createErrorClient{err: errors.New("unexpected create")},
		Recorder: recorder,
		Config:   config.TestConfig(),
	}
	autoApply := true
	nonDestructiveApply := true
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-layer",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastRelevantCommit: "abc123",
				annotations.LastPlanRun:        "plan-run/0",
			},
		},
		Status: configv1alpha1.TerraformLayerStatus{
			LastResult: "Plan: 0 to create, 1 to update, 2 to delete",
		},
	}
	repository := &configv1alpha1.TerraformRepository{
		Spec: configv1alpha1.TerraformRepositorySpec{
			RemediationStrategy: configv1alpha1.RemediationStrategy{
				AutoApply:           &autoApply,
				NonDestructiveApply: &nonDestructiveApply,
			},
		},
	}

	result, run := (&ApplyNeeded{}).getHandler()(context.Background(), reconciler, layer, repository)

	if run != nil {
		t.Fatalf("expected no run when nonDestructiveApply blocks destructive changes")
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.DriftDetection {
		t.Fatalf("expected DriftDetection requeue, got %s", result.RequeueAfter)
	}
	assertEventContains(t, recorder, "nonDestructiveApply is enabled and the plan contains delete actions, Apply run not created")
}

func TestReconcileBlocksDestructiveFreshPlanWhenNonDestructiveApplyEnabled(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := configv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to register config scheme: %s", err)
	}
	if err := coordinationv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to register coordination scheme: %s", err)
	}

	autoApply := true
	nonDestructiveApply := true
	terraformEnabled := true
	planDate := "Mon May  8 11:21:53 UTC 2023"
	now, err := time.Parse(time.UnixDate, planDate)
	if err != nil {
		t.Fatalf("failed to parse test time: %s", err)
	}
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-layer",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit:   "abc123",
				annotations.LastRelevantCommit: "abc123",
				annotations.LastPlanCommit:     "abc123",
				annotations.LastPlanDate:       planDate,
				annotations.LastPlanRun:        "plan-run/0",
				annotations.LastPlanSum:        "plan-sum",
			},
		},
		Spec: configv1alpha1.TerraformLayerSpec{
			Path: "test-layer/",
			Repository: configv1alpha1.TerraformLayerRepository{
				Name:      "test-repo",
				Namespace: "default",
			},
			TerraformConfig: configv1alpha1.TerraformConfig{
				Enabled: &terraformEnabled,
			},
			RemediationStrategy: configv1alpha1.RemediationStrategy{
				AutoApply:           &autoApply,
				NonDestructiveApply: &nonDestructiveApply,
			},
		},
		Status: configv1alpha1.TerraformLayerStatus{
			LastResult: "Plan: 0 to create, 0 to update, 0 to delete",
			LastRun: configv1alpha1.TerraformLayerRun{
				Name: "plan-run",
			},
		},
	}
	repository := &configv1alpha1.TerraformRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repo",
			Namespace: "default",
		},
	}
	planRun := &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "plan-run",
			Namespace: "default",
		},
		Status: configv1alpha1.TerraformRunStatus{
			State: "Succeeded",
		},
	}
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(layer, repository, planRun).
		WithStatusSubresource(&configv1alpha1.TerraformLayer{}).
		Build()
	reconciler := &Reconciler{
		Client: fakeClient,
		Clock:  fixedClock{now: now},
		Config: config.TestConfig(),
		Datastore: planDatastoreClient{
			MockClient: datastore.NewMockClient(),
			plan:       []byte("Plan: 0 to create, 1 to update, 2 to delete"),
		},
		Recorder: record.NewFakeRecorder(1),
	}

	result, err := reconciler.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      layer.Name,
			Namespace: layer.Namespace,
		},
	})

	if err != nil {
		t.Fatalf("expected no reconcile error, got %s", err)
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.DriftDetection {
		t.Fatalf("expected DriftDetection requeue, got %s", result.RequeueAfter)
	}
	updatedLayer := &configv1alpha1.TerraformLayer{}
	if err := fakeClient.Get(context.Background(), types.NamespacedName{Name: layer.Name, Namespace: layer.Namespace}, updatedLayer); err != nil {
		t.Fatalf("failed to fetch updated layer: %s", err)
	}
	if updatedLayer.Status.LastResult != "Plan: 0 to create, 1 to update, 2 to delete" {
		t.Fatalf("expected fresh plan result in status, got %q", updatedLayer.Status.LastResult)
	}
	runs := &configv1alpha1.TerraformRunList{}
	if err := fakeClient.List(context.Background(), runs); err != nil {
		t.Fatalf("failed to list runs: %s", err)
	}
	for _, run := range runs.Items {
		if run.Spec.Action == string(ApplyAction) {
			t.Fatalf("expected no apply run, got %s", run.Name)
		}
	}
}

func TestHasDestructiveChanges(t *testing.T) {
	tt := []struct {
		name       string
		lastResult string
		expected   bool
	}{
		{"NoDelete", "Plan: 3 to create, 1 to update, 0 to delete", false},
		{"OneDelete", "Plan: 0 to create, 0 to update, 1 to delete", true},
		{"MultipleDeletes", "Plan: 0 to create, 1 to update, 12 to delete", true},
		{"NoPlan", "Layer has never been planned", false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := hasDestructiveChanges(tc.lastResult)
			if result != tc.expected {
				t.Fatalf("expected %t, got %t", tc.expected, result)
			}
		})
	}
}

func assertEventContains(t *testing.T, recorder *record.FakeRecorder, want string) {
	t.Helper()

	select {
	case event := <-recorder.Events:
		if !strings.Contains(event, want) {
			t.Fatalf("expected event to contain %q, got %q", want, event)
		}
	default:
		t.Fatalf("expected event containing %q, got none", want)
	}
}

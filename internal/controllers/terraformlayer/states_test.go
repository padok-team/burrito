package terraformlayer

import (
	"context"
	"errors"
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type createErrorClient struct {
	client.Client
	err error
}

func (c createErrorClient) Create(context.Context, client.Object, ...client.CreateOption) error {
	return c.err
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
				annotations.LastRelevantCommit: "abc123",
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
				annotations.LastRelevantCommit: "abc123",
				annotations.LastPlanRun:        "plan-run/0",
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

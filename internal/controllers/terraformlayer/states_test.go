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

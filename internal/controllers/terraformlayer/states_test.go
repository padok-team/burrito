package terraformlayer

import (
	"context"
	"errors"
	"strings"
	"testing"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/repository/providers/mock"
	"github.com/padok-team/burrito/internal/repository/status"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
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

type createRecordingClient struct {
	client.Client
	created []*configv1alpha1.TerraformRun
}

func (c *createRecordingClient) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) error {
	c.created = append(c.created, obj.(*configv1alpha1.TerraformRun))
	return nil
}

func reconcilerWithFakeProvider(client client.Client, provider *mock.APIProvider) *Reconciler {
	return &Reconciler{
		Client:   client,
		Recorder: record.NewFakeRecorder(10),
		Config:   config.TestConfig(),
		APIProviderFactory: func(*configv1alpha1.TerraformRepository) (repositorytypes.APIProvider, error) {
			return provider, nil
		},
	}
}

func TestPlanNeededReportsACommitStatusForAnUnplannedCommit(t *testing.T) {
	recordingClient := &createRecordingClient{}
	provider := &mock.APIProvider{}
	reconciler := reconcilerWithFakeProvider(recordingClient, provider)
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pwet",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "sha-new",
				annotations.LastPlanCommit:   "sha-old",
			},
		},
	}

	_, run := (&PlanNeeded{}).getHandler()(context.Background(), reconciler, layer, &configv1alpha1.TerraformRepository{})

	if run == nil {
		t.Fatalf("expected a run to be created")
	}
	if run.Annotations[annotations.PostCommitStatus] != "true" {
		t.Errorf("expected the run to be marked for commit statuses, got annotations %v", run.Annotations)
	}
	if len(provider.SetStatusCalls) != 1 {
		t.Fatalf("expected one commit status to be posted, got %d", len(provider.SetStatusCalls))
	}
	if got := provider.SetStatusCalls[0]; got.Commit != "sha-new" || got.State != status.StatePending {
		t.Errorf("expected a pending status on sha-new, got %q on %q", got.State, got.Commit)
	}
}

func TestPlanNeededSkipsCommitStatusWhenDisabled(t *testing.T) {
	recordingClient := &createRecordingClient{}
	provider := &mock.APIProvider{}
	reconciler := reconcilerWithFakeProvider(recordingClient, provider)
	reconciler.Config.Controller.CommitStatus.Enabled = false
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pwet",
			Namespace:   "default",
			Annotations: map[string]string{annotations.LastBranchCommit: "sha-new"},
		},
	}

	_, run := (&PlanNeeded{}).getHandler()(context.Background(), reconciler, layer, &configv1alpha1.TerraformRepository{})

	if run == nil {
		t.Fatalf("expected a run to be created")
	}
	if len(provider.SetStatusCalls) != 0 {
		t.Errorf("expected no commit status when the feature is disabled, got %d", len(provider.SetStatusCalls))
	}
}

func TestPlanNeededSkipsCommitStatusOnDriftDetectionReplan(t *testing.T) {
	recordingClient := &createRecordingClient{}
	provider := &mock.APIProvider{}
	reconciler := reconcilerWithFakeProvider(recordingClient, provider)
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pwet",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "sha-same",
				annotations.LastPlanCommit:   "sha-same",
			},
		},
	}

	_, run := (&PlanNeeded{}).getHandler()(context.Background(), reconciler, layer, &configv1alpha1.TerraformRepository{})

	if run == nil {
		t.Fatalf("expected drift detection to still create a run")
	}
	if _, marked := run.Annotations[annotations.PostCommitStatus]; marked {
		t.Errorf("expected a drift detection re-plan not to be marked for commit statuses")
	}
	if len(provider.SetStatusCalls) != 0 {
		t.Errorf("expected no commit status for a drift detection re-plan, got %d", len(provider.SetStatusCalls))
	}
}

func TestApplyNeededReportsACommitStatus(t *testing.T) {
	recordingClient := &createRecordingClient{}
	provider := &mock.APIProvider{}
	reconciler := reconcilerWithFakeProvider(recordingClient, provider)
	autoApply := true
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pwet",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "sha-same",
				annotations.LastPlanCommit:   "sha-same",
				annotations.LastPlanRun:      "plan-run/0",
			},
		},
	}
	repository := &configv1alpha1.TerraformRepository{
		Spec: configv1alpha1.TerraformRepositorySpec{
			RemediationStrategy: configv1alpha1.RemediationStrategy{AutoApply: &autoApply},
		},
	}

	_, run := (&ApplyNeeded{}).getHandler()(context.Background(), reconciler, layer, repository)

	if run == nil {
		t.Fatalf("expected a run to be created")
	}
	if run.Annotations[annotations.PostCommitStatus] != "true" {
		t.Errorf("expected every apply to be marked for commit statuses, got annotations %v", run.Annotations)
	}
	if len(provider.SetStatusCalls) != 1 {
		t.Fatalf("expected one commit status to be posted, got %d", len(provider.SetStatusCalls))
	}
}

type failOnCreateClient struct {
	client.Client
	t *testing.T
}

func (c failOnCreateClient) Create(context.Context, client.Object, ...client.CreateOption) error {
	c.t.Fatalf("expected no TerraformRun to be created")
	return nil
}

func TestPlanNeededSkipsRunWhenLastBranchCommitIsEmpty(t *testing.T) {
	recorder := record.NewFakeRecorder(1)
	reconciler := &Reconciler{
		Client:   failOnCreateClient{t: t},
		Recorder: recorder,
		Config:   config.TestConfig(),
	}
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "pwet",
			Namespace:   "default",
			Annotations: map[string]string{annotations.LastBranchCommit: ""},
		},
	}

	result, run := (&PlanNeeded{}).getHandler()(context.Background(), reconciler, layer, &configv1alpha1.TerraformRepository{})

	if run != nil {
		t.Fatalf("expected no run when the last branch commit annotation is empty")
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue, got %s", result.RequeueAfter)
	}
	assertEventContains(t, recorder, "Layer has no last branch commit annotation, Plan run not created")
}

func TestApplyNeededSkipsRunWhenLastBranchCommitIsEmpty(t *testing.T) {
	recorder := record.NewFakeRecorder(1)
	reconciler := &Reconciler{
		Client:   failOnCreateClient{t: t},
		Recorder: recorder,
		Config:   config.TestConfig(),
	}
	autoApply := true
	layer := &configv1alpha1.TerraformLayer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pwet",
			Namespace: "default",
			Annotations: map[string]string{
				annotations.LastBranchCommit: "",
				annotations.LastPlanRun:      "plan-run/0",
			},
		},
	}
	repository := &configv1alpha1.TerraformRepository{
		Spec: configv1alpha1.TerraformRepositorySpec{
			RemediationStrategy: configv1alpha1.RemediationStrategy{AutoApply: &autoApply},
		},
	}

	result, run := (&ApplyNeeded{}).getHandler()(context.Background(), reconciler, layer, repository)

	if run != nil {
		t.Fatalf("expected no run when the last branch commit annotation is empty")
	}
	if result.RequeueAfter != reconciler.Config.Controller.Timers.OnError {
		t.Fatalf("expected OnError requeue, got %s", result.RequeueAfter)
	}
	assertEventContains(t, recorder, "Layer has no last branch commit annotation, Apply run not created")
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

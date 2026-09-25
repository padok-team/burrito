package terraformrun

import (
	"context"
	"testing"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const stateTestTime = "Mon May  8 11:21:53 UTC 2023"

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func failedRun(lastRun time.Time) *configv1alpha1.TerraformRun {
	return &configv1alpha1.TerraformRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "run",
			Namespace: "default",
		},
		Status: configv1alpha1.TerraformRunStatus{
			State:     "FailureGracePeriod",
			LastRun:   lastRun.Format(time.UnixDate),
			RunnerPod: "runner-pod",
		},
	}
}

func stateTestNow(t *testing.T) time.Time {
	t.Helper()
	now, err := time.Parse(time.UnixDate, stateTestTime)
	if err != nil {
		t.Fatalf("could not parse test time: %s", err)
	}
	return now
}

func TestFailureGracePeriodRequeuesWhenTheGracePeriodEnds(t *testing.T) {
	now := stateTestNow(t)
	cfg := config.TestConfig()
	// The runner pod failed 5 seconds ago and the grace period is 15 seconds,
	// so the run must be requeued 10 seconds from now.
	run := failedRun(now.Add(-5 * time.Second))
	r := &Reconciler{Config: cfg, Clock: fixedClock{now: now}}

	result, _ := (&FailureGracePeriod{}).getHandler()(context.Background(), r, run, nil, nil)

	if result.RequeueAfter != 10*time.Second {
		t.Fatalf("expected the run to be requeued in 10s, when the grace period ends, got %s", result.RequeueAfter)
	}
}

func TestFailureGracePeriodRequeuesWithWaitActionOnceTheGracePeriodIsOver(t *testing.T) {
	now := stateTestNow(t)
	cfg := config.TestConfig()
	// The runner pod failed 60 seconds ago and the grace period is 15 seconds,
	// so it is over and the run must be requeued for a retry.
	run := failedRun(now.Add(-60 * time.Second))
	r := &Reconciler{Config: cfg, Clock: fixedClock{now: now}}

	result, _ := (&FailureGracePeriod{}).getHandler()(context.Background(), r, run, nil, nil)

	if result.RequeueAfter != cfg.Controller.Timers.WaitAction {
		t.Fatalf("expected the run to be requeued after %s, got %s", cfg.Controller.Timers.WaitAction, result.RequeueAfter)
	}
}

func TestFailureGracePeriodRequeuesOnErrorWhenLastRunIsNotParsable(t *testing.T) {
	now := stateTestNow(t)
	cfg := config.TestConfig()
	run := failedRun(now)
	run.Status.LastRun = "not a date"
	r := &Reconciler{Config: cfg, Clock: fixedClock{now: now}}

	result, _ := (&FailureGracePeriod{}).getHandler()(context.Background(), r, run, nil, nil)

	if result.RequeueAfter != cfg.Controller.Timers.OnError {
		t.Fatalf("expected the run to be requeued after %s, got %s", cfg.Controller.Timers.OnError, result.RequeueAfter)
	}
}

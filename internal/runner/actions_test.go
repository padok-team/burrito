package runner

import (
	"errors"
	"testing"
)

func TestApplySummaryReturnsPlaceholderWhenPlanUnavailable(t *testing.T) {
	got := applySummary(nil, errors.New("no plan artifact"))
	want := "Apply Successful"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestApplySummaryReturnsPlaceholderOnUnparseableJSON(t *testing.T) {
	got := applySummary([]byte("not json"), nil)
	want := "Apply Successful"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestApplySummaryReturnsDiffFromAppliedPlan(t *testing.T) {
	planJson := []byte(`{
		"format_version": "1.2",
		"resource_changes": [
			{"change": {"actions": ["create"]}},
			{"change": {"actions": ["create"]}},
			{"change": {"actions": ["update"]}},
			{"change": {"actions": ["delete"]}},
			{"change": {"actions": ["no-op"]}}
		]
	}`)
	got := applySummary(planJson, nil)
	want := "Plan: 2 to create, 1 to update, 1 to delete"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

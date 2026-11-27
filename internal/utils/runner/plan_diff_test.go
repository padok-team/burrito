package runner

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
)

func TestGetDiffReplaceCountsOnce(t *testing.T) {
	plan := &tfjson.Plan{
		ResourceChanges: []*tfjson.ResourceChange{
			{
				Address: "aws_instance.example",
				Change: &tfjson.Change{
					Actions: tfjson.Actions{"create", "delete"},
				},
			},
		},
	}

	hasDiff, destructive, summary := GetDiff(plan)

	if !hasDiff {
		t.Fatalf("expected plan to have changes")
	}

	if !destructive {
		t.Fatalf("expected replacement to be destructive")
	}

	expected := "Plan: 1 to create, 0 to update, 1 to delete"
	if summary != expected {
		t.Fatalf("unexpected summary: expected %q got %q", expected, summary)
	}
}

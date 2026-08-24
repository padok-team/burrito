package runner

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

// Produces a diff summary from the given plan.
// Returns:
//   - whether the plan has any changes
//   - whether the plan contains destructive changes
//   - a human-readable summary
func GetDiff(plan *tfjson.Plan) (bool, bool, string) {
	delete := 0
	create := 0
	update := 0

	for _, res := range plan.ResourceChanges {
		if res.Change.Actions.Create() {
			create++
		}
		if res.Change.Actions.Delete() {
			delete++
		}
		if res.Change.Actions.Update() {
			update++
		}
		if res.Change.Actions.Replace() {
			create++
			delete++
		}
	}

	diff := create+delete+update > 0
	destructive := delete > 0

	return diff, destructive, fmt.Sprintf(
		"Plan: %d to create, %d to update, %d to delete",
		create,
		update,
		delete,
	)
}

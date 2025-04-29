package runner

import (
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

// Produces a diff summary from the given plan
func GetDiff(plan *tfjson.Plan) (bool, string) {
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
	diff := false
	if create+delete+update > 0 {
		diff = true
	}
	return diff, fmt.Sprintf("Plan: %d to create, %d to update, %d to delete", create, update, delete)
}

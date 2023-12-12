package terraform

import (
	"encoding/json"
	"fmt"

	tfjson "github.com/hashicorp/terraform-json"
)

func GetShortDiffFromPlan(bPlan []byte) (bool, string, error) {
	plan := &tfjson.Plan{}
	err := json.Unmarshal(bPlan, plan)
	if err != nil {
		return false, "", err
	}
	delete := 0
	create := 0
	update := 0
	diff := false
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
	}
	if create+delete+update > 0 {
		diff = true
	}
	return diff, fmt.Sprintf("Plan: %d to create, %d to update, %d to delete", create, update, delete), nil
}

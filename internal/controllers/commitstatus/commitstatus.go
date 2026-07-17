// Package commitstatus posts a single commit status scoped to one layer's plan/apply
// action. There is no pull request to aggregate multiple layers under a single status on
// a direct push to the base branch, so every affected layer gets its own status,
// disambiguated by its name.
package commitstatus

import (
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	repositorytypes "github.com/padok-team/burrito/internal/repository/types"
)

const (
	Needed     = "needed"
	InProgress = "in progress"
	Succeeded  = "succeeded"
	Failed     = "failed"
)

// Post sets a commit status for layer's phase (plan or apply), on the given commit.
func Post(provider repositorytypes.APIProvider, repository *configv1alpha1.TerraformRepository, layer *configv1alpha1.TerraformLayer, phase status.Phase, state status.State, outcome string, commit string) error {
	return provider.SetStatus(repository, nil, status.CommitStatus{
		Phase:       phase,
		State:       state,
		Description: fmt.Sprintf("%s/%s %s %s", layer.Namespace, layer.Name, phase, outcome),
		Commit:      commit,
		Context:     fmt.Sprintf("burrito/%s/%s", phase, layer.Name),
	})
}

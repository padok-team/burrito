package terraformrun

import (
	"context"
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	logrus "github.com/sirupsen/logrus"
)

// isPullRequestLayer reports whether layer was generated for an open pull/merge request.
// Those already get an aggregated plan/apply status posted by the terraformpullrequest
// controller once all their layers are done; posting one here too would just clobber it.
func isPullRequestLayer(layer *configv1alpha1.TerraformLayer) bool {
	for _, ref := range layer.OwnerReferences {
		if ref.Kind == "TerraformPullRequest" {
			return true
		}
	}
	return false
}

// setCommitStatusForDirectPush posts a plan/apply commit status for a run against a layer
// that tracks the repository's base branch directly. Commits pushed straight to the base
// branch never go through a TerraformPullRequest, so nothing else posts a status for them.
func (r *Reconciler) setCommitStatusForDirectPush(ctx context.Context, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, state status.State) {
	if isPullRequestLayer(layer) || run.Spec.Layer.Revision == "" {
		return
	}

	provider, err := r.getAPIProvider(repository)
	if err != nil {
		logrus.Warnf("could not get API provider to set commit status for run %s: %s", run.Name, err)
		return
	}

	phase := status.PhasePlan
	if run.Spec.Action == string(ApplyAction) {
		phase = status.PhaseApply
	}
	description := fmt.Sprintf("Burrito %s succeeded", phase)
	if state == status.StateFailure {
		description = fmt.Sprintf("Burrito %s failed", phase)
	}

	s := status.CommitStatus{
		Phase:       phase,
		State:       state,
		Description: description,
		Commit:      run.Spec.Layer.Revision,
		// Several layers can target the same commit on a direct push, with no pull
		// request to aggregate them under a single context: disambiguate by layer.
		Context: fmt.Sprintf("burrito/%s/%s", phase, layer.Name),
	}
	if err := provider.SetStatus(repository, nil, s); err != nil {
		logrus.Warnf("could not set %s commit status for run %s: %s", phase, run.Name, err)
	}
}

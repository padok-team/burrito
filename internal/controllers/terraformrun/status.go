package terraformrun

import (
	"context"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/controllers/commitstatus"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	logrus "github.com/sirupsen/logrus"
)

// postCommitStatus posts a plan/apply commit status scoped to layer for run, best-effort:
// a failure here must not block the reconciliation.
func (r *Reconciler) postCommitStatus(ctx context.Context, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, state status.State, outcome string) {
	if run.Spec.Layer.Revision == "" {
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
	if err := commitstatus.Post(provider, repository, layer, phase, state, run.Spec.Layer.Revision, r.resultMessage(run, layer, outcome)); err != nil {
		logrus.Warnf("could not set %s commit status for run %s: %s", phase, run.Name, err)
	}
}

// resultMessage mirrors the layer's "Last Result" field. While run is still pending or
// running, that field still reflects the previous run, which is the best we have. Once run
// has finished, fetch its own plan/apply summary instead, since the layer's cached field
// won't be refreshed with run's outcome until terraformlayer's next reconciliation.
func (r *Reconciler) resultMessage(run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, outcome string) string {
	if outcome != commitstatus.Succeeded && outcome != commitstatus.Failed {
		return layer.Status.LastResult
	}
	result, err := r.Datastore.GetPlan(layer.Namespace, layer.Name, run.Name, "", "short")
	if err != nil {
		logrus.Warnf("could not get result of run %s for commit status: %s", run.Name, err)
		return "Error getting last Result"
	}
	return string(result)
}

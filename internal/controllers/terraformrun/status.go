package terraformrun

import (
	"context"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/status"
	"github.com/padok-team/burrito/internal/repository/commitstatus"
)

// applySucceeded is what the runner records as a successful apply's result, and the
// fallback wording when the applied diff cannot be recovered.
const applySucceeded = "Apply Successful"

// postCommitStatus posts a plan/apply commit status scoped to layer for run, best-effort:
// a failure here must not block the reconciliation.
func (r *Reconciler) postCommitStatus(ctx context.Context, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, state status.State, outcome string) {
	// Whether a run is worth reporting is decided once, by the layer controller that
	// created it: a drift detection re-plan of an already planned commit is not.
	if run.Annotations[annotations.PostCommitStatus] != "true" {
		return
	}
	if run.Spec.Layer.Revision == "" {
		return
	}
	log := log.WithContext(ctx)

	provider, err := r.getAPIProvider(repository)
	if err != nil {
		log.Warnf("could not get API provider to set commit status for run %s: %s", run.Name, err)
		return
	}

	phase := status.PhasePlan
	if run.Spec.Action == string(ApplyAction) {
		phase = status.PhaseApply
	}
	targetURL := commitstatus.LogsURL(r.Config.Server.PublicURL, layer, run.Name)
	if err := commitstatus.Post(provider, repository, layer, phase, state, run.Spec.Layer.Revision, r.resultMessage(ctx, run, layer, repository, outcome), targetURL); err != nil {
		// Already logged inside Post with more specific context; best-effort, nothing more to do.
		return
	}
}

// resultMessage mirrors the layer's "Last Result" field. While run is still pending or
// running, that field still reflects the previous run, which is the best we have. Once run
// has finished, describe run itself instead, since the layer's cached field won't be
// refreshed with run's outcome until terraformlayer's next reconciliation.
func (r *Reconciler) resultMessage(ctx context.Context, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository, outcome string) string {
	switch {
	case outcome == commitstatus.Failed:
		// A run that failed never got to write a result artifact, so there is nothing to
		// fetch: say which phase failed and let the status's "Details" link lead to the logs.
		if run.Spec.Action == string(ApplyAction) {
			return "Apply failed"
		}
		return "Plan failed"
	case outcome != commitstatus.Succeeded:
		return layer.Status.LastResult
	case run.Spec.Action == string(ApplyAction):
		return r.appliedDiff(ctx, run, layer, repository)
	}
	result, err := r.Datastore.GetPlan(layer.Namespace, layer.Name, run.Name, "", "short")
	if err != nil {
		log.WithContext(ctx).Warnf("could not get result of run %s for commit status: %s", run.Name, err)
		return "Error getting last Result"
	}
	return string(result)
}

// appliedDiff describes what a successful apply changed. An apply produces no diff of its
// own — its "short" artifact is the plain "Apply Successful" that feeds the layer's "Last
// Result" — so the summary is taken from the plan run it applied, whose "short" artifact is
// a single line. Falls back to the plain wording when that plan was not reused
// (applyWithoutPlanArtifact, where the recorded diff would be stale) or is unavailable.
func (r *Reconciler) appliedDiff(ctx context.Context, run *configv1alpha1.TerraformRun, layer *configv1alpha1.TerraformLayer, repository *configv1alpha1.TerraformRepository) string {
	if run.Spec.Artifact.Run == "" || configv1alpha1.GetApplyWithoutPlanArtifactEnabled(repository, layer) {
		return applySucceeded
	}
	shortDiff, err := r.Datastore.GetPlan(layer.Namespace, layer.Name, run.Spec.Artifact.Run, run.Spec.Artifact.Attempt, "short")
	if err != nil {
		log.WithContext(ctx).Warnf("could not get the plan applied by run %s for commit status: %s", run.Name, err)
		return applySucceeded
	}
	if len(shortDiff) == 0 {
		return applySucceeded
	}
	// The plan summary reads "Plan: 1 to create, …"; state it in the past tense instead,
	// since these resources have just been applied.
	return "Applied: " + strings.TrimPrefix(string(shortDiff), "Plan: ")
}

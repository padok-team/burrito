package terraformpullrequest

import (
	"context"
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	logrus "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	DiscoveryNeeded    string = "DiscoveryNeeded"
	Planning           string = "Planning"
	CommentNeeded      string = "CommentNeeded"
	Idle               string = "Idle"
	WaitingForApply    string = "WaitingForApply"
	ApplyCommentNeeded string = "ApplyCommentNeeded"
)

type State struct {
	handler func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result
	Status  configv1alpha1.TerraformPullRequestStatus
}

func (s *State) Handler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result {
	return s.handler(ctx, r, repository, pr, s)
}

func (r *Reconciler) GetState(ctx context.Context, pr *configv1alpha1.TerraformPullRequest) State {
	var state State
	logger := logrus.WithContext(ctx)

	// Check merge state first — bypasses the open-PR lifecycle entirely.
	cMerged, isMerged := r.IsMerged(pr)
	if isMerged {
		mainLayers, err := r.getMainBranchLayers(ctx, pr)
		if err != nil {
			logger.Errorf("failed to get main branch layers for merged pull request %s: %s", pr.Name, err)
		}
		cApplied, areLayersApplied, applyResults := r.AreLayersApplied(ctx, pr, mainLayers)
		state = State{
			Status: configv1alpha1.TerraformPullRequestStatus{
				Conditions:           []metav1.Condition{cMerged, cApplied},
				LastDiscoveredCommit: pr.Status.LastDiscoveredCommit,
				LastCommentedCommit:  pr.Status.LastCommentedCommit,
			},
		}
		if areLayersApplied {
			logger.Infof("merged pull request %s layers have applied, posting apply comment", pr.Name)
			state.handler = makeApplyCommentNeededHandler(applyResults)
			state.Status.State = ApplyCommentNeeded
		} else {
			logger.Infof("merged pull request %s is waiting for layers to apply", pr.Name)
			state.handler = waitingForApplyHandler
			state.Status.State = WaitingForApply
		}
		return state
	}

	c1, isLastCommitDiscovered := r.IsLastCommitDiscovered(pr)
	c2, areLayersStillPlanning := r.AreLayersStillPlanning(pr)
	c3, isCommentUpToDate := r.IsCommentUpToDate(pr)
	conditions := []metav1.Condition{c1, c2, c3}
	state = State{
		Status: configv1alpha1.TerraformPullRequestStatus{
			Conditions:           conditions,
			LastDiscoveredCommit: pr.Status.LastDiscoveredCommit,
			LastCommentedCommit:  pr.Status.LastCommentedCommit,
		},
	}
	switch {
	case !isLastCommitDiscovered:
		logger.Infof("pull request %s needs to be discovered", pr.Name)
		state.handler = discoveryNeededHandler
		state.Status.State = DiscoveryNeeded
	case isLastCommitDiscovered && isCommentUpToDate:
		logger.Infof("pull request %s comment is up to date", pr.Name)
		state.handler = idleHandler
		state.Status.State = Idle
	case isLastCommitDiscovered && areLayersStillPlanning:
		logger.Infof("pull request %s layers are still planning, waiting", pr.Name)
		state.handler = planningHandler
		state.Status.State = Planning
	case isLastCommitDiscovered && !areLayersStillPlanning && !isCommentUpToDate:
		logger.Infof("pull request %s layers have finished, posting comment", pr.Name)
		state.handler = commentNeededHandler
		state.Status.State = CommentNeeded
	default:
		logger.Infof("pull request %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is not an intended behavior.", pr.Name)
		state.handler = idleHandler
		state.Status.State = Idle
	}
	return state
}

func idleHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	return ctrl.Result{}
}

func planningHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

func discoveryNeededHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	err := r.deleteTempLayers(ctx, pr)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to delete temp layers for pull request")
		logrus.Errorf("failed to delete temp layers for pull request %s: %s", pr.Name, err)
	}
	layers, err := r.getAffectedLayers(repository, pr)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to get affected layers for pull request")
		logrus.Errorf("failed to get affected layers for pull request %s: %s", pr.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
	}
	newLayers := generateTempLayers(pr, layers)
	for _, layer := range newLayers {
		err := r.Client.Create(ctx, &layer)
		if err != nil {
			logrus.Errorf("failed to create layer %s: %s", layer.Name, err)
			r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to create layer for pull request")
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		r.Recorder.Event(pr, corev1.EventTypeNormal, "Reconciliation", fmt.Sprintf("Created layer %s", layer.Name))
	}
	state.Status.LastDiscoveredCommit = pr.Annotations[annotations.LastBranchCommit]
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

func commentNeededHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	layers, err := GetLinkedLayers(r.Client, pr)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to get linked layers for pull request")
		logrus.Errorf("failed to get linked layers for pull request %s: %s", pr.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
	}

	provider, err := r.getAPIProvider(repository)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Provider error", "Failed to get API provider for commenting pull request")
		logrus.Errorf("failed to get API provider for commenting pull request %s: %s", pr.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
	}

	c := comment.NewDefaultComment(layers, r.Datastore)
	err = provider.Comment(repository, pr, c)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to comment pull request")
		logrus.Errorf("failed to comment pull request: %s", err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
	}
	r.Recorder.Event(pr, corev1.EventTypeNormal, "Reconciliation", "Commented pull request")
	state.Status.LastCommentedCommit = pr.Annotations[annotations.LastBranchCommit]

	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

// waitingForApplyHandler just waits: each affected layer's own TerraformRun already posts
// its own plan/apply commit status (see internal/controllers/terraformrun).
func waitingForApplyHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

func makeApplyCommentNeededHandler(applyResults []LayerApplyResult) func(context.Context, *Reconciler, *configv1alpha1.TerraformRepository, *configv1alpha1.TerraformPullRequest, *State) ctrl.Result {
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
		provider, err := r.getAPIProvider(repository)
		if err != nil {
			logrus.Errorf("failed to get API provider for merged PR %s. Requeuing: %s", pr.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
		}

		reportedLayers := make([]comment.ApplyReportedLayer, 0, len(applyResults))
		for _, res := range applyResults {
			reportedLayers = append(reportedLayers, comment.ApplyReportedLayer{
				Path:      res.Layer.Spec.Path,
				Succeeded: res.Succeeded,
			})
		}

		applyComment := comment.NewApplyComment(reportedLayers)
		err = provider.Comment(repository, pr, applyComment)
		if err != nil {
			r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to post apply comment on merged pull request")
			logrus.Errorf("an error occurred while posting apply comment on merged pull request %s: %s", pr.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		r.Recorder.Event(pr, corev1.EventTypeNormal, "Reconciliation", "Posted apply comment on merged pull request")

		// Delete the TerraformPullRequest resource now that we're done with it.
		if err := r.Client.Delete(ctx, pr); err != nil {
			logrus.Errorf("failed to delete TerraformPullRequest %s after apply comment: %s", pr.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		return ctrl.Result{}
	}
}

package terraformpullrequest

import (
	"context"
	"fmt"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	repo "github.com/padok-team/burrito/internal/repository"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	DiscoveryNeeded string = "DiscoveryNeeded"
	Planning        string = "Planning"
	CommentNeeded   string = "CommentNeeded"
	Idle            string = "Idle"
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
	log := log.WithContext(ctx)
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
		log.Infof("pull request %s needs to be discovered", pr.Name)
		state.handler = discoveryNeededHandler
		state.Status.State = DiscoveryNeeded
	case isLastCommitDiscovered && isCommentUpToDate:
		log.Infof("pull request %s comment is up to date", pr.Name)
		state.handler = idleHandler
		state.Status.State = Idle
	case isLastCommitDiscovered && areLayersStillPlanning:
		log.Infof("pull request %s layers are still planning, waiting", pr.Name)
		state.handler = planningHandler
		state.Status.State = Planning
	case isLastCommitDiscovered && !areLayersStillPlanning && !isCommentUpToDate:
		log.Infof("pull request %s layers have finished, posting comment", pr.Name)
		state.handler = commentNeededHandler
		state.Status.State = CommentNeeded
	default:
		log.Infof("pull request %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is not an intended behavior.", pr.Name)
		state.handler = idleHandler
		state.Status.State = Idle
	}
	return state
}

func idleHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

func planningHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

func discoveryNeededHandler(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest, state *State) ctrl.Result {
	err := r.deleteTempLayers(ctx, pr)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to delete temp layers for pull request")
		log.Errorf("failed to delete temp layers for pull request %s: %s", pr.Name, err)
	}
	layers, err := r.getAffectedLayers(repository, pr)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to get affected layers for pull request")
		log.Errorf("failed to get affected layers for pull request %s: %s", pr.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
	}
	newLayers := generateTempLayers(pr, layers)
	for _, layer := range newLayers {
		err := r.Client.Create(ctx, &layer)
		if err != nil {
			log.Errorf("failed to create layer %s: %s", layer.Name, err)
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
		log.Errorf("failed to get linked layers for pull request %s: %s", pr.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
	}

	provider, err := repo.GetAPIProviderFromRepository(r.Credentials, repository)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Provider error", "Failed to get API provider for commenting pull request")
		log.Errorf("failed to get API provider for commenting pull request %s: %s", pr.Name, err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
	}

	comment := comment.NewDefaultComment(layers, r.Datastore)
	err = provider.Comment(repository, pr, comment)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Failed to comment pull request")
		log.Errorf("failed to comment pull request: %s", err)
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
	}
	r.Recorder.Event(pr, corev1.EventTypeNormal, "Reconciliation", "Commented pull request")
	state.Status.LastCommentedCommit = pr.Annotations[annotations.LastBranchCommit]
	return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
}

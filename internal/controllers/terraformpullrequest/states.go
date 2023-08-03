package terraformpullrequest

import (
	"context"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State interface {
	getHandler() func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result
}

func (r *Reconciler) GetState(ctx context.Context, pr *configv1alpha1.TerraformPullRequest) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, isLastCommitDiscovered := r.IsLastCommitDiscovered(pr)
	c2, areLayersStillPlanning := r.AreLayersStillPlanning(pr)
	c3, isCommentUpToDate := r.IsCommentUpToDate(pr)
	conditions := []metav1.Condition{c1, c2, c3}
	switch {
	case !isLastCommitDiscovered:
		log.Infof("pull request %s needs to be discovered", pr.Name)
		return &DiscoveryNeeded{}, conditions
	case isLastCommitDiscovered && isCommentUpToDate:
		log.Infof("pull request %s comment is up to date", pr.Name)
		return &Idle{}, conditions
	case isLastCommitDiscovered && areLayersStillPlanning:
		log.Infof("pull request %s layers are still planning, waiting", pr.Name)
		return &Idle{}, conditions
	case isLastCommitDiscovered && !areLayersStillPlanning && !isCommentUpToDate:
		log.Infof("pull request %s layers have finished, posting comment", pr.Name)
		return &CommentNeeded{}, conditions
	default:
		log.Infof("pull request %s is in an unknown state, defaulting to idle. If this happens please file an issue, this is not an intended behavior.", pr.Name)
		return &Idle{}, conditions
	}
}

type Idle struct{}

func (s *Idle) getHandler() func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result {
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result {
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
	}
}

type DiscoveryNeeded struct{}

func (s *DiscoveryNeeded) getHandler() func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result {
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result {
		layers, err := r.getAffectedLayers(repository, pr)
		if err != nil {
			log.Errorf("failed to get affected layers for pull request %s: %s", pr.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		newLayers := generateTempLayers(pr, layers)
		for _, layer := range newLayers {
			err := r.Client.Create(ctx, &layer)
			if errors.IsAlreadyExists(err) {
				log.Infof("layer %s already exists, updating it", layer.Name)
				err = r.Client.Update(ctx, &layer)
				if err != nil {
					log.Errorf("failed to update layer %s: %s", layer.Name, err)
					return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
				}
			}
			if err != nil {
				log.Errorf("failed to create layer %s: %s", layer.Name, err)
				return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
			}
		}
		err = annotations.Add(ctx, r.Client, pr, map[string]string{annotations.LastDiscoveredCommit: pr.Annotations[annotations.LastBranchCommit]})
		if err != nil {
			log.Errorf("failed to add annotation %s to pull request %s: %s", annotations.LastDiscoveredCommit, pr.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
	}
}

type CommentNeeded struct{}

func (s *CommentNeeded) getHandler() func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result {
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository, pr *configv1alpha1.TerraformPullRequest) ctrl.Result {
		layers, err := GetLinkedLayers(r.Client, pr)
		if err != nil {
			log.Errorf("failed to get linked layers for pull request %s: %s", pr.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}

		var provider Provider
		found := false
		for _, p := range r.Providers {
			if p.IsFromProvider(pr) {
				provider = p
				found = true
			}
		}
		if !found {
			log.Infof("failed to get pull request provider. Requeuing")
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
		}

		comment := comment.NewDefaultComment(layers, r.Storage)
		err = provider.Comment(repository, pr, comment)
		if err != nil {
			log.Errorf("an error occurred while commenting pull request: %s", err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		err = annotations.Add(ctx, r.Client, pr, map[string]string{annotations.LastCommentedCommit: pr.Annotations[annotations.LastDiscoveredCommit]})
		if err != nil {
			log.Errorf("failed to add annotation %s to pull request %s: %s", annotations.LastCommentedCommit, pr.Name, err)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.OnError}
		}
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}
	}
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

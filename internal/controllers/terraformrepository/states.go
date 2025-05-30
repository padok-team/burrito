package terraformrepository

import (
	"context"
	"fmt"
	"strings"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	layerCtrl "github.com/padok-team/burrito/internal/controllers/terraformlayer"
	repo "github.com/padok-team/burrito/internal/repository"
	"github.com/padok-team/burrito/internal/repository/types"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	SyncStatusSuccess string = "success"
	SyncStatusFailed  string = "failed"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformRepository) (ctrl.Result, []configv1alpha1.BranchState)

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, repository *configv1alpha1.TerraformRepository) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, IsLastSyncTooOld := r.IsLastSyncTooOld(repository)
	c2, HasLastSyncFailed := r.HasLastSyncFailed(repository)
	conditions := []metav1.Condition{c1, c2}

	if IsLastSyncTooOld || HasLastSyncFailed {
		log.Infof("repository %s needs to be synced", repository.Name)
		return &SyncNeeded{}, conditions
	}

	log.Infof("repository %s is in sync with remote", repository.Name)
	return &Synced{}, conditions
}

type SyncNeeded struct{}

func (s *SyncNeeded) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository) (ctrl.Result, []configv1alpha1.BranchState) {
		log := log.WithContext(ctx)
		branchStates := repository.Status.Branches
		gitProvider, err := repo.GetGitProviderFromRepository(r.Credentials, repository)
		if err != nil {
			r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get git provider: %s", err))
			log.Errorf("failed to get git provider for repo %s/%s: %s", repository.Namespace, repository.Name, err)
			return ctrl.Result{}, branchStates
		}

		// Update the list of layer branches by querying the TerraformLayer resources
		layers, err := r.retrieveManagedLayers(ctx, repository)
		if err != nil {
			r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to list managed layers: %s", err))
			log.Errorf("failed to list managed layers by repo %s/%s: %s", repository.Namespace, repository.Name, err)
			return ctrl.Result{}, branchStates
		}
		if len(layers) == 0 {
			log.Warningf("no managed layers found for repository %s/%s, have you created TerraformLayer resources?", repository.Namespace, repository.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, []configv1alpha1.BranchState{}
		}
		layerBranches := retrieveAllLayerRefs(layers)

		// add in branchStates branches that were not previously managed
		branchStates = mergeBranchesWithBranchState(layerBranches, branchStates)

		// Update datastore with latest revisions for each ref that needs to be synced
		var syncError error
		for _, branch := range branchStates {
			// Filter out branches that have been synced succesfully recently or do not have been requested to sync now
			if lastSync, err := time.Parse(time.UnixDate, branch.LastSyncDate); err == nil {
				syncNow, err := isSyncNowRequested(repository, branch.Name, lastSync)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to parse sync now annotation for ref %s: %s", branch.Name, err))
					continue
				}
				nextSyncTime := lastSync.Add(r.Config.Controller.Timers.RepositorySync)
				now := time.Now()
				if !syncNow && !nextSyncTime.Before(now) && branch.LastSyncStatus == SyncStatusSuccess {
					continue
				}
			}

			latestRev, err := gitProvider.GetLatestRevisionForRef(branch.Name)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get remote revision for ref %s: %s", branch.Name, err))
				log.Errorf("failed to get remote revision for ref %s: %s", branch.Name, err)
				syncError = err
				branchStates = updateBranchState(branchStates, branch.Name, "", SyncStatusFailed)
				continue
			}
			log.Infof("latest revision for repository %s/%s ref %s is %s", repository.Namespace, repository.Name, branch.Name, latestRev)

			isSynced, err := r.Datastore.CheckGitBundle(repository.Namespace, repository.Name, branch.Name, latestRev)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to check stored revision for ref %s: %s", branch.Name, err))
				log.Errorf("failed to check stored revision for ref %s: %s", branch.Name, err)
				syncError = err
				branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusFailed)
				continue
			}

			if isSynced {
				log.Infof("repository %s/%s is in sync with remote for ref %s: rev %s", repository.Namespace, repository.Name, branch.Name, latestRev)
				branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusSuccess)
				syncError = annotateLayers(r.Client, gitProvider, retrieveLayersForRef(branch.Name, layers), latestRev)
				continue
			} else {
				log.Infof("repository %s/%s is out of sync with remote for ref %s. Syncing...", repository.Namespace, repository.Name, branch.Name)
				bundle, err := gitProvider.Bundle(branch.Name)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get revision bundle for ref %s: %s", branch.Name, err))
					log.Errorf("failed to get revision bundle for ref %s: %s", branch.Name, err)
					syncError = err
					branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusFailed)
					continue
				}

				err = r.Datastore.PutGitBundle(repository.Namespace, repository.Name, branch.Name, latestRev, bundle)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to store revision for ref %s: %s", branch.Name, err))
					log.Errorf("failed to store revision for ref %s: %s", branch.Name, err)
					syncError = err
					branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusFailed)
					continue
				}
				log.Infof("stored new bundle for repository %s/%s ref:%s revision:%s", repository.Namespace, repository.Name, branch.Name, latestRev)
				branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusSuccess)

				// Add annotation to trigger a sync for all layers that depend on this branch
				syncError = annotateLayers(r.Client, gitProvider, retrieveLayersForRef(branch.Name, layers), latestRev)
			}
		}
		if syncError != nil {
			return ctrl.Result{}, branchStates
		}

		r.Recorder.Event(repository, corev1.EventTypeNormal, "Reconciliation", "Repository sync completed")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, branchStates
	}
}

type Synced struct{}

func (s *Synced) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository) (ctrl.Result, []configv1alpha1.BranchState) {
		r.Recorder.Event(repository, corev1.EventTypeNormal, "Reconciliation", "Repository is in sync with remote")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.RepositorySync}, repository.Status.Branches
	}
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

func updateBranchState(branchStates []configv1alpha1.BranchState, branch, rev, status string) []configv1alpha1.BranchState {
	for i, b := range branchStates {
		if b.Name == branch {
			branchStates[i].LastSyncDate = time.Now().Format(time.UnixDate)
			branchStates[i].LatestRev = rev
			branchStates[i].LastSyncStatus = status
			return branchStates
		}
	}
	return branchStates
}

func annotateLayers(c client.Client, gitProvider types.GitProvider, layers []configv1alpha1.TerraformLayer, latestRev string) error {
	var err error
	date := time.Now().Format(time.UnixDate)
	for _, layer := range layers {
		ann := map[string]string{}

		// If the layer already has the latest branch commit == latestRev, we skip it
		if currentLastBranch, ok := layer.Annotations[annotations.LastBranchCommit]; !ok || currentLastBranch != latestRev {
			ann[annotations.LastBranchCommit] = latestRev
			ann[annotations.LastBranchCommitDate] = date
		}

		// If the layer does not have a last relevant commit, we set it to the last branch commit
		if currentLastRelevant, ok := layer.Annotations[annotations.LastRelevantCommit]; !ok {
			ann[annotations.LastRelevantCommit] = latestRev
			ann[annotations.LastRelevantCommitDate] = date
		} else {
			// We compare the current last relevant commit with the latest revision
			changes := gitProvider.GetChanges(currentLastRelevant, latestRev)
			if layerCtrl.LayerFilesHaveChanged(layer, changes) {
				log.Infof("layer %s/%s is affected by new revision %s", layer.Namespace, layer.Name, latestRev)
				ann[annotations.LastRelevantCommit] = latestRev
				ann[annotations.LastRelevantCommitDate] = date
			}
		}

		err = annotations.Add(context.TODO(), c, &layer, ann)
		if err != nil {
			log.Errorf("could not add annotation to TerraformLayer %s/%s: %s", layer.Namespace, layer.Name, err)
		} else {
			log.Infof("layer %s/%s annotated with new revision %s", layer.Namespace, layer.Name, latestRev)
		}
	}
	return err
}

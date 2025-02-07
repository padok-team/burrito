package terraformrepository

import (
	"context"
	"fmt"
	"strings"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/utils/gitprovider"
	gt "github.com/padok-team/burrito/internal/utils/gitprovider/types"
	"github.com/padok-team/burrito/internal/utils/typeutils"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
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
		// Initialize git providers for the repository if needed
		if _, ok := r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)]; !ok {
			provider, err := r.initializeProvider(ctx, repository)
			if err != nil {
				log.Errorf("could not initialize provider for repository %s: %s", repository.Name, err)
				return ctrl.Result{}, branchStates
			}
			if provider != nil {
				log.Infof("initialized git provider for repository %s/%s", repository.Namespace, repository.Name)
				r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)] = provider
			}
		}

		// Update the list of layer branches by querying the TerraformLayer resources
		layerBranches, err := r.retrieveLayerBranches(ctx, repository)
		if err != nil {
			r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", "Failed to list managed branches")
			log.Errorf("failed to list managed branches: %s", err)
			return ctrl.Result{}, branchStates
		}
		if len(layerBranches) == 0 {
			log.Warningf("no managed branches found for repository %s/%s, have you created TerraformLayer resources?", repository.Namespace, repository.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, []configv1alpha1.BranchState{}
		}

		// add in branchStates branches that were not previously managed
		branchStates = mergeBranchesWithBranchState(layerBranches, branchStates)

		// Update datastore with latest revisions for each ref that needs to be synced
		var syncError error
		for _, branch := range branchStates {
			// Filter out branches that have been synced succesfully recently or do not have been requested to sync now
			if lastSync, err := time.Parse(time.UnixDate, branch.LastSyncDate); err == nil {
				syncNow, err := isSyncNowRequested(repository, branch.Name, lastSync)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to parse sync now annotation for ref %s", branch.Name))
					continue
				}
				nextSyncTime := lastSync.Add(r.Config.Controller.Timers.RepositorySync)
				now := time.Now()
				if !syncNow && !nextSyncTime.Before(now) && branch.LastSyncStatus == SyncStatusSuccess {
					continue
				}
			}

			latestRev, err := r.getRemoteRevision(repository, branch.Name)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get remote revision for ref %s", branch.Name))
				log.Errorf("failed to get remote revision for ref %s: %s", branch.Name, err)
				syncError = err
				branchStates = updateBranchState(branchStates, branch.Name, "", SyncStatusFailed)
				continue
			}
			log.Infof("latest revision for repository %s/%s ref:%s is %s", repository.Namespace, repository.Name, branch.Name, latestRev)

			isSynced, err := r.Datastore.CheckGitBundle(repository.Namespace, repository.Name, branch.Name, latestRev)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to check stored revision for ref %s", branch.Name))
				log.Errorf("failed to check stored revision for ref %s: %s", branch.Name, err)
				syncError = err
				branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusFailed)
				continue
			}

			if isSynced {
				log.Infof("repository %s/%s is in sync with remote for ref %s: rev %s", repository.Namespace, repository.Name, branch.Name, latestRev)
				branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusSuccess)
				continue
			} else {
				log.Infof("repository %s/%s is out of sync with remote for ref %s. Syncing...", repository.Namespace, repository.Name, branch.Name)
				bundle, err := r.getRevisionBundle(repository, branch.Name, latestRev)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get revision bundle for ref %s", branch.Name))
					log.Errorf("failed to get revision bundle for ref %s: %s", branch.Name, err)
					syncError = err
					branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusFailed)
					continue
				}

				err = r.Datastore.PutGitBundle(repository.Namespace, repository.Name, branch.Name, latestRev, bundle)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to store revision for ref %s", branch.Name))
					log.Errorf("failed to store revision for ref %s: %s", branch.Name, err)
					syncError = err
					branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusFailed)
					continue
				}
				log.Infof("stored new bundle for repository %s/%s ref:%s revision:%s", repository.Namespace, repository.Name, branch.Name, latestRev)
				branchStates = updateBranchState(branchStates, branch.Name, latestRev, SyncStatusSuccess)
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

func (r *Reconciler) initializeProvider(ctx context.Context, repository *configv1alpha1.TerraformRepository) (gitprovider.Provider, error) {
	if repository.Spec.Repository.Url == "" {
		return nil, fmt.Errorf("no repository URL found in TerraformRepository.spec.repository.url for repository %s. Skipping provider initialization", repository.Name)
	}
	var config gitprovider.Config

	if repository.Spec.Repository.SecretName != "" {
		secret := &corev1.Secret{}
		err := r.Client.Get(ctx, types.NamespacedName{
			Name:      repository.Spec.Repository.SecretName,
			Namespace: repository.Namespace,
		}, secret)
		if err != nil {
			log.Errorf("failed to get credentials secret for repository %s: %s", repository.Name, err)
			config = gitprovider.Config{
				URL: repository.Spec.Repository.Url,
			}
		} else {
			config = gitprovider.Config{
				URL:        repository.Spec.Repository.Url,
				EnableMock: secret.Data["enableMock"] != nil && string(secret.Data["enableMock"]) == "true",
				// GitHub App Auth
				AppID:             typeutils.ParseSecretInt64(secret.Data["githubAppId"]),
				AppInstallationID: typeutils.ParseSecretInt64(secret.Data["githubAppInstallationId"]),
				AppPrivateKey:     string(secret.Data["githubAppPrivateKey"]),
				// Token Auth
				GitHubToken: string(secret.Data["githubToken"]),
				GitLabToken: string(secret.Data["gitlabToken"]),
				// Basic Auth
				Username: string(secret.Data["username"]),
				Password: string(secret.Data["password"]),
				// SSH Auth
				SSHPrivateKey: string(secret.Data["sshPrivateKey"]),
			}
		}
	} else {
		log.Infof("no secret configured for repository %s/%s, using empty config", repository.Namespace, repository.Name)
		config = gitprovider.Config{
			URL: repository.Spec.Repository.Url,
		}
	}

	provider, err := gitprovider.New(config, []string{gt.Capabilities.Clone})
	if err != nil {
		log.Errorf("failed to create provider for repository %s: %s", repository.Name, err)
		return nil, err
	}

	err = provider.Init()
	if err != nil {
		log.Errorf("failed to initialize provider for repository %s: %s", repository.Name, err)
		return nil, err
	}
	return provider, nil
}

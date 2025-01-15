package terraformrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	storageerrors "github.com/padok-team/burrito/internal/datastore/storage/error"
	"github.com/padok-team/burrito/internal/utils/gitprovider"
	gt "github.com/padok-team/burrito/internal/utils/gitprovider/types"
	"github.com/padok-team/burrito/internal/utils/typeutils"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformRepository) (ctrl.Result, error)

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
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository) (ctrl.Result, error) {
		log := log.WithContext(ctx)
		// Initialize git providers for the repository if needed
		if _, ok := r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)]; !ok {
			provider, err := r.initializeProvider(ctx, repository)
			if err != nil {
				log.Errorf("could not initialize provider for repository %s: %s", repository.Name, err)
				return ctrl.Result{}, err
			}
			if provider != nil {
				log.Infof("initialized git provider for repository %s/%s", repository.Namespace, repository.Name)
				r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)] = provider
			}
		}

		// Get managed refs
		managedRefs, err := r.listManagedRefs(ctx, repository)
		if err != nil {
			r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", "Failed to list managed refs")
			log.Errorf("failed to list managed refs: %s", err)
			return ctrl.Result{}, err
		}
		if len(managedRefs) == 0 {
			log.Warningf("no managed refs found for repository %s/%s, have you created TerraformLayer resources?", repository.Namespace, repository.Name)
			return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
		}

		// Update datastore with latest revisions for each ref
		var syncError error
		for _, ref := range managedRefs {
			latestRev, err := r.getRemoteRevision(repository, ref)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get remote revision for ref %s", ref))
				log.Errorf("failed to get remote revision for ref %s: %s", ref, err)
				syncError = err
				continue
			}
			log.Infof("latest revision for repository %s/%s ref:%s is %s", repository.Namespace, repository.Name, ref, latestRev)

			storedRev, err := r.Datastore.GetLatestRevision(repository.Namespace, repository.Name, ref)
			var storageErr *storageerrors.StorageError
			if errors.As(err, &storageErr) && storageErr.Nil {
				log.Infof("no stored revision found for ref %s", ref)
			} else if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get stored revision for ref %s", ref))
				log.Errorf("failed to get stored revision for ref %s: %s", ref, err)
				syncError = err
				continue
			}
			if storedRev != "" {
				log.Infof("current stored revision for repository %s/%s ref:%s is %s", repository.Namespace, repository.Name, ref, storedRev)
			}

			if latestRev == storedRev {
				log.Infof("repository %s/%s is in sync with remote for ref %s", repository.Namespace, repository.Name, ref)
				continue
			} else {
				log.Infof("repository %s/%s is out of sync with remote for ref %s. Syncing...", repository.Namespace, repository.Name, ref)
				bundle, err := r.getRevisionBundle(repository, ref, latestRev)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get revision bundle for ref %s", ref))
					log.Errorf("failed to get revision bundle for ref %s: %s", ref, err)
					syncError = err
					continue
				}

				err = r.Datastore.PutGitBundle(repository.Namespace, repository.Name, ref, latestRev, bundle)
				if err != nil {
					r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to store revision for ref %s", ref))
					log.Errorf("failed to store revision for ref %s: %s", ref, err)
					syncError = err
					continue
				}
				log.Infof("stored new bundle for repository %s/%s ref:%s revision:%s", repository.Namespace, repository.Name, ref, latestRev)
			}
		}
		if syncError != nil {
			return ctrl.Result{}, syncError
		}

		r.Recorder.Event(repository, corev1.EventTypeNormal, "Reconciliation", "Repository sync completed")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.WaitAction}, nil
	}
}

type Synced struct{}

func (s *Synced) getHandler() Handler {
	return func(ctx context.Context, r *Reconciler, repository *configv1alpha1.TerraformRepository) (ctrl.Result, error) {
		r.Recorder.Event(repository, corev1.EventTypeNormal, "Reconciliation", "Repository is in sync with remote")
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.RepositorySync}, nil
	}
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
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

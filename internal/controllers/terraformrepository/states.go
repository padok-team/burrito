package terraformrepository

import (
	"context"
	"fmt"
	"strings"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
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

		// Update datastore with latest revisions
		var syncError error
		for ref := range managedRefs {
			rev, err := r.getRemoteRevision(repository, ref)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get remote revision for ref %s", ref))
				log.Errorf("failed to get remote revision for ref %s: %s", ref, err)
				syncError = err
				continue
			}

			// TODO: Download git bundle for this revision
			bundle, err := r.getRevisionBundle(ctx, repository, ref, rev)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to get bundle for ref %s", ref))
				log.Errorf("failed to get bundle for ref %s: %s", ref, err)
				syncError = err
				continue
			}
			err = r.Datastore.StoreRevision(repository.Namespace, repository.Name, ref, rev, bundle)
			if err != nil {
				r.Recorder.Event(repository, corev1.EventTypeWarning, "Reconciliation", fmt.Sprintf("Failed to store revision for ref %s", ref))
				log.Errorf("failed to store revision for ref %s: %s", ref, err)
				syncError = err
				continue
			}
		}
		// Update annotations
		if repository.Annotations == nil {
			repository.Annotations = make(map[string]string)
		}
		if syncError != nil {
			repository.Annotations[annotations.LastSyncStatus] = annotations.SyncStatusFailed
		} else {
			repository.Annotations[annotations.LastSyncStatus] = annotations.SyncStatusSuccess
		}
		repository.Annotations[annotations.LastSyncDate] = time.Now().Format(time.UnixDate)

		if err := r.Client.Update(ctx, repository); err != nil {
			log.Errorf("failed to update repository annotations: %s", err)
			return ctrl.Result{}, err
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
	if repository.Spec.Repository.SecretName == "" {
		log.Debugf("no secret configured for repository %s/%s, skipping provider initialization", repository.Namespace, repository.Name)
		return nil, nil
	}
	secret := &corev1.Secret{}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      repository.Spec.Repository.SecretName,
		Namespace: repository.Namespace,
	}, secret)
	if err != nil {
		log.Errorf("failed to get credentials secret for repository %s: %s", repository.Name, err)
		return nil, err
	}
	config := gitprovider.Config{
		AppID:             typeutils.ParseSecretInt64(secret.Data["githubAppId"]),
		URL:               repository.Spec.Repository.Url,
		AppInstallationID: typeutils.ParseSecretInt64(secret.Data["githubAppInstallationId"]),
		AppPrivateKey:     string(secret.Data["githubAppPrivateKey"]),
		GitHubToken:       string(secret.Data["githubToken"]),
		GitLabToken:       string(secret.Data["gitlabToken"]),
		EnableMock:        secret.Data["enableMock"] != nil && string(secret.Data["enableMock"]) == "true",
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

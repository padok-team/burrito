package terraformrepository

import (
	"context"
	"fmt"
	"strings"
	"time"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Handler func(context.Context, *Reconciler, *configv1alpha1.TerraformRepository) (ctrl.Result, error)

type State interface {
	getHandler() Handler
}

func (r *Reconciler) GetState(ctx context.Context, repository *configv1alpha1.TerraformRepository) (State, []metav1.Condition) {
	log := log.WithContext(ctx)
	c1, IsLastSyncTooOld := r.IsLastSyncTooOld(repository)
	conditions := []metav1.Condition{c1}

	if IsLastSyncTooOld {
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
			rev, err := r.getRemoteRevision(ctx, repository, ref)
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

		// Update last sync date annotation
		if repository.Annotations == nil {
			repository.Annotations = make(map[string]string)
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
		return ctrl.Result{RequeueAfter: r.Config.Controller.Timers.DriftDetection}, nil
	}
}

func getStateString(state State) string {
	t := strings.Split(fmt.Sprintf("%T", state), ".")
	return t[len(t)-1]
}

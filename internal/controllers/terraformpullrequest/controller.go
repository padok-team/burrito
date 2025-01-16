package terraformpullrequest

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	log "github.com/sirupsen/logrus"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/utils/gitprovider"
	gt "github.com/padok-team/burrito/internal/utils/gitprovider/types"
	"github.com/padok-team/burrito/internal/utils/typeutils"
)

// Reconciler reconciles a TerraformPullRequest object
type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Config    *config.Config
	Providers map[string]gitprovider.Provider
	Recorder  record.EventRecorder
	Datastore datastore.Client
}

//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformpullrequests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformpullrequests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=config.terraform.padok.cloud,resources=terraformpullrequests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the TerraformLayer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.WithContext(ctx)
	log.Infof("starting reconciliation for pull request %s/%s ...", req.Namespace, req.Name)
	pr := &configv1alpha1.TerraformPullRequest{}
	err := r.Client.Get(ctx, req.NamespacedName, pr)
	if errors.IsNotFound(err) {
		log.Errorf("resource not found. Ignoring since object must be deleted: %s", err)
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Errorf("failed to get TerraformPullRequest: %s", err)
		return ctrl.Result{}, err
	}
	repository := &configv1alpha1.TerraformRepository{}
	err = r.Client.Get(ctx, types.NamespacedName{
		Name:      pr.Spec.Repository.Name,
		Namespace: pr.Spec.Repository.Namespace,
	}, repository)
	if errors.IsNotFound(err) {
		log.Errorf("repository not found. object must not be configured correctly: %s", err)
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Errorf("failed to get TerraformRepository: %s", err)
		return ctrl.Result{}, err
	}

	if _, ok := r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)]; !ok {
		provider, err := r.initializeProvider(ctx, repository)
		if err != nil {
			log.Errorf("could not initialize provider for repository %s: %s", repository.Name, err)
		}
		if provider != nil {
			r.Providers[fmt.Sprintf("%s/%s", repository.Namespace, repository.Name)] = provider
			log.Infof("initialized webhook handlers for repository %s/%s", repository.Namespace, repository.Name)
		}
	}
	state := r.GetState(ctx, pr)
	result := state.Handler(ctx, r, repository, pr)
	pr.Status = state.Status
	err = r.Client.Status().Update(ctx, pr)
	if err != nil {
		r.Recorder.Event(pr, corev1.EventTypeWarning, "Reconciliation", "Could not update pull request status")
		log.Errorf("could not update pull request %s status: %s", pr.Name, err)
	}
	log.Infof("finished reconciliation cycle for pull request %s/%s", pr.Namespace, pr.Name)
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Providers = make(map[string]gitprovider.Provider)
	err := r.initializeDefaultProviders()
	if err != nil {
		log.Errorf("Some legacy configuration was found, but could not initialize default providers: %s", err)
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformPullRequest{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: r.Config.Controller.MaxConcurrentReconciles}).
		WithEventFilter(ignorePredicate()).
		Complete(r)
}

func GetProviderForPullRequest(pr *configv1alpha1.TerraformPullRequest, r *Reconciler) (gitprovider.Provider, error) {
	for key, p := range r.Providers {
		if fmt.Sprintf("%s/%s", pr.Spec.Repository.Namespace, pr.Spec.Repository.Name) == key {
			return p, nil
		}
	}
	return nil, fmt.Errorf("no provider found for pull request %s", pr.Name)
}

func ignorePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Update only if generation or annotations change, filter out anything else.
			// We only need to check generation or annotations change here, because it is only
			// updated on spec changes. On the other hand RevisionVersion
			// changes also on status changes. We want to omit reconciliation
			// for status updates.
			return (e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()) ||
				cmp.Diff(e.ObjectOld.GetAnnotations(), e.ObjectNew.GetAnnotations()) != ""
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
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

	provider, err := gitprovider.New(config, []string{gt.Capabilities.Comment, gt.Capabilities.Changes})
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

// This function initializes default providers for the controller if user has provided legacy configuration
func (r *Reconciler) initializeDefaultProviders() error {
	log.Warningf("deprecated GitHub/GitLab configuration found. please configure repositories with secrets instead. See https://padok-team.github.io/burrito/operator-manual/git-authentication/#repository-secret for more information.")
	var config = gitprovider.Config{
		AppID:             r.Config.Controller.GithubConfig.AppId,
		AppInstallationID: r.Config.Controller.GithubConfig.InstallationId,
		AppPrivateKey:     r.Config.Controller.GithubConfig.PrivateKey,
		GitHubToken:       r.Config.Controller.GithubConfig.APIToken,
		GitLabToken:       r.Config.Controller.GitlabConfig.APIToken,
		URL:               r.Config.Controller.GitlabConfig.URL,
	}

	providers, err := gitprovider.ListAvailable(config, []string{gt.Capabilities.Changes, gt.Capabilities.Comment})
	if err != nil {
		return err
	}
	for _, provider := range providers {
		providerInstance, err := gitprovider.NewWithName(config, provider)
		if err != nil {
			return err
		}
		r.Providers["default_"+provider] = providerInstance
	}
	return nil
}

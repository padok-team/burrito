package terraformpullrequest

import (
	"context"
	"fmt"
	"strings"

	"github.com/padok-team/burrito/internal/burrito/config"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/comment"
	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/github"
	"github.com/padok-team/burrito/internal/storage"
	"github.com/padok-team/burrito/internal/storage/redis"

	"github.com/padok-team/burrito/internal/controllers/terraformpullrequest/gitlab"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	log "github.com/sirupsen/logrus"

	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
)

type Provider interface {
	Init(*config.Config) error
	IsFromProvider(*configv1alpha1.TerraformPullRequest) bool
	GetChanges(*configv1alpha1.TerraformRepository, *configv1alpha1.TerraformPullRequest) ([]string, error)
	Comment(*configv1alpha1.TerraformRepository, *configv1alpha1.TerraformPullRequest, comment.Comment) error
}

// Reconciler reconciles a TerraformPullRequest object
type Reconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Config    *config.Config
	Providers []Provider
	Storage   storage.Storage
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
	log.Infof("starting reconciliation...")
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
	state, conditions := r.GetState(ctx, pr)
	result := state.getHandler()(ctx, r, repository, pr)

	pr.Status = configv1alpha1.TerraformPullRequestStatus{Conditions: conditions, State: getStateString(state)}
	err = r.Client.Status().Update(ctx, pr)
	if err != nil {
		log.Errorf("could not update pull request %s status: %s", pr.Name, err)
	}
	log.Infof("finished reconciliation cycle for pull request %s", pr.Name)
	return result, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	providers := []Provider{}
	for _, p := range []Provider{&github.Github{}, &gitlab.Gitlab{}} {
		name := strings.Split(fmt.Sprintf("%T", p), ".")
		err := p.Init(r.Config)
		if err != nil {
			log.Warnf("could not initialize provider %s: %s", name, err)
			continue
		}
		log.Infof("provider %s successfully initialized", name)
		providers = append(providers, p)
	}
	r.Providers = providers
	r.Storage = redis.New(r.Config.Redis.URL, r.Config.Redis.Password, r.Config.Redis.Database)
	return ctrl.NewControllerManagedBy(mgr).
		For(&configv1alpha1.TerraformPullRequest{}).
		WithEventFilter(ignorePredicate()).
		Complete(r)
}

func ignorePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

package runner

import (
	"context"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/runner/tools"
	"github.com/padok-team/burrito/internal/utils"
	"github.com/padok-team/burrito/internal/utils/gitprovider"
	gt "github.com/padok-team/burrito/internal/utils/gitprovider/types"
	runnerutils "github.com/padok-team/burrito/internal/utils/runner"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Runner struct {
	config        *config.Config
	exec          tools.BaseExec
	Datastore     datastore.Client
	Client        client.Client
	GitProvider   gitprovider.Provider
	Layer         *configv1alpha1.TerraformLayer
	Run           *configv1alpha1.TerraformRun
	Repository    *configv1alpha1.TerraformRepository
	gitRepository *git.Repository
	workingDir    string
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}

// Entrypoint function of the runner. Initializes the runner and executes its action.
func (r *Runner) Exec() error {
	err := r.initClients()
	if err != nil {
		log.Errorf("error initializing runner clients: %s", err)
		return err
	}

	err = r.Init()
	if err != nil {
		log.Errorf("error initializing runner: %s", err)
		return err
	}

	err = r.ExecInit()
	if err != nil {
		log.Errorf("error executing init: %s", err)
		return err
	}

	return r.ExecAction()
}

// Initialize the runner clients (kubernetes, datastore).
func (r *Runner) initClients() error {
	kubeClient, err := utils.NewK8SClient()
	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err)
		return err
	}
	r.Client = kubeClient

	datastoreClient := datastore.NewDefaultClient(r.config.Datastore)
	r.Datastore = datastoreClient

	return nil
}

// Initialize the runner: retrieve linked resources (layer, run, repository),
// fetch the repository content, install the binaries and configure Hermitcrab mirror.
func (r *Runner) Init() error {
	log.Infof("retrieving linked TerraformLayer and TerraformRepository")
	err := r.GetResources()
	if err != nil {
		log.Errorf("error getting kubernetes resources: %s", err)
		return err
	}

	r.workingDir = filepath.Join(r.config.Runner.RepositoryPath, r.Layer.Spec.Path)

	err = r.initGitProvider()
	if err != nil {
		log.Errorf("error initializing git provider: %s", err)
	}
	log.Info("successfully initialized git provider")
	r.gitRepository, err = r.GitProvider.Clone(r.Repository, r.Layer.Spec.Branch, r.config.Runner.RepositoryPath)
	if err != nil {
		log.Errorf("error fetching repository: %s", err)
		return err
	}
	log.Infof("repository fetched successfully")

	log.Infof("installing binaries...")
	r.exec, err = tools.InstallBinaries(r.Layer, r.Repository, r.config.Runner.RunnerBinaryPath, r.workingDir)
	if err != nil {
		log.Errorf("error installing binaries: %s", err)
		return err
	}

	if r.config.Hermitcrab.Enabled {
		log.Infof("Hermitcrab configuration detected, creating network mirror configuration...")
		return r.EnableHermitcrab()
	}

	return nil
}

// Enable Hermitcrab network mirror configuration.
func (r *Runner) EnableHermitcrab() error {
	log.Infof("Hermitcrab configuration detected, creating network mirror configuration...")
	err := runnerutils.CreateNetworkMirrorConfig(r.config.Runner.RepositoryPath, r.config.Hermitcrab.URL)
	if err != nil {
		log.Errorf("error creating network mirror configuration: %s", err)
	}
	return err
}

// Retrieve linked resources (layer, run, repository) from the Kubernetes API.
func (r *Runner) GetResources() error {
	layer := &configv1alpha1.TerraformLayer{}
	log.Infof("getting layer %s/%s", r.config.Runner.Layer.Namespace, r.config.Runner.Layer.Name)
	err := r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: r.config.Runner.Layer.Namespace,
		Name:      r.config.Runner.Layer.Name,
	}, layer)
	if err != nil {
		return err
	}
	log.Infof("successfully retrieved layer")
	r.Layer = layer

	run := &configv1alpha1.TerraformRun{}
	log.Infof("getting run %s/%s", layer.Namespace, layer.Status.LastRun.Name)
	err = r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: layer.Namespace,
		Name:      r.config.Runner.Run,
	}, run)
	if err != nil {
		return err
	}
	log.Infof("successfully retrieved run")
	r.Run = run

	repository := &configv1alpha1.TerraformRepository{}
	log.Infof("getting repo %s/%s", layer.Spec.Repository.Namespace, layer.Spec.Repository.Name)
	err = r.Client.Get(context.TODO(), types.NamespacedName{
		Namespace: layer.Spec.Repository.Namespace,
		Name:      layer.Spec.Repository.Name,
	}, repository)
	if err != nil {
		return err
	}
	log.Infof("successfully retrieved repo")
	r.Repository = repository
	log.Infof("kubernetes resources successfully retrieved")
	return nil
}

func (r *Runner) initGitProvider() error {
	config := gitprovider.Config{
		URL:               r.Repository.Spec.Repository.Url,
		AppID:             r.config.Runner.Repository.GithubAppId,
		AppInstallationID: r.config.Runner.Repository.GithubAppInstallationId,
		AppPrivateKey:     r.config.Runner.Repository.GithubAppPrivateKey,
		GitHubToken:       r.config.Runner.Repository.GithubToken,
		GitLabToken:       r.config.Runner.Repository.GitlabToken,
		Username:          r.config.Runner.Repository.Username,
		Password:          r.config.Runner.Repository.Password,
		SSHPrivateKey:     r.config.Runner.Repository.SSHPrivateKey,
	}
	provider, err := gitprovider.New(config, []string{gt.Capabilities.Clone})
	if err != nil {
		log.Errorf("error initializing git provider: %s", err)
		return err
	}
	r.GitProvider = provider
	err = r.GitProvider.Init()
	if err != nil {
		log.Errorf("error initializing git provider: %s", err)
		return err
	}
	return nil
}

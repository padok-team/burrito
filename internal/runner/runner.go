package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/runner/tools"
	"github.com/padok-team/burrito/internal/utils"
	runnerutils "github.com/padok-team/burrito/internal/utils/runner"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Runner struct {
	config     *config.Config
	exec       tools.BaseExec
	Datastore  datastore.Client
	Client     client.Client
	Layer      *configv1alpha1.TerraformLayer
	Run        *configv1alpha1.TerraformRun
	Repository *configv1alpha1.TerraformRepository
	workingDir string
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

	err = r.cloneGitBundle()
	if err != nil {
		log.Errorf("error getting git bundle: %s", err)
		return err
	}

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

func (r *Runner) cloneGitBundle() error {
	bundle, err := r.Datastore.GetGitBundle(r.Repository.Namespace, r.Repository.Name, r.Layer.Spec.Branch, r.Run.Spec.Layer.Revision)
	if err != nil {
		log.Errorf("error fetching git bundle from datastore: %s", err)
		return err
	}

	err = os.MkdirAll(r.config.Runner.RepositoryPath, 0755)
	if err != nil {
		log.Errorf("error creating repository directory: %s", err)
	}

	bundlePath := filepath.Join(r.config.Runner.RepositoryPath, fmt.Sprintf("%s-%s.gitbundle", r.Layer.Spec.Branch, r.Run.Spec.Layer.Revision))
	err = os.WriteFile(bundlePath, bundle, 0644)
	if err != nil {
		log.Errorf("error writing git bundle to disk: %s", err)
		return err
	}

	cloneOptions := &git.CloneOptions{
		URL: bundlePath,
	}
	_, err = git.PlainClone(r.config.Runner.RepositoryPath, false, cloneOptions)
	if err != nil {
		log.Errorf("error opening git bundle: %s", err)
		return err
	}

	log.Infof("successfully fetched and opened git bundle from the datastore: repo=%s/%s ref=%s rev=%s", r.Repository.Namespace, r.Repository.Name, r.Layer.Spec.Branch, r.Run.Spec.Layer.Revision)

	return nil
}

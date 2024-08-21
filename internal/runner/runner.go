package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"github.com/padok-team/burrito/internal/utils"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/padok-team/burrito/internal/runner/tools"
	runnerutils "github.com/padok-team/burrito/internal/utils/runner"
)

const RepositoryDir string = "/runner/repository"

type Runner struct {
	config        *config.Config
	exec          tools.TerraformExec
	datastore     datastore.Client
	client        client.Client
	layer         *configv1alpha1.TerraformLayer
	run           *configv1alpha1.TerraformRun
	repository    *configv1alpha1.TerraformRepository
	gitRepository *git.Repository
	workingDir    string
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}

// Entrypoint function of the runner. Initializes the runner, executes the action and updates the layer annotations.
func (r *Runner) Exec() error {
	err := r.init()
	if err != nil {
		log.Errorf("error initializing runner: %s", err)
		return err
	}

	err = r.execInit()
	if err != nil {
		log.Errorf("error executing init: %s", err)
		return err
	}

	ann := map[string]string{}
	ref, _ := r.gitRepository.Head()
	commit := ref.Hash().String()

	switch r.config.Runner.Action {
	case "plan":
		sum, err := r.execPlan()
		if err != nil {
			return err
		}
		ann[annotations.LastPlanDate] = time.Now().Format(time.UnixDate)
		ann[annotations.LastPlanRun] = fmt.Sprintf("%s/%s", r.run.Name, strconv.Itoa(r.run.Status.Retries))
		ann[annotations.LastPlanSum] = sum
		ann[annotations.LastPlanCommit] = commit

	case "apply":
		sum, err := r.execApply()
		if err != nil {
			return err
		}
		ann[annotations.LastApplyDate] = time.Now().Format(time.UnixDate)
		ann[annotations.LastApplySum] = sum
		ann[annotations.LastApplyCommit] = commit
	default:
		return errors.New("unrecognized runner action, if this is happening there might be a version mismatch between the controller and runner")
	}

	err = annotations.Add(context.TODO(), r.client, r.layer, ann)
	if err != nil {
		log.Errorf("could not update terraform layer annotations: %s", err)
		return err
	}
	log.Infof("successfully updated terraform layer annotations")

	return nil
}

// Retrieve linked resources (layer, run, repository) from the Kubernetes API.
func (r *Runner) getResources() error {
	layer := &configv1alpha1.TerraformLayer{}
	log.Infof("getting layer %s/%s", r.config.Runner.Layer.Namespace, r.config.Runner.Layer.Name)
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: r.config.Runner.Layer.Namespace,
		Name:      r.config.Runner.Layer.Name,
	}, layer)
	if err != nil {
		return err
	}
	log.Infof("successfully retrieved layer")
	r.layer = layer
	r.workingDir = filepath.Join(RepositoryDir, r.layer.Spec.Path)

	r.run = &configv1alpha1.TerraformRun{}
	log.Infof("getting run %s/%s", layer.Namespace, layer.Status.LastRun.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: layer.Namespace,
		Name:      r.config.Runner.Run,
	}, r.run)
	if err != nil {
		return err
	}
	log.Infof("successfully retrieved run")

	repository := &configv1alpha1.TerraformRepository{}
	log.Infof("getting repo %s/%s", layer.Spec.Repository.Namespace, layer.Spec.Repository.Name)
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: layer.Spec.Repository.Namespace,
		Name:      layer.Spec.Repository.Name,
	}, repository)
	if err != nil {
		return err
	}
	log.Infof("successfully retrieved repo")
	r.repository = repository

	log.Infof("kubernetes resources successfully retrieved")
	return nil
}

// Initialize the runner's clients, retrieve linked resources (layer, run, repository),
// fetch the repository content, install the binaries and configure Hermitcrab mirror.
func (r *Runner) init() error {
	kubeClient, err := utils.NewK8SClient()
	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err)
		return err
	}
	r.client = kubeClient

	datastoreClient := datastore.NewDefaultClient()
	if r.config.Datastore.TLS {
		log.Info("using TLS for datastore")
		datastoreClient.Scheme = "https"
	}
	r.datastore = datastoreClient

	log.Infof("retrieving linked TerraformLayer and TerraformRepository")
	err = r.getResources()
	if err != nil {
		log.Errorf("error getting kubernetes resources: %s", err)
		return err
	}

	r.gitRepository, err = FetchRepositoryContent(r.repository, r.layer.Spec.Branch, r.config.Runner.Repository)
	if err != nil {
		log.Errorf("error fetching repository: %s", err)
		return err
	}
	log.Infof("repository fetched successfully")

	log.Infof("installing binaries...")
	err = os.Chdir(r.workingDir) // need to cd into the repo to detect tf versions
	if err != nil {
		log.Errorf("error changing directory: %s", err)
		return err
	}
	r.exec, err = tools.InstallBinaries(r.layer, r.repository, r.config.Runner.RunnerBinaryPath)
	if err != nil {
		log.Errorf("error installing binaries: %s", err)
		return err
	}

	if r.config.Hermitcrab.Enabled {
		log.Infof("Hermitcrab configuration detected, creating network mirror configuration...")
		err := runnerutils.CreateNetworkMirrorConfig(RepositoryDir, r.config.Hermitcrab.URL)
		if err != nil {
			log.Errorf("error creating network mirror configuration: %s", err)
		}
	}

	return nil
}

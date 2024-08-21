package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/go-git/go-git/v5"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/internal/annotations"
	"github.com/padok-team/burrito/internal/burrito/config"
	datastore "github.com/padok-team/burrito/internal/datastore/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/padok-team/burrito/internal/runner/tools"
	runnerutils "github.com/padok-team/burrito/internal/utils/runner"
)

const WorkingDir string = "/runner/repository"

type Runner struct {
	config        *config.Config
	exec          tools.TerraformExec
	datastore     datastore.Client
	client        client.Client
	layer         *configv1alpha1.TerraformLayer
	run           *configv1alpha1.TerraformRun
	repository    *configv1alpha1.TerraformRepository
	gitRepository *git.Repository
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}

func (r *Runner) Exec() error {
	client := datastore.NewDefaultClient(r.config.Datastore)
	r.datastore = client
	var commit string
	ann := map[string]string{}

	err := r.init()
	if err != nil {
		log.Errorf("error initializing runner: %s", err)
	}
	if r.gitRepository != nil {
		ref, _ := r.gitRepository.Head()
		commit = ref.Hash().String()
	}

	switch r.config.Runner.Action {
	case "plan":
		sum, err := r.plan()
		ann[annotations.LastPlanDate] = time.Now().Format(time.UnixDate)
		if err == nil {
			ann[annotations.LastPlanCommit] = commit
		}
		ann[annotations.LastPlanRun] = fmt.Sprintf("%s/%s", r.run.Name, strconv.Itoa(r.run.Status.Retries))
		ann[annotations.LastPlanSum] = sum
	case "apply":
		sum, err := r.apply()
		ann[annotations.LastApplyDate] = time.Now().Format(time.UnixDate)
		ann[annotations.LastApplySum] = sum
		if err == nil {
			ann[annotations.LastApplyCommit] = commit
		}
	default:
		err = errors.New("unrecognized runner action, if this is happening there might be a version mismatch between the controller and runner")
	}

	if err != nil {
		log.Errorf("error during runner execution: %s", err)
	}

	annotErr := annotations.Add(context.TODO(), r.client, r.layer, ann)
	if annotErr != nil {
		log.Errorf("could not update terraform layer annotations: %s", err)
	}
	log.Infof("successfully updated terraform layer annotations")

	return err
}

func newK8SClient() (client.Client, error) {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return cl, err
}

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
	return nil
}

func (r *Runner) init() error {
	cl, err := newK8SClient()
	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err)
		return err
	}
	r.client = cl

	log.Infof("retrieving linked TerraformLayer and TerraformRepository")
	err = r.getResources()
	if err != nil {
		log.Errorf("error getting kubernetes resources: %s", err)
		return err
	}
	log.Infof("kubernetes resources successfully retrieved")

	r.gitRepository, err = FetchRepositoryContent(r.repository, r.layer.Spec.Branch, r.config.Runner.Repository)
	if err != nil {
		r.gitRepository = nil // reset git repository for the caller
		log.Errorf("error fetching repository: %s", err)
		return err
	}
	log.Infof("repository fetched successfully")

	log.Infof("installing binaries...")
	err = os.Chdir(fmt.Sprintf("%s/%s", WorkingDir, r.layer.Spec.Path)) // need to cd into the repo to detect tf versions
	if err != nil {
		log.Errorf("error changing directory: %s", err)
		return err
	}
	r.exec, err = tools.InstallBinaries(r.layer, r.repository, r.config.Runner.RunnerBinaryPath)
	if err != nil {
		log.Errorf("error installing binaries: %s", err)
		return err
	}
	log.Infof("binaries successfully installed")

	if r.config.Hermitcrab.Enabled {
		log.Infof("Hermitcrab configuration detected, creating network mirror configuration...")
		err := runnerutils.CreateNetworkMirrorConfig(WorkingDir, r.config.Hermitcrab.URL)
		if err != nil {
			log.Errorf("error creating network mirror configuration: %s", err)
		}
		log.Infof("network mirror configuration created")
	}

	workingDir := fmt.Sprintf("%s/%s", WorkingDir, r.layer.Spec.Path)
	log.Infof("Launching terraform init in %s", workingDir)
	err = r.exec.Init(workingDir)
	if err != nil {
		log.Errorf("error executing terraform init: %s", err)
		return err
	}
	return nil
}

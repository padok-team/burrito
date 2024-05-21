package runner

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	b64 "encoding/base64"
	"encoding/json"

	"github.com/go-git/go-git/v5"
	tfjson "github.com/hashicorp/terraform-json"
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

	"github.com/padok-team/burrito/internal/runner/terraform"
	"github.com/padok-team/burrito/internal/runner/terragrunt"
)

const PlanArtifact string = "/tmp/plan.out"
const WorkingDir string = "/runner/repository"

type Runner struct {
	config        *config.Config
	exec          TerraformExec
	datastore     datastore.Client
	client        client.Client
	layer         *configv1alpha1.TerraformLayer
	run           *configv1alpha1.TerraformRun
	repository    *configv1alpha1.TerraformRepository
	gitRepository *git.Repository
}

type TerraformExec interface {
	Install() error
	Init(string) error
	Plan() error
	Apply(bool) error
	Show(string) ([]byte, error)
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}

func (r *Runner) Exec() error {
	client := datastore.NewDefaultClient()
	if r.config.Datastore.TLS {
		log.Info("using TLS for datastore")
		client.Scheme = "https"
	}
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

func (r *Runner) install() error {
	terraformVersion := configv1alpha1.GetTerraformVersion(r.repository, r.layer)
	terraformExec := terraform.NewTerraform(terraformVersion, PlanArtifact, r.config.Runner.RunnerBinaryPath)
	terraformRuntime := "terraform"
	if configv1alpha1.GetTerragruntEnabled(r.repository, r.layer) {
		terraformRuntime = "terragrunt"
	}
	switch terraformRuntime {
	case "terraform":
		log.Infof("using terraform")
		r.exec = terraformExec
	case "terragrunt":
		log.Infof("using terragrunt")
		r.exec = terragrunt.NewTerragrunt(terraformExec, configv1alpha1.GetTerragruntVersion(r.repository, r.layer), PlanArtifact, r.config.Runner.RunnerBinaryPath)
	}
	err := r.exec.Install()
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) init() error {
	log.Infof("retrieving linked TerraformLayer and TerraformRepository")
	cl, err := newK8SClient()
	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err)
		return err
	}
	r.client = cl
	err = r.getResources()
	if err != nil {
		log.Errorf("error getting kubernetes resources: %s", err)
		return err
	}
	log.Infof("kubernetes resources successfully retrieved")

	log.Infof("cloning repository %s %s branch", r.repository.Spec.Repository.Url, r.layer.Spec.Branch)
	r.gitRepository, err = clone(r.config.Runner.Repository, r.repository.Spec.Repository.Url, r.layer.Spec.Branch, r.layer.Spec.Path)
	if err != nil {
		r.gitRepository = nil // reset git repository for the caller
		log.Errorf("error cloning repository: %s", err)
		return err
	}
	log.Infof("repository cloned successfully")

	log.Infof("installing binaries...")
	err = r.install()
	if err != nil {
		log.Errorf("error installing binaries: %s", err)
		return err
	}
	log.Infof("binaries successfully installed")

	if os.Getenv("HERMITCRAB_ENABLED") == "true" {
		log.Infof("Hermitcrab configuration detected, creating network mirror configuration...")
		err := createNetworkMirrorConfig(os.Getenv("HERMITCRAB_URL"))
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

func (r *Runner) plan() (string, error) {
	log.Infof("starting terraform plan")
	if r.exec == nil {
		err := errors.New("terraform or terragrunt binary not installed")
		return "", err
	}
	err := r.exec.Plan()
	if err != nil {
		log.Errorf("error executing terraform plan: %s", err)
		return "", err
	}
	planJsonBytes, err := r.exec.Show("json")
	if err != nil {
		log.Errorf("error getting terraform plan json: %s", err)
		return "", err
	}
	prettyPlan, err := r.exec.Show("pretty")
	if err != nil {
		log.Errorf("error getting terraform pretty plan: %s", err)
		return "", err
	}
	log.Infof("sending plan to datastore")
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "pretty", prettyPlan)
	if err != nil {
		log.Errorf("could not put pretty plan in datastore: %s", err)
	}
	plan := &tfjson.Plan{}
	err = json.Unmarshal(planJsonBytes, plan)
	if err != nil {
		log.Errorf("error parsing terraform json plan: %s", err)
		return "", err
	}
	_, shortDiff := getDiff(plan)
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "json", planJsonBytes)
	if err != nil {
		log.Errorf("could not put json plan in datastore: %s", err)
	}
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "short", []byte(shortDiff))
	if err != nil {
		log.Errorf("could not put short plan in datastore: %s", err)
	}
	planBin, err := os.ReadFile(PlanArtifact)
	if err != nil {
		log.Errorf("could not read plan output: %s", err)
		return "", err
	}
	sum := sha256.Sum256(planBin)
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "bin", planBin)
	if err != nil {
		log.Errorf("could not put plan binary in cache: %s", err)
		return "", err
	}
	log.Infof("terraform plan ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}

func (r *Runner) apply() (string, error) {
	log.Infof("starting terraform apply")
	if r.exec == nil {
		err := errors.New("terraform or terragrunt binary not installed")
		return "", err
	}
	log.Info("getting plan binary in datastore at key")
	plan, err := r.datastore.GetPlan(r.layer.Namespace, r.layer.Name, r.run.Spec.Artifact.Run, r.run.Spec.Artifact.Attempt, "bin")
	if err != nil {
		log.Errorf("could not get plan artifact: %s", err)
		return "", err
	}
	sum := sha256.Sum256(plan)
	err = os.WriteFile(PlanArtifact, plan, 0644)
	if err != nil {
		log.Errorf("could not write plan artifact to disk: %s", err)
		return "", err
	}
	log.Print("launching terraform apply")
	if configv1alpha1.GetApplyWithoutPlanArtifactEnabled(r.repository, r.layer) {
		log.Infof("applying without reusing plan artifact from previous plan run")
		err = r.exec.Apply(false)
	} else {
		err = r.exec.Apply(true)
	}
	if err != nil {
		log.Errorf("error executing terraform apply: %s", err)
		return "", err
	}
	err = r.datastore.PutPlan(r.layer.Namespace, r.layer.Name, r.run.Name, strconv.Itoa(r.run.Status.Retries), "short", []byte("Apply Successful"))
	if err != nil {
		log.Errorf("could not put short plan in datastore: %s", err)
	}
	log.Infof("terraform apply ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}

func getDiff(plan *tfjson.Plan) (bool, string) {
	delete := 0
	create := 0
	update := 0
	for _, res := range plan.ResourceChanges {
		if res.Change.Actions.Create() {
			create++
		}
		if res.Change.Actions.Delete() {
			delete++
		}
		if res.Change.Actions.Update() {
			update++
		}
	}
	diff := false
	if create+delete+update > 0 {
		diff = true
	}
	return diff, fmt.Sprintf("Plan: %d to create, %d to update, %d to delete", create, update, delete)
}

func createNetworkMirrorConfig(endpoint string) error {
	terraformrcContent := fmt.Sprintf(`
provider_installation {
  network_mirror {
   url = "%s"
  }
}`, endpoint)
	filePath := fmt.Sprintf("%s/config.tfrc", WorkingDir)
	err := os.WriteFile(filePath, []byte(terraformrcContent), 0644)
	if err != nil {
		return err
	}
	err = os.Setenv("TF_CLI_CONFIG_FILE", filePath)
	if err != nil {
		return err
	}
	return nil
}

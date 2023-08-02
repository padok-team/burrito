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
	"github.com/padok-team/burrito/internal/lock"
	"github.com/padok-team/burrito/internal/storage"
	"github.com/padok-team/burrito/internal/storage/redis"
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
const WorkingDir string = "/repository"

type Runner struct {
	config        *config.Config
	exec          TerraformExec
	storage       storage.Storage
	client        client.Client
	layer         *configv1alpha1.TerraformLayer
	repository    *configv1alpha1.TerraformRepository
	gitRepository *git.Repository
}

type TerraformExec interface {
	Install() error
	Init(string) error
	Plan() error
	Apply() error
	Show(string) ([]byte, error)
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}

func (r *Runner) unlock() {
	err := lock.DeleteLock(context.TODO(), r.client, r.layer)
	if err != nil {
		log.Fatalf("could not remove lease lock for terraform layer %s: %s", r.layer.Name, err)
	}
	log.Infof("successfully removed lease lock for terraform layer %s", r.layer.Name)
}

func (r *Runner) updateLayerAnnotations(ann *map[string]string) {
	err := annotations.Add(context.TODO(), r.client, r.layer, *ann)
	if err != nil {
		log.Fatalf("could not update terraform layer annotations: %s", err)
	}
	log.Infof("successfully updated terraform layer annotations")
}

func (r *Runner) incrementLayerFailure(ann *map[string]string) {
	n, ok := r.layer.Annotations[annotations.Failure]
	number := 0
	if ok {
		number, _ = strconv.Atoi(n)
	}
	number++
	(*ann)[annotations.Failure] = strconv.Itoa(number)
}

func (r *Runner) Exec() error {
	var sum string
	var commit string
	ann := map[string]string{}

	err := r.init()
	if err != nil {
		log.Errorf("error initializing runner: %s", err)
		return err
	}
	defer r.updateLayerAnnotations(&ann)
	defer r.unlock()

	err = r.install()
	if err != nil {
		log.Errorf("error installing binaries: %s", err)
		return err
	}
	err = r.clone()
	if err != nil {
		log.Errorf("error cloning repository: %s", err)
		return err
	}
	err = r.tfInit()
	if err != nil {
		log.Errorf("error executing terraform init: %s", err)
		return err
	}

	ref, _ := r.gitRepository.Head()
	commit = ref.Hash().String()

	switch r.config.Runner.Action {
	case "plan":
		sum, err = r.plan()
		ann[annotations.LastPlanDate] = time.Now().Format(time.UnixDate)
		if err == nil {
			ann[annotations.LastPlanCommit] = commit
		}
		ann[annotations.LastPlanSum] = sum
	case "apply":
		sum, err = r.apply()
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
		r.incrementLayerFailure(&ann)
	} else {
		ann[annotations.Failure] = "0"
	}

	// run defers: update annotations & delete lease lock

	return err
}

func (r *Runner) getLayerAndRepository() error {
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
	log.Infof("installing binaries...")
	terraformVersion := configv1alpha1.GetTerraformVersion(r.repository, r.layer)
	terraformExec := terraform.NewTerraform(terraformVersion, PlanArtifact)
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
		r.exec = terragrunt.NewTerragrunt(terraformExec, configv1alpha1.GetTerragruntVersion(r.repository, r.layer), PlanArtifact)
	}
	err := r.exec.Install()
	if err != nil {
		return err
	}
	log.Infof("binaries successfully installed")
	return nil
}

func (r *Runner) init() error {
	log.Infof("initializing runner...")
	r.storage = redis.New(r.config.Redis)
	cl, err := newK8SClient()
	if err != nil {
		log.Errorf("error creating kubernetes client: %s", err)
		return err
	}
	r.client = cl

	log.Infof("retrieving linked TerraformLayer and TerraformRepository")
	err = r.getLayerAndRepository()
	if err != nil {
		log.Errorf("error getting kubernetes resources: %s", err)
		return err
	}
	log.Infof("kubernetes ressources successfully retrieved")

	return nil
}

func (r *Runner) clone() error {
	var err error
	log.Infof("cloning repository %s %s branch", r.repository.Spec.Repository.Url, r.layer.Spec.Branch)
	r.gitRepository, err = clone(r.config.Runner.Repository, r.repository.Spec.Repository.Url, r.layer.Spec.Branch, r.layer.Spec.Path)
	if err != nil {
		r.gitRepository = nil // reset git repository for the caller
		log.Errorf("error cloning repository: %s", err)
		return err
	}
	log.Infof("repository cloned successfully")

	return nil
}

func (r *Runner) tfInit() error {
	workingDir := fmt.Sprintf("%s/%s", WorkingDir, r.layer.Spec.Path)
	log.Infof("Launching terraform init in %s", workingDir)
	err := r.exec.Init(workingDir)
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) plan() (string, error) {
	log.Infof("starting terraform plan")
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
	prettyPlanKey := storage.GenerateKey(storage.LastPrettyPlan, r.layer)
	log.Infof("setting pretty plan into storage at key %s", prettyPlanKey)
	err = r.storage.Set(prettyPlanKey, prettyPlan, 3600)
	if err != nil {
		log.Errorf("could not put pretty plan in cache: %s", err)
	}
	plan := &tfjson.Plan{}
	err = json.Unmarshal(planJsonBytes, plan)
	if err != nil {
		log.Errorf("error parsing terraform json plan: %s", err)
		return "", err
	}
	_, shortDiff := getDiff(plan)
	planJsonKey := storage.GenerateKey(storage.LastPlannedArtifactJson, r.layer)
	log.Infof("setting plan json into storage at key %s", planJsonKey)
	err = r.storage.Set(planJsonKey, planJsonBytes, 3600)
	if err != nil {
		log.Errorf("could not put plan json in cache: %s", err)
	}
	err = r.storage.Set(storage.GenerateKey(storage.LastPlanResult, r.layer), []byte(shortDiff), 3600)
	if err != nil {
		log.Errorf("could not put short plan in cache: %s", err)
	}
	planBin, err := os.ReadFile(PlanArtifact)
	if err != nil {
		log.Errorf("could not read plan output: %s", err)
		return "", err
	}
	sum := sha256.Sum256(planBin)
	planBinKey := storage.GenerateKey(storage.LastPlannedArtifactBin, r.layer)
	log.Infof("setting plan binary into storage at key %s", planBinKey)
	err = r.storage.Set(planBinKey, planBin, 3600)
	if err != nil {
		log.Errorf("could not put plan binary in cache: %s", err)
		return "", err
	}
	log.Infof("terraform plan ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}

func (r *Runner) apply() (string, error) {
	log.Infof("starting terraform apply")
	planBinKey := storage.GenerateKey(storage.LastPlannedArtifactBin, r.layer)
	log.Infof("getting plan binary in cache at key %s", planBinKey)
	plan, err := r.storage.Get(planBinKey)
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
	err = r.exec.Apply()
	if err != nil {
		log.Errorf("error executing terraform apply: %s", err)
		return "", err
	}
	err = r.storage.Set(storage.GenerateKey(storage.LastPlanResult, r.layer), []byte(fmt.Sprintf("Apply: %s", time.Now())), 3600)
	if err != nil {
		log.Errorf("an error occurred during apply result storage: %s", err)
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

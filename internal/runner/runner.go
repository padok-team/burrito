package runner

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	b64 "encoding/base64"
	"encoding/json"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
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
)

const PlanArtifact string = "plan.out"
const WorkingDir string = "/repository"

type Runner struct {
	config     *config.Config
	terraform  *tfexec.Terraform
	storage    storage.Storage
	client     client.Client
	layer      *configv1alpha1.TerraformLayer
	repository *git.Repository
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}

func (r *Runner) Exec() {
	defer func() {
		err := lock.DeleteLock(context.TODO(), r.client, r.layer)
		if err != nil {
			log.Fatalf("could not remove lease lock for terraform layer %s: %s", r.layer.Name, err)
		}
	}()
	var sum string
	err := r.init()
	ann := map[string]string{}
	if err != nil {
		log.Errorf("error initializing runner: %s", err)
	}
	ref, _ := r.repository.Head()
	commit := ref.Hash().String()
	switch r.config.Runner.Action {
	case "plan":
		sum, err = r.plan()
		if err == nil {
			ann[annotations.LastPlanDate] = time.Now().Format(time.UnixDate)
			ann[annotations.LastPlanCommit] = commit
		}
		if sum != "" {
			ann[annotations.LastPlanSum] = sum
		}
	case "apply":
		sum, err = r.apply()
		if err == nil {
			ann[annotations.LastApplyCommit] = commit
			ann[annotations.LastApplySum] = sum
		}
	default:
		err = errors.New("Unrecognized runner action, If this is happening there might be a version mismatch between the controller and runner")
	}
	if err != nil {
		log.Errorf("error during runner execution: %s", err)
		n, ok := r.layer.Annotations[annotations.Failure]
		number := 0
		if ok {
			number, _ = strconv.Atoi(n)
		}
		number++
		ann[annotations.Failure] = strconv.Itoa(number)
	}
	err = annotations.Add(context.TODO(), r.client, *r.layer, ann)
	if err != nil {
		log.Errorf("could not update terraform layer annotations: %s", err)
	}
}

func (r *Runner) init() error {
	r.storage = redis.New(r.config.Redis.URL, r.config.Redis.Password, r.config.Redis.Database)
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	cl, err := client.New(ctrl.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}
	r.client = cl
	layer := &configv1alpha1.TerraformLayer{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: r.config.Runner.Layer.Namespace,
		Name:      r.config.Runner.Layer.Name,
	}, layer)
	if err != nil {
		return err
	}
	r.layer = layer
	log.Infof("initializing runner with Terraform version: %s", r.config.Runner.Version)
	terraformVersion, err := version.NewVersion(r.config.Runner.Version)
	if err != nil {
		return err
	}
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(terraformVersion, nil),
	}
	execPath, err := installer.Install(context.Background())
	if err != nil {
		return err
	}
	log.Infof("cloning repository %s %s branch", r.config.Runner.Repository.URL, r.config.Runner.Branch)
	cloneOptions, err := r.getCloneOptions()
	if err != nil {
		return err
	}
	r.repository, err = git.PlainClone(WorkingDir, false, cloneOptions)
	if err != nil {
		return err
	}
	workingDir := fmt.Sprintf("%s/%s", WorkingDir, r.config.Runner.Path)
	r.terraform, err = tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return err
	}
	r.terraform.SetStdout(os.Stdout)
	r.terraform.SetStderr(os.Stderr)
	log.Infof("Launching terraform init in %s", workingDir)
	err = r.terraform.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) plan() (string, error) {
	log.Infof("starting terraform plan")
	diff, err := r.terraform.Plan(context.Background(), tfexec.Out(PlanArtifact))
	if err != nil {
		log.Errorf("an error occured during terraform plan: %s", err)
		return "", err
	}
	planJson, err := r.terraform.ShowPlanFile(context.TODO(), PlanArtifact)
	if err != nil {
		log.Errorf("an error occured during terraform show: %s", err)
	}
	planJsonBytes, err := json.Marshal(planJson)
	if err != nil {
		log.Errorf("an error occured during json plan parsing: %s", err)
	}
	planJsonKey := storage.GenerateKey(storage.LastPlannedArtifactJson, r.layer)
	log.Infof("setting plan json into storage at key %s", planJsonKey)
	err = r.storage.Set(planJsonKey, planJsonBytes, 3600)
	if err != nil {
		log.Errorf("could not put plan json in cache: %s", err)
	}
	err = r.storage.Set(storage.GenerateKey(storage.LastPlanResult, r.layer), []byte(getShortPlanDiff(planJson)), 3600)
	if err != nil {
		log.Errorf("could not put short plan in cache: %s", err)
	}
	if !diff {
		log.Infof("terraform plan diff empty, no subsequent apply should be launched")
		return "", nil
	}
	plan, err := os.ReadFile(fmt.Sprintf("%s/%s", r.terraform.WorkingDir(), PlanArtifact))
	if err != nil {
		log.Errorf("could not read plan output: %s", err)
		return "", err
	}
	log.Infof("terraform plan ran successfully")
	sum := sha256.Sum256(plan)
	planBinKey := storage.GenerateKey(storage.LastPlannedArtifactBin, r.layer)
	log.Infof("setting plan binary into storage at key %s", planBinKey)
	err = r.storage.Set(planBinKey, plan, 3600)
	if err != nil {
		log.Errorf("could not put plan binary in cache: %s", err)
		return "", err
	}
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
	err = os.WriteFile(fmt.Sprintf("%s/%s", r.terraform.WorkingDir(), PlanArtifact), plan, 0644)
	if err != nil {
		log.Errorf("could not write plan artifact to disk: %s", err)
		return "", err
	}
	log.Print("launching terraform apply")
	err = r.terraform.Apply(context.Background(), tfexec.DirOrPlan(PlanArtifact))
	if err != nil {
		log.Errorf("an error occured during terraform apply: %s", err)
		return "", err
	}
	log.Infof("terraform apply ran successfully")
	return b64.StdEncoding.EncodeToString(sum[:]), nil
}

func getShortPlanDiff(plan *tfjson.Plan) string {
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
	return fmt.Sprintf("Plan: %s to create, %s to update, %s to delete", create, update, delete)
}

func (r *Runner) getCloneOptions() (*git.CloneOptions, error) {
	authMethod := "ssh"
	cloneOptions := &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(r.config.Runner.Branch),
		URL:           r.config.Runner.Repository.URL,
	}
	if strings.Contains(r.config.Runner.Repository.URL, "https://") {
		authMethod = "https"
	}
	log.Infof("clone method is %s", authMethod)
	switch authMethod {
	case "ssh":
		if r.config.Runner.Repository.SSHPrivateKey == "" {
			log.Infof("detected keyless authentication")
			return cloneOptions, nil
		}
		log.Infof("private key found")
		publicKeys, err := ssh.NewPublicKeys("git", []byte(r.config.Runner.Repository.SSHPrivateKey), "")
		if err != nil {
			return cloneOptions, err
		}
		cloneOptions.Auth = publicKeys

	case "https":
		if r.config.Runner.Repository.Username != "" && r.config.Runner.Repository.Password != "" {
			log.Infof("username and password found")
			cloneOptions.Auth = &http.BasicAuth{
				Username: r.config.Runner.Repository.Username,
				Password: r.config.Runner.Repository.Password,
			}
		} else {
			log.Infof("passwordless authentication detected")
		}
	}
	return cloneOptions, nil
}

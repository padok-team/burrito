package runner

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	b64 "encoding/base64"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/padok-team/burrito/annotations"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	"github.com/padok-team/burrito/burrito/config"
	"github.com/padok-team/burrito/cache"
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
	config      *config.Config
	terraform   *tfexec.Terraform
	cache       cache.Cache
	client      client.Client
	layer       *configv1alpha1.TerraformLayer
	annotations map[string]string
	repository  *git.Repository
}

func New(c *config.Config) *Runner {
	return &Runner{
		config:      c,
		annotations: map[string]string{},
	}
}

func (r *Runner) Exec() {
	err := r.init()
	if err != nil {
		log.Fatalf("error initializing runner: %s", err)
	}
	delete(r.layer.Annotations, annotations.Lock)
	ref, _ := r.repository.Head()
	commit := ref.Hash().String()
	switch r.config.Runner.Action {
	case "plan":
		err = r.plan()
		r.annotations[annotations.LastPlanCommit] = commit
	case "apply":
		err = r.apply()
		r.annotations[annotations.LastApplyCommit] = commit
	default:
		err = errors.New("Unrecognized runner action, If this is happening there might be a version mismatch between the controller and runner")
	}
	r.annotations[annotations.Failure] = "0"
	if err != nil {
		log.Fatalf("Error during runner execution: %s", err)
		n, ok := r.layer.Annotations[annotations.Failure]
		number := 0
		if ok {
			number, err = strconv.Atoi(n)
			if err != nil {
				number = 0
			}
		}
		number++
		r.annotations[annotations.Failure] = strconv.Itoa(number)
	}
	for k, v := range r.annotations {
		r.layer.Annotations[k] = v
	}
	err = r.client.Update(context.TODO(), r.layer)
	if err != nil {
		log.Fatalf("Could not update annotations on Layer: %s", err)
	}
}

func (r *Runner) init() error {
	r.cache = cache.NewRedisCache(r.config.Redis.URL, r.config.Redis.Password, r.config.Redis.Database)
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
	log.Printf("Using Terraform version: %s", r.config.Runner.Version)
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
	log.Printf("Cloning repository %s %s branch", r.config.Runner.Repository.URL, r.config.Runner.Branch)
	r.repository, err = git.PlainClone(WorkingDir, false, &git.CloneOptions{
		ReferenceName: plumbing.NewBranchReferenceName(r.config.Runner.Branch),
		URL:           r.config.Runner.Repository.URL,
	})
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
	log.Printf("Launching terraform init in %s", workingDir)
	err = r.terraform.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) plan() error {
	log.Print("Launching terraform plan")
	diff, err := r.terraform.Plan(context.Background(), tfexec.Out(PlanArtifact))
	if err != nil {
		log.Printf("Terraform plan errored: %s", err)
		return err
	}
	planDateKey := cache.GenerateKey(cache.LastPlanDate, r.layer)
	log.Printf("Setting last plan date cache at key %s", planDateKey)
	err = r.cache.Set(planDateKey, []byte(strconv.FormatInt(time.Now().Unix(), 10)), 3600)
	r.annotations[annotations.LastPlanDate] = strconv.FormatInt(time.Now().Unix(), 10)
	if err != nil {
		log.Fatalf("Could not put plan date in cache: %s", err)
		return err
	}
	if !diff {
		log.Printf("Terraform plan diff empty, no subsequent apply should be launched")
		return err
	}
	plan, err := os.ReadFile(fmt.Sprintf("%s/%s", r.terraform.WorkingDir(), PlanArtifact))
	if err != nil {
		log.Fatalf("Could not read plan output: %s", err)
		return err
	}
	log.Print("Terraform plan ran successfully")
	sum := sha256.Sum256(plan)
	planBinKey := cache.GenerateKey(cache.LastPlannedArtifactBin, r.layer)
	log.Printf("Setting plan binary into cache at key %s", planBinKey)
	err = r.cache.Set(planBinKey, plan, 3600)
	if err != nil {
		log.Fatalf("Could not put plan binary in cache: %s", err)
	}
	planSumKey := cache.GenerateKey(cache.LastPlannedArtifact, r.layer)
	log.Printf("Setting plan binary checksum into cache at key %s", planSumKey)
	err = r.cache.Set(planSumKey, []byte(b64.StdEncoding.EncodeToString(sum[:])), 3600)
	r.annotations[annotations.LastPlanSum] = b64.StdEncoding.EncodeToString(sum[:])
	if err != nil {
		log.Fatalf("Could not put plan checksum in cache: %s", err)
	}
	return nil
}

func (r *Runner) apply() error {
	planBinKey := cache.GenerateKey(cache.LastPlannedArtifactBin, r.layer)
	log.Printf("Getting plan binary in cache at key %s", planBinKey)
	plan, err := r.cache.Get(planBinKey)
	if err != nil {
		log.Printf("Could not get plan artifact: %s", err)
		return err
	}
	sum := sha256.Sum256(plan)
	err = os.WriteFile(fmt.Sprintf("%s/%s", r.terraform.WorkingDir(), PlanArtifact), plan, 0644)
	if err != nil {
		log.Printf("Could not write plan artifact to disk: %s", err)
		return err
	}
	log.Print("Launching terraform apply")
	err = r.terraform.Apply(context.Background(), tfexec.DirOrPlan(PlanArtifact))
	if err != nil {
		log.Fatalf("Terraform apply errored: %s", err)
		return err
	}
	log.Print("Terraform apply ran successfully")
	applySumKey := cache.GenerateKey(cache.LastAppliedArtifact, r.layer)
	log.Printf("Setting plan binary checksum into cache at key %s", applySumKey)
	err = r.cache.Set(applySumKey, []byte(b64.StdEncoding.EncodeToString(sum[:])), 3600)
	r.annotations[annotations.LastApplySum] = b64.StdEncoding.EncodeToString(sum[:])
	if err != nil {
		log.Fatalf("Could not put apply checksum in cache: %s", err)
		return err
	}
	return nil
}

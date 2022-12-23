package runner

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/padok-team/burrito/burrito/config"
	"github.com/padok-team/burrito/cache"
)

const PlanArtifact string = "plan.out"
const WorkingDir string = "/repository"

type Runner struct {
	config    *config.Config
	terraform *tfexec.Terraform
	cache     cache.Cache
}

func New(c *config.Config) *Runner {
	return &Runner{
		config: c,
	}
}

func (r *Runner) Exec() {
	r.cache = cache.NewRedisCache(r.config.Redis.URL, r.config.Redis.Password, r.config.Redis.Database)
	defer r.cache.Delete(r.config.Runner.Layer.Lock)
	err := r.init()
	if err != nil {
		log.Fatalf("error initializing runner: %s", err)
	}
	switch r.config.Runner.Action {
	case "plan":
		r.plan()
	case "apply":
		r.apply()
	default:
		log.Fatalf("Unrecognized runner Action")
	}
}

func (r *Runner) init() error {
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
	_, err = git.PlainClone(WorkingDir, false, &git.CloneOptions{
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

func (r *Runner) plan() {
	log.Print("Launching terraform plan")
	diff, err := r.terraform.Plan(context.Background(), tfexec.Out(PlanArtifact))
	if err != nil {
		log.Printf("Terraform plan errored: %s", err)
		return
	}
	log.Printf("Setting last plan date cache at key %s", r.config.Runner.Layer.PlanDate)
	err = r.cache.Set(r.config.Runner.Layer.PlanDate, []byte(strconv.FormatInt(time.Now().Unix(), 10)), 3600)
	if err != nil {
		log.Fatalf("Could not put plan date in cache: %s", err)
	}
	if !diff {
		log.Printf("Terraform plan diff empty, no subsequent apply should be launched")
		return
	}
	plan, err := os.ReadFile(fmt.Sprintf("%s/%s", r.terraform.WorkingDir(), PlanArtifact))
	if err != nil {
		log.Fatalf("Could not read plan output: %s", err)
		return
	}
	log.Print("Terraform plan ran successfully")
	sum := sha256.Sum256(plan)
	log.Printf("Setting plan binary into cache at key %s", r.config.Runner.Layer.PlanBin)
	err = r.cache.Set(r.config.Runner.Layer.PlanBin, plan, 3600)
	if err != nil {
		log.Fatalf("Could not put plan binary in cache: %s", err)
	}
	log.Printf("Setting plan binary checksum into cache at key %s", r.config.Runner.Layer.PlanSum)
	err = r.cache.Set(r.config.Runner.Layer.PlanSum, sum[:], 3600)
	if err != nil {
		log.Fatalf("Could not put plan checksum in cache: %s", err)
	}
}

func (r *Runner) apply() {
	log.Printf("Getting plan binary in cache at key %s", r.config.Runner.Layer.PlanBin)
	plan, err := r.cache.Get(r.config.Runner.Layer.PlanBin)
	if err != nil {
		log.Printf("Could not get plan artifact: %s", err)
		return
	}
	sum := sha256.Sum256(plan)
	err = os.WriteFile(fmt.Sprintf("%s/%s", r.terraform.WorkingDir(), PlanArtifact), plan, 0644)
	if err != nil {
		log.Printf("Could not write plan artifact to disk: %s", err)
		return
	}
	log.Print("Launching terraform apply")
	err = r.terraform.Apply(context.Background(), tfexec.DirOrPlan(PlanArtifact))
	if err != nil {
		log.Fatalf("Terraform apply errored: %s", err)
		return
	}
	log.Print("Terraform apply ran successfully")
	log.Printf("Setting plan binary checksum into cache at key %s", r.config.Runner.Layer.ApplySum)
	err = r.cache.Set(r.config.Runner.Layer.ApplySum, sum[:], 3600)
	if err != nil {
		log.Fatalf("Could not put apply checksum in cache: %s", err)
	}
}

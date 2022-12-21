package runner

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"

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
const WorkingDir string = "/burrito"

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
	r.cache = cache.NewRedisCache(r.config.Redis.URL, r.config.Redis.Password, r.config.Redis.Database)

	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(r.config.Runner.Version)),
	}
	execPath, err := installer.Install(context.Background())
	if err != nil {
		return err
	}
	//TODO: Implement authentication here
	_, err = git.PlainClone(WorkingDir, false, &git.CloneOptions{
		ReferenceName: plumbing.ReferenceName(r.config.Runner.Branch),
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
	err = r.terraform.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) plan() {
	defer r.cache.Delete(r.config.Runner.Layer.Lock)
	_, err := r.terraform.Plan(context.Background(), tfexec.Out(PlanArtifact))
	if err != nil {
		log.Printf("Terraform plan errored: %s", err)
		return
	}
	plan, err := os.ReadFile(fmt.Sprintf("%s/%s", r.terraform.WorkingDir(), PlanArtifact))
	if err != nil {
		log.Fatalf("Could not read plan output: %s", err)
		return
	}
	sum := sha256.Sum256(plan)
	err = r.cache.Set(r.config.Runner.Layer.PlanBin, plan, 3600)
	if err != nil {
		log.Fatalf("Could not put plan binary in cache: %s", err)
	}
	err = r.cache.Set(r.config.Runner.Layer.PlanSum, sum[:], 3600)
	if err != nil {
		log.Fatalf("Could not put plan checksum in cache: %s", err)
	}
	log.Print("Terraform plan ran successfully")
}

func (r *Runner) apply() {
	defer r.cache.Delete(r.config.Runner.Layer.Lock)
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
	err = r.terraform.Apply(context.Background(), tfexec.DirOrPlan(PlanArtifact))
	if err != nil {
		log.Fatalf("Terraform apply errored: %s", err)
		return
	}
	err = r.cache.Set(r.config.Runner.Layer.ApplySum, sum[:], 3600)
	if err != nil {
		log.Fatalf("Could not put apply checksum in cache: %s", err)
	}
	log.Print("Terraform apply ran successfully")
}

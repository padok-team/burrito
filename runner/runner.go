package runner

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
	cache "github.com/padok-team/burrito/cache"
)

const PlanArtifact string = "plan.out"
const WorkingDir string = "/burrito"

type Runner struct {
	Config    *Config
	Out       io.Writer
	Err       io.Writer
	Terraform *tfexec.Terraform
	Cache     cache.Cache
}

type Credentials struct {
	SSH      string `yaml:"ssh,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type Config struct {
	Path        string      `yaml:"path"`
	Branch      string      `yaml:"branch"`
	Version     string      `yaml:"version"`
	Action      string      `yaml:"action"`
	Repository  string      `yaml:"repository"`
	Credentials Credentials `yaml:"credentials"`
	Redis       Redis       `yaml:"redis"`
	Layer       Layer       `yaml:"layer"`
}

type Layer struct {
	Lock     string `yaml:"lock,omitempty"`
	PlanSum  string `yaml:"planSum,omitempty"`
	PlanBin  string `yaml:"planBin,omitempty"`
	ApplySum string `yaml:"applySum,omitempty"`
}

type Redis struct {
	URI string `yaml:"uri"`
}

func New() (*Runner, error) {
	runner := &Runner{
		Config: &Config{},
		Out:    os.Stdout,
		Err:    os.Stderr,
	}
	return runner, nil
}

func (r *Runner) Exec() {
	err := r.init()
	if err != nil {
		log.Fatalf("error initializing runner: %s", err)
	}
	switch r.Config.Action {
	case "plan":
		r.plan()
	case "apply":
		r.apply()
	default:
		log.Fatalf("Unrecognized runner Action")
	}
}

func (r *Runner) init() error {
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(version.NewVersion(r.Config.Version)),
	}
	execPath, err := installer.Install(context.Background())
	if err != nil {
		return err
	}
	//TODO: Implement authentication here
	_, err = git.PlainClone(WorkingDir, false, &git.CloneOptions{
		ReferenceName: plumbing.ReferenceName(r.Config.Branch),
		URL:           r.Config.Repository,
	})
	if err != nil {
		return err
	}
	workingDir := fmt.Sprintf("%s/%s", WorkingDir, r.Config.Path)
	r.Terraform, err = tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return err
	}
	err = r.Terraform.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) plan() {
	_, err := r.Terraform.Plan(context.Background(), tfexec.Out(PlanArtifact))
	cacheErr := r.Cache.Delete(r.Config.Layer.Lock)
	if cacheErr != nil {
		log.Fatalf("Could not delete lock: %s", cacheErr)
		return
	}
	if err != nil {
		log.Printf("Terraform plan errored: %s", err)
		return
	}
	plan, err := os.ReadFile(fmt.Sprintf("%s/%s", r.Terraform.WorkingDir(), PlanArtifact))
	if err != nil {
		log.Fatalf("Could not read plan output: %s", err)
		return
	}
	sum := sha256.Sum256(plan)
	err = r.Cache.Set(r.Config.Layer.PlanBin, plan, 3600)
	if err != nil {
		log.Fatalf("Could not put plan binary in cache: %s", err)
	}
	err = r.Cache.Set(r.Config.Layer.PlanSum, sum[:], 3600)
	if err != nil {
		log.Fatalf("Could not put plan checksum in cache: %s", err)
	}
}

func (r *Runner) apply() {
	plan, err := r.Cache.Get(r.Config.Layer.PlanBin)
	if err != nil {
		log.Printf("Could not get plan artifact: %s", err)
	}
	sum := sha256.Sum256(plan)
	err = os.WriteFile(fmt.Sprintf("%s/%s", r.Terraform.WorkingDir(), PlanArtifact), plan, 0644)
	if err != nil {
		log.Printf("Could not write plan artifact to disk: %s", err)
	}
	err = r.Terraform.Apply(context.Background(), tfexec.DirOrPlan(PlanArtifact))
	cacheErr := r.Cache.Delete(r.Config.Layer.Lock)
	if cacheErr != nil {
		log.Fatalf("Could not delete lock: %s", cacheErr)
		return
	}
	if err != nil {
		log.Fatalf("Terraform apply errored: %s", err)
		return
	}
	err = r.Cache.Set(r.Config.Layer.ApplySum, sum[:], 3600)
	if err != nil {
		log.Fatalf("Could not put apply checksum in cache: %s", err)
	}
}

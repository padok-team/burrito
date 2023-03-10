package runner

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type Terraform struct {
	exec    *tfexec.Terraform
	version string
	path    string
}

func NewTerraform(version string) *Terraform {
	return &Terraform{
		version: version,
	}
}

func (t *Terraform) Install() error {
	terraformVersion, err := version.NewVersion(t.version)
	installer := &releases.ExactVersion{
		Product: product.Terraform,
		Version: version.Must(terraformVersion, nil),
	}
	execPath, err := installer.Install(context.Background())
	if err != nil {
		return err
	}
	t.path = execPath
	return nil
}

func (t *Terraform) Init(workingDir string) error {
	exec, err := tfexec.NewTerraform(workingDir, t.path)
	if err != nil {
		return err
	}
	t.exec = exec
	err = t.exec.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}
	t.exec.SetStdout(os.Stdout)
	t.exec.SetStderr(os.Stderr)
	return nil
}

func (t *Terraform) Plan() error {
	_, err := t.exec.Plan(context.Background(), tfexec.Out(PlanArtifact))
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Apply() error {
	err := t.exec.Apply(context.Background(), tfexec.DirOrPlan(PlanArtifact))
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Show() ([]byte, error) {
	t.exec.SetStdout(io.Discard)
	t.exec.SetStderr(io.Discard)
	planJson, err := t.exec.ShowPlanFile(context.TODO(), PlanArtifact)
	if err != nil {
		return nil, err
	}
	planJsonBytes, err := json.Marshal(planJson)
	if err != nil {
		return nil, err
	}
	t.exec.SetStdout(os.Stdout)
	t.exec.SetStderr(os.Stderr)
	return planJsonBytes, nil
}

package terraform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/terraform-exec/tfexec"
)

type Terraform struct {
	exec             *tfexec.Terraform
	version          string
	ExecPath         string
	planArtifactPath string
}

func NewTerraform(version, planArtifactPath string) *Terraform {
	return &Terraform{
		version:          version,
		planArtifactPath: planArtifactPath,
	}
}

func (t *Terraform) Install() error {
	terraformVersion, err := version.NewVersion(t.version)
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
	t.ExecPath = execPath
	return nil
}

func (t *Terraform) Init(workingDir string) error {
	exec, err := tfexec.NewTerraform(workingDir, t.ExecPath)
	if err != nil {
		return err
	}
	t.exec = exec
	err = t.exec.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Plan(lock bool) error {
	t.verbose()
	options := []tfexec.PlanOption{}
	options = append(options, tfexec.Out(t.planArtifactPath))
	if !lock {
		options = append(options, tfexec.Lock(false))
	}
	_, err := t.exec.Plan(context.Background(), options...)
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Apply() error {
	t.verbose()
	err := t.exec.Apply(context.Background(), tfexec.DirOrPlan(t.planArtifactPath))
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Show(mode string) ([]byte, error) {
	t.silent()
	switch mode {
	case "json":
		planJson, err := t.exec.ShowPlanFile(context.TODO(), t.planArtifactPath)
		if err != nil {
			return nil, err
		}
		planJsonBytes, err := json.Marshal(planJson)
		if err != nil {
			return nil, err
		}
		return planJsonBytes, nil
	case "pretty":
		plan, err := t.exec.ShowPlanFileRaw(context.TODO(), t.planArtifactPath)
		if err != nil {
			return nil, err
		}
		return []byte(plan), nil
	default:
		return nil, errors.New("invalid mode")
	}
}

func (t *Terraform) silent() {
	t.exec.SetStdout(io.Discard)
	t.exec.SetStderr(io.Discard)
}

func (t *Terraform) verbose() {
	t.exec.SetStdout(os.Stdout)
	t.exec.SetStderr(os.Stderr)
}

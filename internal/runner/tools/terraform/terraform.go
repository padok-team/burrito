package terraform

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/hashicorp/terraform-exec/tfexec"
)

type Terraform struct {
	exec     *tfexec.Terraform
	ExecPath string
}

func (t *Terraform) TenvName() string {
	return "terraform"
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

func (t *Terraform) Plan(planArtifactPath string) error {
	t.verbose()
	_, err := t.exec.Plan(context.Background(), tfexec.Out(planArtifactPath))
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Apply(planArtifactPath string) error {
	t.verbose()
	applyOpts := []tfexec.ApplyOption{}
	if planArtifactPath != "" {
		applyOpts = append(applyOpts, tfexec.DirOrPlan(planArtifactPath))
	}

	err := t.exec.Apply(context.Background(), applyOpts...)
	if err != nil {
		return err
	}
	return nil
}

func (t *Terraform) Show(planArtifactPath, mode string) ([]byte, error) {
	t.silent()
	switch mode {
	case "json":
		planJson, err := t.exec.ShowPlanFile(context.TODO(), planArtifactPath)
		if err != nil {
			return nil, err
		}
		planJsonBytes, err := json.Marshal(planJson)
		if err != nil {
			return nil, err
		}
		return planJsonBytes, nil
	case "pretty":
		plan, err := t.exec.ShowPlanFileRaw(context.TODO(), planArtifactPath)
		if err != nil {
			return nil, err
		}
		return []byte(plan), nil
	default:
		return nil, errors.New("invalid mode")
	}
}

func (t *Terraform) GetExecPath() string {
	return t.ExecPath
}

func (t *Terraform) silent() {
	t.exec.SetStdout(io.Discard)
	t.exec.SetStderr(io.Discard)
}

func (t *Terraform) verbose() {
	t.exec.SetStdout(os.Stdout)
	t.exec.SetStderr(os.Stderr)
}

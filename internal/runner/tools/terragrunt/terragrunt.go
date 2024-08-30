package terragrunt

import (
	"errors"
	"os/exec"

	"github.com/padok-team/burrito/internal/runner/tools/terraform"
	"github.com/padok-team/burrito/internal/runner/tools/tofu"
	c "github.com/padok-team/burrito/internal/utils/cmd"
)

type Terragrunt struct {
	ExecPath   string
	WorkingDir string
	Terraform  *terraform.Terraform
	Tofu       *tofu.Tofu
}

func (t *Terragrunt) getDefaultOptions(command string) []string {
	return []string{
		command,
		"--terragrunt-tfpath",
		t.Terraform.ExecPath,
		"--terragrunt-working-dir",
		t.WorkingDir,
		"-no-color",
	}
}

func (t *Terragrunt) Init(workingDir string) error {
	t.WorkingDir = workingDir
	cmd := exec.Command(t.ExecPath, t.getDefaultOptions("init")...)
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Plan(planArtifactPath string) error {
	options := append(t.getDefaultOptions("plan"), "-out", planArtifactPath)
	cmd := exec.Command(t.ExecPath, options...)
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Apply(planArtifactPath string) error {
	options := append(t.getDefaultOptions("apply"), "-auto-approve")
	if planArtifactPath != "" {
		options = append(options, planArtifactPath)
	}

	cmd := exec.Command(t.ExecPath, options...)
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Show(planArtifactPath, mode string) ([]byte, error) {
	options := t.getDefaultOptions("show")
	switch mode {
	case "json":
		options = append(options, "-json", planArtifactPath)
	case "pretty":
		options = append(options, planArtifactPath)
	default:
		return nil, errors.New("invalid mode")
	}
	cmd := exec.Command(t.ExecPath, options...)
	cmd.Dir = t.WorkingDir
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	return output, nil
}

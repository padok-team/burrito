package terragrunt

import (
	"errors"
	"os/exec"

	"github.com/padok-team/burrito/internal/runner/tools/exec/base"
	c "github.com/padok-team/burrito/internal/utils/cmd"
)

type Terragrunt struct {
	ExecPath   string
	WorkingDir string
	BaseExec   base.BaseExec
}

func (t *Terragrunt) TenvName() string {
	return "terragrunt"
}

func (t *Terragrunt) getDefaultOptions(command string) ([]string, error) {
	return []string{
		command,
		"--terragrunt-tfpath",
		t.BaseExec.GetExecPath(),
		"--terragrunt-working-dir",
		t.WorkingDir,
		"-no-color",
	}, nil
}

func (t *Terragrunt) Init(workingDir string) error {
	t.WorkingDir = workingDir
	options, err := t.getDefaultOptions("init")
	if err != nil {
		return err
	}
	cmd := exec.Command(t.ExecPath, options...)
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Plan(planArtifactPath string) error {
	options, err := t.getDefaultOptions("plan")
	if err != nil {
		return err
	}
	options = append(options, "-out", planArtifactPath)
	cmd := exec.Command(t.ExecPath, options...)
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Apply(planArtifactPath string) error {
	options, err := t.getDefaultOptions("apply")
	if err != nil {
		return err
	}
	options = append(options, "-auto-approve")
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
	options, err := t.getDefaultOptions("show")
	if err != nil {
		return nil, err
	}
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

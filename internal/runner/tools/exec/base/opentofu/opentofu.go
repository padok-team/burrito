package opentofu

import (
	"errors"
	"os/exec"

	c "github.com/padok-team/burrito/internal/utils/cmd"
)

// The equivalent of tfexec for OpenTofu is not actively maintained.
// Switch to it when this repo is updated: https://github.com/tofu/tofu-exec

type OpenTofu struct {
	ExecPath   string
	WorkingDir string
}

func (t *OpenTofu) TenvName() string {
	return "tofu"
}

func (t *OpenTofu) Init(workingDir string) error {
	t.WorkingDir = workingDir
	cmd := exec.Command(t.ExecPath, "init", "-upgrade")
	c.Verbose(cmd)
	cmd.Dir = workingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *OpenTofu) Plan(planArtifactPath string) error {
	cmd := exec.Command(t.ExecPath, "plan", "-out", planArtifactPath)
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *OpenTofu) Apply(planArtifactPath string) error {
	var cmd *exec.Cmd
	if planArtifactPath != "" {
		cmd = exec.Command(t.ExecPath, "apply", "-auto-approve", planArtifactPath)
	} else {
		cmd = exec.Command(t.ExecPath, "apply", "-auto-approve")
	}
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *OpenTofu) Show(planArtifactPath, mode string) ([]byte, error) {
	var cmd *exec.Cmd
	switch mode {
	case "json":
		cmd = exec.Command(t.ExecPath, "show", "-json", planArtifactPath)
	case "pretty":
		cmd = exec.Command(t.ExecPath, "show", planArtifactPath)
	default:
		return nil, errors.New("invalid mode")
	}

	cmd.Dir = t.WorkingDir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (t *OpenTofu) GetExecPath() string {
	return t.ExecPath
}

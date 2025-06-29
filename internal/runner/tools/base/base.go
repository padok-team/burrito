package base

import (
	"errors"
	"os/exec"

	c "github.com/padok-team/burrito/internal/utils/cmd"
)

// BaseTool provides common functionality for Terraform and OpenTofu
type BaseTool struct {
	ExecPath   string
	WorkingDir string
	ToolName   string // "terraform" or "tofu"
}

func (t *BaseTool) TenvName() string {
	return t.ToolName
}

func (t *BaseTool) Init(workingDir string) error {
	t.WorkingDir = workingDir
	cmd := exec.Command(t.ExecPath, "init", "-no-color", "-upgrade")
	c.Verbose(cmd)
	cmd.Dir = workingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *BaseTool) Plan(planArtifactPath string) error {
	cmd := exec.Command(t.ExecPath, "plan", "-no-color", "-out", planArtifactPath)
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *BaseTool) Apply(planArtifactPath string) error {
	var cmd *exec.Cmd
	if planArtifactPath != "" {
		cmd = exec.Command(t.ExecPath, "apply", "-no-color", "-auto-approve", planArtifactPath)
	} else {
		cmd = exec.Command(t.ExecPath, "apply", "-no-color", "-auto-approve")
	}
	c.Verbose(cmd)
	cmd.Dir = t.WorkingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *BaseTool) Show(planArtifactPath, mode string) ([]byte, error) {
	var cmd *exec.Cmd
	switch mode {
	case "json":
		cmd = exec.Command(t.ExecPath, "show", "-no-color", "-json", planArtifactPath)
	case "pretty":
		cmd = exec.Command(t.ExecPath, "show", "-no-color", planArtifactPath)
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

func (t *BaseTool) GetExecPath() string {
	return t.ExecPath
}

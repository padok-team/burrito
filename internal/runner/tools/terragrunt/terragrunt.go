package terragrunt

import (
	"errors"
	"os/exec"

	"github.com/blang/semver/v4"
	c "github.com/padok-team/burrito/internal/utils/cmd"
)

type Terragrunt struct {
	ExecPath      string
	WorkingDir    string
	ChildExecPath string
	Version       string
}

func (t *Terragrunt) TenvName() string {
	return "terragrunt"
}

func (t *Terragrunt) getDefaultOptions(command string) ([]string, error) {
	// Parse the version to determine which flags to use
	// Terragrunt 0.73.0 introduced new shortened flags:
	// - --terragrunt-tfpath -> --tf-path
	// - --terragrunt-working-dir -> --working-dir
	version, err := semver.Parse(t.Version)
	if err != nil {
		// If version parsing fails, use legacy flags as fallback
		return []string{
			command,
			"--terragrunt-tfpath",
			t.ChildExecPath,
			"--terragrunt-working-dir",
			t.WorkingDir,
			"-no-color",
		}, nil
	}

	newFlagsVersion := semver.MustParse("0.73.0")

	if version.GTE(newFlagsVersion) {
		// Use new flags for version 0.73.0 and above
		return []string{
			command,
			"--tf-path",
			t.ChildExecPath,
			"--working-dir",
			t.WorkingDir,
			"-no-color",
		}, nil
	} else {
		// Use legacy flags for versions below 0.73.0
		return []string{
			command,
			"--terragrunt-tfpath",
			t.ChildExecPath,
			"--terragrunt-working-dir",
			t.WorkingDir,
			"-no-color",
		}, nil
	}
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

func (t *Terragrunt) GetExecPath() string {
	return t.ExecPath
}

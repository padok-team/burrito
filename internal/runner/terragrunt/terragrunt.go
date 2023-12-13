package terragrunt

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/padok-team/burrito/internal/runner/terraform"
)

const (
	BinWorkDir = "/runner/bin"
)

type Terragrunt struct {
	execPath         string
	planArtifactPath string
	version          string
	workingDir       string
	terraform        *terraform.Terraform
}

func NewTerragrunt(terraformExec *terraform.Terraform, terragruntVersion, planArtifactPath string) *Terragrunt {
	return &Terragrunt{
		version:          terragruntVersion,
		terraform:        terraformExec,
		planArtifactPath: planArtifactPath,
	}
}

func verbose(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
}

func (t *Terragrunt) Install() error {
	err := t.terraform.Install()
	if err != nil {
		return err
	}
	path, err := downloadTerragrunt(t.version)
	if err != nil {
		return err
	}
	t.execPath = path
	return nil
}

func (t *Terragrunt) getDefaultOptions(command string) []string {
	return []string{
		command,
		"--terragrunt-tfpath",
		t.terraform.ExecPath,
		"--terragrunt-working-dir",
		t.workingDir,
		"-no-color",
	}
}

func (t *Terragrunt) Init(workingDir string) error {
	t.workingDir = workingDir
	cmd := exec.Command(t.execPath, t.getDefaultOptions("init")...)
	verbose(cmd)
	cmd.Dir = t.workingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Plan() error {
	options := append(t.getDefaultOptions("plan"), "-out", t.planArtifactPath)
	cmd := exec.Command(t.execPath, options...)
	verbose(cmd)
	cmd.Dir = t.workingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Apply() error {
	options := append(t.getDefaultOptions("apply"), t.planArtifactPath)
	cmd := exec.Command(t.execPath, options...)
	verbose(cmd)
	cmd.Dir = t.workingDir
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (t *Terragrunt) Show(mode string) ([]byte, error) {
	options := t.getDefaultOptions("show")
	switch mode {
	case "json":
		options = append(options, "-json", t.planArtifactPath)
	case "pretty":
		options = append(options, t.planArtifactPath)
	default:
		return nil, errors.New("invalid mode")
	}
	cmd := exec.Command(t.execPath, options...)
	cmd.Dir = t.workingDir
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	return output, nil
}

func downloadTerragrunt(version string) (string, error) {
	cpuArch := runtime.GOARCH

	url := fmt.Sprintf("https://github.com/gruntwork-io/terragrunt/releases/download/v%s/terragrunt_linux_%s", version, cpuArch)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	filename := fmt.Sprintf("%s/terragrunt_%s", BinWorkDir, cpuArch)
	file, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	err = os.Chmod(filename, 0755)
	if err != nil {
		return "", err
	}

	filepath, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	return filepath, nil
}

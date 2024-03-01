package terragrunt

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/padok-team/burrito/internal/runner/terraform"
	log "github.com/sirupsen/logrus"
)

type Terragrunt struct {
	execPath         string
	planArtifactPath string
	version          string
	workingDir       string
	terraform        *terraform.Terraform
	runnerBinaryPath string
}

func NewTerragrunt(terraformExec *terraform.Terraform, terragruntVersion, planArtifactPath string, runnerBinaryPath string) *Terragrunt {
	return &Terragrunt{
		version:          terragruntVersion,
		terraform:        terraformExec,
		planArtifactPath: planArtifactPath,
		runnerBinaryPath: runnerBinaryPath,
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

	path, err := ensureTerragrunt(t.version, t.runnerBinaryPath)
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

func ensureTerragrunt(version string, runnerBinaryPath string) (string, error) {
	runnerBinary := filepath.Join(runnerBinaryPath, "terragrunt")
	info, err := os.Stat(runnerBinary)
	if !os.IsNotExist(err) && !info.IsDir() {
		hash, err := calculateFileSHA256(runnerBinary)
		if err != nil {
			return "", err
		}

		trustedHash, err := getTerragruntSHA256(version)
		if err != nil {
			return "", err
		}

		if hash == trustedHash {
			err = os.Chmod(runnerBinaryPath, 0755)
			if err != nil {
				return "", err
			}
			log.Infof("Terragrunt binary found at %s, using it", runnerBinaryPath)
			return filepath.Abs(runnerBinaryPath)
		}
	}

	log.Infof("Terragrunt binary not found, downloading it. (Consider packaging binaries within your runner image to mitigate eventual network expenses)")
	path, err := downloadTerragrunt(version, runnerBinaryPath)
	if err != nil {
		return "", err
	}

	return path, nil
}

func calculateFileSHA256(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()

	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func getTerragruntSHA256(version string) (string, error) {
	cpuArch := runtime.GOARCH
	response, err := http.Get(fmt.Sprintf("https://github.com/gruntwork-io/terragrunt/releases/download/v%s/SHA256SUMS", version))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		sha := parts[0]
		filename := parts[1]

		if strings.Contains(filename, fmt.Sprintf("linux_%s", cpuArch)) {
			return sha, nil
		}
	}

	return "", errors.New("could not find a hash for this architecture in SHA256SUMS file")
}

func downloadTerragrunt(version string, runnerBinaryPath string) (string, error) {
	cpuArch := runtime.GOARCH

	url := fmt.Sprintf("https://github.com/gruntwork-io/terragrunt/releases/download/v%s/terragrunt_linux_%s", version, cpuArch)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	filename := fmt.Sprintf("%s/terragrunt_%s", runnerBinaryPath, cpuArch)
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

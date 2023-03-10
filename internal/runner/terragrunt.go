package runner

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

type Terragrunt struct {
	path       string
	workingDir string
	version    string
	terraform  *Terraform
}

func NewTerragrunt(terragruntVersion string, terraformVersion string, workingDir string) *Terragrunt {
	return &Terragrunt{
		version:    terragruntVersion,
		terraform:  NewTerraform(terraformVersion),
		workingDir: workingDir,
	}
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
	t.path = path
	return nil
}

func (t *Terragrunt) Init(workingDir string) error {
	return nil
}

func (t *Terragrunt) Plan() (string, error) {
	return "", errors.New("Placeholder")
}

func (t *Terragrunt) Apply() error {
	return errors.New("Placeholder")
}

func (t *Terragrunt) Show() ([]byte, error) {
	return nil, nil
}

func downloadTerragrunt(version string) (string, error) {
	cpuArch := runtime.GOARCH

	url := fmt.Sprintf("https://github.com/gruntwork-io/terragrunt/releases/download/v%s/terragrunt_%s", version, cpuArch)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	filename := fmt.Sprintf("terragrunt_%s", cpuArch)
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

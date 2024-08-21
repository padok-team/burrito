package tools

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclparse"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	tf "github.com/padok-team/burrito/internal/runner/tools/terraform"
	tg "github.com/padok-team/burrito/internal/runner/tools/terragrunt"
	log "github.com/sirupsen/logrus"
	tenvconfig "github.com/tofuutils/tenv/v3/config"
	"github.com/tofuutils/tenv/v3/versionmanager"
	"github.com/tofuutils/tenv/v3/versionmanager/builder"
)

type tenvWrapper = versionmanager.VersionManager

// Creates a `tenv` wrapper for the given tool (Terraform/Terragrunt/OpenTofu)
func newTenvWrapper(binaryPath string, toolName string) (*tenvWrapper, error) {
	conf, err := tenvconfig.InitConfigFromEnv()
	if err != nil {
		return nil, err
	}
	conf.ForceQuiet = true
	conf.RootPath = binaryPath
	conf.InitDisplayer(true)
	hclParser := hclparse.NewParser()
	versionManager := builder.Builders[toolName](&conf, hclParser)

	return &versionManager, nil
}

// detect if the tool is already installed and compatible with the version constraint
// Return the version of the tool found locally, or the version to install
func detect(binaryPath, toolName, versionConstraint string) (string, error) {
	tenvWrapper, err := newTenvWrapper(binaryPath, toolName)
	if err != nil {
		return "", err
	}

	if versionConstraint == "" {
		log.Infof("no version constraint specified for %s, searching for constraint in codebase", toolName)
		versionConstraint, err = tenvWrapper.Resolve("latest-allowed")
		if err != nil {
			return "", err
		}
	}
	version, err := tenvWrapper.Evaluate(versionConstraint, true)
	if err == versionmanager.ErrNoCompatibleLocally {
		log.Infof("compatible %s version %s found remotely", toolName, version)
		return version, nil
	}
	if err != nil {
		return "", err
	}
	// TODO: check integrity of local version!
	log.Infof("found compatible %s version %s already installed", toolName, version)

	return version, nil
}

// install the tool with the given version, do noting if already installed
func install(binaryPath, toolName, version string) error {
	tenvWrapper, err := newTenvWrapper(binaryPath, toolName)
	if err != nil {
		return err
	}
	return tenvWrapper.Install(version)
}

// If not already on the system, install Terraform and, if needed, Terragrunt binaries
func InstallBinaries(layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository, binaryPath, workingDir string) (TerraformExec, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("error getting current working directory: %s", err)
		return nil, err
	}
	err = os.Chdir(workingDir) // need to cd into the repo to detect tf versions
	if err != nil {
		log.Errorf("error changing directory: %s", err)
		return nil, err
	}
	defer os.Chdir(cwd)

	terraformVersion := configv1alpha1.GetTerraformVersion(repo, layer)
	terraformVersion, err = detect(binaryPath, "terraform", terraformVersion)
	if err != nil {
		return nil, err
	}
	if err := install(binaryPath, "terraform", terraformVersion); err != nil {
		return nil, err
	}
	tf := &tf.Terraform{
		ExecPath: filepath.Join(binaryPath, "Terraform", terraformVersion, "terraform"),
	}

	if configv1alpha1.GetTerragruntEnabled(repo, layer) {
		terragruntVersion := configv1alpha1.GetTerragruntVersion(repo, layer)
		terragruntVersion, err := detect(binaryPath, "terragrunt", terragruntVersion)
		if err != nil {
			return nil, err
		}
		if err := install(binaryPath, "terragrunt", terragruntVersion); err != nil {
			return nil, err
		}
		return &tg.Terragrunt{
			ExecPath:  filepath.Join(binaryPath, "Terragrunt", terragruntVersion, "terragrunt"),
			Terraform: tf,
		}, nil
	}

	log.Infof("binaries successfully installed")
	return tf, nil
}

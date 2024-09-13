package tools

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclparse"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	ot "github.com/padok-team/burrito/internal/runner/tools/opentofu"
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
	// OpenTofu is called "tofu" in the context of tenv
	if toolName == "opentofu" {
		toolName = "tofu"
	}
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
func InstallBinaries(layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository, binaryPath, workingDir string) (IacExec, error) {
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
	defer func() {
		err := os.Chdir(cwd)
		if err != nil {
			log.Errorf("error changing directory back to %s: %s", cwd, err)
		}
	}()
	iacTool := configv1alpha1.GetIacTool(repo, layer)
	if iacTool != "terraform" && iacTool != "opentofu" {
		return nil, errors.New("unsupported IaC tool, set spec.iacTool to 'terraform' or 'opentofu'")
	}
	iacToolVersion := configv1alpha1.GetIacToolVersion(repo, layer)
	iacToolVersion, err = detect(binaryPath, iacTool, iacToolVersion)
	if err != nil {
		return nil, err
	}
	if err := install(binaryPath, iacTool, iacToolVersion); err != nil {
		return nil, err
	}
	var iacExec IacExec
	if iacTool == "terraform" {
		log.Infof("using Terraform version %s", iacToolVersion)
		iacExec = &tf.Terraform{
			ExecPath: filepath.Join(binaryPath, "Terraform", iacToolVersion, "terraform"),
		}
	}
	if iacTool == "opentofu" {
		log.Infof("using OpenTofu version %s", iacToolVersion)
		iacExec = &ot.OpenTofu{
			ExecPath: filepath.Join(binaryPath, "OpenTofu", iacToolVersion, "tofu"),
		}
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
		log.Infof("using Terragrunt version %s as wrapper for %s", terragruntVersion, iacTool)
		if iacTool == "terraform" {
			return &tg.Terragrunt{
				ExecPath:  filepath.Join(binaryPath, "Terragrunt", terragruntVersion, "terragrunt"),
				Terraform: iacExec.(*tf.Terraform),
				OpenTofu:  nil,
			}, nil
		} else if iacTool == "opentofu" {
			return &tg.Terragrunt{
				ExecPath:  filepath.Join(binaryPath, "Terragrunt", terragruntVersion, "terragrunt"),
				Terraform: nil,
				OpenTofu:  iacExec.(*ot.OpenTofu),
			}, nil
		}

	}
	return iacExec, nil
}

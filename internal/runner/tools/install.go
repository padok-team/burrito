package tools

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclparse"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	ot "github.com/padok-team/burrito/internal/runner/tools/opentofu"
	tf "github.com/padok-team/burrito/internal/runner/tools/terraform"
	tg "github.com/padok-team/burrito/internal/runner/tools/terragrunt"
	log "github.com/sirupsen/logrus"
	tenvconfig "github.com/tofuutils/tenv/v4/config"
	"github.com/tofuutils/tenv/v4/versionmanager"
	"github.com/tofuutils/tenv/v4/versionmanager/builder"
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
	version, err := tenvWrapper.Evaluate(context.TODO(), versionConstraint, true)
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
	return tenvWrapper.Install(context.TODO(), version)
}

// If not already on the system, install Terraform and, if needed, Terragrunt binaries
func InstallBinaries(layer *configv1alpha1.TerraformLayer, repo *configv1alpha1.TerraformRepository, binaryPath, workingDir string) (BaseExec, error) {
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

	var baseExec BaseExec
	var baseExecVersion string
	if configv1alpha1.GetTerraformEnabled(repo, layer) {
		baseExecVersion, err = detect(binaryPath, "terraform", configv1alpha1.GetTerraformVersion(repo, layer))
		if err != nil {
			return nil, err
		}
		baseExec = tf.NewTerraform(filepath.Join(binaryPath, "Terraform", baseExecVersion, "terraform"))
	} else if configv1alpha1.GetOpenTofuEnabled(repo, layer) {
		baseExecVersion, err = detect(binaryPath, "tofu", configv1alpha1.GetOpenTofuVersion(repo, layer))
		if err != nil {
			return nil, err
		}
		baseExec = ot.NewOpenTofu(filepath.Join(binaryPath, "OpenTofu", baseExecVersion, "tofu"))
	} else {
		return nil, errors.New("Please enable either Terraform or OpenTofu in the repository or layer configuration")
	}
	log.Infof("using %s version %s", baseExec.TenvName(), baseExecVersion)

	if err := install(binaryPath, baseExec.TenvName(), baseExecVersion); err != nil {
		return nil, err
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
		log.Infof("using Terragrunt version %s as wrapper for %s", terragruntVersion, baseExec.TenvName())
		return &tg.Terragrunt{
			ExecPath:      filepath.Join(binaryPath, "Terragrunt", terragruntVersion, "terragrunt"),
			ChildExecPath: baseExec.GetExecPath(),
		}, nil
	}
	return baseExec, nil
}

package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
)

const (
	WorkingDir = "/tmp/burrito/repositories"
	BundleDir  = "/tmp/burrito/gitbundles"
)

// ReferenceName converts a ref string to a plumbing.ReferenceName
// If ref starts with "refs/", use it directly; otherwise assume it's a branch
func ReferenceName(ref string) plumbing.ReferenceName {
	if strings.HasPrefix(ref, "refs/") {
		return plumbing.ReferenceName(ref)
	}
	// Default to branch for backward compatibility
	return plumbing.NewBranchReferenceName(ref)
}

func GetGitBundle(repository *configv1alpha1.TerraformRepository, ref string, revision string, auth transport.AuthMethod) ([]byte, error) {
	repoKey := fmt.Sprintf("%s-%s-%s", repository.Namespace, repository.Name, strings.ReplaceAll(ref, "/", "--"))
	repoDir := filepath.Join(WorkingDir, repoKey)

	// Try to open existing repository
	repo, err := git.PlainOpen(repoDir)
	if err != nil {
		if err != git.ErrRepositoryNotExists {
			return nil, fmt.Errorf("failed to open repository %s: %w", repoKey, err)
		}

		// Clone if it doesn't exist
		log.Infof("Cloning repository %s to %s", repository.Spec.Repository.Url, repoDir)

		cloneOpts := &git.CloneOptions{
			URL:           repository.Spec.Repository.Url,
			Auth:          auth,
			ReferenceName: ReferenceName(ref),
		}

		repo, err = git.PlainClone(repoDir, false, cloneOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository %s: %w", repoKey, err)
		}
	}

	// Fetch latest changes
	fetchOpts := &git.FetchOptions{
		Auth: auth,
	}

	log.Infof("fetching latest changes for repo %s", repoKey)
	err = repo.Fetch(fetchOpts)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Infof("repository %s is already up-to-date", repoKey)
		} else {
			return nil, fmt.Errorf("failed to fetch latest changes: %w", err)
		}
	}

	// Create BundleDir if it doesn't exist
	if _, err := os.Stat(BundleDir); os.IsNotExist(err) {
		if err := os.MkdirAll(BundleDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create BundleDir directory: %v", err)
		}
	}
	bundleDest := filepath.Join(BundleDir, fmt.Sprintf("%s.gitbundle", repoKey))
	bundle, err := createGitBundle(repoDir, bundleDest, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to create bundle: %w", err)
	}

	return bundle, nil
}

// Create a git bundle with `git bundle create` and return the content as a byte array
func createGitBundle(sourceDir, destination, ref string) ([]byte, error) {
	cmd := exec.Command("git", "-C", sourceDir, "bundle", "create", destination, ref)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create git bundle: %v, output: %s", err, string(output))
	}
	data, err := os.ReadFile(destination)
	if err != nil {
		return nil, fmt.Errorf("failed to read git bundle: %v", err)
	}
	return data, nil
}

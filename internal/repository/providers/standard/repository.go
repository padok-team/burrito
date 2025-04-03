package standard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	configv1alpha1 "github.com/padok-team/burrito/api/v1alpha1"
	log "github.com/sirupsen/logrus"
)

type GitProvider struct {
	transport.AuthMethod
	URL            string
	gitRepository  *git.Repository
	workingDir     string
	repositoryPath string
}

func (p *GitProvider) GetLatestRevisionForRef(repository *configv1alpha1.TerraformRepository, ref string) (string, error) {
	// Create an in-memory remote
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repository.Spec.Repository.Url},
	})

	// List references on the remote (equivalent to `git ls-remote <repoURL>`)
	refs, err := remote.List(&git.ListOptions{
		Auth: p.AuthMethod,
	})
	if err != nil {
		return "", fmt.Errorf("failed to list references: %v", err)
	}

	candidates := []string{
		"refs/heads/" + ref,
		"refs/tags/" + ref,
		ref, // in case someone passes the full ref already
	}

	// Look for the ref in the remote’s references
	for _, c := range candidates {
		for _, r := range refs {
			if r.Name().String() == c {
				return r.Hash().String(), nil
			}
		}
	}

	return "", fmt.Errorf("unable to find commit SHA for ref %q in %q", ref, repository.Spec.Repository.Url)
}

func (p *GitProvider) Bundle(ref string) ([]byte, error) {
	if p.gitRepository == nil {
		if err := p.clone(); err != nil {
			return nil, err
		}
	}
	reference, err := p.gitRepository.Reference(plumbing.NewBranchReferenceName(ref), true)
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}
	commit := reference.Hash().String()
	bundleDest := filepath.Join(p.repositoryPath, fmt.Sprintf("%s.gitbundle", commit))
	bundle, err := createGitBundle(p.repositoryPath, bundleDest, commit)
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func (p *GitProvider) clone() error {
	cloneOptions := &git.CloneOptions{
		URL:  p.URL,
		Auth: p.AuthMethod,
	}
	workingDir, err := os.MkdirTemp("", "burrito-repo-*")
	if err != nil {
		return errors.New("failed to create temporary directory")
	}
	p.workingDir = workingDir
	p.repositoryPath = filepath.Join(p.workingDir, "repository")
	log.Infof("cloning repository %s to %s", p.URL, p.repositoryPath)
	p.gitRepository, err = git.PlainClone(p.repositoryPath, false, cloneOptions)
	if err != nil {
		return err
	}
	return nil
}

func (p *GitProvider) GetChanges(previousCommit, currentCommit string) []string {
	if p.gitRepository == nil {
		if err := p.clone(); err != nil {
			log.Errorf("failed to clone repository: %v", err)
			return nil
		}
	}
	c1, err := p.gitRepository.CommitObject(plumbing.NewHash(previousCommit))
	if err != nil {
		log.Errorf("failed to get previous commit: %v", err)
		return nil
	}
	t1, err := c1.Tree()
	if err != nil {
		log.Errorf("failed to get previous tree: %v", err)
	}
	c2, err := p.gitRepository.CommitObject(plumbing.NewHash(currentCommit))
	if err != nil {
		log.Errorf("failed to get current commit: %v", err)
		return nil
	}
	t2, err := c2.Tree()
	if err != nil {
		log.Errorf("failed to get current tree: %v", err)
	}
	changes, err := object.DiffTree(t1, t2)
	if err != nil {
		log.Errorf("failed to get diff: %v", err)
		return nil
	}
	var paths []string
	for _, change := range changes {
		paths = append(paths, change.To.Name)
	}
	return paths
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

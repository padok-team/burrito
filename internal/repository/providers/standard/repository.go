package standard

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	log "github.com/sirupsen/logrus"
)

type GitProvider struct {
	transport.AuthMethod
	RepoURL        string
	gitRepository  *git.Repository
	workingDir     string
	repositoryPath string
}

const remote string = "origin"

func (p *GitProvider) GetLatestRevisionForRef(ref string) (string, error) {
	// Create an in-memory remote
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: remote,
		URLs: []string{p.RepoURL},
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

	return "", fmt.Errorf("unable to find commit SHA for ref %q in %q", ref, p.RepoURL)
}

// getReferenceName converts a ref string to a plumbing.ReferenceName
// If ref starts with "refs/", use it directly; otherwise assume it's a branch
func getReferenceName(ref string) plumbing.ReferenceName {
	if strings.HasPrefix(ref, "refs/") {
		return plumbing.ReferenceName(ref)
	}

	// Default to branch for backward compatibility
	return plumbing.NewBranchReferenceName(ref)
}

// getRemoteReferenceName converts a ref string to a remote plumbing.ReferenceName
func getRemoteReferenceName(ref string) plumbing.ReferenceName {
	if strings.HasPrefix(ref, "refs/") {
		return plumbing.ReferenceName(ref)
	}

	// Default to remote branch
	return plumbing.NewRemoteReferenceName(remote, ref)
}

func (p *GitProvider) Bundle(ref string) ([]byte, error) {
	// Clone repository if it doesn't exist
	if p.gitRepository == nil {
		if err := p.clone(); err != nil {
			return nil, err
		}
	}

	// First, try to get the local branch reference
	localRefName := getReferenceName(ref)
	reference, err := p.gitRepository.Reference(localRefName, true)
	if err != nil {
		// If local branch doesn't exist, try to find the remote reference and create local branch
		remoteRefName := getRemoteReferenceName(ref)
		remoteRef, remoteErr := p.gitRepository.Reference(remoteRefName, true)
		if remoteErr != nil {
			return nil, fmt.Errorf("failed to get reference %s (tried both local %s and remote %s): local_err=%w, remote_err=%v", ref, localRefName, remoteRefName, err, remoteErr)
		}

		// Create a local branch from the remote reference
		log.Infof("creating local branch %s from remote %s", localRefName, remoteRefName)
		localRef := plumbing.NewHashReference(localRefName, remoteRef.Hash())
		err = p.gitRepository.Storer.SetReference(localRef)
		if err != nil {
			return nil, fmt.Errorf("failed to create local branch %s: %w", localRefName, err)
		}
		reference = localRef
	}

	// Checkout the branch
	worktree, err := p.gitRepository.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree for repository %s: %w", p.RepoURL, err)
	}

	// Checkout the specific branch
	checkoutOpts := &git.CheckoutOptions{
		Branch: reference.Name(),
	}

	log.Infof("checking out branch %s", reference.Name())
	err = worktree.Checkout(checkoutOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch %s: %w", reference.Name(), err)
	}

	// Pull latest changes
	pullOpts := &git.PullOptions{
		Auth:       p.AuthMethod,
		RemoteName: remote,
	}

	log.Infof("pulling latest changes for repo %s on branch %s", p.RepoURL, reference.Name())
	err = worktree.Pull(pullOpts)
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			log.Info("repository is already up-to-date")
		} else {
			return nil, fmt.Errorf("failed to pull latest changes for ref %s: %w", reference.Name(), err)
		}
	}

	// Create git bundle
	commit := reference.Hash().String()
	bundleDest := filepath.Join(p.repositoryPath, fmt.Sprintf("%s.gitbundle", commit))
	bundle, err := createGitBundle(p.repositoryPath, bundleDest, ref)
	if err != nil {
		return nil, err
	}
	return bundle, nil
}

func (p *GitProvider) clone() error {
	// Create a consistent directory name based on repository URL hash
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(p.RepoURL)))
	p.workingDir = filepath.Join(os.TempDir(), "burrito-repo-"+hash)
	p.repositoryPath = filepath.Join(p.workingDir, "repository")

	// Check if repository already exists
	if _, err := os.Stat(p.repositoryPath); err == nil {
		// Repository already exists, just open it
		log.Infof("repository already exists at %s, opening existing clone", p.repositoryPath)
		repo, err := git.PlainOpen(p.repositoryPath)
		if err != nil {
			return fmt.Errorf("failed to open existing repository: %w", err)
		}
		p.gitRepository = repo

		// Fetch latest changes from remote, notably to fetch new branches or tags
		log.Infof("fetching latest changes from remote")
		err = repo.Fetch(&git.FetchOptions{
			Auth: p.AuthMethod,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			log.Warnf("failed to fetch latest changes: %v", err)
		}

		return nil
	}

	// Create the working directory
	if err := os.MkdirAll(p.workingDir, 0755); err != nil {
		return fmt.Errorf("failed to create working directory: %w", err)
	}

	cloneOptions := &git.CloneOptions{
		URL:  p.RepoURL,
		Auth: p.AuthMethod,
	}

	log.Infof("cloning repository %s to %s", p.RepoURL, p.repositoryPath)
	repo, err := git.PlainClone(p.repositoryPath, false, cloneOptions)
	if err != nil {
		return err
	}
	p.gitRepository = repo
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

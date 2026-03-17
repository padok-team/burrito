package standard

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testSig() *object.Signature {
	return &object.Signature{Name: "test", Email: "test@example.com", When: time.Now()}
}

// TestBundle_NonDefaultBranch_FastForwards is a regression test for the
// "non-fast-forward update" error when Bundle() is called on a non-default
// branch whose Pull was resolved to the wrong remote ref.
//
// Root cause: Pull() without ReferenceName falls back to the remote's HEAD
// (the default branch). When the default branch has commits not in the
// non-default branch, this is a non-fast-forward relative to the local branch.
//
// Fix: always pass ReferenceName in PullOptions so go-git targets the correct
// remote branch rather than the remote's default branch.
func TestBundle_NonDefaultBranch_FastForwards(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git CLI not available")
	}

	// Set up a "remote" repo with diverged master and feature branches:
	//   master:  A → B  (B has a master-only file)
	//   feature: A → C  (C has a feature-only file; diverges from master at A)
	remoteDir := t.TempDir()
	remoteRepo, err := git.PlainInit(remoteDir, false)
	require.NoError(t, err)

	wt, err := remoteRepo.Worktree()
	require.NoError(t, err)
	sig := testSig()

	// Commit A — initial commit on master
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "README.md"), []byte("initial"), 0644))
	_, err = wt.Add("README.md")
	require.NoError(t, err)
	commitA, err := wt.Commit("initial commit", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	// Create feature branch pointing at A (before master advances)
	require.NoError(t, remoteRepo.Storer.SetReference(
		plumbing.NewHashReference("refs/heads/feature", commitA),
	))

	// Commit B on master (master = A→B, feature = A)
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "master.txt"), []byte("master-only"), 0644))
	_, err = wt.Add("master.txt")
	require.NoError(t, err)
	_, err = wt.Commit("advance master", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	// Commit C on feature (feature = A→C, diverged from master)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/feature"}))
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "feature.txt"), []byte("feature-v1"), 0644))
	_, err = wt.Add("feature.txt")
	require.NoError(t, err)
	_, err = wt.Commit("feature commit", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	// Leave remote worktree on master so the default branch is correct
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/master"}))

	// Clone the remote
	localDir := t.TempDir()
	cloned, err := git.PlainClone(localDir, false, &git.CloneOptions{URL: remoteDir})
	require.NoError(t, err)

	// Build a GitProvider with the cloned repo injected directly (bypasses
	// the clone() call and the hardcoded repositoryDir constant).
	workingDir := t.TempDir()
	p := &GitProvider{
		RepoURL:        remoteDir,
		gitRepository:  cloned,
		repositoryPath: localDir,
		workingDir:     workingDir,
	}

	// First Bundle("feature") call: local feature branch doesn't exist yet.
	// Bundle() creates it from origin/feature (C) and pulls (already up-to-date).
	_, err = p.Bundle("feature")
	require.NoError(t, err, "first Bundle() on non-default branch should succeed")

	// Advance feature on remote: commit D (feature = A→C→D)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/feature"}))
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "feature.txt"), []byte("feature-v2"), 0644))
	_, err = wt.Add("feature.txt")
	require.NoError(t, err)
	_, err = wt.Commit("advance feature", &git.CommitOptions{Author: sig})
	require.NoError(t, err)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/master"}))

	// Second Bundle("feature") call: local feature is at C, remote feature is
	// at D (descendant of C), remote master is at B (not a descendant of C).
	// Without the fix, Pull targets origin/master (B) → non-fast-forward update.
	// With the fix, Pull targets origin/feature (D) → fast-forward C→D.
	_, err = p.Bundle("feature")
	assert.NoError(t, err, "second Bundle() should fast-forward feature to new remote tip, not fail with non-fast-forward update")
}

// TestBundle_NonDefaultBranch_DirectDescendant covers the scenario from the
// original bug report: a non-default branch that is a direct descendant of the
// default branch (feature = master + extra commits on top).
//
// The error in this case is non-obvious: go-git's Pull without ReferenceName
// targets origin/master (B) and tries to fast-forward local feature (C) to B.
// Since B is an ancestor of C (not a descendant), this is a non-fast-forward
// update — go-git refuses to move the branch pointer backwards.
//
// With the fix (ReferenceName set), Pull correctly targets origin/feature (D)
// and fast-forwards C → D.
func TestBundle_NonDefaultBranch_DirectDescendant(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git CLI not available")
	}

	// Set up remote with feature as a direct descendant of master:
	//   master:  A → B
	//   feature: A → B → C  (ahead of master, contains all master commits)
	remoteDir := t.TempDir()
	remoteRepo, err := git.PlainInit(remoteDir, false)
	require.NoError(t, err)

	wt, err := remoteRepo.Worktree()
	require.NoError(t, err)
	sig := testSig()

	// Commit A on master
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "README.md"), []byte("initial"), 0644))
	_, err = wt.Add("README.md")
	require.NoError(t, err)
	_, err = wt.Commit("initial commit", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	// Commit B on master
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "base.txt"), []byte("base"), 0644))
	_, err = wt.Add("base.txt")
	require.NoError(t, err)
	commitB, err := wt.Commit("base commit", &git.CommitOptions{Author: sig})
	require.NoError(t, err)

	// Create feature branch at B (direct descendant starting point)
	require.NoError(t, remoteRepo.Storer.SetReference(
		plumbing.NewHashReference("refs/heads/feature", commitB),
	))

	// Commit C on feature (feature = A→B→C, master = A→B)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/feature"}))
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "feature.txt"), []byte("feature-v1"), 0644))
	_, err = wt.Add("feature.txt")
	require.NoError(t, err)
	commitC, err := wt.Commit("feature commit", &git.CommitOptions{Author: sig})
	require.NoError(t, err)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/master"}))

	// Clone the remote
	localDir := t.TempDir()
	cloned, err := git.PlainClone(localDir, false, &git.CloneOptions{URL: remoteDir})
	require.NoError(t, err)

	workingDir := t.TempDir()
	p := &GitProvider{
		RepoURL:        remoteDir,
		gitRepository:  cloned,
		repositoryPath: localDir,
		workingDir:     workingDir,
	}

	// First Bundle("feature"): creates local feature = C, pulls (already up-to-date)
	_, err = p.Bundle("feature")
	require.NoError(t, err, "first Bundle() on direct-descendant branch should succeed")

	// Advance feature on remote: commit D (feature = A→B→C→D)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/feature"}))
	require.NoError(t, os.WriteFile(filepath.Join(remoteDir, "feature.txt"), []byte("feature-v2"), 0644))
	_, err = wt.Add("feature.txt")
	require.NoError(t, err)
	commitD, err := wt.Commit("advance feature", &git.CommitOptions{Author: sig})
	require.NoError(t, err)
	require.NoError(t, wt.Checkout(&git.CheckoutOptions{Branch: "refs/heads/master"}))

	// Second Bundle("feature"): local feature is at C, remote feature is at D.
	// master is still at B (ancestor of both C and D) — no divergence.
	// Without the fix, Pull targets origin/master (B), which is an ancestor of
	// C, so it returns "already up-to-date" and the bundle is built from C
	// (stale). With the fix, Pull targets origin/feature (D) and fast-forwards.
	_, err = p.Bundle("feature")
	require.NoError(t, err, "second Bundle() on direct-descendant branch should succeed")

	// Verify the local feature branch was advanced to D, not left at C.
	ref, err := cloned.Reference(plumbing.NewBranchReferenceName("feature"), true)
	require.NoError(t, err)
	assert.Equal(t, commitD, ref.Hash(),
		"feature branch should point to the new remote tip D after pull, not the stale C (%s)", commitC)
}

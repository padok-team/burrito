// test-non-default-branch reproduces the Bundle() code path for a non-default
// branch and prints detailed diagnostic output at each step.
//
// It mirrors internal/repository/providers/standard/repository.go exactly so
// the output is directly comparable to what burrito does in production.
//
// The bug: Pull() without ReferenceName targets the remote's default branch
// instead of the requested branch. This causes two failure modes:
//
//   - Silent: if the default branch is already reachable from the local branch
//     (e.g. feature is ahead of main), Pull says "already up-to-date" and
//     returns a stale bundle, missing any new commits on the feature branch.
//
//   - Visible: if the default branch has commits not reachable from the local
//     branch (i.e. the branches have diverged), Pull fails with
//     "non-fast-forward update".
//
// Usage:
//
//	go run ./hack/test-non-default-branch --url <repo-url> --branch <branch> [--token <token>]
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func main() {
	url := flag.String("url", "", "Repository URL (required)")
	branch := flag.String("branch", "", "Non-default branch to test (required)")
	token := flag.String("token", "", "Access token (optional, for private repos)")
	flag.Parse()

	if *url == "" || *branch == "" {
		fmt.Fprintln(os.Stderr, "Usage: go run ./hack/test-non-default-branch --url <url> --branch <branch> [--token <token>]")
		os.Exit(1)
	}

	var auth *githttp.BasicAuth
	if *token != "" {
		auth = &githttp.BasicAuth{Username: "x-token", Password: *token}
	}

	tmpDir, err := os.MkdirTemp("", "burrito-test-*")
	if err != nil {
		fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// =========================================================================
	// STEP 0 — Clone (mirrors GitProvider.clone())
	// =========================================================================
	fmt.Println("=== STEP 0: Clone ===")

	cloneOpts := &git.CloneOptions{
		URL:  *url,
		Auth: auth,
	}
	repo, err := git.PlainClone(tmpDir, false, cloneOpts)
	if err != nil {
		fatalf("✗ clone failed: %v", err)
	}
	fmt.Printf("✓ cloned to %s\n", tmpDir)

	// Print git config (remotes + branch tracking)
	cfg, err := repo.Config()
	if err != nil {
		fmt.Printf("  config: error reading: %v\n", err)
	} else {
		fmt.Println("  config.remotes:")
		for name, r := range cfg.Remotes {
			fmt.Printf("    %s -> URLs=%v  fetch=%v\n", name, r.URLs, r.Fetch)
		}
		fmt.Println("  config.branches:")
		if len(cfg.Branches) == 0 {
			fmt.Println("    (none)")
		}
		for name, b := range cfg.Branches {
			fmt.Printf("    %s -> remote=%s  merge=%s\n", name, b.Remote, b.Merge)
		}
	}

	// Print all refs
	fmt.Println("  refs:")
	iter, err := repo.References()
	if err != nil {
		fmt.Printf("    error listing refs: %v\n", err)
	} else {
		_ = iter.ForEach(func(ref *plumbing.Reference) error {
			fmt.Printf("    %s -> %s\n", ref.Name(), ref.Hash())
			return nil
		})
	}

	// Print HEAD and capture default branch tip for later comparison
	head, err := repo.Head()
	if err != nil {
		fmt.Printf("  HEAD: error: %v\n", err)
	} else {
		fmt.Printf("  HEAD: %s -> %s\n", head.Name(), head.Hash())
	}
	var defaultBranchTip plumbing.Hash
	if head != nil {
		defaultBranchTip = head.Hash()
	}

	// =========================================================================
	// STEP 1 — Local ref lookup (mirrors Bundle() line 95)
	// =========================================================================
	fmt.Printf("\n=== STEP 1: local ref refs/heads/%s ===\n", *branch)

	localRefName := plumbing.NewBranchReferenceName(*branch)
	localRef, err := repo.Reference(localRefName, true)
	if err != nil {
		fmt.Printf("✗ ref not found: %v\n", err)
	} else {
		fmt.Printf("✓ %s -> %s\n", localRef.Name(), localRef.Hash())
	}

	// =========================================================================
	// STEP 2 — Remote ref lookup (mirrors Bundle() line 99)
	// =========================================================================
	fmt.Printf("\n=== STEP 2: remote ref refs/remotes/origin/%s ===\n", *branch)

	remoteRefName := plumbing.NewRemoteReferenceName("origin", *branch)
	remoteRef, err := repo.Reference(remoteRefName, true)
	if err != nil {
		fmt.Printf("✗ remote ref not found: %v\n", err)
		fmt.Println("  available remote refs:")
		iter2, _ := repo.References()
		_ = iter2.ForEach(func(r *plumbing.Reference) error {
			if strings.HasPrefix(r.Name().String(), "refs/remotes/") {
				fmt.Printf("    %s -> %s\n", r.Name(), r.Hash())
			}
			return nil
		})
		os.Exit(1)
	}
	fmt.Printf("✓ %s -> %s\n", remoteRef.Name(), remoteRef.Hash())

	// =========================================================================
	// STEP 3 — SetReference (mirrors Bundle() line 107)
	// =========================================================================
	fmt.Printf("\n=== STEP 3: SetReference refs/heads/%s -> %s ===\n", *branch, remoteRef.Hash())

	newLocalRef := plumbing.NewHashReference(localRefName, remoteRef.Hash())
	if err := repo.Storer.SetReference(newLocalRef); err != nil {
		fatalf("✗ SetReference failed: %v", err)
	}
	fmt.Printf("✓ created local branch (no upstream tracking config)\n")

	// Confirm no branch tracking was set
	cfg2, _ := repo.Config()
	if b, ok := cfg2.Branches[*branch]; ok {
		fmt.Printf("  branch tracking: remote=%s merge=%s\n", b.Remote, b.Merge)
	} else {
		fmt.Println("  branch tracking: (none — this is the suspected root cause)")
	}

	// =========================================================================
	// STEP 4 — Checkout (mirrors Bundle() line 126)
	// =========================================================================
	fmt.Printf("\n=== STEP 4: Checkout refs/heads/%s ===\n", *branch)

	worktree, err := repo.Worktree()
	if err != nil {
		fatalf("✗ Worktree() failed: %v", err)
	}

	checkoutOpts := &git.CheckoutOptions{
		Branch: localRefName,
	}
	if err := worktree.Checkout(checkoutOpts); err != nil {
		fatalf("✗ Checkout failed: %v", err)
	}
	fmt.Println("✓ checked out")

	// Print HEAD after checkout
	head2, _ := repo.Head()
	fmt.Printf("  HEAD after checkout: %s -> %s\n", head2.Name(), head2.Hash())

	// =========================================================================
	// STEP 5 — Pull (mirrors Bundle() line 138 — the critical step)
	//
	// Local branch was just set to the remote tip (Step 3), so the pull has
	// nothing to fetch for this branch. The failure mode here depends on what
	// go-git picks as the merge target without tracking config:
	//
	//   "already up-to-date": go-git targeted the default branch and its tip
	//   is already reachable from local (e.g. feature is ahead of main).
	//   This is the SILENT bug: if the feature branch later gains new commits,
	//   future pulls will still say "already up-to-date" and return stale bundles.
	//
	//   "non-fast-forward update": go-git targeted the default branch and its
	//   tip is NOT reachable from local (branches have diverged).
	//   This is the VISIBLE bug reported by users.
	// =========================================================================
	fmt.Println("\n=== STEP 5: Pull (RemoteName=origin, no ReferenceName) ===")

	pullOpts := &git.PullOptions{
		Auth:       auth,
		RemoteName: "origin",
	}
	err = worktree.Pull(pullOpts)
	if err == nil || err == git.NoErrAlreadyUpToDate {
		fmt.Printf("✓ success (%v)\n", err)
	} else {
		fmt.Printf("✗ %v\n", err)
	}

	// =========================================================================
	// STEP 6 — Simulate second Bundle() call.
	//
	// Rewinds local branch to its parent to simulate "local is behind remote"
	// (i.e. the remote feature branch gained a commit since the last Bundle).
	// Then pulls to check whether go-git fast-forwards to the correct commit.
	//
	// NOTE: if the branch is a direct descendant of the default branch, the
	// parent commit equals the default branch's tip. In that case, Step 6a
	// will show "already up-to-date" (the silent bug — go-git targets the
	// default branch, whose tip is reachable from the parent, so it thinks
	// there is nothing to do and misses the new feature commits). The visible
	// "non-fast-forward" error only appears when the default branch has commits
	// that are NOT reachable from the rewound local branch.
	//
	// Step 6b shows the fix in both cases: with ReferenceName set, go-git
	// always targets the correct remote branch and fast-forwards properly.
	// =========================================================================
	fmt.Printf("\n=== STEP 6: Simulate second Bundle() call (rewind local branch to parent) ===\n")

	commit, err := repo.CommitObject(remoteRef.Hash())
	if err != nil {
		fmt.Printf("  skipped: cannot get commit object: %v\n", err)
	} else if len(commit.ParentHashes) == 0 {
		fmt.Printf("  skipped: branch tip has no parent (root commit)\n")
	} else {
		parentHash := commit.ParentHashes[0]
		fmt.Printf("  rewinding refs/heads/%s: %s → parent %s\n", *branch, remoteRef.Hash(), parentHash)
		if !defaultBranchTip.IsZero() {
			if parentHash == defaultBranchTip {
				fmt.Printf("  ⚠ parent == default branch tip (%s): Step 6a will show the silent bug\n", defaultBranchTip)
				fmt.Printf("    (go-git targets the default branch, whose tip is already reachable — stale bundle)\n")
			} else {
				fmt.Printf("  default branch tip: %s\n", defaultBranchTip)
			}
		}

		rewound := plumbing.NewHashReference(localRefName, parentHash)
		if err := repo.Storer.SetReference(rewound); err != nil {
			fmt.Printf("  ✗ SetReference (rewind) failed: %v\n", err)
		} else {
			fmt.Printf("  ✓ rewound\n")

			fmt.Printf("\n=== STEP 6a: Pull after rewind (RemoteName=origin, no ReferenceName) ===\n")
			err = worktree.Pull(&git.PullOptions{
				Auth:       auth,
				RemoteName: "origin",
			})
			if err == nil || err == git.NoErrAlreadyUpToDate {
				fmt.Printf("  ✓ success (%v)\n", err)
				h, _ := repo.Reference(localRefName, true)
				if h != nil && h.Hash() != remoteRef.Hash() {
					fmt.Printf("  ⚠ local branch is at %s, not remote tip %s — stale bundle!\n", h.Hash(), remoteRef.Hash())
				}
			} else {
				fmt.Printf("  ✗ %v\n", err)
			}

			// Rewind again for 6b comparison
			if err2 := repo.Storer.SetReference(rewound); err2 == nil {
				fmt.Printf("\n=== STEP 6b: Pull after rewind (ReferenceName=refs/heads/%s) — the fix ===\n", *branch)
				err = worktree.Pull(&git.PullOptions{
					Auth:          auth,
					RemoteName:    "origin",
					ReferenceName: localRefName,
				})
				if err == nil || err == git.NoErrAlreadyUpToDate {
					h, _ := repo.Reference(localRefName, true)
					if h != nil && h.Hash() == remoteRef.Hash() {
						fmt.Printf("  ✓ success — local branch correctly advanced to remote tip %s\n", remoteRef.Hash())
					} else {
						fmt.Printf("  ✓ success (%v)\n", err)
					}
				} else {
					fmt.Printf("  ✗ %v\n", err)
				}
			}
		}
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

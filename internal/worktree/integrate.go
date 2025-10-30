package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Integrate rebases a worktree branch onto the default branch, fast-forward merges it,
// and removes the corresponding worktree/branch when the merge succeeds.
func Integrate(branch string) error {
	repo, err := GetRepoRoot()
	if err != nil {
		return err
	}

	worktrees, err := parseWorktrees()
	if err != nil {
		return err
	}

	var target *Worktree
	for _, wt := range worktrees {
		if wt.Branch == branch {
			wtCopy := wt
			target = &wtCopy
			break
		}
	}
	if target == nil {
		return fmt.Errorf("worktree '%s' not found", branch)
	}

	if _, err := os.Stat(target.Path); err != nil {
		return fmt.Errorf("worktree path %s is not accessible: %w", target.Path, err)
	}

	defaultBranch := detectDefaultBranch(repo)
	if branch == defaultBranch {
		return fmt.Errorf("branch '%s' is already the default branch", branch)
	}

	primaryPath, err := getPrimaryWorktreePath(repo)
	if err != nil {
		return err
	}

	if err := ensureCleanWorktree(target.Path); err != nil {
		return fmt.Errorf("%s has uncommitted changes: %w", target.Path, err)
	}
	if err := ensureCleanWorktree(primaryPath); err != nil {
		return fmt.Errorf("%s has uncommitted changes: %w", primaryPath, err)
	}

	remoteName := "origin"
	useRemote := hasRemote(primaryPath, remoteName)

	if err := rebaseWorktreeOntoDefault(target.Path, defaultBranch, remoteName, useRemote); err != nil {
		return err
	}

	if err := fastForwardDefaultBranch(primaryPath, branch, defaultBranch, remoteName, useRemote); err != nil {
		return err
	}

	if err := RemoveWithOptions(branch, RemoveOptions{DeleteBranch: true}); err != nil {
		return err
	}

	fmt.Printf("Integrated %s into %s and removed worktree.\n", branch, defaultBranch)
	return nil
}

func getPrimaryWorktreePath(repo string) (string, error) {
	cmd := exec.Command("git", "-C", repo, "rev-parse", "--git-common-dir")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to determine git common dir: %w", err)
	}

	commonDir := strings.TrimSpace(string(output))
	if !filepath.IsAbs(commonDir) {
		commonDir = filepath.Join(repo, commonDir)
	}

	return filepath.Dir(commonDir), nil
}

func ensureCleanWorktree(path string) error {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(output)) != "" {
		return fmt.Errorf("working tree is dirty")
	}
	return nil
}

func hasRemote(path, remote string) bool {
	cmd := exec.Command("git", "-C", path, "remote")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(line) == remote {
			return true
		}
	}
	return false
}

func rebaseWorktreeOntoDefault(path, defaultBranch, remote string, useRemote bool) error {
	target := defaultBranch
	if useRemote {
		if err := runGitCommandStreaming(path, "fetch", remote, "--prune"); err != nil {
			return fmt.Errorf("failed to fetch %s: %w", remote, err)
		}
		target = fmt.Sprintf("%s/%s", remote, defaultBranch)
	}

	if err := runGitCommandStreaming(path, "rebase", target); err != nil {
		return fmt.Errorf("rebase onto %s failed: %w", target, err)
	}
	return nil
}

func fastForwardDefaultBranch(path, branch, defaultBranch, remote string, useRemote bool) error {
	if useRemote {
		if err := runGitCommandStreaming(path, "fetch", remote, "--prune"); err != nil {
			return fmt.Errorf("failed to fetch %s: %w", remote, err)
		}
	}

	if err := runGitCommandStreaming(path, "checkout", defaultBranch); err != nil {
		return fmt.Errorf("failed to checkout %s: %w", defaultBranch, err)
	}

	if useRemote {
		if err := runGitCommandStreaming(path, "pull", "--rebase", remote, defaultBranch); err != nil {
			return fmt.Errorf("pull --rebase failed: %w", err)
		}
	}

	if err := runGitCommandStreaming(path, "merge", "--ff-only", branch); err != nil {
		return fmt.Errorf("fast-forward merge failed: %w", err)
	}

	return nil
}

func runGitCommandStreaming(path string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

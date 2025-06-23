package worktree

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tobiase/worktree-utils/internal/interactive"
)

const (
	mainBranchName   = "main"
	masterBranchName = "master"
)

// SelectBranchInteractively presents a fuzzy finder for branch selection
func SelectBranchInteractively() (string, error) {
	branches, err := GetAvailableBranches()
	if err != nil {
		return "", err
	}

	if len(branches) == 0 {
		return "", fmt.Errorf("no worktrees available")
	}

	if len(branches) == 1 {
		return branches[0], nil
	}

	// Use interactive selection with preview
	return interactive.SelectStringWithPreview(
		branches,
		"Select worktree:",
		createBranchPreview,
	)
}

// createBranchPreview creates preview content for a branch selection
func createBranchPreview(i, width, height int) string {
	worktrees, err := GetWorktreeInfo()
	if err != nil || i < 0 || i >= len(worktrees) {
		return "Unable to load worktree information"
	}

	wt := worktrees[i]

	var preview strings.Builder
	preview.WriteString(fmt.Sprintf("Branch: %s\n", wt.Branch))
	preview.WriteString(fmt.Sprintf("Path:   %s\n", wt.Path))

	// Add branch comparison info
	branchInfo, err := getBranchInfo(wt.Branch)
	if err == nil && branchInfo != "" {
		preview.WriteString(fmt.Sprintf("Status: %s\n", branchInfo))
	}

	preview.WriteString("\n")

	// Add git status if available
	gitStatus, err := getGitStatus(wt.Path)
	if err == nil && gitStatus != "" {
		preview.WriteString(fmt.Sprintf("Working directory: %s\n\n", gitStatus))
	} else {
		preview.WriteString("Working directory: Clean\n\n")
	}

	// Add recent commits
	gitLog, err := getGitLog(wt.Branch, 5)
	if err != nil {
		preview.WriteString("Recent commits: (unable to load)")
	} else {
		preview.WriteString("Recent commits:\n")
		preview.WriteString(gitLog)
	}

	return preview.String()
}

// getGitLog returns recent commit messages for a branch (without hashes)
func getGitLog(branch string, count int) (string, error) {
	repo, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	// Get just commit messages, not hashes
	cmd := exec.Command("git", "-C", repo, "log",
		"--pretty=format:• %s",
		fmt.Sprintf("-%d", count),
		branch)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

// getBranchInfo returns branch comparison information
func getBranchInfo(branch string) (string, error) {
	repo, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	// Skip comparison for main branch
	if branch == mainBranchName || branch == masterBranchName {
		return "Main branch", nil
	}

	// Check if main exists, fallback to master
	mainBranch := mainBranchName
	cmd := exec.Command("git", "-C", repo, "rev-parse", "--verify", mainBranchName)
	if cmd.Run() != nil {
		cmd = exec.Command("git", "-C", repo, "rev-parse", "--verify", "master")
		if cmd.Run() != nil {
			return "No main/master branch found", nil
		}
		mainBranch = "master"
	}

	// Get ahead/behind count
	cmd = exec.Command("git", "-C", repo, "rev-list", "--left-right", "--count",
		fmt.Sprintf("%s...%s", mainBranch, branch))
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	counts := strings.Fields(strings.TrimSpace(string(output)))
	if len(counts) != 2 {
		return "Up to date", nil
	}

	behind := counts[0]
	ahead := counts[1]

	if ahead == "0" && behind == "0" {
		return fmt.Sprintf("Up to date with %s", mainBranch), nil
	} else if ahead == "0" {
		return fmt.Sprintf("%s commits behind %s", behind, mainBranch), nil
	} else if behind == "0" {
		return fmt.Sprintf("%s commits ahead of %s", ahead, mainBranch), nil
	}
	return fmt.Sprintf("%s ahead, %s behind %s", ahead, behind, mainBranch), nil
}

// getGitStatus returns the working directory status for a worktree
func getGitStatus(worktreePath string) (string, error) {
	cmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	status := string(output)
	if status == "" {
		return "Clean working directory", nil
	}

	// Count changes
	lines := strings.Split(strings.TrimSpace(status), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return "Clean working directory", nil
	}

	return fmt.Sprintf("%d files with changes", len(lines)), nil
}

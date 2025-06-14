package worktree

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tobiase/worktree-utils/internal/interactive"
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
	preview.WriteString(fmt.Sprintf("Path:   %s\n\n", wt.Path))

	// Add git log for the branch
	gitLog, err := getGitLog(wt.Branch, 5)
	if err != nil {
		preview.WriteString("Recent commits: (unable to load)\n")
	} else {
		preview.WriteString("Recent commits:\n")
		preview.WriteString(gitLog)
	}

	// Add git status if available
	gitStatus, err := getGitStatus(wt.Path)
	if err == nil && gitStatus != "" {
		preview.WriteString("\nWorking directory status:\n")
		preview.WriteString(gitStatus)
	}

	return preview.String()
}

// getGitLog returns recent commits for a branch
func getGitLog(branch string, count int) (string, error) {
	repo, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "-C", repo, "log",
		"--oneline",
		fmt.Sprintf("-%d", count),
		branch)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(output), nil
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

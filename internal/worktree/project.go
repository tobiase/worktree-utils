package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tobiase/worktree-utils/internal/config"
)

// GetGitRemote returns the origin remote URL for the current repository
func GetGitRemote() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return "", nil // No remote is OK
	}
	return strings.TrimSpace(string(output)), nil
}

// GetRelativePath returns the path relative to the repository root
func GetRelativePath(absolutePath string) (string, error) {
	repoRoot, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(repoRoot, absolutePath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// SmartNewWorktree creates or switches to a worktree intelligently based on branch state
// This is the "Do What I Mean" implementation that handles all cases:
// 1. Branch doesn't exist -> Create branch + worktree + switch
// 2. Branch exists, no worktree -> Create worktree + switch
// 3. Branch + worktree exist -> Just switch
func SmartNewWorktree(branch string, baseBranch string, cfg *config.Manager) (string, error) {
	branchExists := checkBranchExists(branch)
	worktreeExists := checkWorktreeExists(branch)

	if worktreeExists {
		// Case 3: Both branch and worktree exist - just switch
		return Go(branch)
	}

	if branchExists {
		// Case 2: Branch exists but no worktree - create worktree only
		fmt.Printf("Branch '%s' exists, creating worktree...\n", branch)
		return createWorktreeForExistingBranch(branch, cfg)
	}

	// Case 1: Branch doesn't exist - create branch + worktree
	fmt.Printf("Creating new branch '%s' and worktree...\n", branch)
	return createBranchAndWorktree(branch, baseBranch, cfg)
}

// createWorktreeForExistingBranch creates a worktree for an existing branch (old Add behavior)
func createWorktreeForExistingBranch(branch string, cfg *config.Manager) (string, error) {
	repo, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	worktreeBase, err := GetWorktreeBase()
	if err != nil {
		return "", err
	}

	// Use project-specific worktree base if configured
	if cfg != nil && cfg.GetCurrentProject() != nil {
		if projectBase := cfg.GetCurrentProject().Settings.WorktreeBase; projectBase != "" {
			worktreeBase = projectBase
		}
	}

	// Create worktree base directory if it doesn't exist
	if err := os.MkdirAll(worktreeBase, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory: %v", err)
	}

	worktreePath := filepath.Join(worktreeBase, branch)
	cmd := exec.Command("git", "-C", repo, "worktree", "add", worktreePath, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create worktree: %v", err)
	}

	return worktreePath, nil
}

// createBranchAndWorktree creates a new branch and its worktree (old NewWorktree behavior)
func createBranchAndWorktree(branch string, baseBranch string, cfg *config.Manager) (string, error) {
	repo, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	worktreeBase, err := GetWorktreeBase()
	if err != nil {
		return "", err
	}

	// Use project-specific worktree base if configured
	if cfg != nil && cfg.GetCurrentProject() != nil {
		if projectBase := cfg.GetCurrentProject().Settings.WorktreeBase; projectBase != "" {
			worktreeBase = projectBase
		}
	}

	// Create worktree base directory if it doesn't exist
	if err := os.MkdirAll(worktreeBase, 0755); err != nil {
		return "", fmt.Errorf("failed to create worktree directory: %v", err)
	}

	worktreePath := filepath.Join(worktreeBase, branch)

	// Create the worktree with new branch
	args := []string{"-C", repo, "worktree", "add", worktreePath}

	if baseBranch != "" {
		// Create new branch from base
		args = append(args, "-b", branch, baseBranch)
	} else {
		// Create new branch from current HEAD
		args = append(args, "-b", branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create worktree: %v", err)
	}

	return worktreePath, nil
}

// NewWorktree is kept for backwards compatibility but now uses SmartNewWorktree
// NOTE: No backwards compatibility needed per user request - this can be removed
func NewWorktree(branch string, baseBranch string, cfg *config.Manager) (string, error) {
	return SmartNewWorktree(branch, baseBranch, cfg)
}

// CopyEnvFile copies .env files from current directory to the same relative path in target worktree
func CopyEnvFile(targetBranch string, recursive bool) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Get repository root
	repoRoot, err := GetRepoRoot()
	if err != nil {
		return err
	}

	// Get relative path from repo root
	relPath, err := filepath.Rel(repoRoot, currentDir)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %v", err)
	}

	// Find target worktree
	worktrees, err := parseWorktrees()
	if err != nil {
		return err
	}

	var targetPath string
	for _, wt := range worktrees {
		if wt.Branch == targetBranch {
			targetPath = wt.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("worktree '%s' not found", targetBranch)
	}

	// Construct target directory
	targetDir := filepath.Join(targetPath, relPath)

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	if recursive {
		// Copy all .env files recursively
		return copyEnvFilesRecursive(currentDir, targetDir)
	}

	// Copy single .env file
	sourceEnv := filepath.Join(currentDir, ".env")
	targetEnv := filepath.Join(targetDir, ".env")

	if _, err := os.Stat(sourceEnv); os.IsNotExist(err) {
		return fmt.Errorf("no .env file found in current directory")
	}

	return copyFile(sourceEnv, targetEnv)
}

// copyFile copies a single file preserving permissions
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Get file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// Read file content
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to destination with same permissions
	if err := os.WriteFile(dst, content, sourceInfo.Mode()); err != nil {
		return err
	}

	fmt.Printf("Copied %s to %s\n", src, dst)
	return nil
}

// copyEnvFilesRecursive copies all .env files from source to target directory
func copyEnvFilesRecursive(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a .env file
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".env") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Construct target path
		targetPath := filepath.Join(targetDir, relPath)

		// Ensure target directory exists
		targetFileDir := filepath.Dir(targetPath)
		if err := os.MkdirAll(targetFileDir, 0755); err != nil {
			return err
		}

		// Copy the file
		return copyFile(path, targetPath)
	})
}

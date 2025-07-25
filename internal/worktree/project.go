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

	// Ensure target directory exists
	targetDir := filepath.Dir(dst)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
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

		// Skip if not a .env file - must start with ".env" followed by nothing or a dot
		if info.IsDir() {
			return nil
		}

		// Check if it's a valid .env file pattern
		name := info.Name()
		if name != ".env" && !strings.HasPrefix(name, ".env.") {
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

// SyncEnvFiles copies .env files from current directory to target worktree(s)
func SyncEnvFiles(targetBranch string, recursive bool, syncAll bool) error {
	if syncAll {
		return syncEnvToAllWorktrees(recursive)
	}
	return CopyEnvFile(targetBranch, recursive)
}

// syncEnvToAllWorktrees copies current .env files to all other worktrees
func syncEnvToAllWorktrees(recursive bool) error {
	worktrees, err := parseWorktrees()
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	// Determine which worktree we're currently in
	var currentWorktree string
	for _, wt := range worktrees {
		if strings.HasPrefix(currentDir, wt.Path) {
			currentWorktree = wt.Branch
			break
		}
	}

	var syncCount int
	var errors []string

	for _, wt := range worktrees {
		// Skip the current worktree
		if wt.Branch == currentWorktree {
			continue
		}

		err := CopyEnvFile(wt.Branch, recursive)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", wt.Branch, err))
		} else {
			syncCount++
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to sync to some worktrees:\n%s", strings.Join(errors, "\n"))
	}

	fmt.Printf("✓ Synced .env files to %d worktrees\n", syncCount)
	return nil
}

// DiffEnvFiles shows differences between current .env and target worktree
func DiffEnvFiles(targetBranch string) error {
	// Get current directory and .env file
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	repoRoot, err := GetRepoRoot()
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(repoRoot, currentDir)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %v", err)
	}

	sourceEnv := filepath.Join(currentDir, ".env")
	if _, err := os.Stat(sourceEnv); os.IsNotExist(err) {
		return fmt.Errorf("no .env file found in current directory")
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

	targetEnv := filepath.Join(targetPath, relPath, ".env")
	if _, err := os.Stat(targetEnv); os.IsNotExist(err) {
		fmt.Printf("Target worktree '%s' has no .env file at %s\n", targetBranch, relPath)
		return nil
	}

	// Read both files
	sourceContent, err := os.ReadFile(sourceEnv)
	if err != nil {
		return fmt.Errorf("failed to read source .env: %v", err)
	}

	targetContent, err := os.ReadFile(targetEnv)
	if err != nil {
		return fmt.Errorf("failed to read target .env: %v", err)
	}

	// Simple diff output
	if string(sourceContent) == string(targetContent) {
		fmt.Printf("✓ .env files are identical\n")
		return nil
	}

	fmt.Printf("Differences between current .env and %s worktree:\n", targetBranch)
	fmt.Printf("Current (%s):\n", relPath)
	fmt.Printf("%s\n", sourceContent)
	fmt.Printf("\nTarget (%s):\n", targetBranch)
	fmt.Printf("%s\n", targetContent)

	return nil
}

// ListEnvFiles shows all .env files across all worktrees
func ListEnvFiles() error {
	worktrees, err := parseWorktrees()
	if err != nil {
		return err
	}

	fmt.Printf("%-20s %-30s %s\n", "Worktree", "Path", "Status")
	fmt.Printf("%-20s %-30s %s\n", "--------", "----", "------")

	for _, wt := range worktrees {
		// Look for .env files in this worktree
		envPath := filepath.Join(wt.Path, ".env")
		if info, err := os.Stat(envPath); err == nil {
			size := info.Size()
			fmt.Printf("%-20s %-30s %d bytes\n", wt.Branch, ".env", size)
		}

		// Look for .env files recursively
		err := filepath.Walk(wt.Path, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return nil // Skip errors, continue walking
			}

			if info.IsDir() || !strings.HasSuffix(info.Name(), ".env") {
				return nil
			}

			// Skip the root .env we already handled
			if path == envPath {
				return nil
			}

			relPath, err := filepath.Rel(wt.Path, path)
			if err != nil {
				return nil
			}

			size := info.Size()
			fmt.Printf("%-20s %-30s %d bytes\n", wt.Branch, relPath, size)
			return nil
		})

		if err != nil {
			fmt.Printf("%-20s %-30s error: %v\n", wt.Branch, "N/A", err)
		}
	}

	return nil
}

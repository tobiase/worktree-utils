package worktree

import (
	"fmt"
	"path/filepath"

	"github.com/tobiase/worktree-utils/internal/config"
	"github.com/tobiase/worktree-utils/internal/filesystem"
	"github.com/tobiase/worktree-utils/internal/git"
	"github.com/tobiase/worktree-utils/internal/shell"
)

// Service provides worktree operations using injected dependencies
type Service struct {
	git   git.Client
	fs    filesystem.Filesystem
	shell shell.Shell
}

// NewService creates a new worktree service with the given dependencies
func NewService(git git.Client, fs filesystem.Filesystem, shell shell.Shell) *Service {
	return &Service{
		git:   git,
		fs:    fs,
		shell: shell,
	}
}

// GetRepoRoot returns the root directory of the git repository
func (s *Service) GetRepoRoot() (string, error) {
	output, err := s.git.RevParse("--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("not inside a Git repository")
	}
	return output, nil
}

// GetWorktreeBase returns the base directory for worktrees
func (s *Service) GetWorktreeBase() (string, error) {
	repo, err := s.GetRepoRoot()
	if err != nil {
		return "", err
	}

	repoName := filepath.Base(repo)
	repoParent := filepath.Dir(repo)
	return filepath.Join(repoParent, repoName+"-worktrees"), nil
}

// List displays all worktrees
func (s *Service) List() ([]Worktree, error) {
	gitWorktrees, err := s.git.WorktreeList()
	if err != nil {
		return nil, err
	}

	worktrees := make([]Worktree, len(gitWorktrees))
	for i, gw := range gitWorktrees {
		worktrees[i] = Worktree{
			Path:   gw.Path,
			Branch: gw.Branch,
		}
	}

	return worktrees, nil
}

// Add creates a new worktree
func (s *Service) Add(branch string, cfg *config.Manager) error {
	repo, err := s.GetRepoRoot()
	if err != nil {
		return err
	}

	worktreeBase, err := s.GetWorktreeBase()
	if err != nil {
		return err
	}

	// Use project-specific worktree base if configured
	if cfg != nil && cfg.GetCurrentProject() != nil {
		if projectBase := cfg.GetCurrentProject().Settings.WorktreeBase; projectBase != "" {
			worktreeBase = projectBase
		}
	}

	// Create worktree base directory if it doesn't exist
	if err := s.fs.MkdirAll(worktreeBase, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %v", err)
	}

	worktreePath := filepath.Join(worktreeBase, branch)
	if err := s.git.WorktreeAdd(worktreePath, branch); err != nil {
		return err
	}

	// Run setup automation if configured
	if cfg != nil && cfg.GetCurrentProject() != nil && cfg.GetCurrentProject().Setup != nil {
		fmt.Printf("Running setup automation for new worktree...\n")
		if err := s.runWorktreeSetup(repo, worktreePath, cfg.GetCurrentProject().Setup); err != nil {
			fmt.Printf("Warning: Setup automation failed: %v\n", err)
			// Don't fail the entire operation, just warn
		}
	}

	return nil
}

// Remove deletes a worktree by branch name or path
func (s *Service) Remove(target string) error {
	// First, try to find the worktree by branch name
	worktrees, err := s.List()
	if err != nil {
		return err
	}

	var worktreePath string
	for _, wt := range worktrees {
		if wt.Branch == target {
			worktreePath = wt.Path
			break
		}
	}

	// If not found by branch name, check if target is a path
	if worktreePath == "" {
		// Check if target is an absolute path that exists
		if filepath.IsAbs(target) {
			if s.fs.Exists(target) {
				worktreePath = target
			}
		} else {
			// Try as a relative path from repo root
			repo, err := s.GetRepoRoot()
			if err == nil {
				testPath := filepath.Join(repo, target)
				if s.fs.Exists(testPath) {
					worktreePath = testPath
				}
			}
		}
	}

	if worktreePath == "" {
		return fmt.Errorf("worktree '%s' not found", target)
	}

	return s.git.WorktreeRemove(worktreePath)
}

// CheckBranchExists checks if a branch exists in the repository
func (s *Service) CheckBranchExists(branch string) bool {
	err := s.git.ShowRef("refs/heads/" + branch)
	return err == nil
}

// CheckWorktreeExists checks if a worktree exists for the given branch
func (s *Service) CheckWorktreeExists(branch string) bool {
	worktrees, err := s.List()
	if err != nil {
		return false
	}

	for _, wt := range worktrees {
		if wt.Branch == branch {
			return true
		}
	}
	return false
}

// GetAvailableBranches returns a list of branch names from existing worktrees
func (s *Service) GetAvailableBranches() ([]string, error) {
	worktrees, err := s.List()
	if err != nil {
		return nil, err
	}

	branches := make([]string, len(worktrees))
	for i, wt := range worktrees {
		branches[i] = wt.Branch
	}

	return branches, nil
}

// runWorktreeSetup executes setup automation for a newly created worktree
func (s *Service) runWorktreeSetup(repoRoot, worktreePath string, setup *config.SetupConfig) error {
	if setup == nil {
		return nil
	}

	// Create directories
	for _, dir := range setup.CreateDirectories {
		dirPath := filepath.Join(worktreePath, dir)
		if err := s.fs.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
		fmt.Printf("Created directory: %s\n", dir)
	}

	// Copy files
	for _, copyFile := range setup.CopyFiles {
		sourcePath := filepath.Join(repoRoot, copyFile.Source)
		targetPath := filepath.Join(worktreePath, copyFile.Target)

		// Ensure target directory exists
		targetDir := filepath.Dir(targetPath)
		if err := s.fs.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory for %s: %v", copyFile.Target, err)
		}

		// Check if source file exists
		if !s.fs.Exists(sourcePath) {
			fmt.Printf("Warning: Source file %s not found, skipping\n", copyFile.Source)
			continue
		}

		// Copy the file
		if err := s.copyFile(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %v", copyFile.Source, copyFile.Target, err)
		}
		fmt.Printf("Copied: %s → %s\n", copyFile.Source, copyFile.Target)
	}

	// Run commands
	for _, cmdConfig := range setup.Commands {
		cmdDir := filepath.Join(worktreePath, cmdConfig.Directory)

		// Ensure command directory exists
		if !s.fs.Exists(cmdDir) {
			fmt.Printf("Warning: Command directory %s not found, skipping command: %s\n",
				cmdConfig.Directory, cmdConfig.Command)
			continue
		}

		fmt.Printf("Running: %s (in %s)\n", cmdConfig.Command, cmdConfig.Directory)

		// Execute command through shell
		if err := s.shell.RunInDir(cmdDir, "sh", "-c", cmdConfig.Command); err != nil {
			return fmt.Errorf("command failed in %s: %s (%v)",
				cmdConfig.Directory, cmdConfig.Command, err)
		}
	}

	return nil
}

// copyFile copies a single file preserving permissions
func (s *Service) copyFile(src, dst string) error {
	// Get file info for permissions
	info, err := s.fs.Stat(src)
	if err != nil {
		return err
	}

	// Read file content
	content, err := s.fs.ReadFile(src)
	if err != nil {
		return err
	}

	// Write to destination with same permissions
	return s.fs.WriteFile(dst, content, info.Mode())
}

// GetGitRemote returns the origin remote URL for the current repository
func (s *Service) GetGitRemote() (string, error) {
	remote, err := s.git.GetRemoteURL("origin")
	if err != nil {
		return "", nil // No remote is OK
	}
	return remote, nil
}

// GetRelativePath returns the path relative to the repository root
func (s *Service) GetRelativePath(absolutePath string) (string, error) {
	repoRoot, err := s.GetRepoRoot()
	if err != nil {
		return "", err
	}

	relPath, err := filepath.Rel(repoRoot, absolutePath)
	if err != nil {
		return "", err
	}

	return relPath, nil
}

// Go returns the path to change to based on index or branch name
func (s *Service) Go(target string) (string, error) {
	worktrees, err := s.List()
	if err != nil {
		return "", err
	}

	if len(worktrees) == 0 {
		return "", fmt.Errorf("no worktrees exist")
	}

	// Try to parse as index first
	if index := parseIndex(target); index >= 0 {
		if index < len(worktrees) {
			return worktrees[index].Path, nil
		}
		return "", fmt.Errorf("index %d out of range (0..%d)", index, len(worktrees)-1)
	}

	// Try to match by branch name
	for _, wt := range worktrees {
		if wt.Branch == target {
			return wt.Path, nil
		}
	}

	return "", fmt.Errorf("branch '%s' not found among worktrees", target)
}

// parseIndex attempts to parse a string as an integer index
func parseIndex(s string) int {
	var index int
	if _, err := fmt.Sscanf(s, "%d", &index); err == nil {
		return index
	}
	return -1
}

// ResolveBranchName delegates to the package-level ResolveBranchName function
func (s *Service) ResolveBranchName(input string, availableBranches []string) (string, error) {
	return ResolveBranchName(input, availableBranches)
}

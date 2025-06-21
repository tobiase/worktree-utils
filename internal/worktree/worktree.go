package worktree

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/tobiase/worktree-utils/internal/config"
)

type Worktree struct {
	Path   string
	Branch string
}

// GetRepoRoot returns the root directory of the git repository
func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not inside a Git repository")
	}
	return strings.TrimSpace(string(output)), nil
}

// GetWorktreeBase returns the base directory for worktrees
func GetWorktreeBase() (string, error) {
	repo, err := GetRepoRoot()
	if err != nil {
		return "", err
	}

	repoName := filepath.Base(repo)
	repoParent := filepath.Dir(repo)
	return filepath.Join(repoParent, repoName+"-worktrees"), nil
}

// parseWorktrees parses git worktree list output
func parseWorktrees() ([]Worktree, error) {
	repo, err := GetRepoRoot()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "-C", repo, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %v", err)
	}

	var worktrees []Worktree
	var currentPath string

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch refs/heads/") {
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			if currentPath != "" && branch != "" {
				worktrees = append(worktrees, Worktree{
					Path:   currentPath,
					Branch: branch,
				})
			}
		} else if currentPath != "" && line == "" {
			// Empty line indicates end of worktree entry - skip worktrees without proper branch info
			currentPath = ""
		}
	}

	return worktrees, scanner.Err()
}

// List displays all worktrees
func List() error {
	worktrees, err := parseWorktrees()
	if err != nil {
		return err
	}

	if len(worktrees) == 0 {
		fmt.Println("wt: no worktrees found.")
		return nil
	}

	fmt.Printf("%-5s %-20s %s\n", "Index", "Branch", "Path")
	for i, wt := range worktrees {
		fmt.Printf("%-5d %-20s %s\n", i, wt.Branch, wt.Path)
	}

	return nil
}

// Add creates a new worktree
func Add(branch string, cfg *config.Manager) error {
	repo, err := GetRepoRoot()
	if err != nil {
		return err
	}

	worktreeBase, err := GetWorktreeBase()
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
	if err := os.MkdirAll(worktreeBase, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %v", err)
	}

	worktreePath := filepath.Join(worktreeBase, branch)
	cmd := exec.Command("git", "-C", repo, "worktree", "add", worktreePath, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// Run setup automation if configured
	if cfg != nil && cfg.GetCurrentProject() != nil && cfg.GetCurrentProject().Setup != nil {
		fmt.Printf("Running setup automation for new worktree...\n")
		if err := runWorktreeSetup(repo, worktreePath, cfg.GetCurrentProject().Setup); err != nil {
			fmt.Printf("Warning: Setup automation failed: %v\n", err)
			// Don't fail the entire operation, just warn
		}
	}

	return nil
}

// runWorktreeSetup executes setup automation for a newly created worktree
func runWorktreeSetup(repoRoot, worktreePath string, setup *config.SetupConfig) error {
	// Create directories
	for _, dir := range setup.CreateDirectories {
		dirPath := filepath.Join(worktreePath, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
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
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create target directory for %s: %v", copyFile.Target, err)
		}

		// Check if source file exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			fmt.Printf("Warning: Source file %s not found, skipping\n", copyFile.Source)
			continue
		}

		// Copy the file
		if err := copyFileForSetup(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %v", copyFile.Source, copyFile.Target, err)
		}
		fmt.Printf("Copied: %s → %s\n", copyFile.Source, copyFile.Target)
	}

	// Run commands
	for _, cmdConfig := range setup.Commands {
		cmdDir := filepath.Join(worktreePath, cmdConfig.Directory)

		// Ensure command directory exists
		if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
			fmt.Printf("Warning: Command directory %s not found, skipping command: %s\n", cmdConfig.Directory, cmdConfig.Command)
			continue
		}

		fmt.Printf("Running: %s (in %s)\n", cmdConfig.Command, cmdConfig.Directory)

		// Execute command through shell for proper handling of complex commands
		// The command string is passed as a single argument to the shell, preventing injection
		// of additional commands through shell metacharacters in the configuration
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", cmdConfig.Command)
		} else {
			// Use sh -c to execute the command as a single shell argument
			// This is safe because cmdConfig.Command comes from a config file,
			// not from user input at runtime
			cmd = exec.Command("sh", "-c", cmdConfig.Command)
		}
		cmd.Dir = cmdDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed in %s: %s (%v)", cmdConfig.Directory, cmdConfig.Command, err)
		}
	}

	return nil
}

// copyFileForSetup copies a single file for setup automation
func copyFileForSetup(src, dst string) error {
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
	return os.WriteFile(dst, content, sourceInfo.Mode())
}

// RunSetup executes setup automation for an existing worktree
func RunSetup(repoRoot, worktreePath string, setup *config.SetupConfig) error {
	return runWorktreeSetup(repoRoot, worktreePath, setup)
}

// Remove deletes a worktree by branch name or path
func Remove(target string) error {
	repo, err := GetRepoRoot()
	if err != nil {
		return err
	}

	// First, try to find the worktree by branch name
	worktrees, err := parseWorktrees()
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
			if _, err := os.Stat(target); err == nil {
				worktreePath = target
			}
		} else {
			// Try as a relative path from repo root
			testPath := filepath.Join(repo, target)
			if _, err := os.Stat(testPath); err == nil {
				worktreePath = testPath
			}
		}
	}

	if worktreePath == "" {
		return fmt.Errorf("worktree '%s' not found", target)
	}

	cmd := exec.Command("git", "-C", repo, "worktree", "remove", worktreePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetAvailableBranches returns a list of branch names from existing worktrees
func GetAvailableBranches() ([]string, error) {
	worktrees, err := parseWorktrees()
	if err != nil {
		return nil, err
	}

	branches := make([]string, len(worktrees))
	for i, wt := range worktrees {
		branches[i] = wt.Branch
	}

	return branches, nil
}

// GetWorktreeInfo returns worktree information for interactive selection
func GetWorktreeInfo() ([]Worktree, error) {
	return parseWorktrees()
}

// Go returns the path to change to based on index or branch name
func Go(target string) (string, error) {
	worktrees, err := parseWorktrees()
	if err != nil {
		return "", err
	}

	if len(worktrees) == 0 {
		return "", fmt.Errorf("no worktrees exist")
	}

	// Try to parse as index first
	if index, err := strconv.Atoi(target); err == nil {
		if index >= 0 && index < len(worktrees) {
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

// checkBranchExists checks if a branch exists in the repository
func checkBranchExists(branch string) bool {
	repo, err := GetRepoRoot()
	if err != nil {
		return false
	}

	cmd := exec.Command("git", "-C", repo, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	return cmd.Run() == nil
}

// checkWorktreeExists checks if a worktree exists for the given branch
func checkWorktreeExists(branch string) bool {
	worktrees, err := parseWorktrees()
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

// ResolveBranchName attempts to resolve a partial branch name to a full branch name
// Returns the resolved name, or original input if no unique match found
func ResolveBranchName(input string, availableBranches []string) (string, error) {
	// Exact match first
	for _, branch := range availableBranches {
		if branch == input {
			return branch, nil
		}
	}

	// Find matches using different strategies
	var exactMatches []string
	var prefixMatches []string
	var containsMatches []string

	inputLower := strings.ToLower(input)

	for _, branch := range availableBranches {
		branchLower := strings.ToLower(branch)

		if branchLower == inputLower {
			exactMatches = append(exactMatches, branch)
		} else if strings.HasPrefix(branchLower, inputLower) {
			prefixMatches = append(prefixMatches, branch)
		} else if strings.Contains(branchLower, inputLower) {
			containsMatches = append(containsMatches, branch)
		}
	}

	// Return single exact match
	if len(exactMatches) == 1 {
		return exactMatches[0], nil
	}

	// Return single prefix match
	if len(prefixMatches) == 1 {
		return prefixMatches[0], nil
	}

	// If multiple matches, return error with suggestions
	allMatches := append(exactMatches, prefixMatches...)
	allMatches = append(allMatches, containsMatches...)

	if len(allMatches) == 0 {
		return "", fmt.Errorf("branch '%s' not found", input)
	}

	if len(allMatches) > 1 {
		suggestions := suggestSimilarBranches(input, allMatches)
		return "", fmt.Errorf("branch '%s' is ambiguous. Did you mean:\n%s", input, strings.Join(suggestions, "\n"))
	}

	return allMatches[0], nil
}

// suggestSimilarBranches returns formatted suggestions for similar branch names
func suggestSimilarBranches(input string, candidates []string) []string {
	// Sort by relevance (shorter branches first, then alphabetical)
	sort.Slice(candidates, func(i, j int) bool {
		lenI, lenJ := len(candidates[i]), len(candidates[j])
		if lenI == lenJ {
			return candidates[i] < candidates[j]
		}
		return lenI < lenJ
	})

	// Format as numbered list (limit to 5 suggestions)
	suggestions := make([]string, 0, min(5, len(candidates)))
	for i, branch := range candidates {
		if i >= 5 {
			break
		}
		suggestions = append(suggestions, fmt.Sprintf("  %d) %s", i+1, branch))
	}

	return suggestions
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

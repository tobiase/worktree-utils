package worktree

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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

// getWorktreeBase returns the base directory for worktrees
func getWorktreeBase() (string, error) {
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
			worktrees = append(worktrees, Worktree{
				Path:   currentPath,
				Branch: branch,
			})
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
func Add(branch string) error {
	repo, err := GetRepoRoot()
	if err != nil {
		return err
	}
	
	worktreeBase, err := getWorktreeBase()
	if err != nil {
		return err
	}
	
	// Create worktree base directory if it doesn't exist
	if err := os.MkdirAll(worktreeBase, 0755); err != nil {
		return fmt.Errorf("failed to create worktree directory: %v", err)
	}
	
	worktreePath := filepath.Join(worktreeBase, branch)
	cmd := exec.Command("git", "-C", repo, "worktree", "add", worktreePath, branch)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}

// Remove deletes a worktree
func Remove(branch string) error {
	repo, err := GetRepoRoot()
	if err != nil {
		return err
	}
	
	worktreeBase, err := getWorktreeBase()
	if err != nil {
		return err
	}
	
	worktreePath := filepath.Join(worktreeBase, branch)
	cmd := exec.Command("git", "-C", repo, "worktree", "remove", worktreePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
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
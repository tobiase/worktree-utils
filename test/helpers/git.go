package helpers

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// runGitCommand runs a git command with consistent locale
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// Set LANG=C to ensure consistent output across locales
	cmd.Env = append(os.Environ(), "LANG=C")
	return cmd.Run()
}

// CreateTestRepo creates a temporary git repository for testing
func CreateTestRepo(t *testing.T) (repoPath string, cleanup func()) {
	t.Helper()

	// Create temp directory
	tempDir, err := os.MkdirTemp("", "wt-test-repo-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo with main as default branch
	if err := runGitCommand(tempDir, "init", "--initial-branch=main"); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user for commits
	if err := runGitCommand(tempDir, "config", "user.name", "Test User"); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to configure git user.name: %v", err)
	}

	if err := runGitCommand(tempDir, "config", "user.email", "test@example.com"); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to configure git user.email: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test Repo\n"), 0644); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create test file: %v", err)
	}

	if err := runGitCommand(tempDir, "add", "."); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to stage files: %v", err)
	}

	if err := runGitCommand(tempDir, "commit", "-m", "Initial commit"); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to create initial commit: %v", err)
	}

	cleanup = func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// AddTestWorktree creates a worktree in the expected location
func AddTestWorktree(t *testing.T, repoPath, branch string) (worktreePath string, err error) {
	t.Helper()

	// Get repo name and parent
	repoName := filepath.Base(repoPath)
	repoParent := filepath.Dir(repoPath)
	worktreeBase := filepath.Join(repoParent, repoName+"-worktrees")

	// Create worktree base directory
	if err := os.MkdirAll(worktreeBase, 0755); err != nil {
		return "", err
	}

	// Create worktree
	worktreePath = filepath.Join(worktreeBase, branch)
	cmd := exec.Command("git", "-C", repoPath, "worktree", "add", "-b", branch, worktreePath)
	// Set LANG=C to ensure consistent output across locales
	cmd.Env = append(os.Environ(), "LANG=C")
	if err := cmd.Run(); err != nil {
		return "", err
	}

	return worktreePath, nil
}

// CreateBareRepo creates a bare git repository for testing
func CreateBareRepo(t *testing.T) (repoPath string, cleanup func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "wt-test-bare-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	if err := runGitCommand(tempDir, "init", "--bare", "--initial-branch=main"); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to init bare repo: %v", err)
	}

	cleanup = func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

// GetGitOutput runs a git command and returns the output
func GetGitOutput(t *testing.T, repoPath string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	// Set LANG=C to ensure consistent output across locales
	cmd.Env = append(os.Environ(), "LANG=C")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Git command failed: %v", err)
	}
	return string(output)
}

// CreateTestBranch creates a new branch with a commit
func CreateTestBranch(t *testing.T, repoPath, branch string) {
	t.Helper()

	// Create and checkout branch
	cmd := exec.Command("git", "-C", repoPath, "checkout", "-b", branch)
	// Set LANG=C to ensure consistent output across locales
	cmd.Env = append(os.Environ(), "LANG=C")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create branch %s: %v", branch, err)
	}

	// Create a unique file for this branch
	testFile := filepath.Join(repoPath, branch+".txt")
	if err := os.WriteFile(testFile, []byte("Content for "+branch), 0644); err != nil {
		t.Fatalf("Failed to create branch file: %v", err)
	}

	// Commit the change
	cmd = exec.Command("git", "-C", repoPath, "add", ".")
	// Set LANG=C to ensure consistent output across locales
	cmd.Env = append(os.Environ(), "LANG=C")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to stage branch file: %v", err)
	}

	cmd = exec.Command("git", "-C", repoPath, "commit", "-m", "Add "+branch+" file")
	// Set LANG=C to ensure consistent output across locales
	cmd.Env = append(os.Environ(), "LANG=C")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to commit branch file: %v", err)
	}
}

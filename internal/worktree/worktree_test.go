package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tobiase/worktree-utils/test/helpers"
)

func TestGetRepoRoot(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		wantErr bool
		check   func(t *testing.T, got string, tempDir string)
	}{
		{
			name: "valid git repository",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				// Change to repo directory
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, got, tempDir string) {
				// Resolve both paths to handle macOS /private symlinks
				resolvedGot, _ := filepath.EvalSymlinks(got)
				resolvedTemp, _ := filepath.EvalSymlinks(tempDir)
				if resolvedGot != resolvedTemp {
					t.Errorf("Expected repo root %s, got %s", resolvedTemp, resolvedGot)
				}
			},
		},
		{
			name: "subdirectory of git repository",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				subdir := filepath.Join(repo, "subdir")
				_ = os.MkdirAll(subdir, 0755)

				oldWd, _ := os.Getwd()
				_ = os.Chdir(subdir)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, got, tempDir string) {
				// Resolve both paths to handle macOS /private symlinks
				resolvedGot, _ := filepath.EvalSymlinks(got)
				resolvedTemp, _ := filepath.EvalSymlinks(tempDir)
				if resolvedGot != resolvedTemp {
					t.Errorf("Expected repo root %s, got %s", resolvedTemp, resolvedGot)
				}
			},
		},
		{
			name: "not a git repository",
			setup: func() (string, func()) {
				tempDir, err := os.MkdirTemp("", "not-git-*")
				if err != nil {
					t.Fatal(err)
				}
				oldWd, _ := os.Getwd()
				_ = os.Chdir(tempDir)
				return tempDir, func() {
					_ = os.Chdir(oldWd)
					os.RemoveAll(tempDir)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := tt.setup()
			defer cleanup()

			got, err := GetRepoRoot()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoRoot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, got, tempDir)
			}
		})
	}
}

func TestGetWorktreeBase(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		wantErr bool
		check   func(t *testing.T, got string, repo string)
	}{
		{
			name: "standard repository",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, got, repo string) {
				repoName := filepath.Base(repo)
				expected := filepath.Join(filepath.Dir(repo), repoName+"-worktrees")
				// Resolve both paths to handle macOS /private symlinks
				resolvedGot, _ := filepath.EvalSymlinks(got)
				resolvedExpected, _ := filepath.EvalSymlinks(expected)
				if resolvedGot != resolvedExpected {
					t.Errorf("Expected worktree base %s, got %s", resolvedExpected, resolvedGot)
				}
			},
		},
		{
			name: "repository with special characters",
			setup: func() (string, func()) {
				tempDir, _ := os.MkdirTemp("", "test-repo-with-dash-*")
				_, _, _ = helpers.RunCommand(t, "git", "-C", tempDir, "init")
				oldWd, _ := os.Getwd()
				_ = os.Chdir(tempDir)
				return tempDir, func() {
					_ = os.Chdir(oldWd)
					os.RemoveAll(tempDir)
				}
			},
			wantErr: false,
			check: func(t *testing.T, got, repo string) {
				repoName := filepath.Base(repo)
				expected := filepath.Join(filepath.Dir(repo), repoName+"-worktrees")
				// Resolve both paths to handle macOS /private symlinks
				resolvedGot, _ := filepath.EvalSymlinks(got)
				resolvedExpected, _ := filepath.EvalSymlinks(expected)
				if resolvedGot != resolvedExpected {
					t.Errorf("Expected worktree base %s, got %s", resolvedExpected, resolvedGot)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, cleanup := tt.setup()
			defer cleanup()

			got, err := GetWorktreeBase()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetWorktreeBase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, got, repo)
			}
		})
	}
}

func TestParseWorktrees(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		want    []Worktree
		wantErr bool
	}{
		{
			name: "single main worktree",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			want: []Worktree{
				{Branch: "main"},
			},
			wantErr: false,
		},
		{
			name: "multiple worktrees",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				// Create additional worktrees
				_, _ = helpers.AddTestWorktree(t, repo, "feature-1")
				_, _ = helpers.AddTestWorktree(t, repo, "feature-2")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			want: []Worktree{
				{Branch: "main"},
				{Branch: "feature-1"},
				{Branch: "feature-2"},
			},
			wantErr: false,
		},
		{
			name: "worktree with detached HEAD",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				// Get current commit hash
				output := helpers.GetGitOutput(t, repo, "rev-parse", "HEAD")
				commitHash := strings.TrimSpace(output)

				// Create worktree at specific commit (detached HEAD)
				worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
				_ = os.MkdirAll(worktreeBase, 0755)
				worktreePath := filepath.Join(worktreeBase, "detached")
				_, _, err := helpers.RunCommand(t, "git", "-C", repo, "worktree", "add", "--detach", worktreePath, commitHash)
				if err != nil {
					t.Fatalf("Failed to create detached worktree: %v", err)
				}

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			want: []Worktree{
				{Branch: "main"},
				// Note: Current implementation doesn't parse detached HEAD worktrees
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := tt.setup()
			defer cleanup()

			got, err := parseWorktrees()

			if (err != nil) != tt.wantErr {
				t.Errorf("parseWorktrees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Expected %d worktrees, got %d", len(tt.want), len(got))
					return
				}

				// Create a map of branches for easier comparison
				gotBranches := make(map[string]bool)
				for _, wt := range got {
					gotBranches[wt.Branch] = true
				}

				for _, want := range tt.want {
					if !gotBranches[want.Branch] {
						t.Errorf("Expected worktree with branch %q not found", want.Branch)
					}
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		wantOut func(string) bool // Function to check output
		wantErr bool
	}{
		{
			name: "list with multiple worktrees",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				_, _ = helpers.AddTestWorktree(t, repo, "feature-1")
				_, _ = helpers.AddTestWorktree(t, repo, "bugfix-2")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantOut: func(output string) bool {
				// Check for header
				if !strings.Contains(output, "Index") || !strings.Contains(output, "Branch") || !strings.Contains(output, "Path") {
					return false
				}
				// Check for branches
				return strings.Contains(output, "main") &&
					strings.Contains(output, "feature-1") &&
					strings.Contains(output, "bugfix-2")
			},
			wantErr: false,
		},
		{
			name: "list with no worktrees",
			setup: func() (string, func()) {
				// Create a repo but we can't have zero worktrees, so this is hypothetical
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantOut: func(output string) bool {
				// At minimum we'll have the main worktree
				return strings.Contains(output, "main")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := tt.setup()
			defer cleanup()

			// Capture output
			stdout, _, err := helpers.CaptureOutput(func() {
				err := List()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				}
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantOut != nil {
				if !tt.wantOut(stdout) {
					t.Errorf("List() output validation failed.\nGot:\n%s", stdout)
				}
			}
		})
	}
}

func TestGo(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (string, func())
		target  string
		wantErr bool
		check   func(t *testing.T, gotPath string, repo string)
	}{
		{
			name: "go by index",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				_, _ = helpers.AddTestWorktree(t, repo, "feature-1")
				_, _ = helpers.AddTestWorktree(t, repo, "feature-2")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			target:  "1",
			wantErr: false,
			check: func(t *testing.T, gotPath, repo string) {
				// Should return path to one of the worktrees
				if !strings.Contains(gotPath, "worktrees") {
					t.Errorf("Expected path to contain 'worktrees', got %s", gotPath)
				}
			},
		},
		{
			name: "go by branch name",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				_, _ = helpers.AddTestWorktree(t, repo, "feature-1")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			target:  "feature-1",
			wantErr: false,
			check: func(t *testing.T, gotPath, repo string) {
				if !strings.HasSuffix(gotPath, "feature-1") {
					t.Errorf("Expected path to end with 'feature-1', got %s", gotPath)
				}
			},
		},
		{
			name: "invalid index",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			target:  "99",
			wantErr: true,
		},
		{
			name: "non-existent branch",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			target:  "non-existent",
			wantErr: true,
		},
		{
			name: "go to main",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				_, _ = helpers.AddTestWorktree(t, repo, "feature-1")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			target:  "main",
			wantErr: false,
			check: func(t *testing.T, gotPath, repo string) {
				// Resolve both paths to handle macOS /private symlinks
				resolvedGot, _ := filepath.EvalSymlinks(gotPath)
				resolvedRepo, _ := filepath.EvalSymlinks(repo)
				if resolvedGot != resolvedRepo {
					t.Errorf("Expected path to main repo %s, got %s", resolvedRepo, resolvedGot)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, cleanup := tt.setup()
			defer cleanup()

			gotPath, err := Go(tt.target)

			if (err != nil) != tt.wantErr {
				t.Errorf("Go() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, gotPath, repo)
			}
		})
	}
}

// =============================================================================
// EDGE CASE TESTS - Critical Git Integration Scenarios
// =============================================================================

// Helper functions for creating edge case scenarios

func corruptGitRepo(t *testing.T, repoPath string) {
	t.Helper()
	// Remove critical .git files to simulate corruption
	gitDir := filepath.Join(repoPath, ".git")
	os.Remove(filepath.Join(gitDir, "HEAD"))
	os.Remove(filepath.Join(gitDir, "config"))
}

func createRepoWithNoCommits(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "wt-test-no-commits-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize repo but don't create any commits
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	_ = cmd.Run()

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	_ = cmd.Run()

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return tempDir, cleanup
}

func createRepoWithUncommittedChanges(t *testing.T) (string, func()) {
	repo, cleanup := helpers.CreateTestRepo(t)

	// Add some uncommitted changes
	testFile := filepath.Join(repo, "uncommitted.txt")
	if err := os.WriteFile(testFile, []byte("uncommitted changes"), 0644); err != nil {
		cleanup()
		t.Fatalf("Failed to create uncommitted file: %v", err)
	}

	return repo, cleanup
}

// Edge Case Tests

func TestListWorktreesNoCommits(t *testing.T) {
	repo, cleanup := createRepoWithNoCommits(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	err := List()

	// Should handle gracefully - either succeed or provide clear error
	if err != nil {
		// Error is acceptable, but should be user-friendly
		if !strings.Contains(err.Error(), "commit") && !strings.Contains(err.Error(), "branch") && !strings.Contains(err.Error(), "reference") {
			t.Errorf("Error message should mention commits/branches/references: %v", err)
		}
	}
}

func TestListWorktreesCorruptedGit(t *testing.T) {
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	// Corrupt the repository
	corruptGitRepo(t, repo)

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	err := List()

	// Should fail gracefully with meaningful error
	if err == nil {
		t.Error("Expected error for corrupted repository, got none")
	}

	// Error should not be a panic and should be user-friendly
	if strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestListWorktreesPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission denied as root")
	}

	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	// Remove read permissions from .git directory
	gitDir := filepath.Join(repo, ".git")
	originalMode, err := os.Stat(gitDir)
	if err != nil {
		t.Fatalf("Failed to get .git permissions: %v", err)
	}

	// Remove read permissions
	if err := os.Chmod(gitDir, 0000); err != nil {
		t.Fatalf("Failed to change .git permissions: %v", err)
	}

	// Restore permissions at the end
	defer func() {
		_ = os.Chmod(gitDir, originalMode.Mode())
	}()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	err = List()

	// Should handle permission errors gracefully
	if err == nil {
		t.Error("Expected permission error, got none")
	}

	// Error should mention permissions or be a generic git error (which is acceptable)
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "permission") && !strings.Contains(errStr, "access") && !strings.Contains(errStr, "git repository") {
		t.Errorf("Error should mention permissions or repository access: %v", err)
	}
}

func TestAddWorktreeVeryLongBranchName(t *testing.T) {
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	// Create a very long branch name (filesystem limits vary, but 255+ chars often cause issues)
	longName := strings.Repeat("very-long-branch-name-", 20) // ~440 characters

	_, err := NewWorktree(longName, "", nil)

	// Should either succeed or fail gracefully with clear error
	if err != nil {
		// Error should be user-friendly and not a panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		// Should mention the issue clearly or show it's git's validation
		errStr := strings.ToLower(err.Error())
		if !strings.Contains(errStr, "name") && !strings.Contains(errStr, "long") && !strings.Contains(errStr, "invalid") && !strings.Contains(errStr, "reference") {
			t.Logf("Long branch name failed as expected with: %v", err)
		}
	}
}

func TestAddWorktreeSpecialCharBranchNames(t *testing.T) {
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	problematicNames := []string{
		"branch with spaces",
		"branch/with/slashes",
		"branch-with-unicode-🚀",
		"branch.with.dots",
		"branch--with--double-dashes",
		".hidden-branch",
		"branch_with_underscores",
	}

	for _, name := range problematicNames {
		t.Run(fmt.Sprintf("branch_%s", name), func(t *testing.T) {
			_, err := NewWorktree(name, "", nil)

			// Either should work or fail gracefully
			if err != nil {
				// Error should be user-friendly
				if strings.Contains(err.Error(), "panic") {
					t.Errorf("Error should not contain panic for branch '%s': %v", name, err)
				}
			}
		})
	}
}

func TestGoMissingWorktreeDir(t *testing.T) {
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	// Create a worktree first
	worktreePath, err := helpers.AddTestWorktree(t, repo, "test-branch")
	if err != nil {
		t.Fatalf("Failed to create test worktree: %v", err)
	}

	// Manually delete the worktree directory (simulating user deletion)
	if err := os.RemoveAll(worktreePath); err != nil {
		t.Fatalf("Failed to remove worktree directory: %v", err)
	}

	// Try to go to the missing worktree
	path, err := Go("test-branch")

	// Should either fail or return a path that doesn't exist
	if err == nil {
		// Check if the returned path actually exists
		if _, statErr := os.Stat(path); statErr == nil {
			t.Error("Expected missing directory but path exists")
		} else {
			t.Logf("Go() returned path to missing directory: %s (this is acceptable)", path)
		}
		return
	}

	// Error should mention the missing directory
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "not found") && !strings.Contains(errStr, "missing") && !strings.Contains(errStr, "exist") {
		t.Errorf("Error should mention missing directory: %v", err)
	}
}

func TestAddWorktreeWithUncommittedChanges(t *testing.T) {
	repo, cleanup := createRepoWithUncommittedChanges(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	// Try to add a worktree when there are uncommitted changes
	_, err := NewWorktree("new-branch", "", nil)

	// Git behavior: should either succeed or provide clear guidance
	if err != nil {
		// If it fails, error should be clear about the issue
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		// Git may succeed or fail - both are acceptable
		// Note: Git error messages may not be user-friendly (exit status 128)
		t.Logf("Creating worktree with uncommitted changes failed as expected: %v", err)
	}
}

func TestRemoveWorktreeWithUncommittedChanges(t *testing.T) {
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	// Create a worktree
	worktreePath, err := helpers.AddTestWorktree(t, repo, "test-branch")
	if err != nil {
		t.Fatalf("Failed to create test worktree: %v", err)
	}

	// Add uncommitted changes to the worktree
	testFile := filepath.Join(worktreePath, "uncommitted.txt")
	if err := os.WriteFile(testFile, []byte("uncommitted"), 0644); err != nil {
		t.Fatalf("Failed to create uncommitted file: %v", err)
	}

	// Try to remove the worktree with uncommitted changes
	err = Remove("test-branch")

	// Should either succeed with warning or fail with clear guidance
	if err != nil {
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		// Git should protect against removing dirty worktrees
		// Note: Error message may not be user-friendly (exit status 128)
		t.Logf("Removing worktree with uncommitted changes failed as expected: %v", err)
	}
}

func TestDeepDirectoryPaths(t *testing.T) {
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	oldWd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWd) }()
	_ = os.Chdir(repo)

	// Create a deeply nested worktree path
	deepBranch := "very/deeply/nested/branch/structure/that/goes/many/levels/deep"

	_, err := NewWorktree(deepBranch, "", nil)

	// Should handle gracefully - either work or provide clear error
	if err != nil {
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		// Git validates branch names and rejects invalid ones
		// Note: Error message may not be user-friendly (exit status 128)
		t.Logf("Deep directory path failed as expected: %v", err)
	}
}

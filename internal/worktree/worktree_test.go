package worktree

import (
	"fmt"
	"os"
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
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				os.MkdirAll(subdir, 0755)

				oldWd, _ := os.Getwd()
				os.Chdir(subdir)
				return repo, func() {
					os.Chdir(oldWd)
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
				os.Chdir(tempDir)
				return tempDir, func() {
					os.Chdir(oldWd)
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
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				helpers.RunCommand(t, "git", "-C", tempDir, "init")
				oldWd, _ := os.Getwd()
				os.Chdir(tempDir)
				return tempDir, func() {
					os.Chdir(oldWd)
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
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				helpers.AddTestWorktree(t, repo, "feature-1")
				helpers.AddTestWorktree(t, repo, "feature-2")

				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				os.MkdirAll(worktreeBase, 0755)
				worktreePath := filepath.Join(worktreeBase, "detached")
				_, _, err := helpers.RunCommand(t, "git", "-C", repo, "worktree", "add", "--detach", worktreePath, commitHash)
				if err != nil {
					t.Fatalf("Failed to create detached worktree: %v", err)
				}

				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				helpers.AddTestWorktree(t, repo, "feature-1")
				helpers.AddTestWorktree(t, repo, "bugfix-2")

				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				helpers.AddTestWorktree(t, repo, "feature-1")
				helpers.AddTestWorktree(t, repo, "feature-2")

				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				helpers.AddTestWorktree(t, repo, "feature-1")

				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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
				helpers.AddTestWorktree(t, repo, "feature-1")

				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
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

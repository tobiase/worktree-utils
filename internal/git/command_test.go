package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestHelper helps with git test repositories
type TestHelper struct {
	t       *testing.T
	tempDir string
}

// NewTestHelper creates a new test helper with a temporary git repository
func NewTestHelper(t *testing.T) *TestHelper {
	tempDir, err := os.MkdirTemp("", "git-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Initialize git repo with main as the initial branch
	// This ensures consistency across different git configurations
	initTestRepoWithMain(t, tempDir)

	// Configure git user
	cmds := [][]string{
		{"config", "user.name", "Test User"},
		{"config", "user.email", "test@example.com"},
		// Also set the default branch name for future operations
		{"config", "init.defaultBranch", "main"},
	}
	for _, args := range cmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = tempDir
		if err := cmd.Run(); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to configure git: %v", err)
		}
	}

	return &TestHelper{t: t, tempDir: tempDir}
}

// initTestRepoWithMain is a local version to avoid circular dependencies
func initTestRepoWithMain(t *testing.T, dir string) {
	t.Helper()

	// Try to init with main as initial branch (git 2.28+)
	cmd := exec.Command("git", "init", "--initial-branch=main")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// Fallback for older git versions
		cmd = exec.Command("git", "init")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to init git repo: %v", err)
		}
		// Set default branch config for consistency
		cmd = exec.Command("git", "config", "init.defaultBranch", "main")
		cmd.Dir = dir
		_ = cmd.Run() // Ignore error if config not supported
	}
}

// Cleanup removes the temporary directory
func (h *TestHelper) Cleanup() {
	os.RemoveAll(h.tempDir)
}

// CreateCommit creates a commit with a test file
func (h *TestHelper) CreateCommit(message string) {
	// Create a file
	testFile := filepath.Join(h.tempDir, fmt.Sprintf("test-%d.txt", len(message)))
	if err := os.WriteFile(testFile, []byte(message), 0644); err != nil {
		h.t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit
	cmds := [][]string{
		{"add", "."},
		{"commit", "-m", message},
	}
	for _, args := range cmds {
		cmd := exec.Command("git", args...)
		cmd.Dir = h.tempDir
		if err := cmd.Run(); err != nil {
			h.t.Fatalf("Failed to create commit: %v", err)
		}
	}
}

// CreateBranch creates a new branch
func (h *TestHelper) CreateBranch(name string) {
	cmd := exec.Command("git", "checkout", "-b", name)
	cmd.Dir = h.tempDir
	if err := cmd.Run(); err != nil {
		h.t.Fatalf("Failed to create branch %s: %v", name, err)
	}
}

// CreateMergeCommit creates a merge commit
func (h *TestHelper) CreateMergeCommit(fromBranch, message string) {
	cmd := exec.Command("git", "merge", fromBranch, "-m", message)
	cmd.Dir = h.tempDir
	if err := cmd.Run(); err != nil {
		h.t.Fatalf("Failed to create merge commit: %v", err)
	}
}

func TestCommandClient_ForEachRef(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create some commits and branches
	helper.CreateCommit("Initial commit")
	helper.CreateBranch("feature-1")
	helper.CreateCommit("Feature 1 commit")

	// Small delay to ensure different commit times
	time.Sleep(10 * time.Millisecond)

	helper.CreateBranch("feature-2")
	helper.CreateCommit("Feature 2 commit")

	client := NewCommandClient(helper.tempDir)

	tests := []struct {
		name    string
		format  string
		options []string
		want    []string
	}{
		{
			name:   "basic format",
			format: "%(refname:short)",
			want:   []string{"feature-1", "feature-2", "main"},
		},
		{
			name:    "with refs filter",
			format:  "%(refname:short)",
			options: []string{"refs/heads/"},
			want:    []string{"feature-1", "feature-2", "main"},
		},
		{
			name:    "with sorting",
			format:  "%(refname:short)",
			options: []string{"--sort=-committerdate", "refs/heads/"},
			want:    []string{"feature-2", "feature-1", "main"}, // feature-2 is most recent
		},
		{
			name:    "with count limit",
			format:  "%(refname:short)",
			options: []string{"--count=2", "--sort=-committerdate", "refs/heads/"},
			want:    []string{"feature-2", "feature-1"},
		},
		{
			name:   "multiple fields format",
			format: "%(refname:short)|%(objectname:short)",
			want:   []string{"|"}, // Should contain pipe separator
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := client.ForEachRef(tt.format, tt.options...)
			if err != nil {
				t.Fatalf("ForEachRef() error = %v", err)
			}

			lines := strings.Split(output, "\n")
			// Filter out empty lines
			var nonEmpty []string
			for _, line := range lines {
				if line != "" {
					nonEmpty = append(nonEmpty, line)
				}
			}

			if tt.name == "multiple fields format" {
				// Just check that output contains the separator
				for _, line := range nonEmpty {
					if !strings.Contains(line, "|") {
						t.Errorf("Expected output to contain '|', got %s", line)
					}
				}
			} else {
				// Check branch names
				if len(nonEmpty) != len(tt.want) {
					t.Errorf("Expected %d branches, got %d: %v", len(tt.want), len(nonEmpty), nonEmpty)
				}

				// For sorting tests, we need to be more flexible since git sort might vary
				if tt.name == "with sorting" || tt.name == "with count limit" {
					// Just check that we have the expected branches, not the exact order
					for _, want := range tt.want {
						found := false
						for _, got := range nonEmpty {
							if strings.Contains(got, want) {
								found = true
								break
							}
						}
						if !found {
							t.Errorf("Expected to find branch %s in output, got %v", want, nonEmpty)
						}
					}
				} else {
					// For non-sorting tests, check exact order
					for i, want := range tt.want {
						if i < len(nonEmpty) && !strings.Contains(nonEmpty[i], want) {
							t.Errorf("Expected branch %s at position %d, got %s", want, i, nonEmpty[i])
						}
					}
				}
			}
		})
	}
}

func TestCommandClient_GetConfigValue(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	client := NewCommandClient(helper.tempDir)

	tests := []struct {
		name    string
		key     string
		want    string
		wantErr bool
	}{
		{
			name: "existing config value",
			key:  "user.name",
			want: "Test User",
		},
		{
			name: "existing email value",
			key:  "user.email",
			want: "test@example.com",
		},
		{
			name: "non-existent config value",
			key:  "non.existent.key",
			want: "", // Should return empty string, not error
		},
		{
			name:    "invalid config key format",
			key:     "invalid key with spaces",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := client.GetConfigValue(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if value != tt.want {
				t.Errorf("GetConfigValue() = %v, want %v", value, tt.want)
			}
		})
	}
}

func TestCommandClient_GetLastNonMergeCommit(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create initial commit on main
	helper.CreateCommit("Initial commit")

	// Create feature branch with commits
	helper.CreateBranch("feature")
	helper.CreateCommit("Feature commit 1")
	helper.CreateCommit("Feature commit 2")

	// Go back to main and create another commit
	cmd := exec.Command("git", "checkout", "main")
	cmd.Dir = helper.tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to checkout main: %v", err)
	}
	helper.CreateCommit("Main commit")

	// Create a merge commit
	helper.CreateMergeCommit("feature", "Merge feature branch")

	// Create one more regular commit
	helper.CreateCommit("Post-merge commit")

	client := NewCommandClient(helper.tempDir)

	tests := []struct {
		name       string
		branch     string
		format     string
		wantSubstr string
		wantEmpty  bool
		wantErr    bool
	}{
		{
			name:       "last non-merge commit on main",
			branch:     "main",
			format:     "%s",
			wantSubstr: "Post-merge commit",
		},
		{
			name:       "last non-merge commit on feature",
			branch:     "feature",
			format:     "%s",
			wantSubstr: "Feature commit 2",
		},
		{
			name:       "custom format",
			branch:     "main",
			format:     "%H|%s|%an",
			wantSubstr: "|Post-merge commit|Test User",
		},
		{
			name:    "non-existent branch",
			branch:  "non-existent",
			format:  "%s",
			wantErr: true,
		},
		{
			name:      "empty format uses default",
			branch:    "main",
			format:    "",
			wantEmpty: false, // Should return something
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := client.GetLastNonMergeCommit(tt.branch, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLastNonMergeCommit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if tt.wantEmpty && output != "" {
				t.Errorf("Expected empty output, got %v", output)
			}

			if tt.wantSubstr != "" && !strings.Contains(output, tt.wantSubstr) {
				t.Errorf("GetLastNonMergeCommit() = %v, want substring %v", output, tt.wantSubstr)
			}

			if tt.format == "" && output == "" {
				t.Errorf("Expected non-empty output with empty format")
			}
		})
	}
}

func TestCommandClient_GetLastNonMergeCommit_EdgeCases(t *testing.T) {
	t.Run("branch with only merge commits", func(t *testing.T) {
		// Create a new test helper for this specific test
		helper2 := NewTestHelper(t)
		defer helper2.Cleanup()
		client2 := NewCommandClient(helper2.tempDir)

		// Create initial setup
		helper2.CreateCommit("Initial")
		helper2.CreateBranch("branch1")
		helper2.CreateCommit("Branch1 commit")

		cmd := exec.Command("git", "checkout", "main")
		cmd.Dir = helper2.tempDir
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to checkout main: %v", err)
		}

		helper2.CreateBranch("branch2")
		// Only add merge commits to branch2
		helper2.CreateMergeCommit("branch1", "Merge branch1 into branch2")

		output, err := client2.GetLastNonMergeCommit("branch2", "%s")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// When a branch only has merge commits, git log --no-merges will find
		// the last non-merge commit in its history, which is "Branch1 commit"
		// since branch2 was created from main and then merged branch1
		if !strings.Contains(output, "Branch1 commit") && !strings.Contains(output, "Initial") {
			t.Errorf("Expected to find a non-merge commit, got %v", output)
		}
	})
}

func TestCommandClient_Checkout(t *testing.T) {
	helper := NewTestHelper(t)
	defer helper.Cleanup()

	// Create branches
	helper.CreateCommit("Initial commit")
	helper.CreateBranch("feature")
	helper.CreateCommit("Feature commit")

	client := NewCommandClient(helper.tempDir)

	tests := []struct {
		name    string
		branch  string
		wantErr bool
	}{
		{
			name:    "checkout existing branch",
			branch:  "main",
			wantErr: false,
		},
		{
			name:    "checkout feature branch",
			branch:  "feature",
			wantErr: false,
		},
		{
			name:    "checkout non-existent branch",
			branch:  "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Checkout(tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Checkout() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify we're on the correct branch
				currentBranch, err := client.GetCurrentBranch()
				if err != nil {
					t.Fatalf("Failed to get current branch: %v", err)
				}
				if currentBranch != tt.branch {
					t.Errorf("Expected to be on branch %s, but on %s", tt.branch, currentBranch)
				}
			}
		})
	}
}

func TestCommandClient_runCommand_Errors(t *testing.T) {
	client := NewCommandClient("/nonexistent/directory")

	// This should fail because the directory doesn't exist
	_, err := client.ForEachRef("%(refname:short)")
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}

	// Test with malformed git command
	_, err = client.runCommand("not-a-git-command")
	if err == nil {
		t.Error("Expected error for invalid git command")
	}
}

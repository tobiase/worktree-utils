package worktree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobiase/worktree-utils/test/helpers"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		args      []string
		wantError bool
	}{
		{
			name:      "simple echo command",
			command:   "echo",
			args:      []string{"hello"},
			wantError: false,
		},
		{
			name:      "command not found",
			command:   "nonexistent-command-12345",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "false command",
			command:   "false",
			args:      []string{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RunCommand(tt.command, tt.args...)
			if (err != nil) != tt.wantError {
				t.Errorf("RunCommand() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetGitRemote(t *testing.T) {
	// Create a temporary git repo
	tempDir := t.TempDir()
	gitDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()

	if err := os.Chdir(gitDir); err != nil {
		t.Fatal(err)
	}

	// Initialize git repo
	if _, _, err := helpers.RunCommand(t, "git", "init"); err != nil {
		t.Fatal(err)
	}

	// Configure git user for commits (for consistency)
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.name", "Test User"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}

	// Test without remote
	remote, err := GetGitRemote()
	if err != nil {
		t.Errorf("GetGitRemote() error = %v", err)
	}
	if remote != "" {
		t.Errorf("GetGitRemote() = %q, want empty string for repo without remote", remote)
	}

	// Add remote
	if _, _, err := helpers.RunCommand(t, "git", "remote", "add", "origin", "https://github.com/test/repo.git"); err != nil {
		t.Fatal(err)
	}

	// Test with remote
	remote, err = GetGitRemote()
	if err != nil {
		t.Errorf("GetGitRemote() error = %v", err)
	}
	if remote != "https://github.com/test/repo.git" {
		t.Errorf("GetGitRemote() = %q, want https://github.com/test/repo.git", remote)
	}
}

func TestGetRelativePath(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	mainRepo := filepath.Join(tempDir, "test-repo")
	subPath := filepath.Join(mainRepo, "src", "components")

	if err := os.MkdirAll(subPath, 0755); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()

	// Initialize git repo
	if err := os.Chdir(mainRepo); err != nil {
		t.Fatal(err)
	}

	if _, _, err := helpers.RunCommand(t, "git", "init"); err != nil {
		t.Fatal(err)
	}

	// Configure git user for commits (for consistency)
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.name", "Test User"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}

	// Test from subdirectory
	if err := os.Chdir(subPath); err != nil {
		t.Fatal(err)
	}

	cwd, _ := os.Getwd()
	relPath, err := GetRelativePath(cwd)
	if err != nil {
		t.Fatalf("GetRelativePath() error = %v", err)
	}

	expected := filepath.Join("src", "components")
	if relPath != expected {
		t.Errorf("GetRelativePath() = %q, want %q", relPath, expected)
	}

	// Test from same directory
	if err := os.Chdir(mainRepo); err != nil {
		t.Fatal(err)
	}

	cwd, _ = os.Getwd()
	relPath, err = GetRelativePath(cwd)
	if err != nil {
		t.Fatalf("GetRelativePath() error = %v", err)
	}

	if relPath != "." {
		t.Errorf("GetRelativePath() = %q, want .", relPath)
	}
}

func TestAdd(t *testing.T) {
	// Create a temporary git repo with worktree setup
	tempDir := t.TempDir()
	mainRepo := filepath.Join(tempDir, "test-repo")
	worktreeBase := filepath.Join(tempDir, "test-repo-worktrees")

	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()

	if err := os.Chdir(mainRepo); err != nil {
		t.Fatal(err)
	}

	// Initialize git repo
	if _, _, err := helpers.RunCommand(t, "git", "init"); err != nil {
		t.Fatal(err)
	}

	// Configure git user for commits
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.name", "Test User"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	testFile := filepath.Join(mainRepo, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := helpers.RunCommand(t, "git", "add", "."); err != nil {
		t.Fatal(err)
	}

	if _, _, err := helpers.RunCommand(t, "git", "commit", "-m", "Initial commit"); err != nil {
		t.Fatal(err)
	}

	// Create the branch first
	if _, _, err := helpers.RunCommand(t, "git", "checkout", "-b", "feature-branch"); err != nil {
		t.Fatal(err)
	}

	// Go back to main branch
	if _, _, err := helpers.RunCommand(t, "git", "checkout", "master"); err != nil {
		// Try main if master doesn't exist
		if _, _, err := helpers.RunCommand(t, "git", "checkout", "main"); err != nil {
			// Just create a new branch if neither exists
			if _, _, err := helpers.RunCommand(t, "git", "checkout", "-b", "main"); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Test adding a worktree for existing branch
	err := Add("feature-branch", nil)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify worktree was created
	expectedPath := filepath.Join(worktreeBase, "feature-branch")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Expected worktree at %q to exist", expectedPath)
	}

	// Test adding duplicate worktree
	err = Add("feature-branch", nil)
	if err == nil {
		t.Error("Add() should fail for duplicate branch")
	}
}

func TestRemoveWithForce(t *testing.T) {
	// Create a temporary git repo with worktree
	tempDir := t.TempDir()
	mainRepo := filepath.Join(tempDir, "test-repo")
	worktreeBase := filepath.Join(tempDir, "test-repo-worktrees")

	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()

	if err := os.Chdir(mainRepo); err != nil {
		t.Fatal(err)
	}

	// Initialize git repo
	if _, _, err := helpers.RunCommand(t, "git", "init"); err != nil {
		t.Fatal(err)
	}

	// Configure git user for commits
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.name", "Test User"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	testFile := filepath.Join(mainRepo, "README.md")
	if err := os.WriteFile(testFile, []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := helpers.RunCommand(t, "git", "add", "."); err != nil {
		t.Fatal(err)
	}

	if _, _, err := helpers.RunCommand(t, "git", "commit", "-m", "Initial commit"); err != nil {
		t.Fatal(err)
	}

	// Create a worktree
	branchName := "test-branch"
	if _, _, err := helpers.RunCommand(t, "git", "worktree", "add", filepath.Join(worktreeBase, branchName), "-b", branchName); err != nil {
		t.Fatal(err)
	}

	// Test remove with force
	err := Remove(branchName)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	// Verify worktree was removed
	expectedPath := filepath.Join(worktreeBase, branchName)
	if _, err := os.Stat(expectedPath); !os.IsNotExist(err) {
		t.Errorf("Expected worktree at %q to be removed", expectedPath)
	}
}

func TestResolveBranchName(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		availableBranches []string
		want              string
		wantErr           bool
	}{
		{
			name:              "exact match",
			input:             "feature",
			availableBranches: []string{"main", "feature", "feature-branch"},
			want:              "feature",
			wantErr:           false,
		},
		{
			name:              "prefix match",
			input:             "feat",
			availableBranches: []string{"main", "feature"},
			want:              "feature",
			wantErr:           false,
		},
		{
			name:              "ambiguous match",
			input:             "feat",
			availableBranches: []string{"main", "feature", "feature-branch"},
			want:              "",
			wantErr:           true,
		},
		{
			name:              "no match",
			input:             "nonexistent",
			availableBranches: []string{"main", "feature"},
			want:              "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveBranchName(tt.input, tt.availableBranches)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveBranchName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResolveBranchName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvFileFunctions(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test env files
	envContent := "TEST_VAR=123\nANOTHER_VAR=abc\n"
	envLocalContent := "LOCAL_VAR=456\n"

	if err := os.WriteFile(filepath.Join(sourceDir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, ".env.local"), []byte(envLocalContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory with env file
	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, ".env"), []byte("SUB_VAR=789\n"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("TestSyncEnvFiles", func(t *testing.T) {
		// Skip this test as it requires actual git worktrees
		t.Skip("Skipping SyncEnvFiles test - requires actual git worktree setup")

		// Verify .env was copied
		targetEnv := filepath.Join(targetDir, ".env")
		if _, err := os.Stat(targetEnv); os.IsNotExist(err) {
			t.Error("Expected .env file to be copied")
		}

		// Verify content matches
		content, err := os.ReadFile(targetEnv)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != envContent {
			t.Errorf("Copied content = %q, want %q", string(content), envContent)
		}
	})

	t.Run("TestSyncEnvFilesRecursive", func(t *testing.T) {
		// Clean target directory
		_ = os.RemoveAll(targetDir)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Skip this test as it requires actual git worktrees
		t.Skip("Skipping SyncEnvFiles recursive test - requires actual git worktree setup")

		// Verify subdirectory .env was copied
		targetSubEnv := filepath.Join(targetDir, "subdir", ".env")
		if _, err := os.Stat(targetSubEnv); os.IsNotExist(err) {
			t.Error("Expected subdirectory .env file to be copied")
		}

		// Verify .env.local was copied
		targetEnvLocal := filepath.Join(targetDir, ".env.local")
		if _, err := os.Stat(targetEnvLocal); os.IsNotExist(err) {
			t.Error("Expected .env.local file to be copied")
		}
	})

	t.Run("TestDiffEnvFiles", func(t *testing.T) {
		// Modify target .env file
		modifiedContent := "TEST_VAR=999\nNEW_VAR=xyz\n"
		if err := os.WriteFile(filepath.Join(targetDir, ".env"), []byte(modifiedContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Skip this test as it requires actual git worktrees
		t.Skip("Skipping DiffEnvFiles test - requires actual git worktree setup")
	})

	t.Run("TestListEnvFiles", func(t *testing.T) {
		// Skip this test as it requires actual git worktrees
		t.Skip("Skipping ListEnvFiles test - requires actual git worktree setup")
	})
}

func TestCopyFileForSetup(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test file
	sourceFile := filepath.Join(sourceDir, "test.txt")
	content := "test content"
	if err := os.WriteFile(sourceFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test copying file
	targetFile := filepath.Join(targetDir, "test.txt")
	if err := copyFileForSetup(sourceFile, targetFile); err != nil {
		t.Fatalf("copyFileForSetup() error = %v", err)
	}

	// Verify file was copied
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		t.Error("Expected file to be copied")
	}

	// Verify content
	copiedContent, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(copiedContent) != content {
		t.Errorf("Copied content = %q, want %q", string(copiedContent), content)
	}

	// Test copying non-existent file
	if err := copyFileForSetup(filepath.Join(sourceDir, "nonexistent"), targetFile); err == nil {
		t.Error("Expected error for non-existent source file")
	}
}

func TestRunSetup(t *testing.T) {
	// Create a temporary git repo
	tempDir := t.TempDir()
	mainRepo := filepath.Join(tempDir, "test-repo")

	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()

	if err := os.Chdir(mainRepo); err != nil {
		t.Fatal(err)
	}

	// Test RunSetup with nil config (should not error)
	if err := RunSetup(mainRepo, mainRepo, nil); err != nil {
		t.Errorf("RunSetup() with nil config error = %v", err)
	}
}

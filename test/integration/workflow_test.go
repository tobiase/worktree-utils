package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tobiase/worktree-utils/test/helpers"
)

// TestCompleteWorkflow tests a complete workflow from creating a worktree to removing it
func TestCompleteWorkflow(t *testing.T) {
	// Build the binary
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	// Create test repository
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	// Change to repo directory
	oldWd, _ := os.Getwd()
	os.Chdir(repo)
	defer os.Chdir(oldWd)

	// Test workflow steps
	t.Run("list initial worktrees", func(t *testing.T) {
		output := runCommand(t, binPath, "list")
		if !strings.Contains(output, "main") {
			t.Error("Expected to see main branch in list")
		}
	})

	t.Run("create new worktree", func(t *testing.T) {
		output := runCommand(t, binPath, "new", "feature-test", "--base", "HEAD")
		if !strings.Contains(output, "CD:") {
			t.Errorf("Expected CD: prefix for new command, got: %s", output)
		}
		
		// Verify worktree was created
		worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
		worktreePath := filepath.Join(worktreeBase, "feature-test")
		if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
			t.Error("Worktree directory was not created")
		}
	})

	t.Run("list shows new worktree", func(t *testing.T) {
		output := runCommand(t, binPath, "list")
		if !strings.Contains(output, "feature-test") {
			t.Error("Expected to see feature-test branch in list")
		}
		// Should show index numbers
		if !strings.Contains(output, "0") && !strings.Contains(output, "1") {
			t.Error("Expected to see index numbers in list")
		}
	})

	t.Run("go to worktree by name", func(t *testing.T) {
		output := runCommand(t, binPath, "go", "feature-test")
		if !strings.Contains(output, "CD:") {
			t.Error("Expected CD: prefix for go command")
		}
		if !strings.Contains(output, "feature-test") {
			t.Error("Expected path to contain feature-test")
		}
	})

	t.Run("go to worktree by index", func(t *testing.T) {
		output := runCommand(t, binPath, "go", "0")
		if !strings.Contains(output, "CD:") {
			t.Error("Expected CD: prefix for go command")
		}
	})

	t.Run("go back to main", func(t *testing.T) {
		output := runCommand(t, binPath, "go", "main")
		if !strings.Contains(output, "CD:") {
			t.Error("Expected CD: prefix for go command")
		}
		// Should return to original repo path
		if !strings.Contains(output, repo) {
			t.Error("Expected to return to main repo path")
		}
	})

	t.Run("remove worktree", func(t *testing.T) {
		// First go back to main to avoid removing current worktree
		runCommand(t, binPath, "go", "main")
		
		// Now remove the feature worktree
		_ = runCommand(t, binPath, "rm", "feature-test")
		
		// Verify it was removed
		worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
		worktreePath := filepath.Join(worktreeBase, "feature-test")
		if _, err := os.Stat(worktreePath); !os.IsNotExist(err) {
			t.Error("Worktree directory was not removed")
		}
	})

	t.Run("list no longer shows removed worktree", func(t *testing.T) {
		output := runCommand(t, binPath, "list")
		if strings.Contains(output, "feature-test") {
			t.Error("Expected feature-test to be removed from list")
		}
	})
}

// TestEnvCopyWorkflow tests the env-copy functionality
func TestEnvCopyWorkflow(t *testing.T) {
	// Build the binary
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	// Create test repository
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	// Create .env file
	envContent := "DATABASE_URL=postgres://localhost/test\nAPI_KEY=secret123\n"
	envPath := filepath.Join(repo, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a worktree
	oldWd, _ := os.Getwd()
	os.Chdir(repo)
	defer os.Chdir(oldWd)

	// Use helper to create worktree with new branch
	helpers.AddTestWorktree(t, repo, "feature-env")

	t.Run("copy env file to worktree", func(t *testing.T) {
		output := runCommand(t, binPath, "env-copy", "feature-env")
		
		// Should show success message
		if !strings.Contains(output, "Copied") {
			t.Errorf("Expected success message, got: %s", output)
		}

		// Verify file was copied
		worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
		copiedEnv := filepath.Join(worktreeBase, "feature-env", ".env")
		
		content, err := os.ReadFile(copiedEnv)
		if err != nil {
			t.Errorf("Failed to read copied .env file: %v", err)
		}
		
		if string(content) != envContent {
			t.Errorf("Copied content doesn't match original")
		}
	})

	t.Run("env-copy from subdirectory", func(t *testing.T) {
		// Create subdirectory with .env
		subdir := filepath.Join(repo, "src", "api")
		os.MkdirAll(subdir, 0755)
		
		subEnvContent := "SUBDIRECTORY_ENV=true\n"
		subEnvPath := filepath.Join(subdir, ".env")
		if err := os.WriteFile(subEnvPath, []byte(subEnvContent), 0644); err != nil {
			t.Fatal(err)
		}

		// Change to subdirectory
		os.Chdir(subdir)
		
		_ = runCommand(t, binPath, "env-copy", "feature-env")
		
		// Should maintain relative path
		worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
		copiedSubEnv := filepath.Join(worktreeBase, "feature-env", "src", "api", ".env")
		
		if _, err := os.Stat(copiedSubEnv); os.IsNotExist(err) {
			t.Error("Subdirectory .env was not copied to correct location")
		}
	})
}

// TestProjectCommands tests project-specific command functionality
func TestProjectCommands(t *testing.T) {
	// Build the binary
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	// Create test repository
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	// Create mock home directory for config
	mockHome := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", mockHome)
	defer os.Setenv("HOME", oldHome)

	// Initialize project
	oldWd, _ := os.Getwd()
	os.Chdir(repo)
	defer os.Chdir(oldWd)

	t.Run("initialize project", func(t *testing.T) {
		_ = runCommand(t, binPath, "project", "init", "testproject")
		
		// Verify config was created
		configPath := filepath.Join(mockHome, ".config", "wt", "projects", "testproject.yaml")
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("Project config file was not created")
		}
	})

	t.Run("project commands available in project directory", func(t *testing.T) {
		// Modify the project config to add custom commands
		configPath := filepath.Join(mockHome, ".config", "wt", "projects", "testproject.yaml")
		// Resolve symlinks for macOS
		resolvedRepo, _ := filepath.EvalSymlinks(repo)
		config := `name: testproject
match:
  paths:
    - ` + resolvedRepo + `
    - ` + repo + `
commands:
  api:
    description: "Go to API"
    target: "services/api"
  docs:
    description: "Go to docs"
    target: "docs"
`
		if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
			t.Fatal(err)
		}

		// Create the target directories
		os.MkdirAll(filepath.Join(repo, "services", "api"), 0755)
		os.MkdirAll(filepath.Join(repo, "docs"), 0755)

		// Test custom navigation command
		output := runCommand(t, binPath, "api")
		if !strings.Contains(output, "CD:") {
			t.Errorf("Expected CD: prefix for project navigation command, got: %s", output)
		}
		if !strings.Contains(output, "services/api") {
			t.Errorf("Expected path to contain services/api, got: %s", output)
		}
	})
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	// Build the binary
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	tests := []struct {
		name        string
		args        []string
		setup       func() func()
		expectError bool
		errorMsg    string
	}{
		{
			name:        "invalid command",
			args:        []string{"invalid-cmd"},
			expectError: true,
			errorMsg:    "unknown command",
		},
		{
			name:        "go without arguments outside repo",
			args:        []string{"go"},
			expectError: true,
			errorMsg:    "not inside a Git repository",
			setup: func() func() {
				oldWd, _ := os.Getwd()
				tmpDir := t.TempDir()
				os.Chdir(tmpDir)
				return func() { os.Chdir(oldWd) }
			},
		},
		{
			name:        "remove non-existent worktree",
			args:        []string{"rm", "non-existent"},
			expectError: true,
			errorMsg:    "not found",
			setup: func() func() {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return func() {
					os.Chdir(oldWd)
					cleanup()
				}
			},
		},
		{
			name:        "add existing branch",
			args:        []string{"add", "main"},
			expectError: true,
			errorMsg:    "already",
			setup: func() func() {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return func() {
					os.Chdir(oldWd)
					cleanup()
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cleanup func()
			if tt.setup != nil {
				cleanup = tt.setup()
				defer cleanup()
			}

			cmd := exec.Command(binPath, tt.args...)
			// Set LANG=C to ensure consistent output across locales
			cmd.Env = append(os.Environ(), "LANG=C")
			output, err := cmd.CombinedOutput()

			if !tt.expectError {
				if err != nil {
					t.Errorf("Unexpected error: %v\nOutput: %s", err, output)
				}
			} else {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(string(output), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, string(output))
				}
			}
		})
	}
}


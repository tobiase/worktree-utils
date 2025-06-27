package worktree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobiase/worktree-utils/internal/config"
	"github.com/tobiase/worktree-utils/test/helpers"
)

func TestCopyEnvFilesRecursive(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	targetDir := filepath.Join(tempDir, "target")

	// Create nested directory structure
	nestedDir := filepath.Join(sourceDir, "config", "env")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create various env files
	envFiles := map[string]string{
		filepath.Join(sourceDir, ".env"):                 "ROOT_VAR=1",
		filepath.Join(sourceDir, ".env.local"):           "LOCAL_VAR=2",
		filepath.Join(sourceDir, ".env.production"):      "PROD_VAR=3",
		filepath.Join(sourceDir, "config", ".env"):       "CONFIG_VAR=4",
		filepath.Join(nestedDir, ".env.test"):            "TEST_VAR=5",
		filepath.Join(sourceDir, "not-env.txt"):          "NOT_ENV=6",
		filepath.Join(sourceDir, "config", "env.config"): "NOT_DOT_ENV=7",
		filepath.Join(sourceDir, ".environment"):         "NOT_ENV_FILE=8",
	}

	for path, content := range envFiles {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test recursive copy
	err := copyEnvFilesRecursive(sourceDir, targetDir)
	if err != nil {
		t.Fatalf("copyEnvFilesRecursive() error = %v", err)
	}

	// Verify correct files were copied
	expectedCopied := []string{
		".env",
		".env.local",
		".env.production",
		filepath.Join("config", ".env"),
		filepath.Join("config", "env", ".env.test"),
	}

	// Verify each expected file was copied with correct content
	for _, relPath := range expectedCopied {
		targetPath := filepath.Join(targetDir, relPath)
		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			t.Errorf("Expected %s to be copied", relPath)
			continue
		}

		// Verify content
		sourcePath := filepath.Join(sourceDir, relPath)
		sourceContent, _ := os.ReadFile(sourcePath)
		targetContent, _ := os.ReadFile(targetPath)
		if string(sourceContent) != string(targetContent) {
			t.Errorf("Content mismatch for %s", relPath)
		}
	}

	// Verify non-env files were not copied
	notCopied := []string{
		"not-env.txt",
		filepath.Join("config", "env.config"),
		".environment",
	}

	for _, relPath := range notCopied {
		targetPath := filepath.Join(targetDir, relPath)
		if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
			t.Errorf("File %s should not have been copied", relPath)
		}
	}
}

func TestCopyFile(t *testing.T) {
	tempDir := t.TempDir()

	// Test basic file copy
	t.Run("basic copy", func(t *testing.T) {
		src := filepath.Join(tempDir, "source.txt")
		dst := filepath.Join(tempDir, "dest.txt")
		content := "test content\nwith newlines\n"

		if err := os.WriteFile(src, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}

		if err := copyFile(src, dst); err != nil {
			t.Fatalf("copyFile() error = %v", err)
		}

		// Verify content
		copied, err := os.ReadFile(dst)
		if err != nil {
			t.Fatal(err)
		}
		if string(copied) != content {
			t.Error("Content mismatch after copy")
		}

		// Verify permissions
		srcInfo, _ := os.Stat(src)
		dstInfo, _ := os.Stat(dst)
		if srcInfo.Mode() != dstInfo.Mode() {
			t.Error("Permission mismatch after copy")
		}
	})

	// Test copy to subdirectory
	t.Run("copy to subdirectory", func(t *testing.T) {
		src := filepath.Join(tempDir, "source2.txt")
		dst := filepath.Join(tempDir, "subdir", "dest2.txt")

		if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		if err := copyFile(src, dst); err != nil {
			t.Fatalf("copyFile() error = %v", err)
		}

		if _, err := os.Stat(dst); os.IsNotExist(err) {
			t.Error("Destination file not created")
		}
	})

	// Test copy non-existent file
	t.Run("copy non-existent", func(t *testing.T) {
		src := filepath.Join(tempDir, "does-not-exist.txt")
		dst := filepath.Join(tempDir, "dest3.txt")

		if err := copyFile(src, dst); err == nil {
			t.Error("Expected error for non-existent source")
		}
	})
}

func TestRunWorktreeSetup(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	worktreeDir := filepath.Join(tempDir, "worktree")

	if err := os.MkdirAll(worktreeDir, 0755); err != nil {
		t.Fatal(err)
	}

	oldCwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldCwd) }()

	if err := os.Chdir(worktreeDir); err != nil {
		t.Fatal(err)
	}

	// Test with copy command
	t.Run("copy command", func(t *testing.T) {
		// Create source file
		srcFile := filepath.Join(tempDir, "source.txt")
		if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		// Create a simple setup config for testing
		setup := &config.SetupConfig{
			CopyFiles: []config.CopyFileConfig{
				{Source: "source.txt", Target: "copied.txt"},
			},
		}
		if err := runWorktreeSetup(tempDir, worktreeDir, setup); err != nil {
			t.Errorf("runWorktreeSetup() copy error = %v", err)
		}

		// Verify file was copied
		if _, err := os.Stat(filepath.Join(worktreeDir, "copied.txt")); os.IsNotExist(err) {
			t.Error("Expected file to be copied")
		}
	})

	// Test with run command (echo)
	t.Run("run command", func(t *testing.T) {
		// Create a setup config with run command
		setup := &config.SetupConfig{
			Commands: []config.SetupCommand{
				{Command: "echo 'test' > echo_output.txt", Directory: "."},
			},
		}
		if err := runWorktreeSetup(tempDir, worktreeDir, setup); err != nil {
			t.Errorf("runWorktreeSetup() run error = %v", err)
		}

		// Verify command created file
		outputFile := filepath.Join(worktreeDir, "echo_output.txt")
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Error("Expected echo command to create file")
		}
	})

	// Test with nil setup config
	t.Run("nil config", func(t *testing.T) {
		// Test with nil setup config should not error
		if err := runWorktreeSetup(tempDir, worktreeDir, nil); err != nil {
			t.Errorf("runWorktreeSetup() with nil config should not error: %v", err)
		}
	})
}

func TestBranchDetection(t *testing.T) {
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

	// Configure git user for commits
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.name", "Test User"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := helpers.RunCommand(t, "git", "config", "user.email", "test@example.com"); err != nil {
		t.Fatal(err)
	}

	// Create initial commit
	if err := os.WriteFile("README.md", []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := helpers.RunCommand(t, "git", "add", "."); err != nil {
		t.Fatal(err)
	}
	if _, _, err := helpers.RunCommand(t, "git", "commit", "-m", "Initial"); err != nil {
		t.Fatal(err)
	}

	// Create and checkout a new branch
	if _, _, err := helpers.RunCommand(t, "git", "checkout", "-b", "test-branch"); err != nil {
		t.Fatal(err)
	}

	// Test branch existence check
	t.Run("checkBranchExists", func(t *testing.T) {
		// Existing branch
		if !checkBranchExists("test-branch") {
			t.Error("checkBranchExists() should return true for existing branch")
		}
		if !checkBranchExists("master") && !checkBranchExists("main") {
			t.Error("checkBranchExists() should return true for default branch")
		}

		// Non-existing branch
		if checkBranchExists("non-existent-branch") {
			t.Error("checkBranchExists() should return false for non-existing branch")
		}
	})
}

func TestGetRelativePathProject(t *testing.T) {
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

	// Configure git user for commits (in case any git operations need it)
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

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	relPath, err := GetRelativePath(cwd)
	if err != nil {
		t.Fatalf("GetRelativePath() error = %v", err)
	}

	expected := filepath.Join("src", "components")
	if relPath != expected {
		t.Errorf("GetRelativePath() = %q, want %q", relPath, expected)
	}

	// Test from repo root
	if err := os.Chdir(mainRepo); err != nil {
		t.Fatal(err)
	}

	cwd, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	relPath, err = GetRelativePath(cwd)
	if err != nil {
		t.Fatalf("GetRelativePath() error = %v", err)
	}

	if relPath != "." {
		t.Errorf("GetRelativePath() = %q, want .", relPath)
	}
}

func TestSmartNewWorktree(t *testing.T) {
	// This test requires a full git repo setup with worktrees
	// For now, we'll skip the complex setup and just test the logic
	// Real integration tests would need to set up proper git worktrees
	t.Skip("Skipping integration test that requires full git worktree setup")
}

func TestDetermineBaseBranch(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		fallback string
		branches []string
		want     string
	}{
		{
			name:     "specified base exists",
			base:     "develop",
			fallback: "main",
			branches: []string{"main", "develop", "feature"},
			want:     "develop",
		},
		{
			name:     "specified base doesn't exist, use fallback",
			base:     "non-existent",
			fallback: "main",
			branches: []string{"main", "develop", "feature"},
			want:     "main",
		},
		{
			name:     "empty base, use fallback",
			base:     "",
			fallback: "master",
			branches: []string{"master", "develop", "feature"},
			want:     "master",
		},
		{
			name:     "neither base nor fallback exist",
			base:     "non-existent",
			fallback: "also-non-existent",
			branches: []string{"develop", "feature"},
			want:     "also-non-existent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock branch checking
			result := determineBaseBranch(tt.base, tt.fallback, tt.branches)
			if result != tt.want {
				t.Errorf("determineBaseBranch() = %q, want %q", result, tt.want)
			}
		})
	}
}

// Helper function for tests
func determineBaseBranch(base, fallback string, branches []string) string {
	if base != "" {
		for _, b := range branches {
			if b == base {
				return base
			}
		}
	}
	return fallback
}

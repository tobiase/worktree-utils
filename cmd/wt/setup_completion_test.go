package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupCompletionCommand(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectError   bool
		expectOutput  []string
		expectFiles   []string
		expectNoFiles []string
	}{
		{
			name:         "setup with completion auto",
			args:         []string{"setup", "--completion", "auto"},
			expectError:  false,
			expectOutput: []string{"✓ Setup complete!"},
			expectFiles:  []string{}, // No files created with process substitution
		},
		{
			name:         "setup with completion bash",
			args:         []string{"setup", "--completion", "bash"},
			expectError:  false,
			expectOutput: []string{"✓ Setup complete!"},
			expectFiles:  []string{}, // No files created with process substitution
		},
		{
			name:         "setup with completion zsh",
			args:         []string{"setup", "--completion", "zsh"},
			expectError:  false,
			expectOutput: []string{"✓ Setup complete!"},
			expectFiles:  []string{}, // No files created with process substitution
		},
		{
			name:          "setup with completion none",
			args:          []string{"setup", "--completion", "none"},
			expectError:   false,
			expectOutput:  []string{"✓ Setup complete!"},
			expectFiles:   []string{},
			expectNoFiles: []string{".config/wt/completion.bash", ".config/wt/completions/_wt"},
		},
		{
			name:          "setup with no-completion flag",
			args:          []string{"setup", "--no-completion"},
			expectError:   false,
			expectOutput:  []string{"✓ Setup complete!"},
			expectFiles:   []string{},
			expectNoFiles: []string{".config/wt/completion.bash", ".config/wt/completions/_wt"},
		},
		{
			name:         "setup default (with completion)",
			args:         []string{"setup"},
			expectError:  false,
			expectOutput: []string{"✓ Setup complete!"},
			expectFiles:  []string{}, // No files created with process substitution
		},
		{
			name:         "setup with invalid completion option",
			args:         []string{"setup", "--completion", "invalid"},
			expectError:  true,
			expectOutput: []string{"Error: invalid completion option 'invalid'"},
		},
		{
			name:         "setup with completion missing value",
			args:         []string{"setup", "--completion"},
			expectError:  true,
			expectOutput: []string{"Error: --completion requires a value"},
		},
		{
			name:         "setup with unknown option",
			args:         []string{"setup", "--unknown"},
			expectError:  true,
			expectOutput: []string{"Unknown setup option: --unknown"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()

			// Build test binary
			binaryPath, cleanup := createSetupTestBinary(t)
			defer cleanup()

			// Set up environment
			oldHome := os.Getenv("HOME")
			defer os.Setenv("HOME", oldHome)
			os.Setenv("HOME", tempDir)

			// Create .bashrc to satisfy shell config detection
			bashrcPath := filepath.Join(tempDir, ".bashrc")
			if err := os.WriteFile(bashrcPath, []byte("# test bashrc\n"), 0644); err != nil {
				t.Fatal(err)
			}

			// Run command
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Check error expectation
			hasError := err != nil
			if hasError != tt.expectError {
				t.Errorf("Expected error=%v, got error=%v. Output: %s", tt.expectError, hasError, outputStr)
			}

			// Check expected output strings
			for _, expected := range tt.expectOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, outputStr)
				}
			}

			// If error was expected, don't check files
			if tt.expectError {
				return
			}

			// Check expected files exist
			for _, file := range tt.expectFiles {
				path := filepath.Join(tempDir, file)
				if _, err := os.Stat(path); err != nil {
					t.Errorf("Expected file %s not found: %v", file, err)
				}
			}

			// Check expected files don't exist
			for _, file := range tt.expectNoFiles {
				path := filepath.Join(tempDir, file)
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("File %s should not exist but was found", file)
				}
			}
		})
	}
}

// TestSetupCompletionCheck removed - no longer relevant with process substitution approach

func TestSetupCompletionUsage(t *testing.T) {
	// Build test binary
	binaryPath, cleanup := createSetupTestBinary(t)
	defer cleanup()

	// Test that unknown setup option shows usage
	cmd := exec.Command(binaryPath, "setup", "--help")
	output, err := cmd.CombinedOutput()

	// Command should fail with unknown option
	if err == nil {
		t.Error("Expected setup --help to fail with unknown option error")
	}

	outputStr := string(output)

	// Check usage information is shown
	expectedUsageParts := []string{
		"Usage: wt setup [options]",
		"--completion <shell>",
		"--no-completion",
		"--check",
		"--uninstall",
		"Examples:",
		"wt setup                      # Install with auto-detected completion",
		"wt setup --completion bash",
		"wt setup --no-completion",
	}

	for _, expected := range expectedUsageParts {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Setup usage missing expected part: %s", expected)
		}
	}
}

func TestSetupCompletionInMainHelp(t *testing.T) {
	// Build test binary
	binaryPath, cleanup := createSetupTestBinary(t)
	defer cleanup()

	// Set temporary HOME to avoid interfering with real installation
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	defer os.Setenv("HOME", oldHome)
	os.Setenv("HOME", tempDir)

	// Test that main help mentions completion in setup command
	cmd := exec.Command(binaryPath, "help")
	output, _ := cmd.CombinedOutput()

	outputStr := string(output)

	// Check that setup command mentions completion
	if !strings.Contains(outputStr, "Install wt to ~/.local/bin with shell completion") {
		t.Error("Main help should mention completion in setup command description")
	}

	if !strings.Contains(outputStr, "--completion <shell>, --no-completion") {
		t.Error("Main help should mention completion options for setup command")
	}
}

// createSetupTestBinary builds the binary for setup testing
func createSetupTestBinary(t *testing.T) (string, func()) {
	t.Helper()
	tempDir := t.TempDir()

	// Build the actual binary for testing
	binaryPath := filepath.Join(tempDir, "wt-test")

	// Build command
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = getProjectDir()

	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return binaryPath, cleanup
}

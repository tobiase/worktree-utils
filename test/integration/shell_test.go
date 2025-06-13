package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tobiase/worktree-utils/test/helpers"
)

// TestShellWrapper tests the shell wrapper functionality
func TestShellWrapper(t *testing.T) {
	// Build the binary first
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	tests := []struct {
		name           string
		args           []string
		setup          func(t *testing.T) (string, func())
		expectedOutput string
		expectedType   string // "CD", "EXEC", "OUTPUT", "ERROR"
		wantError      bool
	}{
		{
			name: "CD prefix for go command",
			args: []string{"go", "main"},
			setup: func(t *testing.T) (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
					cleanup()
				}
			},
			expectedType: "CD",
		},
		{
			name: "EXEC prefix for venv command",
			args: []string{"venv"},
			setup: func(t *testing.T) (string, func()) {
				// Create a test repo with project config
				repo, cleanup := helpers.CreateTestRepo(t)
				
				// Create a temporary HOME directory
				tempHome := filepath.Join(os.TempDir(), "wt-test-home")
				os.MkdirAll(tempHome, 0755)
				
				// Create config directory structure
				configDir := filepath.Join(tempHome, ".config", "wt", "projects")
				os.MkdirAll(configDir, 0755)
				
				// Create project config with virtualenv
				// Resolve symlinks for macOS compatibility
				resolvedRepo, _ := filepath.EvalSymlinks(repo)
				projectConfig := `name: testproject
match:
  paths:
    - ` + resolvedRepo + `
    - ` + repo + `
virtualenv:
  name: .venv
  auto_commands: true
`
				configPath := filepath.Join(configDir, "testproject.yaml")
				os.WriteFile(configPath, []byte(projectConfig), 0644)
				
				// Create the virtualenv directory and activation script
				venvBin := filepath.Join(repo, ".venv", "bin")
				os.MkdirAll(venvBin, 0755)
				
				// Create a dummy activate script
				activateScript := filepath.Join(venvBin, "activate")
				os.WriteFile(activateScript, []byte("#!/bin/bash\n# Dummy activate script\n"), 0755)
				
				oldWd, _ := os.Getwd()
				oldHome := os.Getenv("HOME")
				os.Chdir(repo)
				os.Setenv("HOME", tempHome)
				
				return repo, func() {
					os.Chdir(oldWd)
					os.Setenv("HOME", oldHome)
					os.RemoveAll(tempHome)
					cleanup()
				}
			},
			expectedType:   "EXEC",
			expectedOutput: "source",
		},
		{
			name: "regular output for list command",
			args: []string{"list"},
			setup: func(t *testing.T) (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				os.Chdir(repo)
				return repo, func() {
					os.Chdir(oldWd)
					cleanup()
				}
			},
			expectedType:   "OUTPUT",
			expectedOutput: "Index",
		},
		{
			name: "error output for invalid command",
			args: []string{"invalid-command"},
			setup: func(t *testing.T) (string, func()) {
				return "", func() {}
			},
			expectedType: "ERROR",
			wantError:    true,
		},
		{
			name: "version command",
			args: []string{"version"},
			setup: func(t *testing.T) (string, func()) {
				return "", func() {}
			},
			expectedType:   "OUTPUT",
			expectedOutput: "dev", // Default version
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := tt.setup(t)
			defer cleanup()

			// Run the command
			cmd := exec.Command(binPath, tt.args...)
			// Set LANG=C to ensure consistent output across locales
			cmd.Env = append(os.Environ(), "LANG=C")
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check error status
			if (err != nil) != tt.wantError {
				t.Errorf("Command error = %v, wantError %v", err, tt.wantError)
			}

			// Get output
			output := stdout.String()
			errorOutput := stderr.String()
			

			// Verify output type
			switch tt.expectedType {
			case "CD":
				if !strings.HasPrefix(output, "CD:") {
					t.Errorf("Expected CD: prefix, got %q", output)
				}
			case "EXEC":
				if !strings.HasPrefix(output, "EXEC:") {
					t.Errorf("Expected EXEC: prefix, got %q", output)
				}
				if tt.expectedOutput != "" && !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
				}
			case "OUTPUT":
				if strings.HasPrefix(output, "CD:") || strings.HasPrefix(output, "EXEC:") {
					t.Errorf("Expected regular output, got special prefix: %q", output)
				}
				if tt.expectedOutput != "" && !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("Expected output to contain %q, got %q", tt.expectedOutput, output)
				}
			case "ERROR":
				if errorOutput == "" && output == "" {
					t.Error("Expected error output, got none")
				}
			}
		})
	}
}

// TestShellWrapperScript tests that the shell wrapper script handles prefixes correctly
func TestShellWrapperScript(t *testing.T) {
	// Skip if not on a Unix-like system
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}

	// Build test binary
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	// Create a mock binary that outputs specific strings
	mockBin := createMockBinary(t, map[string]mockResponse{
		"go main":     {output: "CD:/test/path", exitCode: 0},
		"venv":        {output: "EXEC:source .venv/bin/activate", exitCode: 0},
		"list":        {output: "Regular output", exitCode: 0},
		"error":       {output: "Error message", exitCode: 1, isError: true},
		"shell-init":  {output: getShellWrapper(), exitCode: 0},
	})
	defer os.Remove(mockBin)

	// Test the shell wrapper with different inputs
	tests := []struct {
		name         string
		command      string
		checkScript  string
		expectOutput string
		expectError  bool
	}{
		{
			name:    "CD prefix changes directory",
			command: "wt go main",
			checkScript: `pwd`,
			expectOutput: "/test/path",
		},
		{
			name:    "EXEC prefix executes command",
			command: "wt venv",
			checkScript: `echo "Command would execute: source .venv/bin/activate"`,
			expectOutput: "Command would execute",
		},
		{
			name:    "Regular output is printed",
			command: "wt list",
			checkScript: `true`,
			expectOutput: "Regular output",
		},
		{
			name:        "Errors go to stderr",
			command:     "wt error",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test script
			script := fmt.Sprintf(`#!/bin/bash
export WT_BIN="%s"

# Source the shell wrapper
eval "$($WT_BIN shell-init)"

# Run the command
%s

# Run check if provided
%s
`, mockBin, tt.command, tt.checkScript)

			scriptFile := filepath.Join(t.TempDir(), "test.sh")
			if err := os.WriteFile(scriptFile, []byte(script), 0755); err != nil {
				t.Fatal(err)
			}

			// Execute the script
			cmd := exec.Command("bash", scriptFile)
			// Set LANG=C to ensure consistent output across locales
			cmd.Env = append(os.Environ(), "LANG=C")
			output, err := cmd.CombinedOutput()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v\nOutput: %s", err, output)
				}
			}

			if tt.expectOutput != "" && !strings.Contains(string(output), tt.expectOutput) {
				t.Errorf("Expected output to contain %q, got %q", tt.expectOutput, string(output))
			}
		})
	}
}

// mockResponse defines what a mock command should return
type mockResponse struct {
	output   string
	exitCode int
	isError  bool
}

// createMockBinary creates a test binary that returns predefined responses
func createMockBinary(t *testing.T, responses map[string]mockResponse) string {
	t.Helper()

	mockCode := `package main

import (
	"fmt"
	"os"
	"strings"
)

var responses = map[string]struct{
	output   string
	exitCode int
	isError  bool
}{
`
	// Add responses to the code
	for cmd, resp := range responses {
		mockCode += fmt.Sprintf("\t%q: {output: %q, exitCode: %d, isError: %v},\n", 
			cmd, resp.output, resp.exitCode, resp.isError)
	}

	mockCode += `}

func main() {
	args := strings.Join(os.Args[1:], " ")
	
	if resp, ok := responses[args]; ok {
		if resp.isError {
			fmt.Fprint(os.Stderr, resp.output)
		} else {
			fmt.Print(resp.output)
		}
		os.Exit(resp.exitCode)
	}
	
	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args)
	os.Exit(1)
}
`

	// Write and compile the mock
	mockDir := t.TempDir()
	mockFile := filepath.Join(mockDir, "mock.go")
	if err := os.WriteFile(mockFile, []byte(mockCode), 0644); err != nil {
		t.Fatal(err)
	}

	mockBin := filepath.Join(mockDir, "mock-wt")
	cmd := exec.Command("go", "build", "-o", mockBin, mockFile)
	// Set LANG=C for consistent build output
	cmd.Env = append(os.Environ(), "LANG=C")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}

	return mockBin
}


// getShellWrapper returns the shell wrapper script (matching main.go)
func getShellWrapper() string {
	return `# Shell function to handle CD: and EXEC: prefixes
wt() {
  output=$("${WT_BIN:-wt-bin}" "$@" 2>&1)
  exit_code=$?
  
  if [ $exit_code -eq 0 ]; then
    if [[ "$output" == "CD:"* ]]; then
      cd "${output#CD:}"
    elif [[ "$output" == "EXEC:"* ]]; then
      eval "${output#EXEC:}"
    else
      [ -n "$output" ] && echo "$output"
    fi
  else
    echo "$output" >&2
    return $exit_code
  fi
}
`
}
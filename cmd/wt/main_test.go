package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tobiase/worktree-utils/internal/config"
	"github.com/tobiase/worktree-utils/test/helpers"
)

// captureOutput captures stdout and stderr during function execution
func captureOutput(f func() error) (stdout, stderr string, err error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr

	err = f()

	wOut.Close()
	wErr.Close()

	var bufOut, bufErr bytes.Buffer
	_, _ = io.Copy(&bufOut, rOut)
	_, _ = io.Copy(&bufErr, rErr)

	return bufOut.String(), bufErr.String(), err
}

func TestHandleVersionCommand(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		buildTime   string
		wantContain string
	}{
		{
			name:        "with version and build time",
			version:     "1.2.3",
			buildTime:   "2024-01-01",
			wantContain: "wt version 1.2.3 (built 2024-01-01)",
		},
		{
			name:        "development version",
			version:     "dev",
			buildTime:   "",
			wantContain: "wt version dev",
		},
		{
			name:        "empty version",
			version:     "",
			buildTime:   "",
			wantContain: "wt version dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version = tt.version
			buildTime = tt.buildTime

			stdout, _, err := captureOutput(func() error {
				handleVersionCommand([]string{})
				return nil
			})

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(stdout, tt.wantContain) {
				t.Errorf("expected output to contain %q, got %q", tt.wantContain, stdout)
			}
		})
	}
}

func TestShellInitCommand(t *testing.T) {
	// Initialize config manager
	configMgr := &config.Manager{}

	stdout, _, err := captureOutput(func() error {
		runCommand("shell-init", []string{}, configMgr)
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should output shell initialization script
	if !strings.Contains(stdout, "wt()") {
		t.Error("expected shell function definition")
	}
	if !strings.Contains(stdout, "CD:") {
		t.Error("expected CD: prefix handling")
	}
	if !strings.Contains(stdout, "EXEC:") {
		t.Error("expected EXEC: prefix handling")
	}
}

func TestHandleCompletionCommand(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		wantShell string
	}{
		{
			name:      "bash completion",
			args:      []string{"bash"},
			wantShell: "bash",
		},
		{
			name:      "zsh completion",
			args:      []string{"zsh"},
			wantShell: "zsh",
		},
		{
			name:      "no shell specified",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "unsupported shell",
			args:      []string{"fish"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that would cause exit - test them in integration tests
			if tt.wantError && (len(tt.args) == 0 || (len(tt.args) > 0 && tt.args[0] == "fish")) {
				t.Skip("Skipping test that exits - covered by integration tests")
				return
			}

			// Initialize config manager
			configMgr := &config.Manager{}

			stdout, stderr, _ := captureOutput(func() error {
				handleCompletionCommand(tt.args, configMgr)
				return nil
			})

			if tt.wantError && stderr == "" {
				t.Error("expected error output")
			}
			if !tt.wantError && stderr != "" {
				t.Errorf("unexpected error: %s", stderr)
			}
			if !tt.wantError && tt.wantShell != "" {
				if !strings.Contains(stdout, tt.wantShell) {
					t.Errorf("expected %s completion output", tt.wantShell)
				}
			}
		})
	}
}

func TestHandleProjectCommand(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	// Create a git repo
	gitDir := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(gitDir); err != nil {
		t.Fatal(err)
	}
	_, _, err := helpers.RunCommand(t, "git", "init")
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = helpers.RunCommand(t, "git", "remote", "add", "origin", "https://github.com/test/repo.git")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "init project",
			args:      []string{"init", "testproject"},
			wantError: false,
		},
		{
			name:      "no subcommand",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "help flag",
			args:      []string{"--help"},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set config dir to temp
			origConfigDir := os.Getenv("XDG_CONFIG_HOME")
			os.Setenv("XDG_CONFIG_HOME", tempDir)
			defer os.Setenv("XDG_CONFIG_HOME", origConfigDir)

			// Initialize config manager
			configMgr := &config.Manager{}

			// Save original osExit and capture exit code
			oldExit := osExit
			exitCode := -1
			osExit = func(code int) {
				exitCode = code
			}
			defer func() {
				osExit = oldExit
			}()

			stdout, stderr, _ := captureOutput(func() error {
				handleProjectCommand(tt.args, configMgr)
				return nil
			})

			if tt.wantError {
				if exitCode != 1 && stderr == "" {
					t.Error("expected error exit or error output")
				}
			} else {
				if exitCode == 1 || (stderr != "" && (len(tt.args) == 0 || !strings.Contains(tt.args[0], "help"))) {
					t.Errorf("unexpected error: exit=%d, stderr=%s", exitCode, stderr)
				}
			}
			if len(tt.args) > 0 && tt.args[0] == "init" && !tt.wantError {
				if !strings.Contains(stdout, "Project 'testproject' initialized") {
					t.Error("expected success message")
				}
			}
		})
	}
}

func TestResolveCommandAlias(t *testing.T) {
	tests := []struct {
		cmd      string
		expected string
	}{
		{"ls", "list"},
		{"switch", "go"},
		{"s", "go"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			result := resolveCommandAlias(tt.cmd)
			if result != tt.expected {
				t.Errorf("resolveCommandAlias(%q) = %q, want %q", tt.cmd, result, tt.expected)
			}
		})
	}
}

// Remove this test as it tests internal implementation details
// The functionality is covered by integration tests

func TestHandleEnvCommand(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	// Setup test worktree structure
	mainRepo := filepath.Join(tempDir, "test-repo")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(mainRepo); err != nil {
		t.Fatal(err)
	}
	_, _, err := helpers.RunCommand(t, "git", "init")
	if err != nil {
		t.Fatal(err)
	}

	// Create .env file
	envContent := "TEST_VAR=123\n"
	if err := os.WriteFile(".env", []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name      string
		args      []string
		wantError bool
	}{
		{
			name:      "no subcommand shows usage",
			args:      []string{},
			wantError: true,
		},
		{
			name:      "help flag",
			args:      []string{"--help"},
			wantError: false,
		},
		{
			name:      "sync subcommand needs target",
			args:      []string{"sync"},
			wantError: true,
		},
		{
			name:      "list subcommand",
			args:      []string{"list"},
			wantError: false,
		},
		{
			name:      "invalid subcommand",
			args:      []string{"invalid"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original osExit
			oldExit := osExit
			exitCode := -1
			osExit = func(code int) {
				exitCode = code
			}
			defer func() {
				osExit = oldExit
			}()

			stdout, stderr, _ := captureOutput(func() error {
				handleEnvCommand(tt.args)
				return nil
			})

			if tt.wantError {
				if exitCode != 1 && stderr == "" {
					t.Error("expected error exit or error output")
				}
			} else {
				if exitCode == 1 || (stderr != "" && len(tt.args) > 0 && !strings.Contains(tt.args[0], "help")) {
					t.Errorf("unexpected error: exit=%d, stderr=%s", exitCode, stderr)
				}
			}
			if len(tt.args) > 0 && tt.args[0] == "list" {
				if !strings.Contains(stdout, ".env") {
					t.Error("expected .env file in list output")
				}
			}
			if tt.wantError && exitCode != 1 {
				t.Errorf("expected exit code 1, got %d", exitCode)
			}
		})
	}
}

func TestPrintErrorAndExit(t *testing.T) {
	// Save original os.Exit
	oldExit := osExit
	exitCode := -1
	osExit = func(code int) {
		exitCode = code
	}
	defer func() {
		osExit = oldExit
	}()

	tests := []struct {
		name     string
		format   string
		args     []interface{}
		wantExit int
		wantMsg  string
	}{
		{
			name:     "simple error",
			format:   "test error",
			args:     []interface{}{},
			wantExit: 1,
			wantMsg:  "wt: test error\n",
		},
		{
			name:     "formatted error",
			format:   "error: %s",
			args:     []interface{}{"failed"},
			wantExit: 1,
			wantMsg:  "wt: error: failed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitCode = -1
			_, stderr, _ := captureOutput(func() error {
				printErrorAndExit(tt.format, tt.args...)
				return nil
			})

			if exitCode != tt.wantExit {
				t.Errorf("expected exit code %d, got %d", tt.wantExit, exitCode)
			}
			if stderr != tt.wantMsg {
				t.Errorf("expected stderr %q, got %q", tt.wantMsg, stderr)
			}
		})
	}
}

// Test help flag handling across commands
func TestHelpFlagHandling(t *testing.T) {
	commands := []struct {
		name    string
		handler func([]string)
		args    []string
	}{
		{"list", handleListCommand, []string{"--help"}},
		{"list", handleListCommand, []string{"-h"}},
		{"new", func(args []string) { handleNewCommand(args, &config.Manager{}) }, []string{"--help"}},
		{"new", func(args []string) { handleNewCommand(args, &config.Manager{}) }, []string{"-h"}},
		{"go", handleGoCommand, []string{"--help"}},
		{"go", handleGoCommand, []string{"-h"}},
		{"rm", handleRemoveCommand, []string{"--help"}},
		{"rm", handleRemoveCommand, []string{"-h"}},
		{"setup", handleSetupCommand, []string{"--help"}},
		{"env", handleEnvCommand, []string{"--help"}},
		{"project", func(args []string) { handleProjectCommand(args, &config.Manager{}) }, []string{"--help"}},
		{"completion", func(args []string) { handleCompletionCommand(args, &config.Manager{}) }, []string{"--help"}},
		{"update", handleUpdateCommand, []string{"--help"}},
	}

	for _, cmd := range commands {
		t.Run(fmt.Sprintf("%s_%s", cmd.name, cmd.args[0]), func(t *testing.T) {
			// Help content is initialized in init()

			stdout, stderr, _ := captureOutput(func() error {
				cmd.handler(cmd.args)
				return nil
			})

			// Should show help without error
			if stderr != "" && !strings.Contains(stderr, "usage:") {
				t.Errorf("unexpected error for %s: %s", cmd.name, stderr)
			}

			// Should contain command name or usage info
			output := stdout + stderr
			if !strings.Contains(strings.ToLower(output), cmd.name) &&
				!strings.Contains(strings.ToLower(output), "usage") {
				t.Errorf("help output should contain command name or usage for %s", cmd.name)
			}
		})
	}
}

func TestHandleNewCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectError    bool
		expectCDOutput bool
		expectedOutput string
	}{
		{
			name:        "no arguments shows usage",
			args:        []string{},
			expectError: true,
		},
		{
			name:           "help flag shows help",
			args:           []string{"--help"},
			expectError:    false,
			expectCDOutput: false,
		},
		{
			name:           "create new branch without --no-switch outputs CD:",
			args:           []string{"test-branch"},
			expectError:    false,
			expectCDOutput: true,
		},
		{
			name:           "create new branch with --no-switch outputs path",
			args:           []string{"test-branch", "--no-switch"},
			expectError:    false,
			expectCDOutput: false,
			expectedOutput: "Created worktree at",
		},
		{
			name:           "create new branch with base branch",
			args:           []string{"test-branch", "--base", "main"},
			expectError:    false,
			expectCDOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that would actually create worktrees
			if !tt.expectError && tt.name != "help flag shows help" {
				t.Skip("Skipping test that would create actual worktree")
			}

			// Capture output
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			// Capture exit code
			oldExit := osExit
			exitCode := -1
			osExit = func(code int) {
				exitCode = code
			}

			// Run command
			handleNewCommand(tt.args, &config.Manager{})

			// Restore
			osExit = oldExit
			w.Close()
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			output, _ := io.ReadAll(r)
			outputStr := string(output)

			// Check expectations
			if tt.expectError && exitCode != 1 {
				t.Errorf("Expected exit code 1, got %d", exitCode)
			}

			if tt.expectCDOutput && !strings.Contains(outputStr, "CD:") {
				t.Errorf("Expected CD: in output, got: %s", outputStr)
			}

			if tt.expectedOutput != "" && !strings.Contains(outputStr, tt.expectedOutput) {
				t.Errorf("Expected '%s' in output, got: %s", tt.expectedOutput, outputStr)
			}
		})
	}
}

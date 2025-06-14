package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCompletionCommand(t *testing.T) {
	// Build the binary for testing
	binaryPath, cleanup := createCompletionTestBinary(t)
	defer cleanup()

	testCases := []struct {
		name        string
		args        []string
		expectError bool
		contains    []string
	}{
		{
			name:        "bash completion",
			args:        []string{"completion", "bash"},
			expectError: false,
			contains:    []string{"#!/bin/bash", "_wt_completion", "complete -F"},
		},
		{
			name:        "zsh completion",
			args:        []string{"completion", "zsh"},
			expectError: false,
			contains:    []string{"#compdef wt", "_wt()", "_arguments"},
		},
		{
			name:        "no arguments",
			args:        []string{"completion"},
			expectError: true,
			contains:    []string{"Usage:", "bash|zsh"},
		},
		{
			name:        "unsupported shell",
			args:        []string{"completion", "fish"},
			expectError: true,
			contains:    []string{"unsupported shell", "fish"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			cmd.Env = []string{"HOME=" + t.TempDir()}

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			hasError := err != nil
			if hasError != tc.expectError {
				t.Errorf("Expected error=%v, got error=%v. Output: %s", tc.expectError, hasError, outputStr)
			}

			for _, expected := range tc.contains {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't. Output: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestCompletionBashGeneration(t *testing.T) {
	// Build the binary for testing
	binaryPath, cleanup := createCompletionTestBinary(t)
	defer cleanup()

	cmd := exec.Command(binaryPath, "completion", "bash")
	cmd.Env = []string{"HOME=" + t.TempDir()}

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to generate bash completion: %v", err)
	}

	completion := string(output)

	// Test structure and essential components
	requiredComponents := []string{
		"#!/bin/bash",
		"_wt_completion()",
		"_init_completion",
		"complete -F _wt_completion wt",
		"_wt_complete_branches",
		"COMPREPLY=",
		"compgen -W",
	}

	for _, component := range requiredComponents {
		if !strings.Contains(completion, component) {
			t.Errorf("Bash completion missing required component: %s", component)
		}
	}

	// Test that all main commands are included
	expectedCommands := []string{
		"list", "add", "rm", "go", "new", "env-copy",
		"project", "setup", "update", "version", "help", "completion",
	}

	commandList := extractCommandList(completion)
	for _, cmd := range expectedCommands {
		if !strings.Contains(commandList, cmd) {
			t.Errorf("Bash completion missing command: %s", cmd)
		}
	}

	// Test aliases are included
	expectedAliases := []string{"ls", "switch", "s"}
	for _, alias := range expectedAliases {
		if !strings.Contains(commandList, alias) {
			t.Errorf("Bash completion missing alias: %s", alias)
		}
	}
}

func TestCompletionZshGeneration(t *testing.T) {
	// Build the binary for testing
	binaryPath, cleanup := createCompletionTestBinary(t)
	defer cleanup()

	cmd := exec.Command(binaryPath, "completion", "zsh")
	cmd.Env = []string{"HOME=" + t.TempDir()}

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to generate zsh completion: %v", err)
	}

	completion := string(output)

	// Test structure and essential components
	requiredComponents := []string{
		"#compdef wt",
		"_wt()",
		"_arguments -C",
		"_wt_commands()",
		"_wt_branches()",
		"_describe",
	}

	for _, component := range requiredComponents {
		if !strings.Contains(completion, component) {
			t.Errorf("Zsh completion missing required component: %s", component)
		}
	}

	// Test command descriptions are included
	expectedDescriptions := []string{
		"List all worktrees",
		"Add a new worktree",
		"Remove a worktree",
		"Switch to a worktree",
		"Generate shell completion scripts",
	}

	for _, desc := range expectedDescriptions {
		if !strings.Contains(completion, desc) {
			t.Errorf("Zsh completion missing description: %s", desc)
		}
	}
}

func TestCompletionHelp(t *testing.T) {
	// Build the binary for testing
	binaryPath, cleanup := createCompletionTestBinary(t)
	defer cleanup()

	cmd := exec.Command(binaryPath, "completion")
	cmd.Env = []string{"HOME=" + t.TempDir()}

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Should exit with error when no shell specified
	if err == nil {
		t.Error("Expected completion command to fail when no shell specified")
	}

	// Should show helpful usage information
	expectedHelpParts := []string{
		"Usage: wt completion <bash|zsh>",
		"Generate shell completion scripts",
		"Examples:",
		"wt completion bash",
		"wt completion zsh",
		"eval",
	}

	for _, part := range expectedHelpParts {
		if !strings.Contains(outputStr, part) {
			t.Errorf("Completion help missing expected part: %s", part)
		}
	}
}

func TestCompletionInUsage(t *testing.T) {
	// Build the binary for testing
	binaryPath, cleanup := createCompletionTestBinary(t)
	defer cleanup()

	cmd := exec.Command(binaryPath, "help")
	cmd.Env = []string{"HOME=" + t.TempDir()}

	output, _ := cmd.CombinedOutput()
	outputStr := string(output)

	// help command might exit with non-zero, that's OK

	// Should include completion in usage
	if !strings.Contains(outputStr, "completion") {
		t.Error("Help output should mention completion command")
	}

	if !strings.Contains(outputStr, "shell completion scripts") {
		t.Error("Help output should describe completion command")
	}
}

func TestCompletionWithProjectContext(t *testing.T) {
	// This test verifies that completion generation works even without a project context
	// In a real git repository with project configuration, it would include project commands

	// Build the binary for testing
	binaryPath, cleanup := createCompletionTestBinary(t)
	defer cleanup()

	cmd := exec.Command(binaryPath, "completion", "bash")
	cmd.Env = []string{"HOME=" + t.TempDir()}

	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to generate completion with project context: %v", err)
	}

	completion := string(output)

	// Should generate valid completion even without project
	if !strings.Contains(completion, "_wt_completion") {
		t.Error("Should generate valid completion function")
	}

	// Should handle the case where no project commands exist gracefully
	if !strings.Contains(completion, "complete -F _wt_completion wt") {
		t.Error("Should register completion function")
	}
}

// Helper function to extract the command list from bash completion
func extractCommandList(completion string) string {
	lines := strings.Split(completion, "\n")
	for _, line := range lines {
		if strings.Contains(line, "local commands=") {
			start := strings.Index(line, "\"")
			end := strings.LastIndex(line, "\"")
			if start != -1 && end != -1 && end > start {
				return line[start+1 : end]
			}
		}
	}
	return ""
}

// createCompletionTestBinary builds the binary for completion testing
func createCompletionTestBinary(t *testing.T) (string, func()) {
	t.Helper()
	tempDir := t.TempDir()

	// Build the actual binary for testing
	binaryPath := tempDir + "/wt-test"

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

const projectDir = "/Users/tobias/Projects/worktree-utils/cmd/wt"

func getProjectDir() string {
	return projectDir
}

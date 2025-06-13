package integration

import (
	"os"
	"strings"
	"testing"

	"github.com/tobiase/worktree-utils/test/helpers"
)

// TestCommandAliases tests that command aliases work correctly
func TestCommandAliases(t *testing.T) {
	// Build the binary
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	// Create test repository
	repo, cleanup := helpers.CreateTestRepo(t)
	defer cleanup()

	// Change to repo directory
	oldWd, _ := os.Getwd()
	_ = os.Chdir(repo)
	defer func() { _ = os.Chdir(oldWd) }()

	// Create a test worktree
	_, _ = helpers.AddTestWorktree(t, repo, "test-branch")

	tests := []struct {
		name         string
		args         []string
		expectedLike []string // Command it should behave like
		checkOutput  func(t *testing.T, output string)
	}{
		{
			name:         "ls alias for list",
			args:         []string{"ls"},
			expectedLike: []string{"list"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "Index") || !strings.Contains(output, "Branch") {
					t.Error("Expected list output with Index and Branch headers")
				}
				if !strings.Contains(output, "main") {
					t.Error("Expected to see main branch")
				}
			},
		},
		{
			name:         "switch alias for go",
			args:         []string{"switch", "test-branch"},
			expectedLike: []string{"go", "test-branch"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "CD:") {
					t.Error("Expected CD: prefix for switch command")
				}
				if !strings.Contains(output, "test-branch") {
					t.Error("Expected path to contain test-branch")
				}
			},
		},
		{
			name:         "s alias for go",
			args:         []string{"s", "main"},
			expectedLike: []string{"go", "main"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "CD:") {
					t.Error("Expected CD: prefix for s command")
				}
			},
		},
		{
			name:         "switch with no args",
			args:         []string{"switch"},
			expectedLike: []string{"go"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "CD:") {
					t.Error("Expected CD: prefix")
				}
				// Should go to repo root
				if !strings.Contains(output, repo) {
					t.Error("Expected to go to repo root")
				}
			},
		},
		{
			name:         "s with index",
			args:         []string{"s", "0"},
			expectedLike: []string{"go", "0"},
			checkOutput: func(t *testing.T, output string) {
				if !strings.Contains(output, "CD:") {
					t.Error("Expected CD: prefix")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := runCommand(t, binPath, tt.args...)

			if tt.checkOutput != nil {
				tt.checkOutput(t, output)
			}

			// If we have an expected equivalent command, run it and compare
			if len(tt.expectedLike) > 0 {
				expectedOutput := runCommand(t, binPath, tt.expectedLike...)

				// For commands that output paths, just check the prefix
				if strings.HasPrefix(output, "CD:") && strings.HasPrefix(expectedOutput, "CD:") {
					// Both are CD commands, that's enough
					return
				}

				// For other commands, outputs should be similar
				if !strings.Contains(output, "CD:") && output != expectedOutput {
					t.Errorf("Alias output differs from original command.\nAlias: %s\nOriginal: %s", output, expectedOutput)
				}
			}
		})
	}
}

// TestAliasesInHelp verifies that aliases are shown in help text
func TestAliasesInHelp(t *testing.T) {
	// Build the binary
	binPath := buildTestBinary(t)
	defer os.Remove(binPath)

	output := runCommand(t, binPath, "help")

	// Check that aliases are mentioned
	aliasesInHelp := []string{
		"ls",
		"switch",
		"s",
	}

	for _, alias := range aliasesInHelp {
		if !strings.Contains(output, alias) {
			t.Errorf("Expected help text to mention alias %q", alias)
		}
	}
}

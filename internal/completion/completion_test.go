package completion

import (
	"strings"
	"testing"

	"github.com/tobiase/worktree-utils/internal/config"
)

func TestGetCompletionData(t *testing.T) {
	data := GetCompletionData(nil)

	// Test core commands are present
	expectedCommands := []string{
		"list", "add", "rm", "go", "new", "env-copy",
		"project", "setup", "update", "version", "help",
		"completion", "shell-init",
	}

	if len(data.Commands) < len(expectedCommands) {
		t.Errorf("Expected at least %d commands, got %d", len(expectedCommands), len(data.Commands))
	}

	commandNames := make(map[string]bool)
	for _, cmd := range data.Commands {
		commandNames[cmd.Name] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Expected command '%s' not found", expected)
		}
	}

	// Test aliases
	expectedAliases := map[string]string{
		"ls":     "list",
		"switch": "go",
		"s":      "go",
	}

	for alias, target := range expectedAliases {
		if data.Aliases[alias] != target {
			t.Errorf("Expected alias '%s' -> '%s', got '%s'", alias, target, data.Aliases[alias])
		}
	}
}

func TestGetCommandByName(t *testing.T) {
	data := GetCompletionData(nil)

	testCases := []struct {
		name        string
		expectFound bool
	}{
		{"list", true},
		{"ls", true}, // alias
		{"go", true},
		{"switch", true}, // alias
		{"s", true},      // alias
		{"nonexistent", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := data.GetCommandByName(tc.name)
			found := cmd != nil

			if found != tc.expectFound {
				t.Errorf("Expected found=%v for command '%s', got %v", tc.expectFound, tc.name, found)
			}

			// For aliases, verify they resolve to the correct command
			if found && tc.name != cmd.Name {
				expectedTarget := data.Aliases[tc.name]
				if cmd.Name != expectedTarget {
					t.Errorf("Alias '%s' should resolve to '%s', got '%s'", tc.name, expectedTarget, cmd.Name)
				}
			}
		})
	}
}

func TestGetAllCommandNames(t *testing.T) {
	data := GetCompletionData(nil)
	names := data.GetAllCommandNames()

	// Should include core commands
	expectedNames := []string{"list", "add", "rm", "go", "new", "completion"}
	for _, expected := range expectedNames {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected command name '%s' not found in GetAllCommandNames()", expected)
		}
	}

	// Should include aliases
	expectedAliases := []string{"ls", "switch", "s"}
	for _, expected := range expectedAliases {
		found := false
		for _, name := range names {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected alias '%s' not found in GetAllCommandNames()", expected)
		}
	}
}

func TestNormalizeCommandName(t *testing.T) {
	data := GetCompletionData(nil)

	testCases := []struct {
		input    string
		expected string
	}{
		{"list", "list"},
		{"ls", "list"},
		{"go", "go"},
		{"switch", "go"},
		{"s", "go"},
		{"nonexistent", "nonexistent"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := data.NormalizeCommandName(tc.input)
			if result != tc.expected {
				t.Errorf("Expected NormalizeCommandName('%s') = '%s', got '%s'", tc.input, tc.expected, result)
			}
		})
	}
}

func TestSplitCompletionLine(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{"wt list", []string{"wt", "list"}},
		{"wt go branch", []string{"wt", "go", "branch"}},
		{"wt new test --base main", []string{"wt", "new", "test", "--base", "main"}},
		{"", []string{}},
		{"   wt   list   ", []string{"wt", "list"}},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := SplitCompletionLine(tc.input)
			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d parts, got %d: %v", len(tc.expected), len(result), result)
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("Expected part %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestGenerateBashCompletion(t *testing.T) {
	data := GetCompletionData(nil)
	bash := GenerateBashCompletion(nil)

	// Test basic structure
	requiredParts := []string{
		"#!/bin/bash",
		"_wt_completion()",
		"_init_completion",
		"complete -F _wt_completion wt",
		"_wt_complete_branches()",
	}

	for _, part := range requiredParts {
		if !strings.Contains(bash, part) {
			t.Errorf("Bash completion missing required part: %s", part)
		}
	}

	// Test that all commands are included
	for _, cmd := range data.Commands {
		if !strings.Contains(bash, cmd.Name) {
			t.Errorf("Bash completion missing command: %s", cmd.Name)
		}
	}

	// Test that aliases are included
	for alias := range data.Aliases {
		if !strings.Contains(bash, alias) {
			t.Errorf("Bash completion missing alias: %s", alias)
		}
	}
}

func TestGenerateZshCompletion(t *testing.T) {
	data := GetCompletionData(nil)
	zsh := GenerateZshCompletion(nil)

	// Test basic structure
	requiredParts := []string{
		"#compdef wt",
		"_wt()",
		"_arguments -C",
		"_wt_commands()",
		"_wt_branches()",
	}

	for _, part := range requiredParts {
		if !strings.Contains(zsh, part) {
			t.Errorf("Zsh completion missing required part: %s", part)
		}
	}

	// Test that commands are included with descriptions
	for _, cmd := range data.Commands {
		// Check command name is present
		if !strings.Contains(zsh, cmd.Name) {
			t.Errorf("Zsh completion missing command: %s", cmd.Name)
		}
		// Check description is present
		if cmd.Description != "" && !strings.Contains(zsh, cmd.Description) {
			t.Errorf("Zsh completion missing description for %s: %s", cmd.Name, cmd.Description)
		}
	}
}

func TestCommandFlags(t *testing.T) {
	data := GetCompletionData(nil)

	// Test specific commands have expected flags
	testCases := []struct {
		command       string
		expectedFlags []string
	}{
		{"new", []string{"--base"}},
		{"env-copy", []string{"--recursive"}},
		{"setup", []string{"--check", "--uninstall"}},
		{"update", []string{"--check", "--force"}},
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			cmd := data.GetCommandByName(tc.command)
			if cmd == nil {
				t.Fatalf("Command '%s' not found", tc.command)
			}

			flagNames := make([]string, len(cmd.Flags))
			for i, flag := range cmd.Flags {
				flagNames[i] = flag.Name
			}

			for _, expectedFlag := range tc.expectedFlags {
				found := false
				for _, flag := range flagNames {
					if flag == expectedFlag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Command '%s' missing expected flag: %s", tc.command, expectedFlag)
				}
			}
		})
	}
}

func TestCommandArguments(t *testing.T) {
	data := GetCompletionData(nil)

	// Test commands that should have worktree branch arguments
	worktreeBranchCommands := []string{"rm", "go", "env-copy"}
	for _, cmdName := range worktreeBranchCommands {
		t.Run(cmdName, func(t *testing.T) {
			cmd := data.GetCommandByName(cmdName)
			if cmd == nil {
				t.Fatalf("Command '%s' not found", cmdName)
			}

			if len(cmd.Args) == 0 {
				t.Errorf("Command '%s' should have arguments", cmdName)
				return
			}

			firstArg := cmd.Args[0]
			if firstArg.Type != ArgWorktreeBranch && cmdName != "go" { // go is optional
				t.Errorf("Command '%s' first argument should be worktree branch type, got %v", cmdName, firstArg.Type)
			}
		})
	}
}

func TestGetCompletionCandidates(t *testing.T) {
	// Create test data with mock branches
	data := &CompletionData{
		AvailableBranches: []string{"main", "feature-1", "feature-2"},
		ProjectCommands:   []string{"api", "frontend", "backend"},
	}

	testCases := []struct {
		argType  ArgumentType
		expected []string
	}{
		{ArgBranch, []string{"main", "feature-1", "feature-2"}},
		{ArgProject, []string{"api", "frontend", "backend"}},
		{ArgString, []string{}},
		{ArgFile, []string{}},
		// Note: ArgWorktreeBranch is tested separately since it calls git commands
	}

	for _, tc := range testCases {
		t.Run(string(rune(tc.argType)), func(t *testing.T) {
			result := data.GetCompletionCandidates([]string{}, tc.argType)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d candidates, got %d", len(tc.expected), len(result))
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("Expected candidate %d to be '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}

	// Test ArgWorktreeBranch separately (may fail if not in git repo)
	t.Run("ArgWorktreeBranch", func(t *testing.T) {
		result := data.GetCompletionCandidates([]string{}, ArgWorktreeBranch)
		// In test environment, this might return empty list due to git command failure
		// That's expected and acceptable
		if result == nil {
			t.Error("Expected slice, got nil")
		}
	})
}

func TestParseWorktreeBranches(t *testing.T) {
	testOutput := `worktree /Users/user/repo
HEAD 1234567

worktree /Users/user/repo-worktrees/feature-1
HEAD 2345678
branch refs/heads/feature-1

worktree /Users/user/repo-worktrees/feature-2
HEAD 3456789
branch refs/heads/feature-2

worktree /Users/user/repo-worktrees/main-copy
HEAD 4567890
branch refs/heads/main
`

	result := parseWorktreeBranches(testOutput)
	expected := []string{"feature-1", "feature-2", "main"}

	if len(result) != len(expected) {
		t.Errorf("Expected %d branches, got %d: %v", len(expected), len(result), result)
		return
	}

	for i, expectedBranch := range expected {
		if result[i] != expectedBranch {
			t.Errorf("Expected branch %d to be '%s', got '%s'", i, expectedBranch, result[i])
		}
	}
}

func TestProjectCommandsIntegration(t *testing.T) {
	// Create a mock config manager with project commands
	configMgr, err := config.NewManager()
	if err != nil {
		t.Fatal(err)
	}

	// For testing, we'll manually set the current project
	// This is a bit of a hack since we can't easily create a full project setup
	data := GetCompletionData(configMgr)

	// Verify that completion data structure supports project commands
	// Note: ProjectCommands will be empty slice when no project is loaded, which is expected

	// Test completion generation includes project command handling
	bash := GenerateBashCompletion(configMgr)
	if !strings.Contains(bash, "_wt_complete_project_commands") && len(data.ProjectCommands) == 0 {
		t.Log("No project commands found (expected when no project is loaded)")
	}

	zsh := GenerateZshCompletion(configMgr)
	if !strings.Contains(zsh, "_wt_commands") {
		t.Error("Zsh completion should contain command completion function")
	}

	// Verify the structure can handle project commands when they exist
	testData := &CompletionData{
		Commands:        data.Commands,
		Aliases:         data.Aliases,
		ProjectCommands: []string{"api", "frontend", "backend"},
	}

	projectNames := testData.GetAllCommandNames()
	expectedProjectCommands := []string{"api", "frontend", "backend"}

	for _, expected := range expectedProjectCommands {
		found := false
		for _, name := range projectNames {
			if name == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected project command '%s' in GetAllCommandNames()", expected)
		}
	}
}

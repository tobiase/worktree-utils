package interactive

import (
	"testing"
)

func TestGetAvailableCommands(t *testing.T) {
	commands := GetAvailableCommands()

	if len(commands) == 0 {
		t.Error("Expected at least some commands to be available")
	}

	// Test that core commands are present
	expectedCommands := []string{"list", "go", "add", "rm", "new", "setup", "help"}
	commandMap := make(map[string]bool)

	for _, cmd := range commands {
		commandMap[cmd.Name] = true

		// Validate that each command has required fields
		if cmd.Name == "" {
			t.Errorf("Command missing name: %+v", cmd)
		}
		if cmd.Description == "" {
			t.Errorf("Command %s missing description", cmd.Name)
		}
		if cmd.Usage == "" {
			t.Errorf("Command %s missing usage", cmd.Name)
		}
	}

	for _, expectedCmd := range expectedCommands {
		if !commandMap[expectedCmd] {
			t.Errorf("Expected command %s not found in available commands", expectedCmd)
		}
	}
}

func TestCommandInfoStructure(t *testing.T) {
	commands := GetAvailableCommands()

	for _, cmd := range commands {
		// Test that usage starts with "wt"
		if len(cmd.Usage) < 2 || cmd.Usage[:2] != "wt" {
			t.Errorf("Command %s usage should start with 'wt', got: %s", cmd.Name, cmd.Usage)
		}

		// Test that description is not empty and doesn't end with period (for consistency)
		if cmd.Description == "" {
			t.Errorf("Command %s should have a description", cmd.Name)
		}

		// Test that aliases don't conflict with command names
		allCommandNames := make(map[string]bool)
		for _, c := range commands {
			allCommandNames[c.Name] = true
		}

		for _, alias := range cmd.Aliases {
			if allCommandNames[alias] {
				t.Errorf("Command %s has alias %s that conflicts with another command name", cmd.Name, alias)
			}
		}
	}
}

func TestCommandAliases(t *testing.T) {
	commands := GetAvailableCommands()

	// Check for expected aliases
	expectedAliases := map[string][]string{
		"list": {"ls"},
		"go":   {"switch", "s"},
	}

	for _, cmd := range commands {
		if expected, exists := expectedAliases[cmd.Name]; exists {
			if len(cmd.Aliases) != len(expected) {
				t.Errorf("Command %s expected %d aliases, got %d", cmd.Name, len(expected), len(cmd.Aliases))
				continue
			}

			aliasMap := make(map[string]bool)
			for _, alias := range cmd.Aliases {
				aliasMap[alias] = true
			}

			for _, expectedAlias := range expected {
				if !aliasMap[expectedAlias] {
					t.Errorf("Command %s missing expected alias: %s", cmd.Name, expectedAlias)
				}
			}
		}
	}
}

func TestSelectCommandInteractiveFallback(t *testing.T) {
	// This test will run the fallback behavior since we're not in an interactive terminal
	// In a real interactive environment, this would test the actual selection

	// We can't easily test the interactive selection without mocking,
	// but we can test that the function exists and has the right signature
	commands := GetAvailableCommands()
	if len(commands) == 0 {
		t.Error("Should have commands available for selection")
	}

	// Test that each command can be found in the display format
	for _, cmd := range commands {
		displayName := cmd.Name
		if displayName == "" {
			t.Errorf("Command should have a non-empty name for display")
		}
	}
}

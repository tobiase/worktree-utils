package help

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout captures stdout during function execution
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestShowCommandHelp(t *testing.T) {

	tests := []struct {
		name        string
		command     string
		wantContain []string
		wantMissing []string
	}{
		{
			name:    "list command help",
			command: "list",
			wantContain: []string{
				"NAME",
				"wt list",
				"List all worktrees",
				"USAGE",
				"EXAMPLES",
				"SEE ALSO",
			},
		},
		{
			name:    "new command help",
			command: "new",
			wantContain: []string{
				"NAME",
				"wt new",
				"Smart worktree creation",
				"USAGE",
				"OPTIONS",
				"EXAMPLES",
			},
		},
		{
			name:    "go command help",
			command: "go",
			wantContain: []string{
				"NAME",
				"wt go",
				"Switch to a worktree",
				"ALIASES",
				"OPTIONS",
				"--fuzzy",
			},
		},
		{
			name:    "rm command help",
			command: "rm",
			wantContain: []string{
				"NAME",
				"wt rm",
				"Remove a worktree",
				"OPTIONS",
			},
		},
		{
			name:    "integrate command help",
			command: "integrate",
			wantContain: []string{
				"NAME",
				"wt integrate",
				"Rebase a worktree branch",
				"OPTIONS",
			},
		},
		{
			name:    "env command help",
			command: "env",
			wantContain: []string{
				"NAME",
				"wt env",
				"Unified environment file management",
				"SUBCOMMANDS",
				"sync",
				"diff",
				"list",
			},
		},
		{
			name:    "setup command help",
			command: "setup",
			wantContain: []string{
				"NAME",
				"wt setup",
				"Install wt with shell integration",
				"OPTIONS",
			},
		},
		{
			name:    "project command help",
			command: "project",
			wantContain: []string{
				"NAME",
				"wt project",
				"Manage project configuration",
				"SUBCOMMANDS",
				"init",
				"setup",
			},
		},
		{
			name:    "completion command help",
			command: "completion",
			wantContain: []string{
				"NAME",
				"wt completion",
				"Generate shell completion",
				"USAGE",
				"bash",
				"zsh",
			},
		},
		{
			name:    "update command help",
			command: "update",
			wantContain: []string{
				"NAME",
				"wt update",
				"Check for and install updates",
				"OPTIONS",
			},
		},
		{
			name:    "version command help",
			command: "version",
			wantContain: []string{
				"NAME",
				"wt version",
				"Show version information",
			},
		},
		// Unknown command test removed - it exits and can't be tested this way
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(func() {
				ShowCommandHelp(tt.command)
			})

			for _, want := range tt.wantContain {
				if !strings.Contains(output, want) {
					t.Errorf("ShowCommandHelp(%q) output missing %q", tt.command, want)
				}
			}

			for _, missing := range tt.wantMissing {
				if strings.Contains(output, missing) {
					t.Errorf("ShowCommandHelp(%q) output should not contain %q", tt.command, missing)
				}
			}
		})
	}
}

func TestHasHelpFlag(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		commandName string
		wantHelp    bool
	}{
		{
			name:        "help long flag",
			args:        []string{"--help"},
			commandName: "list",
			wantHelp:    true,
		},
		{
			name:        "help short flag",
			args:        []string{"-h"},
			commandName: "list",
			wantHelp:    true,
		},
		{
			name:        "help flag with other args",
			args:        []string{"feature-branch", "--help"},
			commandName: "new",
			wantHelp:    true,
		},
		{
			name:        "no help flag",
			args:        []string{"feature-branch"},
			commandName: "new",
			wantHelp:    false,
		},
		{
			name:        "empty args",
			args:        []string{},
			commandName: "list",
			wantHelp:    false,
		},
		{
			name:        "similar but not help flag",
			args:        []string{"--helpful", "-helper"},
			commandName: "list",
			wantHelp:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output since HasHelpFlag shows help when found
			output := captureStdout(func() {
				result := HasHelpFlag(tt.args, tt.commandName)
				if result != tt.wantHelp {
					t.Errorf("HasHelpFlag(%v, %q) = %v, want %v", tt.args, tt.commandName, result, tt.wantHelp)
				}
			})

			// If help was shown, verify output contains command name
			if tt.wantHelp && !strings.Contains(output, tt.commandName) {
				t.Errorf("Help output should contain command name %q", tt.commandName)
			}
		})
	}
}

func TestCommandHelpMapContent(t *testing.T) {
	expectedCommands := []string{
		"list", "go", "rm", "integrate", "new", "env", "env-copy",
		"project", "completion", "setup", "update", "version",
	}

	for _, cmd := range expectedCommands {
		if _, exists := commandHelpMap[cmd]; !exists {
			t.Errorf("Expected command %q to be in help map", cmd)
		}
	}

	// Verify help content structure
	listHelp := commandHelpMap["list"]
	if listHelp.Name != "list" {
		t.Errorf("list help name = %q, want %q", listHelp.Name, "list")
	}
	if listHelp.Usage == "" {
		t.Error("list help usage should not be empty")
	}
	if listHelp.Description == "" {
		t.Error("list help description should not be empty")
	}
	if len(listHelp.Examples) == 0 {
		t.Error("list help should have examples")
	}
}

// Remove ShowUsage test as the function doesn't exist

func TestCommandHelpCompleteness(t *testing.T) {

	// Every command should have complete help information
	for cmd, help := range commandHelpMap {
		if help.Name == "" {
			t.Errorf("Command %q missing Name", cmd)
		}
		if help.Usage == "" {
			t.Errorf("Command %q missing Usage", cmd)
		}
		if help.Description == "" {
			t.Errorf("Command %q missing Description", cmd)
		}
		// Most commands should have examples (version might not)
		if len(help.Examples) == 0 && cmd != "version" {
			t.Errorf("Command %q missing Examples", cmd)
		}
	}
}

func TestHelpFormatting(t *testing.T) {

	// Test that help output is properly formatted
	output := captureStdout(func() {
		ShowCommandHelp("new")
	})

	// Check for proper spacing and formatting
	lines := strings.Split(output, "\n")

	// Should have empty lines between sections
	nameIdx := -1
	usageIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "NAME") {
			nameIdx = i
		}
		if strings.HasPrefix(line, "USAGE") {
			usageIdx = i
		}
	}

	if nameIdx == -1 || usageIdx == -1 {
		t.Error("Missing NAME or USAGE section")
		return
	}

	// Should have empty line between sections
	if usageIdx-nameIdx < 3 {
		t.Error("Sections should be separated by empty lines")
	}
}

func TestSubcommandHelp(t *testing.T) {

	// Test commands with subcommands
	envHelp := commandHelpMap["env"]
	if len(envHelp.Subcommands) == 0 {
		t.Error("env command should have subcommands")
	}

	expectedSubcommands := []string{"sync", "diff", "list"}
	for _, sub := range expectedSubcommands {
		found := false
		for _, subCmd := range envHelp.Subcommands {
			if strings.Contains(subCmd, sub) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("env command missing subcommand %q", sub)
		}
	}
}

func TestAliasesInHelp(t *testing.T) {

	// Commands with aliases should show them
	tests := []struct {
		command string
		aliases []string
	}{
		{"go", []string{"switch", "s"}},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			output := captureStdout(func() {
				ShowCommandHelp(tt.command)
			})

			if len(tt.aliases) > 0 && !strings.Contains(output, "ALIASES") {
				t.Errorf("Command %q should show ALIASES section", tt.command)
			}

			for _, alias := range tt.aliases {
				if !strings.Contains(output, alias) {
					t.Errorf("Command %q help missing alias %q", tt.command, alias)
				}
			}
		})
	}
}

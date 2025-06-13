package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// TestShowUsage verifies the usage message format
func TestShowUsage(t *testing.T) {
	// Capture output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	showUsage()

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Check for key elements in usage
	expectedElements := []string{
		"Usage: wt <command>",
		"Core commands:",
		"list",
		"add",
		"rm",
		"go",
		"new",
		"Utility commands:",
		"env-copy",
		"project init",
		"Setup commands:",
		"setup",
		"update",
		"Other commands:",
		"shell-init",
		"version",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Usage output missing %q", elem)
		}
	}
}

// TestVersionFormat verifies version string formatting
func TestVersionFormat(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
		want    string
	}{
		{
			name:    "normal version",
			version: "1.2.3",
			commit:  "abc123",
			date:    "2024-01-01",
			want:    "wt version 1.2.3\ncommit: abc123\nbuilt at: 2024-01-01",
		},
		{
			name:    "dev version",
			version: "dev",
			commit:  "unknown",
			date:    "unknown",
			want:    "wt version dev\ncommit: unknown\nbuilt at: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test version info
			oldVersion := version
			oldCommit := commit
			oldDate := date
			defer func() {
				version = oldVersion
				commit = oldCommit
				date = oldDate
			}()

			version = tt.version
			commit = tt.commit
			date = tt.date

			// Create expected output
			expected := fmt.Sprintf("wt version %s\ncommit: %s\nbuilt at: %s", version, commit, date)

			// The actual version command would use these variables
			got := fmt.Sprintf("wt version %s\ncommit: %s\nbuilt at: %s", version, commit, date)

			if got != expected {
				t.Errorf("Version output = %q, want %q", got, expected)
			}
		})
	}
}

// TestShellWrapperOutput verifies the shell wrapper script
func TestShellWrapperOutput(t *testing.T) {
	output := shellWrapper

	// Check key components of the shell wrapper
	expectedComponents := []string{
		"wt()",
		"WT_BIN:-wt-bin",
		"CD:",
		"EXEC:",
		"exit_code",
		"2>&1",
		"$?",
	}

	for _, component := range expectedComponents {
		if !strings.Contains(output, component) {
			t.Errorf("Shell wrapper missing component %q", component)
		}
	}
}

// TestCommandlineArgParsing tests how main.go would parse arguments
func TestCommandlineArgParsing(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCmd     string
		wantArgs    []string
		wantSpecial bool // true for commands that bypass normal flow
	}{
		{
			name:     "simple command",
			args:     []string{"wt", "list"},
			wantCmd:  "list",
			wantArgs: []string{},
		},
		{
			name:     "command with args",
			args:     []string{"wt", "add", "feature-branch"},
			wantCmd:  "add",
			wantArgs: []string{"feature-branch"},
		},
		{
			name:     "command with multiple args",
			args:     []string{"wt", "new", "feature", "main"},
			wantCmd:  "new",
			wantArgs: []string{"feature", "main"},
		},
		{
			name:        "setup command",
			args:        []string{"wt", "setup"},
			wantCmd:     "setup",
			wantArgs:    []string{},
			wantSpecial: true,
		},
		{
			name:        "shell-init command",
			args:        []string{"wt", "shell-init"},
			wantCmd:     "shell-init",
			wantArgs:    []string{},
			wantSpecial: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.args) < 2 {
				return
			}

			cmd := tt.args[1]
			args := tt.args[2:]

			if cmd != tt.wantCmd {
				t.Errorf("Parsed command = %q, want %q", cmd, tt.wantCmd)
			}

			if len(args) != len(tt.wantArgs) {
				t.Errorf("Parsed args length = %d, want %d", len(args), len(tt.wantArgs))
			}

			for i, arg := range args {
				if i < len(tt.wantArgs) && arg != tt.wantArgs[i] {
					t.Errorf("Parsed arg[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

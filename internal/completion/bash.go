package completion

import (
	"fmt"
	"strings"

	"github.com/tobiase/worktree-utils/internal/config"
)

// GenerateBashCompletion generates a bash completion script
func GenerateBashCompletion(configMgr *config.Manager) string {
	data := GetCompletionData(configMgr)

	var builder strings.Builder

	// Header
	builder.WriteString("#!/bin/bash\n")
	builder.WriteString("# Bash completion for wt (worktree-utils)\n")
	builder.WriteString("# Generated automatically - do not edit manually\n\n")

	// Main completion function
	builder.WriteString("_wt_completion() {\n")
	builder.WriteString("    local cur prev words cword\n")
	builder.WriteString("    _init_completion || return\n\n")

	// First argument - command completion
	builder.WriteString("    # Complete commands on first argument\n")
	builder.WriteString("    if [[ $cword -eq 1 ]]; then\n")
	builder.WriteString("        local commands=\"")
	builder.WriteString(strings.Join(data.GetAllCommandNames(), " "))
	builder.WriteString("\"\n")
	builder.WriteString("        COMPREPLY=($(compgen -W \"$commands\" -- \"$cur\"))\n")
	builder.WriteString("        return\n")
	builder.WriteString("    fi\n\n")

	// Context-sensitive completion
	builder.WriteString("    # Context-sensitive completion based on command\n")
	builder.WriteString("    local command=\"${words[1]}\"\n")
	builder.WriteString("    case \"$command\" in\n")

	// Add completion for each command
	for _, cmd := range data.Commands {
		if len(cmd.Args) > 0 || len(cmd.Flags) > 0 {
			builder.WriteString(fmt.Sprintf("        %s)\n", cmd.Name))
			generateBashCommandCompletion(&builder, cmd, data)
			builder.WriteString("            ;;\n")
		}
	}

	// Handle aliases
	for alias, target := range data.Aliases {
		if targetCmd := data.GetCommandByName(target); targetCmd != nil && (len(targetCmd.Args) > 0 || len(targetCmd.Flags) > 0) {
			builder.WriteString(fmt.Sprintf("        %s)\n", alias))
			generateBashCommandCompletion(&builder, *targetCmd, data)
			builder.WriteString("            ;;\n")
		}
	}

	builder.WriteString("    esac\n")
	builder.WriteString("}\n\n")

	// Helper functions
	generateBashHelperFunctions(&builder, data)

	// Register completion
	builder.WriteString("# Register completion for wt command\n")
	builder.WriteString("complete -F _wt_completion wt\n")

	return builder.String()
}

// generateBashCommandCompletion generates completion logic for a specific command
func generateBashCommandCompletion(builder *strings.Builder, cmd Command, data *CompletionData) {
	// Handle flags
	if len(cmd.Flags) > 0 {
		builder.WriteString("            # Complete flags\n")
		builder.WriteString("            if [[ \"$cur\" == -* ]]; then\n")
		builder.WriteString("                local flags=\"")

		var flags []string
		for _, flag := range cmd.Flags {
			flags = append(flags, flag.Name)
		}
		builder.WriteString(strings.Join(flags, " "))
		builder.WriteString("\"\n")
		builder.WriteString("                COMPREPLY=($(compgen -W \"$flags\" -- \"$cur\"))\n")
		builder.WriteString("                return\n")
		builder.WriteString("            fi\n")
	}

	// Handle arguments based on position and type
	if len(cmd.Args) > 0 {
		arg := cmd.Args[0] // Handle first argument for now
		switch arg.Type {
		case ArgBranch:
			builder.WriteString("            # Complete branch names\n")
			builder.WriteString("            _wt_complete_branches\n")
		case ArgWorktreeBranch:
			builder.WriteString("            # Complete worktree branch names\n")
			builder.WriteString("            _wt_complete_worktree_branches\n")
		case ArgString:
			if cmd.Name == "completion" {
				builder.WriteString("            # Complete shell types\n")
				builder.WriteString("            COMPREPLY=($(compgen -W \"bash zsh\" -- \"$cur\"))\n")
			} else if cmd.Name == "project" {
				builder.WriteString("            # Complete project subcommands\n")
				builder.WriteString("            COMPREPLY=($(compgen -W \"init\" -- \"$cur\"))\n")
			}
		}
	}

	// Special handling for commands with complex argument patterns
	switch cmd.Name {
	case "new":
		builder.WriteString("            # Handle --base flag value completion\n")
		builder.WriteString("            if [[ \"$prev\" == \"--base\" ]]; then\n")
		builder.WriteString("                _wt_complete_branches\n")
		builder.WriteString("                return\n")
		builder.WriteString("            fi\n")
	}
}

// generateBashHelperFunctions generates helper functions for bash completion
func generateBashHelperFunctions(builder *strings.Builder, data *CompletionData) {
	// Branch completion helper
	builder.WriteString("# Helper function to complete branch names\n")
	builder.WriteString("_wt_complete_branches() {\n")
	builder.WriteString("    local branches\n")
	builder.WriteString("    # Try to get branches from wt command\n")
	builder.WriteString("    if command -v wt >/dev/null 2>&1; then\n")
	builder.WriteString("        branches=$(wt list 2>/dev/null | awk '{print $2}' | grep -v '^$' | sort -u)\n")
	builder.WriteString("        if [[ -n \"$branches\" ]]; then\n")
	builder.WriteString("            COMPREPLY=($(compgen -W \"$branches\" -- \"$cur\"))\n")
	builder.WriteString("            return\n")
	builder.WriteString("        fi\n")
	builder.WriteString("    fi\n")
	builder.WriteString("    \n")
	builder.WriteString("    # Fallback to git branches if available\n")
	builder.WriteString("    if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then\n")
	builder.WriteString("        branches=$(git branch --format='%(refname:short)' 2>/dev/null)\n")
	builder.WriteString("        COMPREPLY=($(compgen -W \"$branches\" -- \"$cur\"))\n")
	builder.WriteString("    fi\n")
	builder.WriteString("}\n\n")

	// Worktree branch completion helper (only existing worktrees)
	builder.WriteString("# Helper function to complete worktree branch names\n")
	builder.WriteString("_wt_complete_worktree_branches() {\n")
	builder.WriteString("    local branches\n")
	builder.WriteString("    # Get branches from wt list (only existing worktrees)\n")
	builder.WriteString("    if command -v wt >/dev/null 2>&1; then\n")
	builder.WriteString("        branches=$(wt list 2>/dev/null | awk '{print $2}' | grep -v '^$' | sort -u)\n")
	builder.WriteString("        COMPREPLY=($(compgen -W \"$branches\" -- \"$cur\"))\n")
	builder.WriteString("    fi\n")
	builder.WriteString("}\n\n")

	// Project command completion helper
	if len(data.ProjectCommands) > 0 {
		builder.WriteString("# Helper function to complete project commands\n")
		builder.WriteString("_wt_complete_project_commands() {\n")
		builder.WriteString("    local project_commands=\"")
		builder.WriteString(strings.Join(data.ProjectCommands, " "))
		builder.WriteString("\"\n")
		builder.WriteString("    COMPREPLY=($(compgen -W \"$project_commands\" -- \"$cur\"))\n")
		builder.WriteString("}\n\n")
	}
}

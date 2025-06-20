package completion

import (
	"fmt"
	"strings"

	"github.com/tobiase/worktree-utils/internal/config"
)

// GenerateZshCompletion generates a zsh completion script
func GenerateZshCompletion(configMgr *config.Manager) string {
	data := GetCompletionData(configMgr)

	var builder strings.Builder

	// Header with proper compdef directive
	builder.WriteString("#compdef wt\n")
	builder.WriteString("# Zsh completion for wt (worktree-utils)\n")
	builder.WriteString("# Generated automatically - do not edit manually\n\n")

	// Main completion function
	builder.WriteString("_wt() {\n")
	builder.WriteString("    local context state line\n")
	builder.WriteString("    typeset -A opt_args\n\n")

	// Argument specification
	builder.WriteString("    _arguments -C \\\n")
	builder.WriteString("        '1: :_wt_commands' \\\n")
	builder.WriteString("        '*:: :->args'\n\n")

	// State handling
	builder.WriteString("    case $state in\n")
	builder.WriteString("        args)\n")
	builder.WriteString("            case $words[1] in\n")

	// Add completion for each command
	for _, cmd := range data.Commands {
		if len(cmd.Args) > 0 || len(cmd.Flags) > 0 {
			builder.WriteString(fmt.Sprintf("                %s)\n", cmd.Name))
			generateZshCommandCompletion(&builder, cmd, data)
			builder.WriteString("                    ;;\n")
		}
	}

	// Handle aliases
	for alias, target := range data.Aliases {
		if targetCmd := data.GetCommandByName(target); targetCmd != nil && (len(targetCmd.Args) > 0 || len(targetCmd.Flags) > 0) {
			builder.WriteString(fmt.Sprintf("                %s)\n", alias))
			generateZshCommandCompletion(&builder, *targetCmd, data)
			builder.WriteString("                    ;;\n")
		}
	}

	builder.WriteString("            esac\n")
	builder.WriteString("            ;;\n")
	builder.WriteString("    esac\n")
	builder.WriteString("}\n\n")

	// Helper functions
	generateZshHelperFunctions(&builder, data)

	return builder.String()
}

// generateZshCommandCompletion generates completion logic for a specific command
func generateZshCommandCompletion(builder *strings.Builder, cmd Command, data *CompletionData) {
	switch cmd.Name {
	case "go", "rm", "env-copy":
		builder.WriteString("                    _wt_worktree_branches\n")
	case "new":
		builder.WriteString("                    _wt_new_args\n")
	case "completion":
		builder.WriteString("                    _wt_shells\n")
	case "project":
		builder.WriteString("                    _wt_project_args\n")
	case "setup":
		builder.WriteString("                    _wt_setup_args\n")
	case "update":
		builder.WriteString("                    _wt_update_args\n")
	default:
		if len(cmd.Flags) > 0 {
			builder.WriteString("                    _wt_flags\n")
		}
	}
}

// generateZshHelperFunctions generates helper functions for zsh completion
func generateZshHelperFunctions(builder *strings.Builder, data *CompletionData) {
	// Commands completion
	builder.WriteString("_wt_commands() {\n")
	builder.WriteString("    local commands=(\n")

	for _, cmd := range data.Commands {
		builder.WriteString(fmt.Sprintf("        '%s:%s'\n", cmd.Name, cmd.Description))
	}

	// Add aliases with descriptions
	for alias, target := range data.Aliases {
		if targetCmd := data.GetCommandByName(target); targetCmd != nil {
			builder.WriteString(fmt.Sprintf("        '%s:Alias for %s - %s'\n", alias, target, targetCmd.Description))
		}
	}

	// Add project commands
	for _, projectCmd := range data.ProjectCommands {
		builder.WriteString(fmt.Sprintf("        '%s:Project-specific command'\n", projectCmd))
	}

	builder.WriteString("    )\n")
	builder.WriteString("    _describe 'commands' commands\n")
	builder.WriteString("}\n\n")

	// Branch completion
	builder.WriteString("_wt_branches() {\n")
	builder.WriteString("    local branches=()\n")
	builder.WriteString("    \n")
	builder.WriteString("    # Try to get branches from wt command\n")
	builder.WriteString("    if command -v wt >/dev/null 2>&1; then\n")
	builder.WriteString("        local wt_output\n")
	builder.WriteString("        wt_output=$(wt list 2>/dev/null)\n")
	builder.WriteString("        if [[ $? -eq 0 && -n \"$wt_output\" ]]; then\n")
	builder.WriteString("            branches=(${(f)\"$(echo \"$wt_output\" | awk '{print $2}' | grep -v '^$' | sort -u)\"})\n")
	builder.WriteString("        fi\n")
	builder.WriteString("    fi\n")
	builder.WriteString("    \n")
	builder.WriteString("    # Fallback to git branches if available\n")
	builder.WriteString("    if [[ ${#branches[@]} -eq 0 ]] && command -v git >/dev/null 2>&1; then\n")
	builder.WriteString("        if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then\n")
	builder.WriteString("            branches=(${(f)\"$(git branch --format='%(refname:short)' 2>/dev/null)\"})\n")
	builder.WriteString("        fi\n")
	builder.WriteString("    fi\n")
	builder.WriteString("    \n")
	builder.WriteString("    if [[ ${#branches[@]} -gt 0 ]]; then\n")
	builder.WriteString("        _describe 'branches' branches\n")
	builder.WriteString("    else\n")
	builder.WriteString("        _message 'no branches found'\n")
	builder.WriteString("    fi\n")
	builder.WriteString("}\n\n")

	// Worktree branches (only existing worktrees)
	builder.WriteString("_wt_worktree_branches() {\n")
	builder.WriteString("    local branches=()\n")
	builder.WriteString("    \n")
	builder.WriteString("    # Get branches from wt list (only existing worktrees)\n")
	builder.WriteString("    if command -v wt >/dev/null 2>&1; then\n")
	builder.WriteString("        local wt_output\n")
	builder.WriteString("        wt_output=$(wt list 2>/dev/null)\n")
	builder.WriteString("        if [[ $? -eq 0 && -n \"$wt_output\" ]]; then\n")
	builder.WriteString("            branches=(${(f)\"$(echo \"$wt_output\" | awk '{print $2}' | grep -v '^$' | sort -u)\"})\n")
	builder.WriteString("        fi\n")
	builder.WriteString("    fi\n")
	builder.WriteString("    \n")
	builder.WriteString("    if [[ ${#branches[@]} -gt 0 ]]; then\n")
	builder.WriteString("        _describe 'worktree branches' branches\n")
	builder.WriteString("    else\n")
	builder.WriteString("        _message 'no worktrees found'\n")
	builder.WriteString("    fi\n")
	builder.WriteString("}\n\n")

	// Shell types for completion command
	builder.WriteString("_wt_shells() {\n")
	builder.WriteString("    local shells=(\n")
	builder.WriteString("        'bash:Generate bash completion'\n")
	builder.WriteString("        'zsh:Generate zsh completion'\n")
	builder.WriteString("    )\n")
	builder.WriteString("    _describe 'shells' shells\n")
	builder.WriteString("}\n\n")

	// New command arguments
	builder.WriteString("_wt_new_args() {\n")
	builder.WriteString("    _arguments \\\n")
	builder.WriteString("        '--base[Base branch]:branch:_wt_branches' \\\n")
	builder.WriteString("        '1:new branch name:_message \"branch name\"'\n")
	builder.WriteString("}\n\n")

	// Project command arguments
	builder.WriteString("_wt_project_args() {\n")
	builder.WriteString("    local subcommands=(\n")
	builder.WriteString("        'init:Initialize project configuration'\n")
	builder.WriteString("    )\n")
	builder.WriteString("    _describe 'project subcommands' subcommands\n")
	builder.WriteString("}\n\n")

	// Setup command arguments
	builder.WriteString("_wt_setup_args() {\n")
	builder.WriteString("    local options=(\n")
	builder.WriteString("        '--check:Check installation status'\n")
	builder.WriteString("        '--uninstall:Uninstall wt'\n")
	builder.WriteString("    )\n")
	builder.WriteString("    _describe 'setup options' options\n")
	builder.WriteString("}\n\n")

	// Update command arguments
	builder.WriteString("_wt_update_args() {\n")
	builder.WriteString("    local options=(\n")
	builder.WriteString("        '--check:Check for updates only'\n")
	builder.WriteString("        '--force:Force update even if up to date'\n")
	builder.WriteString("    )\n")
	builder.WriteString("    _describe 'update options' options\n")
	builder.WriteString("}\n\n")

	// Register completion function (only if compdef is available)
	builder.WriteString("# Register the completion function\n")
	builder.WriteString("if command -v compdef >/dev/null 2>&1; then\n")
	builder.WriteString("    compdef _wt wt\n")
	builder.WriteString("else\n")
	builder.WriteString("    # Completion system not initialized yet\n")
	builder.WriteString("    autoload -Uz compinit && compinit\n")
	builder.WriteString("    compdef _wt wt\n")
	builder.WriteString("fi\n")

}

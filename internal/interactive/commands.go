package interactive

import (
	"fmt"
	"strings"
)

// CommandInfo represents a command with its description and aliases
type CommandInfo struct {
	Name        string
	Description string
	Aliases     []string
	Usage       string
}

// GetAvailableCommands returns the list of available commands
func GetAvailableCommands() []CommandInfo {
	return []CommandInfo{
		{
			Name:        "list",
			Description: "List all worktrees",
			Aliases:     []string{"ls"},
			Usage:       "wt list",
		},
		{
			Name:        "go",
			Description: "Switch to a worktree",
			Aliases:     []string{"switch", "s"},
			Usage:       "wt go [branch] [--fuzzy]",
		},
		{
			Name:        "new",
			Description: "Smart worktree creation (handles all branch states)",
			Aliases:     []string{},
			Usage:       "wt new <branch> [--base <branch>]",
		},
		{
			Name:        "rm",
			Description: "Remove a worktree",
			Aliases:     []string{},
			Usage:       "wt rm [branch] [--fuzzy]",
		},
		{
			Name:        "env-copy",
			Description: "Copy .env files to another worktree",
			Aliases:     []string{},
			Usage:       "wt env-copy [branch] [--fuzzy] [--recursive]",
		},
		{
			Name:        "project",
			Description: "Manage project configuration",
			Aliases:     []string{},
			Usage:       "wt project init <name>",
		},
		{
			Name:        "setup",
			Description: "Install wt with shell completion",
			Aliases:     []string{},
			Usage:       "wt setup [--completion <shell>] [--no-completion]",
		},
		{
			Name:        "update",
			Description: "Check and install updates",
			Aliases:     []string{},
			Usage:       "wt update [--check] [--force]",
		},
		{
			Name:        "completion",
			Description: "Generate shell completion scripts",
			Aliases:     []string{},
			Usage:       "wt completion <bash|zsh>",
		},
		{
			Name:        "version",
			Description: "Show version information",
			Aliases:     []string{},
			Usage:       "wt version",
		},
		{
			Name:        "help",
			Description: "Show help information",
			Aliases:     []string{},
			Usage:       "wt help",
		},
	}
}

// SelectCommandInteractively presents a fuzzy finder for command selection
func SelectCommandInteractively() (string, error) {
	commands := GetAvailableCommands()

	// Create display strings for commands
	displayItems := make([]string, len(commands))
	for i, cmd := range commands {
		display := cmd.Name
		if len(cmd.Aliases) > 0 {
			display += fmt.Sprintf(" (%s)", strings.Join(cmd.Aliases, ", "))
		}
		displayItems[i] = display
	}

	// Use interactive selection with preview
	result, err := SelectStringWithPreview(
		displayItems,
		"Select command:",
		func(i, width, height int) string {
			if i < 0 || i >= len(commands) {
				return "No command selected"
			}

			cmd := commands[i]
			var preview strings.Builder

			preview.WriteString(fmt.Sprintf("Command: %s\n", cmd.Name))
			preview.WriteString(fmt.Sprintf("Description: %s\n\n", cmd.Description))
			preview.WriteString(fmt.Sprintf("Usage: %s\n", cmd.Usage))

			if len(cmd.Aliases) > 0 {
				preview.WriteString(fmt.Sprintf("Aliases: %s\n", strings.Join(cmd.Aliases, ", ")))
			}

			// Add some helpful tips based on the command
			switch cmd.Name {
			case "go":
				preview.WriteString("\nTip: Run without arguments to select from worktrees interactively")
			case "rm":
				preview.WriteString("\nTip: Run without arguments to select worktree to remove interactively")
			case "env-copy":
				preview.WriteString("\nTip: Run without arguments to select target worktree interactively")
			case "new":
				preview.WriteString("\nTip: Use --base to create from a specific branch")
			case "setup":
				preview.WriteString("\nTip: Run this once to install shell completion and integration")
			}

			return preview.String()
		},
	)

	if err != nil {
		return "", err
	}

	// Extract the command name from the display string
	parts := strings.Split(result, " ")
	return parts[0], nil
}

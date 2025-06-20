package help

import (
	"fmt"
	"os"
	"strings"
)

// Help flag constants
const (
	helpFlag      = "--help"
	helpFlagShort = "-h"
)

// CommandHelp represents help information for a command
type CommandHelp struct {
	Name        string
	Usage       string
	Description string
	Subcommands []string
	Examples    []string
	Flags       []FlagHelp
	Aliases     []string
	SeeAlso     []string
}

// FlagHelp represents help for a command flag
type FlagHelp struct {
	Flag        string
	ShortFlag   string
	Description string
	Example     string
}

// ShowCommandHelp displays detailed help for a specific command
func ShowCommandHelp(commandName string) {
	help, exists := commandHelpMap[commandName]
	if !exists {
		fmt.Fprintf(os.Stderr, "No help available for command '%s'\n", commandName)
		os.Exit(1)
	}

	fmt.Printf("NAME\n")
	fmt.Printf("    wt %s - %s\n\n", help.Name, help.Description)

	fmt.Printf("USAGE\n")
	fmt.Printf("    %s\n\n", help.Usage)

	if len(help.Aliases) > 0 {
		fmt.Printf("ALIASES\n")
		fmt.Printf("    %s\n\n", strings.Join(help.Aliases, ", "))
	}

	if len(help.Subcommands) > 0 {
		fmt.Printf("SUBCOMMANDS\n")
		for _, subcommand := range help.Subcommands {
			fmt.Printf("    %s\n", subcommand)
		}
		fmt.Printf("\n")
	}

	if len(help.Flags) > 0 {
		fmt.Printf("OPTIONS\n")
		for _, flag := range help.Flags {
			flagStr := fmt.Sprintf("    %s", flag.Flag)
			if flag.ShortFlag != "" {
				flagStr += fmt.Sprintf(", %s", flag.ShortFlag)
			}
			fmt.Printf("%-20s %s\n", flagStr, flag.Description)
			if flag.Example != "" {
				fmt.Printf("%-20s Example: %s\n", "", flag.Example)
			}
		}
		fmt.Printf("\n")
	}

	if len(help.Examples) > 0 {
		fmt.Printf("EXAMPLES\n")
		for _, example := range help.Examples {
			fmt.Printf("    %s\n", example)
		}
		fmt.Printf("\n")
	}

	if len(help.SeeAlso) > 0 {
		fmt.Printf("SEE ALSO\n")
		fmt.Printf("    %s\n\n", strings.Join(help.SeeAlso, ", "))
	}
}

// HasHelpFlag checks if help flags are present in args and shows help if found
func HasHelpFlag(args []string, commandName string) bool {
	for _, arg := range args {
		if arg == helpFlag || arg == helpFlagShort {
			ShowCommandHelp(commandName)
			return true
		}
	}
	return false
}

// commandHelpMap contains help information for all commands
var commandHelpMap = map[string]CommandHelp{
	"list": {
		Name:        "list",
		Usage:       "wt list",
		Description: "List all worktrees with their index, branch name, and path",
		Examples: []string{
			"wt list              # Show all worktrees",
			"wt ls                # Same as above (alias)",
		},
		Aliases: []string{"ls"},
		SeeAlso: []string{"wt go", "wt new"},
	},
	"ls": {
		Name:        "list",
		Usage:       "wt ls",
		Description: "List all worktrees with their index, branch name, and path",
		Examples: []string{
			"wt ls                # Show all worktrees",
			"wt list              # Same as above",
		},
		Aliases: []string{"list"},
		SeeAlso: []string{"wt go", "wt new"},
	},
	"go": {
		Name:        "go",
		Usage:       "wt go [branch|index] [options]",
		Description: "Switch to a worktree by branch name or index. Supports fuzzy matching.",
		Examples: []string{
			"wt go                # Go to repository root",
			"wt go main           # Switch to main branch worktree",
			"wt go mai            # Fuzzy match to 'main'",
			"wt go 0              # Switch to first worktree by index",
			"wt go feat --fuzzy   # Force interactive selection for 'feat'",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--fuzzy",
				ShortFlag:   "-f",
				Description: "Force interactive selection even for unique matches",
				Example:     "wt go main -f",
			},
		},
		Aliases: []string{"switch", "s"},
		SeeAlso: []string{"wt list", "wt new"},
	},
	"switch": {
		Name:        "go",
		Usage:       "wt switch [branch|index] [options]",
		Description: "Switch to a worktree by branch name or index. Supports fuzzy matching.",
		Examples: []string{
			"wt switch main       # Switch to main branch worktree",
			"wt s feat            # Short alias for switch with fuzzy matching",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--fuzzy",
				ShortFlag:   "-f",
				Description: "Force interactive selection even for unique matches",
			},
		},
		Aliases: []string{"go", "s"},
		SeeAlso: []string{"wt list", "wt new"},
	},
	"s": {
		Name:        "go",
		Usage:       "wt s [branch|index] [options]",
		Description: "Switch to a worktree by branch name or index. Supports fuzzy matching.",
		Examples: []string{
			"wt s main            # Switch to main branch worktree",
			"wt s 1               # Switch to second worktree",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--fuzzy",
				ShortFlag:   "-f",
				Description: "Force interactive selection even for unique matches",
			},
		},
		Aliases: []string{"go", "switch"},
		SeeAlso: []string{"wt list", "wt new"},
	},
	"new": {
		Name:        "new",
		Usage:       "wt new <branch> [options]",
		Description: "Smart worktree creation that handles all branch states intelligently",
		Examples: []string{
			"wt new feature               # Create new branch + worktree",
			"wt new existing-branch       # Create worktree for existing branch",
			"wt new feature --base main   # Create new branch from main",
			"wt new feature --no-switch   # Create worktree without switching to it",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--base <branch>",
				Description: "Base branch for new branch creation",
				Example:     "wt new feature --base develop",
			},
			{
				Flag:        "--no-switch",
				Description: "Create worktree without switching to it",
				Example:     "wt new feature --no-switch",
			},
		},
		SeeAlso: []string{"wt go", "wt rm"},
	},
	"rm": {
		Name:        "rm",
		Usage:       "wt rm [branch] [options]",
		Description: "Remove a worktree. Supports fuzzy matching for branch names.",
		Examples: []string{
			"wt rm feature        # Remove worktree for 'feature' branch",
			"wt rm feat           # Fuzzy match to remove 'feature' branch",
			"wt rm --fuzzy        # Interactive selection of worktree to remove",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--fuzzy",
				ShortFlag:   "-f",
				Description: "Force interactive selection for worktree to remove",
				Example:     "wt rm -f",
			},
		},
		SeeAlso: []string{"wt list", "wt new"},
	},
	"env-copy": {
		Name:        "env-copy",
		Usage:       "wt env-copy [branch] [options]",
		Description: "Copy .env files from current directory to target worktree",
		Examples: []string{
			"wt env-copy feature          # Copy .env to feature branch worktree",
			"wt env-copy feat --recursive # Copy all .env* files recursively",
			"wt env-copy --fuzzy          # Interactive selection of target",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--recursive",
				Description: "Copy all .env* files recursively",
				Example:     "wt env-copy feature --recursive",
			},
			{
				Flag:        "--fuzzy",
				ShortFlag:   "-f",
				Description: "Force interactive selection of target worktree",
				Example:     "wt env-copy -f",
			},
		},
		SeeAlso: []string{"wt env", "wt go", "wt list"},
	},
	"env": {
		Name:        "env",
		Usage:       "wt env <subcommand> [options]",
		Description: "Unified environment file management across worktrees",
		Subcommands: []string{
			"sync     Copy .env files to target worktree(s)",
			"diff     Show differences between .env files",
			"list     List all .env files across worktrees",
		},
		Examples: []string{
			"wt env sync feature          # Copy .env files to feature worktree",
			"wt env sync --all            # Sync .env to all other worktrees",
			"wt env diff main             # Show differences with main worktree",
			"wt env list                  # List all .env files across worktrees",
			"wt env                       # Interactive environment operations",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--all",
				Description: "Apply operation to all worktrees (sync only)",
				Example:     "wt env sync --all",
			},
			{
				Flag:        "--recursive",
				Description: "Include all .env* files recursively",
				Example:     "wt env sync feature --recursive",
			},
			{
				Flag:        "--fuzzy",
				ShortFlag:   "-f",
				Description: "Force interactive selection of target worktree",
				Example:     "wt env sync -f",
			},
		},
		SeeAlso: []string{"wt env-copy", "wt go", "wt list"},
	},
	"project": {
		Name:        "project",
		Usage:       "wt project <subcommand> [options]",
		Description: "Manage project configuration for custom commands and settings",
		Subcommands: []string{
			"init     Initialize project configuration",
			"setup    Manage worktree setup automation",
		},
		Examples: []string{
			"wt project init myproject    # Initialize project configuration",
			"wt project setup run         # Run setup automation for current worktree",
			"wt project setup show        # Show configured setup steps",
		},
		SeeAlso: []string{"wt new"},
	},
	"setup": {
		Name:        "setup",
		Usage:       "wt setup [options]",
		Description: "Install wt with shell integration and completion",
		Examples: []string{
			"wt setup                     # Install with auto-detected completion",
			"wt setup --completion bash   # Install with bash completion",
			"wt setup --no-completion     # Install without completion",
			"wt setup --check             # Check installation status",
			"wt setup --uninstall         # Remove wt from system",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--completion <shell>",
				Description: "Install completion for specified shell (auto|bash|zsh|none)",
				Example:     "wt setup --completion zsh",
			},
			{
				Flag:        "--no-completion",
				Description: "Skip completion installation",
			},
			{
				Flag:        "--check",
				Description: "Check installation status without installing",
			},
			{
				Flag:        "--uninstall",
				Description: "Remove wt from system",
			},
		},
		SeeAlso: []string{"wt completion"},
	},
	"update": {
		Name:        "update",
		Usage:       "wt update [options]",
		Description: "Check for and install updates from GitHub releases",
		Examples: []string{
			"wt update               # Check and install latest version",
			"wt update --check       # Check for updates without installing",
			"wt update --force       # Force update even if already latest",
		},
		Flags: []FlagHelp{
			{
				Flag:        "--check",
				Description: "Check for updates without installing",
			},
			{
				Flag:        "--force",
				Description: "Force update even if already on latest version",
			},
		},
		SeeAlso: []string{"wt version"},
	},
	"completion": {
		Name:        "completion",
		Usage:       "wt completion <shell>",
		Description: "Generate shell completion scripts for bash or zsh",
		Examples: []string{
			"wt completion bash >> ~/.bashrc     # Install bash completion",
			"wt completion zsh >> ~/.zshrc       # Install zsh completion",
			"eval \"$(wt completion bash)\"        # Load completion in current session",
		},
		SeeAlso: []string{"wt setup"},
	},
	"version": {
		Name:        "version",
		Usage:       "wt version",
		Description: "Show version information including build details",
		Examples: []string{
			"wt version           # Show current version",
		},
		SeeAlso: []string{"wt update"},
	},
}

package cli

import (
	"github.com/tobiase/worktree-utils/internal/help"
)

// CommandHandler is a function that handles a command with parsed flags
type CommandHandler func(flags *FlagSet) error

// WrapCommand wraps a command handler with common functionality:
// 1. Checks for help flag
// 2. Parses flags
// 3. Calls the handler
// 4. Handles errors consistently
func WrapCommand(name string, handler CommandHandler) func([]string) {
	return func(args []string) {
		// Check for help flag first
		if help.HasHelpFlag(args, name) {
			return
		}

		// Parse flags
		flags := ParseFlags(args)

		// Call the handler
		if err := handler(flags); err != nil {
			ExitWithError("%v", err)
		}
	}
}

// SimpleCommandHandler is a function that handles a command with raw arguments
type SimpleCommandHandler func(args []string) error

// WrapSimpleCommand wraps a simple command handler with help checking
func WrapSimpleCommand(name string, handler SimpleCommandHandler) func([]string) {
	return func(args []string) {
		// Check for help flag first
		if help.HasHelpFlag(args, name) {
			return
		}

		// Call the handler
		if err := handler(args); err != nil {
			ExitWithError("%v", err)
		}
	}
}

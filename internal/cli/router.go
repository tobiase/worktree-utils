package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// SubcommandRouter routes subcommands to their handlers
type SubcommandRouter struct {
	Name        string
	Commands    map[string]func([]string)
	Usage       string
	Interactive func() // Optional interactive handler when no subcommand provided
}

// NewSubcommandRouter creates a new subcommand router
func NewSubcommandRouter(name string, usage string) *SubcommandRouter {
	return &SubcommandRouter{
		Name:     name,
		Commands: make(map[string]func([]string)),
		Usage:    usage,
	}
}

// AddCommand adds a subcommand handler
func (r *SubcommandRouter) AddCommand(name string, handler func([]string)) {
	r.Commands[name] = handler
}

// SetInteractive sets the interactive handler
func (r *SubcommandRouter) SetInteractive(handler func()) {
	r.Interactive = handler
}

// Route routes to the appropriate subcommand handler
func (r *SubcommandRouter) Route(args []string) {
	if len(args) == 0 {
		if r.Interactive != nil {
			r.Interactive()
			return
		}
		r.ShowUsage()
		return
	}

	subcommand := args[0]
	subargs := args[1:]

	handler, exists := r.Commands[subcommand]
	if !exists {
		// Try to find a similar command
		suggestions := r.findSimilarCommands(subcommand)
		if len(suggestions) > 0 {
			fmt.Fprintf(os.Stderr, "wt %s: unknown subcommand '%s'\n", r.Name, subcommand)
			fmt.Fprintf(os.Stderr, "\nDid you mean:\n")
			for _, suggestion := range suggestions {
				fmt.Fprintf(os.Stderr, "  %s\n", suggestion)
			}
		} else {
			fmt.Fprintf(os.Stderr, "wt %s: unknown subcommand '%s'\n", r.Name, subcommand)
		}
		fmt.Fprintln(os.Stderr)
		r.ShowUsage()
		os.Exit(1)
	}

	handler(subargs)
}

// ShowUsage displays the usage message
func (r *SubcommandRouter) ShowUsage() {
	fmt.Fprintln(os.Stderr, r.Usage)

	if len(r.Commands) > 0 {
		fmt.Fprintln(os.Stderr, "\nAvailable subcommands:")

		// Sort command names for consistent output
		var names []string
		for name := range r.Commands {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
	}

	os.Exit(1)
}

// findSimilarCommands finds commands similar to the input
func (r *SubcommandRouter) findSimilarCommands(input string) []string {
	var suggestions []string
	inputLower := strings.ToLower(input)

	for cmd := range r.Commands {
		cmdLower := strings.ToLower(cmd)

		// Check if command starts with input
		if strings.HasPrefix(cmdLower, inputLower) {
			suggestions = append(suggestions, cmd)
			continue
		}

		// Check if command contains input
		if strings.Contains(cmdLower, inputLower) {
			suggestions = append(suggestions, cmd)
		}
	}

	// Sort suggestions
	sort.Strings(suggestions)

	// Limit to 3 suggestions
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return suggestions
}

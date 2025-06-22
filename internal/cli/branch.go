package cli

import (
	"fmt"
	"os"

	"github.com/tobiase/worktree-utils/internal/worktree"
)

// ResolveBranchArgument resolves a branch name from command arguments
// It handles:
// 1. If no branch provided, use interactive selection
// 2. If branch provided, resolve with fuzzy matching if enabled
// 3. Returns the resolved branch name or exits on error
func ResolveBranchArgument(args []string, useFuzzy bool, usageMessage string) string {
	// Parse flags to get positional arguments
	flags := ParseFlags(args)

	var target string
	if len(flags.Positional) > 0 {
		target = flags.Positional[0]
	}

	if target == "" {
		// Use interactive selection
		return SelectBranchInteractively(useFuzzy, usageMessage)
	}

	// Resolve branch name with fuzzy matching
	branches, err := worktree.GetAvailableBranches()
	if err != nil {
		ExitWithError("%v", err)
	}

	resolvedTarget, err := worktree.ResolveBranchName(target, branches)
	if err != nil {
		ExitWithError("%v", err)
	}

	return resolvedTarget
}

// SelectBranchInteractively wraps the worktree interactive selection with error handling
func SelectBranchInteractively(useFuzzy bool, prompt string) string {
	// Note: Currently the interactive selection always uses fuzzy matching
	// The useFuzzy parameter is preserved for API compatibility
	result, err := worktree.SelectBranchInteractively()
	if err != nil {
		if err.Error() == "no worktrees exist" {
			fmt.Fprintln(os.Stderr, prompt)
		} else {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		}
		os.Exit(1)
	}
	return result
}

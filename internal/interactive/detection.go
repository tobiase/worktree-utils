package interactive

import (
	"os"

	"github.com/mattn/go-isatty"
)

// IsInteractive returns true if the current session supports interactive features
func IsInteractive() bool {
	return isTerminal() && !isDisabled() && !isCIEnvironment()
}

// isTerminal checks if both stdin and stdout are connected to a terminal
func isTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) && isatty.IsTerminal(os.Stdout.Fd())
}

// isDisabled checks if interactive features are explicitly disabled
func isDisabled() bool {
	return os.Getenv("DISABLE_FUZZY") == "true" ||
		os.Getenv("WT_NO_INTERACTIVE") == "true" ||
		os.Getenv("NO_COLOR") != "" // Respect NO_COLOR convention
}

// isCIEnvironment checks if running in a CI/automated environment
func isCIEnvironment() bool {
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"BUILDKITE",
	}

	for _, env := range ciVars {
		if os.Getenv(env) != "" {
			return true
		}
	}

	return false
}

// ShouldUseFuzzy determines if fuzzy finding should be enabled for a given context
func ShouldUseFuzzy(itemCount int, explicitFlag bool) bool {
	// If explicitly requested via flag, honor it (if interactive)
	if explicitFlag {
		return IsInteractive()
	}

	// Auto-enable fuzzy finding when there are multiple items and we're interactive
	return itemCount > 1 && IsInteractive()
}

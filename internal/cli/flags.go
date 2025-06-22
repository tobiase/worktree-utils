package cli

import (
	"strings"
)

// FlagSet holds parsed command line flags and arguments
type FlagSet struct {
	// Common boolean flags
	Fuzzy      bool
	Help       bool
	Recursive  bool
	All        bool
	NoSwitch   bool
	Force      bool
	Check      bool
	Completion bool

	// String flags
	Base string

	// Positional arguments (non-flag arguments)
	Positional []string
}

// ParseFlags parses command line arguments into a FlagSet
func ParseFlags(args []string) *FlagSet {
	flags := &FlagSet{
		Positional: []string{},
	}

	skipNext := false
	for i, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}

		// Handle boolean and string flags
		handled, skip := parseFlag(arg, i, args, flags)
		if handled {
			if skip {
				skipNext = true
			}
			continue
		}

		// Handle combined short flags or positional arguments
		if strings.HasPrefix(arg, "-") && len(arg) > 2 && !strings.HasPrefix(arg, "--") {
			parseCombinedShortFlags(arg, flags)
		} else {
			flags.Positional = append(flags.Positional, arg)
		}
	}

	return flags
}

// parseFlag handles individual flag parsing
func parseFlag(arg string, index int, args []string, flags *FlagSet) (handled bool, skipNext bool) {
	// Boolean flags
	boolFlags := map[string]*bool{
		"--fuzzy":      &flags.Fuzzy,
		"-f":           &flags.Fuzzy,
		"--help":       &flags.Help,
		"-h":           &flags.Help,
		"--recursive":  &flags.Recursive,
		"-r":           &flags.Recursive,
		"--all":        &flags.All,
		"-a":           &flags.All,
		"--no-switch":  &flags.NoSwitch,
		"--force":      &flags.Force,
		"--check":      &flags.Check,
		"--completion": &flags.Completion,
	}

	if flagPtr, ok := boolFlags[arg]; ok {
		*flagPtr = true
		return true, false
	}

	// String flags with values
	if arg == "--base" || arg == "-b" {
		if index+1 < len(args) && !strings.HasPrefix(args[index+1], "-") {
			flags.Base = args[index+1]
			return true, true
		}
		return true, false
	}

	// Check for --flag=value syntax
	if strings.HasPrefix(arg, "--") && strings.Contains(arg, "=") {
		parts := strings.SplitN(arg, "=", 2)
		if parts[0] == "--base" {
			flags.Base = parts[1]
			return true, false
		}
	}

	return false, false
}

// parseCombinedShortFlags handles combined short flags like -rf
func parseCombinedShortFlags(arg string, flags *FlagSet) {
	shortFlags := map[rune]*bool{
		'f': &flags.Fuzzy,
		'h': &flags.Help,
		'r': &flags.Recursive,
		'a': &flags.All,
	}

	for _, c := range arg[1:] {
		if flagPtr, ok := shortFlags[c]; ok {
			*flagPtr = true
		}
	}
}

// HasFlag checks if a specific flag was set in the arguments
func HasFlag(args []string, flagName string) bool {
	for _, arg := range args {
		if arg == flagName {
			return true
		}
		// Handle --flag=value syntax
		if strings.HasPrefix(arg, flagName+"=") {
			return true
		}
	}
	return false
}

// GetFlagValue returns the value of a flag that takes a parameter
func GetFlagValue(args []string, flagName string, shortFlag string) (string, bool) {
	for i, arg := range args {
		// Check long flag
		if arg == flagName && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			return args[i+1], true
		}
		// Check short flag
		if shortFlag != "" && arg == shortFlag && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
			return args[i+1], true
		}
		// Check --flag=value syntax
		if strings.HasPrefix(arg, flagName+"=") {
			return strings.TrimPrefix(arg, flagName+"="), true
		}
	}
	return "", false
}

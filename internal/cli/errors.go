package cli

import (
	"fmt"
	"os"
)

// ExitWithError prints an error message and exits with status 1
func ExitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "wt: "+format+"\n", args...)
	os.Exit(1)
}

// HandleError checks if an error occurred and exits if it did
func HandleError(err error, context string) {
	if err != nil {
		ExitWithError("%s: %v", context, err)
	}
}

// ExitWithUsage prints a usage message and exits with status 1
func ExitWithUsage(usage string) {
	fmt.Fprintln(os.Stderr, usage)
	os.Exit(1)
}

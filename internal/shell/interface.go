package shell

import "io"

// Shell defines the interface for shell command execution
type Shell interface {
	// Run executes a command and returns an error if it fails
	Run(name string, args ...string) error

	// RunWithOutput executes a command and returns its output
	RunWithOutput(name string, args ...string) (string, error)

	// RunInDir executes a command in a specific directory
	RunInDir(dir, name string, args ...string) error

	// RunInDirWithOutput executes a command in a specific directory and returns output
	RunInDirWithOutput(dir, name string, args ...string) (string, error)

	// RunWithIO executes a command with custom stdin/stdout/stderr
	RunWithIO(stdin io.Reader, stdout, stderr io.Writer, name string, args ...string) error
}

package shell

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CommandExecutor implements Shell using os/exec
type CommandExecutor struct {
	defaultStdout io.Writer
	defaultStderr io.Writer
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor() *CommandExecutor {
	return &CommandExecutor{
		defaultStdout: os.Stdout,
		defaultStderr: os.Stderr,
	}
}

// Run executes a command and returns an error if it fails
func (e *CommandExecutor) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = e.defaultStdout
	cmd.Stderr = e.defaultStderr
	return cmd.Run()
}

// RunWithOutput executes a command and returns its output
func (e *CommandExecutor) RunWithOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("command failed: %v\nstderr: %s", err, string(exitErr.Stderr))
		}
		return "", err
	}
	return string(output), nil
}

// RunInDir executes a command in a specific directory
func (e *CommandExecutor) RunInDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = e.defaultStdout
	cmd.Stderr = e.defaultStderr
	return cmd.Run()
}

// RunInDirWithOutput executes a command in a specific directory and returns output
func (e *CommandExecutor) RunInDirWithOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("command failed: %v\nstderr: %s", err, string(exitErr.Stderr))
		}
		return "", err
	}
	return string(output), nil
}

// RunWithIO executes a command with custom stdin/stdout/stderr
func (e *CommandExecutor) RunWithIO(stdin io.Reader, stdout, stderr io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// RunShellCommand executes a command through the shell for complex commands
func (e *CommandExecutor) RunShellCommand(command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Stdout = e.defaultStdout
	cmd.Stderr = e.defaultStderr
	return cmd.Run()
}

// RunShellCommandInDir executes a shell command in a specific directory
func (e *CommandExecutor) RunShellCommandInDir(dir, command string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = dir
	cmd.Stdout = e.defaultStdout
	cmd.Stderr = e.defaultStderr
	return cmd.Run()
}

// Which checks if a command exists in PATH
func Which(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// CombineOutput runs a command and returns combined stdout and stderr
func CombineOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return strings.TrimSpace(buf.String()), err
}

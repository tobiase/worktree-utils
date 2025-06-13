package helpers

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// CommandRecorder records commands that would be executed
type CommandRecorder struct {
	mu       sync.Mutex
	Commands [][]string
	Outputs  map[string]string // Map of command -> output
	Errors   map[string]error  // Map of command -> error
}

// NewCommandRecorder creates a new command recorder
func NewCommandRecorder() *CommandRecorder {
	return &CommandRecorder{
		Commands: [][]string{},
		Outputs:  make(map[string]string),
		Errors:   make(map[string]error),
	}
}

// Record records a command execution
func (c *CommandRecorder) Record(args []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Commands = append(c.Commands, args)
}

// SetOutput sets the output for a specific command
func (c *CommandRecorder) SetOutput(command, output string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Outputs[command] = output
}

// SetError sets the error for a specific command
func (c *CommandRecorder) SetError(command string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Errors[command] = err
}

// GetOutput returns the configured output for a command
func (c *CommandRecorder) GetOutput(args []string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	key := strings.Join(args, " ")
	if err, ok := c.Errors[key]; ok {
		return "", err
	}
	
	if output, ok := c.Outputs[key]; ok {
		return output, nil
	}
	
	return "", nil
}

// CaptureOutput captures stdout and stderr during function execution
func CaptureOutput(fn func()) (stdout, stderr string, err error) {
	// Save current stdout/stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	
	// Create pipes
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	
	// Redirect stdout/stderr
	os.Stdout = wOut
	os.Stderr = wErr
	
	// Capture output in goroutines
	outChan := make(chan string)
	errChan := make(chan string)
	
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rOut)
		outChan <- buf.String()
	}()
	
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, rErr)
		errChan <- buf.String()
	}()
	
	// Run the function
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
			}
		}()
		fn()
	}()
	
	// Close write ends
	wOut.Close()
	wErr.Close()
	
	// Restore stdout/stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	
	// Get captured output
	stdout = <-outChan
	stderr = <-errChan
	
	return stdout, stderr, err
}

// RunCommand runs a command and returns output, error output, and error
func RunCommand(t *testing.T, name string, args ...string) (string, string, error) {
	t.Helper()
	
	cmd := exec.Command(name, args...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	return stdout.String(), stderr.String(), err
}

// MockExec creates a mock executable that returns predefined output
func MockExec(t *testing.T, responses map[string]string) string {
	t.Helper()
	
	// Create temp file for mock executable
	tmpFile, err := os.CreateTemp("", "mock-exec-*")
	if err != nil {
		t.Fatalf("Failed to create mock executable: %v", err)
	}
	
	// Write mock script
	script := `#!/bin/sh
args="$@"
case "$args" in
`
	for args, output := range responses {
		script += fmt.Sprintf("  \"%s\")\n    echo '%s'\n    ;;\n", args, output)
	}
	
	script += `  *)
    echo "Unexpected arguments: $args" >&2
    exit 1
    ;;
esac
`
	
	if _, err := tmpFile.WriteString(script); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to write mock script: %v", err)
	}
	
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to close mock file: %v", err)
	}
	
	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		t.Fatalf("Failed to make mock executable: %v", err)
	}
	
	return tmpFile.Name()
}

// AssertCommand verifies that a command was called with specific arguments
func AssertCommand(t *testing.T, recorder *CommandRecorder, expected []string) {
	t.Helper()
	
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	
	for _, cmd := range recorder.Commands {
		if len(cmd) == len(expected) {
			match := true
			for i := range cmd {
				if cmd[i] != expected[i] {
					match = false
					break
				}
			}
			if match {
				return
			}
		}
	}
	
	t.Errorf("Expected command not found: %v\nRecorded commands: %v", expected, recorder.Commands)
}
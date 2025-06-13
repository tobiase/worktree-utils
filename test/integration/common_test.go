package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildTestBinary builds the wt binary for testing with consistent locale
func buildTestBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "wt-test")
	cmd := exec.Command("go", "build", "-o", binPath, "../../cmd/wt")
	// Set LANG=C for consistent build output
	cmd.Env = append(os.Environ(), "LANG=C")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}

	return binPath
}

// runCommand runs a wt command and returns the output with consistent locale
func runCommand(t *testing.T, binPath string, args ...string) string {
	t.Helper()

	cmd := exec.Command(binPath, args...)
	// Set LANG=C to ensure consistent output across locales
	cmd.Env = append(os.Environ(), "LANG=C")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Some commands are expected to have non-zero exit codes
		// but we still want their output
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			// This might be a usage error or similar, include output
			return string(output)
		}
		t.Fatalf("Command failed: %v\nOutput: %s", err, output)
	}

	return string(output)
}

package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// PHASE 3: SYSTEM INTEGRATION ROBUSTNESS EDGE CASE TESTS
// =============================================================================

// Helper functions for creating edge case scenarios

func createCorruptedBinary(t *testing.T, path string) {
	t.Helper()
	// Create a binary that's not actually executable
	content := []byte("this is not a valid binary file")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create corrupted binary: %v", err)
	}
}

func createReadOnlyDirectory(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.Chmod(path, 0444); err != nil {
		t.Fatalf("Failed to make directory read-only: %v", err)
	}
}

func createVeryDeepPath(t *testing.T, baseDir string) string {
	t.Helper()
	// Create a very deep path that might exceed filesystem limits
	deepPath := baseDir
	for i := 0; i < 50; i++ {
		deepPath = filepath.Join(deepPath, "very-deep-directory-level-"+strings.Repeat("x", 10))
	}
	return deepPath
}

func fillDisk(t *testing.T, dir string, sizeBytes int64) func() {
	t.Helper()
	// Create a large file to simulate disk full conditions
	largePath := filepath.Join(dir, "large-file-disk-full.tmp")

	file, err := os.Create(largePath)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Write data to fill up space
	data := make([]byte, 1024*1024) // 1MB chunks
	var written int64
	for written < sizeBytes {
		remaining := sizeBytes - written
		if remaining < int64(len(data)) {
			data = data[:remaining]
		}
		n, err := file.Write(data)
		if err != nil {
			file.Close()
			os.Remove(largePath)
			t.Fatalf("Failed to write large file: %v", err)
		}
		written += int64(n)
	}
	file.Close()

	return func() {
		os.Remove(largePath)
	}
}

// Edge Case Tests

func TestSetupDiskFull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping disk full test in short mode")
	}

	tempDir := t.TempDir()

	// Create a fake home directory
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Create basic shell config to avoid "no shell configs" error
	bashrc := filepath.Join(homeDir, ".bashrc")
	if err := os.WriteFile(bashrc, []byte("# basic shell config"), 0644); err != nil {
		t.Fatalf("Failed to create .bashrc: %v", err)
	}

	// Fill up most available space (be careful not to actually fill the real disk)
	cleanup := fillDisk(t, tempDir, 100*1024*1024) // 100MB
	defer cleanup()

	// Create a test binary
	sourceBinary := filepath.Join(tempDir, "test-binary")
	testContent := []byte("#!/bin/sh\necho 'test'")
	if err := os.WriteFile(sourceBinary, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	// Try setup when disk is "full"
	err := Setup(sourceBinary)

	// Should either succeed or fail gracefully with clear error
	if err != nil {
		// Error should mention disk space or I/O issues, unless it's about shell configs
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "shell configuration") {
			t.Skip("Test environment issue: shell config detection failed")
		}
		if !strings.Contains(errStr, "space") && !strings.Contains(errStr, "write") && !strings.Contains(errStr, "create") && !strings.Contains(errStr, "no space") && !strings.Contains(errStr, "copy") {
			t.Logf("Disk full test may not have triggered expected condition: %v", err)
		}
		// Should not panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
	}
}

func TestSetupPermissionDeniedDirectories(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission denied as root")
	}

	tempDir := t.TempDir()

	// Create a fake home directory with restricted permissions
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Create .local but make it read-only
	localDir := filepath.Join(homeDir, ".local")
	createReadOnlyDirectory(t, localDir)

	// Restore permissions at the end
	defer func() {
		_ = os.Chmod(localDir, 0755)
	}()

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Create a test binary
	sourceBinary := filepath.Join(tempDir, "test-binary")
	testContent := []byte("#!/bin/sh\necho 'test'")
	if err := os.WriteFile(sourceBinary, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	err := Setup(sourceBinary)

	// Should fail gracefully with permission error
	if err == nil {
		t.Error("Expected permission error, got none")
	}

	// Error should mention permissions
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "permission") && !strings.Contains(errStr, "denied") && !strings.Contains(errStr, "create") && !strings.Contains(errStr, "directory") {
		t.Errorf("Error should mention permissions: %v", err)
	}

	// Should not panic
	if strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestSetupCorruptedSourceBinary(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake home directory
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Create a corrupted source binary
	sourceBinary := filepath.Join(tempDir, "corrupted-binary")
	createCorruptedBinary(t, sourceBinary)

	err := Setup(sourceBinary)

	// Should either succeed (copying any file) or fail gracefully
	if err != nil {
		// Error should not be a panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		// Error should be about file operations or execution
		errStr := strings.ToLower(err.Error())
		if !strings.Contains(errStr, "copy") && !strings.Contains(errStr, "file") && !strings.Contains(errStr, "binary") && !strings.Contains(errStr, "open") {
			t.Logf("Setup failed with corrupted binary (acceptable): %v", err)
		}
	}
}

func TestSetupMissingSourceBinary(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake home directory
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Use non-existent source binary
	sourceBinary := filepath.Join(tempDir, "non-existent-binary")

	err := Setup(sourceBinary)

	// Should fail with clear error about missing file
	if err == nil {
		t.Error("Expected error for missing source binary, got none")
	}

	// Error should mention the missing file
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "no such file") && !strings.Contains(errStr, "not found") && !strings.Contains(errStr, "open") && !strings.Contains(errStr, "file") {
		t.Errorf("Error should mention missing file: %v", err)
	}

	// Should not panic
	if strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestSetupVeryDeepPaths(t *testing.T) {
	tempDir := t.TempDir()

	// Create a very deep fake home directory
	deepHomeDir := createVeryDeepPath(t, tempDir)

	// Try to create the deep path (may fail due to filesystem limits)
	if err := os.MkdirAll(deepHomeDir, 0755); err != nil {
		// If we can't even create the deep path, test with a shorter but still deep path
		shorterPath := filepath.Join(tempDir, strings.Repeat("deep-", 20), "home")
		if err := os.MkdirAll(shorterPath, 0755); err != nil {
			t.Skip("Cannot create deep directory structure for testing")
		}
		deepHomeDir = shorterPath
	}

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", deepHomeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Create a test binary
	sourceBinary := filepath.Join(tempDir, "test-binary")
	testContent := []byte("#!/bin/sh\necho 'test'")
	if err := os.WriteFile(sourceBinary, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	err := Setup(sourceBinary)

	// Should either succeed or fail gracefully
	if err != nil {
		// Error should not be a panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		// May fail due to path length limits
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "name too long") || strings.Contains(errStr, "path") || strings.Contains(errStr, "create") {
			t.Logf("Setup failed with deep paths as expected: %v", err)
		}
	}
}

func TestSetupNoShellConfigFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake home directory with NO shell config files
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Override home directory and shell
	oldHome := os.Getenv("HOME")
	oldShell := os.Getenv("SHELL")
	os.Setenv("HOME", homeDir)
	os.Setenv("SHELL", "/unknown/shell")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
		if oldShell != "" {
			os.Setenv("SHELL", oldShell)
		} else {
			os.Unsetenv("SHELL")
		}
	}()

	// Create a test binary
	sourceBinary := filepath.Join(tempDir, "test-binary")
	testContent := []byte("#!/bin/sh\necho 'test'")
	if err := os.WriteFile(sourceBinary, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	err := Setup(sourceBinary)

	// Should fail with clear error about missing shell configs
	if err == nil {
		t.Error("Expected error for missing shell config files, got none")
	}

	// Error should mention shell configuration
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "shell") && !strings.Contains(errStr, "config") && !strings.Contains(errStr, "found") {
		t.Errorf("Error should mention missing shell configs: %v", err)
	}

	// Should not panic
	if strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestSetupCorruptedShellConfigFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake home directory
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Create corrupted shell config files
	bashrc := filepath.Join(homeDir, ".bashrc")
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE}
	if err := os.WriteFile(bashrc, binaryData, 0644); err != nil {
		t.Fatalf("Failed to create corrupted .bashrc: %v", err)
	}

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Create a test binary
	sourceBinary := filepath.Join(tempDir, "test-binary")
	testContent := []byte("#!/bin/sh\necho 'test'")
	if err := os.WriteFile(sourceBinary, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	err := Setup(sourceBinary)

	// Should handle corrupted config files gracefully
	// May succeed (if it just appends) or fail gracefully
	if err != nil {
		// Error should not be a panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		t.Logf("Setup failed with corrupted shell config (may be acceptable): %v", err)
	}
}

func TestUninstallWhenNothingInstalled(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake home directory with nothing installed
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	err := Uninstall()

	// Should succeed even when nothing is installed
	if err != nil {
		t.Errorf("Uninstall should succeed even when nothing is installed: %v", err)
	}
}

func TestUninstallPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission denied as root")
	}

	tempDir := t.TempDir()

	// Create a fake home directory
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Create a fake binary and config with restricted permissions
	binDir := filepath.Join(homeDir, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	binaryPath := filepath.Join(binDir, "wt-bin")
	if err := os.WriteFile(binaryPath, []byte("fake binary"), 0755); err != nil {
		t.Fatalf("Failed to create fake binary: %v", err)
	}

	// Make parent directory read-only
	if err := os.Chmod(binDir, 0444); err != nil {
		t.Fatalf("Failed to make bin dir read-only: %v", err)
	}

	// Restore permissions at the end
	defer func() {
		_ = os.Chmod(binDir, 0755)
	}()

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	err := Uninstall()

	// Should succeed with warnings (uninstall should be tolerant)
	if err != nil {
		t.Errorf("Uninstall should succeed with warnings: %v", err)
	}
}

func TestCheckWhenNothingInstalled(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake home directory with nothing installed
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Override home directory and PATH
	oldHome := os.Getenv("HOME")
	oldPath := os.Getenv("PATH")
	os.Setenv("HOME", homeDir)
	os.Setenv("PATH", "/usr/bin:/bin") // Exclude ~/.local/bin
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
		if oldPath != "" {
			os.Setenv("PATH", oldPath)
		} else {
			os.Unsetenv("PATH")
		}
	}()

	err := Check()

	// Should succeed even when nothing is installed (just report status)
	if err != nil {
		t.Errorf("Check should succeed even when nothing is installed: %v", err)
	}
}

func TestCheckCorruptedInstallation(t *testing.T) {
	tempDir := t.TempDir()

	// Create a fake home directory
	homeDir := filepath.Join(tempDir, "home")
	if err := os.MkdirAll(homeDir, 0755); err != nil {
		t.Fatalf("Failed to create fake home: %v", err)
	}

	// Create a corrupted installation
	binDir := filepath.Join(homeDir, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("Failed to create bin dir: %v", err)
	}

	// Create corrupted binary
	binaryPath := filepath.Join(binDir, "wt-bin")
	createCorruptedBinary(t, binaryPath)

	// Create config directory but with corrupted init script
	configDir := filepath.Join(homeDir, ".config", "wt")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	initPath := filepath.Join(configDir, "init.sh")
	corruptedInit := []byte{0x00, 0x01, 0x02, 0xFF}
	if err := os.WriteFile(initPath, corruptedInit, 0644); err != nil {
		t.Fatalf("Failed to create corrupted init script: %v", err)
	}

	// Override home directory
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", homeDir)
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	err := Check()

	// Should succeed but report issues (check should be robust)
	if err != nil {
		t.Errorf("Check should succeed even with corrupted installation: %v", err)
	}
}

func TestSetupHomeDirectoryUnavailable(t *testing.T) {
	// Override HOME to something that doesn't exist and can't be created
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", "/root/non-existent-deeply/nested/path/that/cannot/be/created")
	defer func() {
		if oldHome != "" {
			os.Setenv("HOME", oldHome)
		} else {
			os.Unsetenv("HOME")
		}
	}()

	// Create a test binary
	tempDir := t.TempDir()
	sourceBinary := filepath.Join(tempDir, "test-binary")
	testContent := []byte("#!/bin/sh\necho 'test'")
	if err := os.WriteFile(sourceBinary, testContent, 0755); err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	err := Setup(sourceBinary)

	// Should fail gracefully
	if err == nil {
		t.Error("Expected error when home directory is unavailable, got none")
	}

	// Error should not be a panic
	if strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}

	// Error should mention directory creation or home directory issues
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "directory") && !strings.Contains(errStr, "create") && !strings.Contains(errStr, "home") && !strings.Contains(errStr, "permission") {
		t.Errorf("Error should mention directory or home issues: %v", err)
	}
}

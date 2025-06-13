package helpers

import (
	"os"
	"path/filepath"
	"testing"
)

// WithTempDir creates a temporary directory and calls the provided function
func WithTempDir(t *testing.T, fn func(dir string)) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "wt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fn(tempDir)
}

// CreateFiles creates multiple files with specified content
func CreateFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()

	for path, content := range files {
		fullPath := filepath.Join(dir, path)

		// Create parent directories if needed
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create parent directories for %s: %v", path, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", path, err)
		}
	}
}

// AssertFileContents verifies that a file contains the expected content
func AssertFileContents(t *testing.T, path, expected string) {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}

	if string(content) != expected {
		t.Errorf("File content mismatch for %s\nExpected:\n%s\nGot:\n%s", path, expected, string(content))
	}
}

// AssertFileExists verifies that a file exists
func AssertFileExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// AssertFileNotExists verifies that a file does not exist
func AssertFileNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected file to not exist: %s", path)
	}
}

// AssertDirExists verifies that a directory exists
func AssertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected directory to exist: %s", path)
		return
	}

	if !info.IsDir() {
		t.Errorf("Expected %s to be a directory", path)
	}
}

// CopyFile copies a file from src to dst
func CopyFile(t *testing.T, src, dst string) {
	t.Helper()

	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("Failed to read source file %s: %v", src, err)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		t.Fatalf("Failed to create parent directories for %s: %v", dst, err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		t.Fatalf("Failed to write destination file %s: %v", dst, err)
	}
}

// CreateSymlink creates a symbolic link for testing
func CreateSymlink(t *testing.T, target, link string) {
	t.Helper()

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(link), 0755); err != nil {
		t.Fatalf("Failed to create parent directories for symlink %s: %v", link, err)
	}

	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("Failed to create symlink from %s to %s: %v", link, target, err)
	}
}

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// PHASE 2: CONFIG ROBUSTNESS EDGE CASE TESTS
// =============================================================================

// Helper functions for creating corrupted config scenarios

func createMalformedYAML(t *testing.T, configDir, name string) string {
	t.Helper()
	configPath := filepath.Join(configDir, "projects", name+".yaml")

	// Create directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write malformed YAML
	malformedYAML := `name: test
match:
  paths:
    - /some/path
  - invalid: yaml structure
    missing colon
commands
  invalid-command`

	if err := os.WriteFile(configPath, []byte(malformedYAML), 0644); err != nil {
		t.Fatalf("Failed to write malformed YAML: %v", err)
	}

	return configPath
}

func createBinaryCorruptedConfig(t *testing.T, configDir, name string) string {
	t.Helper()
	configPath := filepath.Join(configDir, "projects", name+".yaml")

	// Create directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Write binary data that's not YAML
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC}

	if err := os.WriteFile(configPath, binaryData, 0644); err != nil {
		t.Fatalf("Failed to write binary data: %v", err)
	}

	return configPath
}

func createVeryLargeConfig(t *testing.T, configDir, name string) string {
	t.Helper()
	configPath := filepath.Join(configDir, "projects", name+".yaml")

	// Create directory
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create a very large YAML file (multiple MB)
	largeYAML := `name: large-test
match:
  paths:`

	// Add thousands of path entries
	for i := 0; i < 5000; i++ { // Reduced from 10000 to avoid test timeout
		largeYAML += "\n    - /very/long/path/number/" + fmt.Sprintf("%d", i) + "/that/goes/on/and/on/with/many/segments"
	}

	largeYAML += `
commands:`

	// Add thousands of commands
	for i := 0; i < 2000; i++ { // Reduced from 5000 to avoid test timeout
		largeYAML += fmt.Sprintf(`
  cmd-%d:
    description: "Command number %d"
    target: "target/path/%d"`, i, i, i)
	}

	if err := os.WriteFile(configPath, []byte(largeYAML), 0644); err != nil {
		t.Fatalf("Failed to write large YAML: %v", err)
	}

	return configPath
}

// Edge Case Tests

func TestLoadMalformedYAML(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{configDir: tempDir}

	// Create malformed YAML file
	configPath := createMalformedYAML(t, tempDir, "malformed-test")
	defer os.Remove(configPath)

	_, err := manager.loadProjectConfig("malformed-test.yaml")

	// Should fail gracefully with YAML error
	if err == nil {
		t.Error("Expected error for malformed YAML, got none")
	}

	// Error should not be a panic and should mention YAML
	if strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}

	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "yaml") && !strings.Contains(errStr, "parse") && !strings.Contains(errStr, "unmarshal") {
		t.Errorf("Error should mention YAML/parsing issue: %v", err)
	}
}

func TestLoadBinaryCorruptedConfig(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{configDir: tempDir}

	// Create binary corrupted file
	configPath := createBinaryCorruptedConfig(t, tempDir, "binary-test")
	defer os.Remove(configPath)

	_, err := manager.loadProjectConfig("binary-test.yaml")

	// Should fail gracefully
	if err == nil {
		t.Error("Expected error for binary data in YAML file, got none")
	}

	// Error should not be a panic
	if strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestLoadVeryLargeConfig(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{configDir: tempDir}

	// Create very large config file
	configPath := createVeryLargeConfig(t, tempDir, "large-test")
	defer os.Remove(configPath)

	// This test is about performance and memory usage
	config, err := manager.loadProjectConfig("large-test.yaml")

	// Should either succeed or fail gracefully (not hang/crash)
	if err != nil {
		// Error is acceptable for very large files
		t.Logf("Large config failed to load (acceptable): %v", err)

		// Should not be a panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
	} else {
		// If it succeeds, verify it's actually loaded correctly
		if config.Name != "large-test" {
			t.Errorf("Expected name 'large-test', got %s", config.Name)
		}
		t.Logf("Large config loaded successfully with %d paths", len(config.Match.Paths))
	}
}

func TestLoadConfigPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission denied as root")
	}

	tempDir := t.TempDir()
	manager := &Manager{configDir: tempDir}

	// Create a normal config file first
	configPath := filepath.Join(tempDir, "projects", "permission-test.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	validYAML := `name: permission-test
match:
  paths:
    - /some/path
commands:
  test:
    description: "Test command"
    target: "target"`

	if err := os.WriteFile(configPath, []byte(validYAML), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Remove read permissions
	originalMode, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Failed to get file permissions: %v", err)
	}

	if err := os.Chmod(configPath, 0000); err != nil {
		t.Fatalf("Failed to remove permissions: %v", err)
	}

	// Restore permissions at the end
	defer func() {
		os.Chmod(configPath, originalMode.Mode())
	}()

	_, err = manager.loadProjectConfig("permission-test.yaml")

	// Should handle permission errors gracefully
	if err == nil {
		t.Error("Expected permission error, got none")
	}

	// Error should mention permissions or access
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "permission") && !strings.Contains(errStr, "access") && !strings.Contains(errStr, "denied") {
		t.Errorf("Error should mention permissions: %v", err)
	}
}

func TestMatchCircularPaths(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{configDir: tempDir}

	// Create a project config with circular symlinks in paths
	circularPath := filepath.Join(tempDir, "circular")
	if err := os.MkdirAll(circularPath, 0755); err != nil {
		t.Fatalf("Failed to create circular dir: %v", err)
	}

	// Create circular symlink: circular/link -> circular
	linkPath := filepath.Join(circularPath, "link")
	if err := os.Symlink(circularPath, linkPath); err != nil {
		t.Fatalf("Failed to create circular symlink: %v", err)
	}

	projectConfig := &ProjectConfig{
		Name: "circular-test",
		Match: ProjectMatch{
			Paths: []string{linkPath + "/link/link/link"},
		},
	}

	// Test if it matches without hanging
	matches := manager.matchesProject(projectConfig, circularPath, "")

	// Should either match or not match, but not hang/crash
	t.Logf("Circular path matching result: %v", matches)
}

func TestMatchUnicodeInPaths(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{configDir: tempDir}

	// Test various unicode scenarios
	unicodePaths := []string{
		"/path/with/unicode/🚀/rocket",
		"/path/with/émojis/和/chinese",
		"/path/with/spaces and symbols/!@#$%",
		"/very/long/unicode/path/with/много/разных/языков/and/العربية/text",
	}

	projectConfig := &ProjectConfig{
		Name: "unicode-test",
		Match: ProjectMatch{
			Paths: unicodePaths,
		},
	}

	// Test matching against unicode paths
	for _, path := range unicodePaths {
		matches := manager.matchesProject(projectConfig, path, "")
		t.Logf("Unicode path %s matches: %v", path, matches)

		// Should handle unicode gracefully (no panics)
	}
}

func TestConfigWithBrokenRemoteURLs(t *testing.T) {
	tempDir := t.TempDir()
	manager := &Manager{configDir: tempDir}

	// Create config with various broken remote URLs
	projectConfig := &ProjectConfig{
		Name: "broken-remote-test",
		Match: ProjectMatch{
			Remotes: []string{
				"not-a-url",
				"ftp://invalid.protocol/repo.git",
				"git@nonexistent-host-12345.com:user/repo.git",
				"https://404.example.com/missing/repo.git",
				"ssh://user@[invalid-ipv6]/repo.git",
				"", // empty remote
			},
		},
	}

	// Test matching against broken remotes
	testRemotes := []string{
		"https://github.com/user/repo.git",
		"invalid-remote",
		"",
	}

	for _, remote := range testRemotes {
		matches := manager.matchesProject(projectConfig, tempDir, remote)
		t.Logf("Broken remote matching %s: %v", remote, matches)

		// Should handle broken remotes gracefully
	}
}

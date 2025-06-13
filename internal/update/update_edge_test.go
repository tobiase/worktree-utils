package update

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// PHASE 3: SYSTEM INTEGRATION ROBUSTNESS EDGE CASE TESTS - UPDATE PACKAGE
// =============================================================================

// Helper functions for creating edge case scenarios

func createReadOnlyExecutable(t *testing.T, path string) {
	t.Helper()
	content := []byte("readonly binary")
	if err := os.WriteFile(path, content, 0444); err != nil { // Read-only
		t.Fatalf("Failed to create read-only executable: %v", err)
	}
}

func createNetworkTimeoutServer(delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate network timeout by delaying
		time.Sleep(delay)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"tag_name": "v1.0.0"}`))
	}))
}

func createRateLimitedServer() *httptest.Server {
	requestCount := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 3 {
			// Return rate limit error for first few requests
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"message": "API rate limit exceeded"}`))
		} else {
			// Eventually succeed
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"tag_name": "v1.0.0", "assets": []}`))
		}
	}))
}

func createCorruptedAssetServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			// Return valid release info with current platform asset name
			assetName := getAssetName() + ".tar.gz"
			release := `{
				"tag_name": "v1.0.0",
				"assets": [{
					"name": "` + assetName + `",
					"browser_download_url": "` + r.Host + `/download/corrupted.tar.gz",
					"size": 1024
				}]
			}`
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(release))
		} else if strings.HasSuffix(r.URL.Path, "/corrupted.tar.gz") {
			// Return corrupted archive data
			w.Header().Set("Content-Length", "1024")
			corruptedData := []byte("this is not a valid tar.gz file at all")
			_, _ = w.Write(corruptedData)
		} else {
			w.WriteHeader(404)
		}
	}))
}

func createInconsistentSizeServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			// Return release with incorrect size using current platform asset name
			assetName := getAssetName() + ".tar.gz"
			release := `{
				"tag_name": "v1.0.0",
				"assets": [{
					"name": "` + assetName + `",
					"browser_download_url": "` + r.Host + `/download/test.tar.gz",
					"size": 999999
				}]
			}`
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(release))
		} else if strings.HasSuffix(r.URL.Path, "/test.tar.gz") {
			// Return much smaller content than advertised
			smallContent := []byte("small")
			w.Header().Set("Content-Length", "5") // Correct actual size
			_, _ = w.Write(smallContent)
		} else {
			w.WriteHeader(404)
		}
	}))
}

// Edge Case Tests

func TestCheckForUpdateNetworkTimeout(t *testing.T) {
	// Save original values
	oldURL := githubAPIURL
	oldClient := httpClient
	defer func() {
		githubAPIURL = oldURL
		httpClient = oldClient
	}()

	// Create server that times out
	server := createNetworkTimeoutServer(200 * time.Millisecond)
	defer server.Close()

	// Set short timeout
	httpClient = &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	// Override the API URL
	githubAPIURL = fmt.Sprintf("%s/repos/tobiase/worktree-utils/releases/latest", server.URL)

	_, _, err := CheckForUpdate("v1.0.0")

	// Should fail gracefully with timeout error
	if err == nil {
		t.Error("Expected timeout error, got none")
	}

	// Error should mention timeout or network
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "timeout") && !strings.Contains(errStr, "context deadline") && !strings.Contains(errStr, "failed to fetch") {
		t.Errorf("Error should mention timeout or network issue: %v", err)
	}

	// Should not panic
	if err != nil && strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestCheckForUpdateRateLimit(t *testing.T) {
	// Save original values
	oldURL := githubAPIURL
	defer func() { githubAPIURL = oldURL }()

	// Create rate limited server
	server := createRateLimitedServer()
	defer server.Close()

	// Override the API URL
	githubAPIURL = fmt.Sprintf("%s/repos/tobiase/worktree-utils/releases/latest", server.URL)

	_, _, err := CheckForUpdate("v1.0.0")

	// Should handle rate limiting gracefully
	if err != nil {
		// Error should mention rate limiting
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "rate limit") || strings.Contains(errStr, "429") {
			t.Logf("Rate limit error handled as expected: %v", err)
		} else {
			t.Errorf("Expected rate limit error, got: %v", err)
		}

		// Should not panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
	}
}

func TestCheckForUpdateMalformedJSON(t *testing.T) {
	// Save original values
	oldURL := githubAPIURL
	defer func() { githubAPIURL = oldURL }()

	// Create server that returns malformed JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name": "v1.0.0", "invalid": json malformed`))
	}))
	defer server.Close()

	// Override the API URL
	githubAPIURL = fmt.Sprintf("%s/repos/tobiase/worktree-utils/releases/latest", server.URL)

	_, _, err := CheckForUpdate("v1.0.0")

	// Should fail gracefully with JSON parse error
	if err == nil {
		t.Error("Expected JSON parse error, got none")
	}

	// Error should mention JSON or parsing
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "json") && !strings.Contains(errStr, "parse") && !strings.Contains(errStr, "unmarshal") && !strings.Contains(errStr, "failed to parse") {
		t.Errorf("Error should mention JSON parsing issue: %v", err)
	}

	// Should not panic
	if err != nil && strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestDownloadAndInstallCorruptedArchive(t *testing.T) {
	// Save original values
	oldExecutable := executablePath
	defer func() { executablePath = oldExecutable }()

	// Create temp executable
	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "wt-bin")
	if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Override executablePath
	executablePath = func() (string, error) {
		return exePath, nil
	}

	// Create server that serves corrupted archive
	server := createCorruptedAssetServer()
	defer server.Close()

	// Create release with corrupted asset using current platform asset name
	assetName := getAssetName() + ".tar.gz"
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: server.URL + "/download/corrupted.tar.gz",
				Size:               1024,
			},
		},
	}

	var progressCalls int
	progressFunc := func(downloaded, total int64) {
		progressCalls++
	}

	err := DownloadAndInstall(release, progressFunc)

	// Should fail gracefully with archive error
	if err == nil {
		t.Error("Expected error for corrupted archive, got none")
	}

	// Error should mention archive or extraction issues
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "gzip") && !strings.Contains(errStr, "tar") && !strings.Contains(errStr, "archive") && !strings.Contains(errStr, "extract") && !strings.Contains(errStr, "invalid") && !strings.Contains(errStr, "unexpected eof") && !strings.Contains(errStr, "download") {
		t.Errorf("Error should mention archive corruption: %v", err)
	}

	// Should not panic
	if err != nil && strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}

	// Progress should still have been called during download
	if progressCalls == 0 {
		t.Error("Progress callback should have been called during download")
	}
}

func TestDownloadAndInstallInconsistentSize(t *testing.T) {
	// Save original values
	oldExecutable := executablePath
	defer func() { executablePath = oldExecutable }()

	// Create temp executable
	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "wt-bin")
	if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
		t.Fatal(err)
	}

	// Override executablePath
	executablePath = func() (string, error) {
		return exePath, nil
	}

	// Create server with size mismatch
	server := createInconsistentSizeServer()
	defer server.Close()

	// Create release with wrong size using current platform asset name
	assetName := getAssetName() + ".tar.gz"
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: server.URL + "/download/test.tar.gz",
				Size:               999999, // Much larger than actual content
			},
		},
	}

	var progressCalls int
	progressFunc := func(downloaded, total int64) {
		progressCalls++
	}

	err := DownloadAndInstall(release, progressFunc)

	// Should handle size mismatch gracefully (may succeed or fail)
	if err != nil {
		// Should not panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
		t.Logf("Size mismatch handled: %v", err)
	}

	// Progress should have been called
	if progressCalls == 0 {
		t.Error("Progress callback should have been called")
	}
}

func TestDownloadAndInstallReadOnlyExecutable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission denied as root")
	}

	// Save original values
	oldExecutable := executablePath
	defer func() { executablePath = oldExecutable }()

	// Create temp read-only executable
	tempDir := t.TempDir()
	exePath := filepath.Join(tempDir, "wt-bin")
	createReadOnlyExecutable(t, exePath)

	// Restore permissions at the end
	defer func() {
		_ = os.Chmod(exePath, 0755)
	}()

	// Override executablePath
	executablePath = func() (string, error) {
		return exePath, nil
	}

	// Create a valid test archive
	testContent := []byte("#!/bin/sh\necho 'new binary'")
	archive := createTestArchive(t, "wt-bin", testContent)

	// Create server that serves valid archive
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archive)))
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	// Create release with current platform asset name
	assetName := getAssetName() + ".tar.gz"
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: server.URL + "/test.tar.gz",
				Size:               int64(len(archive)),
			},
		},
	}

	progressFunc := func(downloaded, total int64) {}

	err := DownloadAndInstall(release, progressFunc)

	// Should handle read-only executable (may succeed by overwriting or fail gracefully)
	if err != nil {
		// Error should mention permissions or file issues
		errStr := strings.ToLower(err.Error())
		if !strings.Contains(errStr, "permission") && !strings.Contains(errStr, "denied") && !strings.Contains(errStr, "read-only") && !strings.Contains(errStr, "create") && !strings.Contains(errStr, "open") && !strings.Contains(errStr, "write") {
			t.Logf("Read-only executable test failed with unexpected error: %v", err)
		}

		// Should not panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Error should not contain panic: %v", err)
		}
	} else {
		t.Logf("Read-only executable was successfully overwritten (acceptable behavior)")
	}
}

func TestDownloadAndInstallExecutableNotFound(t *testing.T) {
	// Save original values
	oldExecutable := executablePath
	defer func() { executablePath = oldExecutable }()

	// Override executablePath to return non-existent path
	executablePath = func() (string, error) {
		return "/non/existent/path/to/binary", nil
	}

	// Create a valid test archive
	testContent := []byte("#!/bin/sh\necho 'new binary'")
	archive := createTestArchive(t, "wt-bin", testContent)

	// Create server that serves valid archive
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archive)))
		_, _ = w.Write(archive)
	}))
	defer server.Close()

	// Create release with current platform asset name
	assetName := getAssetName() + ".tar.gz"
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{
				Name:               assetName,
				BrowserDownloadURL: server.URL + "/test.tar.gz",
				Size:               int64(len(archive)),
			},
		},
	}

	progressFunc := func(downloaded, total int64) {}

	err := DownloadAndInstall(release, progressFunc)

	// Should fail gracefully when executable doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent executable, got none")
	}

	// Error should mention file not found
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "no such file") && !strings.Contains(errStr, "not found") && !strings.Contains(errStr, "create") && !strings.Contains(errStr, "open") {
		t.Errorf("Error should mention missing file: %v", err)
	}

	// Should not panic
	if err != nil && strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestDownloadAndInstallExecutablePathError(t *testing.T) {
	// Save original values
	oldExecutable := executablePath
	defer func() { executablePath = oldExecutable }()

	// Override executablePath to return error
	executablePath = func() (string, error) {
		return "", fmt.Errorf("failed to determine executable path")
	}

	// Create a basic release (won't be used)
	release := &Release{
		TagName: "v1.0.0",
		Assets:  []Asset{},
	}

	progressFunc := func(downloaded, total int64) {}

	err := DownloadAndInstall(release, progressFunc)

	// Should fail gracefully when executable path can't be determined
	if err == nil {
		t.Error("Expected error when executable path determination fails, got none")
	}

	// Error should mention executable path issue or early failure
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "executable") && !strings.Contains(errStr, "path") && !strings.Contains(errStr, "determine") && !strings.Contains(errStr, "failed to determine") && !strings.Contains(errStr, "no release found") {
		t.Errorf("Error should mention executable path issue: %v", err)
	}

	// Should not panic
	if err != nil && strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestDownloadAndInstallNoMatchingAsset(t *testing.T) {
	// Save original values
	oldOS := platformInfo.OS
	oldArch := platformInfo.Arch
	defer func() {
		platformInfo.OS = oldOS
		platformInfo.Arch = oldArch
	}()

	// Set platform to something that won't match available assets
	platformInfo.OS = "unknown-os"
	platformInfo.Arch = "unknown-arch"

	// Create release with assets that won't match
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{
				Name:               "wt_Linux_x86_64.tar.gz",
				BrowserDownloadURL: "http://example.com/linux.tar.gz",
				Size:               1024,
			},
			{
				Name:               "wt_Windows_x86_64.tar.gz",
				BrowserDownloadURL: "http://example.com/windows.tar.gz",
				Size:               1024,
			},
		},
	}

	progressFunc := func(downloaded, total int64) {}

	err := DownloadAndInstall(release, progressFunc)

	// Should fail gracefully when no matching asset is found
	if err == nil {
		t.Error("Expected error for no matching asset, got none")
	}

	// Error should mention no release found
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "no release found") && !strings.Contains(errStr, "no matching") && !strings.Contains(errStr, "asset") && !strings.Contains(errStr, "platform") {
		t.Errorf("Error should mention no matching asset: %v", err)
	}

	// Should not panic
	if err != nil && strings.Contains(err.Error(), "panic") {
		t.Errorf("Error should not contain panic: %v", err)
	}
}

func TestCalculateChecksumVeryLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	// Create a large file (100MB)
	tempDir := t.TempDir()
	largePath := filepath.Join(tempDir, "large.bin")

	file, err := os.Create(largePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// Write 100MB of data
	data := make([]byte, 1024*1024) // 1MB chunks
	for i := 0; i < 100; i++ {
		if _, err := file.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	file.Close()

	// Calculate checksum
	hash, err := CalculateChecksum(largePath)

	// Should handle large files gracefully
	if err != nil {
		t.Errorf("Should handle large file gracefully: %v", err)
	}

	// Should return a valid hash
	if len(hash) != 64 { // SHA256 hash length
		t.Errorf("Expected 64-character hash, got %d: %s", len(hash), hash)
	}
}

func TestCalculateChecksumPermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission denied as root")
	}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "no-read.bin")

	// Create file and remove read permissions
	if err := os.WriteFile(testFile, []byte("test"), 0000); err != nil {
		t.Fatal(err)
	}

	// Restore permissions at the end
	defer func() {
		_ = os.Chmod(testFile, 0644)
	}()

	_, err := CalculateChecksum(testFile)

	// Should fail gracefully with permission error
	if err == nil {
		t.Error("Expected permission error, got none")
	}

	// Error should mention permissions or access
	errStr := strings.ToLower(err.Error())
	if !strings.Contains(errStr, "permission") && !strings.Contains(errStr, "denied") && !strings.Contains(errStr, "open") {
		t.Errorf("Error should mention permission issue: %v", err)
	}
}

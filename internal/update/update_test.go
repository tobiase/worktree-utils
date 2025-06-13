package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Create a mock GitHub release response

// Create a test tar.gz archive with a binary
func createTestArchive(t *testing.T, binaryName string, content []byte) []byte {
	var buf bytes.Buffer

	// Create gzip writer
	gw := gzip.NewWriter(&buf)

	// Create tar writer
	tw := tar.NewWriter(gw)

	// Add file to archive
	header := &tar.Header{
		Name:    binaryName,
		Mode:    0755,
		Size:    int64(len(content)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		t.Fatal(err)
	}

	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}

	// Close writers
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func TestCheckForUpdate(t *testing.T) {
	// Save original values
	oldURL := githubAPIURL
	defer func() { githubAPIURL = oldURL }()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/tobiase/worktree-utils/releases/latest" {
			w.WriteHeader(404)
			return
		}

		// Return mock release
		release := Release{
			TagName:     "v1.2.0",
			Name:        "Release v1.2.0",
			Body:        "Bug fixes and improvements",
			PublishedAt: time.Now().Format(time.RFC3339),
			Assets: []Asset{
				{
					Name:               "wt_Darwin_all.tar.gz",
					BrowserDownloadURL: r.Host + "/download/wt_Darwin_all.tar.gz",
					Size:               1024,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Override the API URL for testing
	githubAPIURL = fmt.Sprintf("%s/repos/tobiase/worktree-utils/releases/latest", server.URL)

	tests := []struct {
		name           string
		currentVersion string
		wantUpdate     bool
		wantError      bool
	}{
		{
			name:           "update available",
			currentVersion: "v1.0.0",
			wantUpdate:     true,
		},
		{
			name:           "already on latest",
			currentVersion: "v1.2.0",
			wantUpdate:     false,
		},
		{
			name:           "dev version",
			currentVersion: "dev",
			wantUpdate:     false,
		},
		{
			name:           "version without v prefix",
			currentVersion: "1.0.0",
			wantUpdate:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release, hasUpdate, err := CheckForUpdate(tt.currentVersion)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if hasUpdate != tt.wantUpdate {
					t.Errorf("Expected hasUpdate=%v, got %v", tt.wantUpdate, hasUpdate)
				}

				if release == nil {
					t.Error("Expected release to be non-nil")
				} else if release.TagName != "v1.2.0" {
					t.Errorf("Expected tag v1.2.0, got %s", release.TagName)
				}
			}
		})
	}
}

func TestCheckForUpdateErrors(t *testing.T) {
	// Save original values
	oldURL := githubAPIURL
	oldClient := httpClient
	defer func() {
		githubAPIURL = oldURL
		httpClient = oldClient
	}()

	tests := []struct {
		name      string
		handler   http.HandlerFunc
		setupFunc func()
		wantError string
	}{
		{
			name: "server error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(500)
			},
			wantError: "GitHub API returned status 500",
		},
		{
			name: "invalid JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, _ = w.Write([]byte("invalid json"))
			},
			wantError: "failed to parse release info",
		},
		{
			name: "network timeout",
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Simulate timeout by not responding
				time.Sleep(100 * time.Millisecond)
			},
			setupFunc: func() {
				httpClient = &http.Client{
					Timeout: 50 * time.Millisecond,
				}
			},
			wantError: "failed to fetch release info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			// Override the API URL
			githubAPIURL = fmt.Sprintf("%s/repos/tobiase/worktree-utils/releases/latest", server.URL)

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			_, _, err := CheckForUpdate("v1.0.0")

			if err == nil {
				t.Error("Expected error but got none")
			} else if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Expected error containing %q, got %q", tt.wantError, err.Error())
			}
		})
	}
}

func TestGetAssetName(t *testing.T) {
	// Save original values
	oldOS := platformInfo.OS
	oldArch := platformInfo.Arch
	defer func() {
		platformInfo.OS = oldOS
		platformInfo.Arch = oldArch
	}()

	tests := []struct {
		name     string
		goos     string
		goarch   string
		expected string
	}{
		{
			name:     "macOS universal binary",
			goos:     "darwin",
			goarch:   "amd64",
			expected: "wt_Darwin_all",
		},
		{
			name:     "macOS arm64",
			goos:     "darwin",
			goarch:   "arm64",
			expected: "wt_Darwin_all",
		},
		{
			name:     "Linux amd64",
			goos:     "linux",
			goarch:   "amd64",
			expected: "wt_Linux_x86_64",
		},
		{
			name:     "Linux arm64",
			goos:     "linux",
			goarch:   "arm64",
			expected: "wt_Linux_arm64",
		},
		{
			name:     "Windows amd64",
			goos:     "windows",
			goarch:   "amd64",
			expected: "wt_Windows_x86_64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Override platform values
			platformInfo.OS = tt.goos
			platformInfo.Arch = tt.goarch

			result := getAssetName()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDownloadAndInstall(t *testing.T) {
	// Create test archives for both binary names
	testBinaryContent := []byte("#!/bin/sh\necho 'test binary v2.0.0'")
	archiveWtBin := createTestArchive(t, "wt-bin", testBinaryContent)
	archiveWorktreeUtils := createTestArchive(t, "worktree-utils", testBinaryContent)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "test-wt-bin.tar.gz"):
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archiveWtBin)))
			_, _ = w.Write(archiveWtBin)
		case strings.HasSuffix(r.URL.Path, "test-worktree-utils.tar.gz"):
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archiveWorktreeUtils)))
			_, _ = w.Write(archiveWorktreeUtils)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	assetName := getAssetName()
	tests := []struct {
		name      string
		release   *Release
		wantError bool
		errorMsg  string
	}{
		{
			name: "successful install with exact match",
			release: &Release{
				TagName: "v2.0.0",
				Assets: []Asset{
					{
						Name:               fmt.Sprintf("%s.tar.gz", assetName),
						BrowserDownloadURL: server.URL + "/download/test-wt-bin.tar.gz",
						Size:               int64(len(archiveWtBin)),
					},
				},
			},
			wantError: false,
		},
		{
			name: "successful install with versioned asset name",
			release: &Release{
				TagName: "v2.0.0",
				Assets: []Asset{
					{
						Name:               fmt.Sprintf("wt_2.0.0_%s.tar.gz", assetName[3:]), // Remove "wt_" prefix and add version
						BrowserDownloadURL: server.URL + "/download/test-wt-bin.tar.gz",
						Size:               int64(len(archiveWtBin)),
					},
				},
			},
			wantError: false,
		},
		{
			name: "successful install with worktree-utils binary name",
			release: &Release{
				TagName: "v2.0.0",
				Assets: []Asset{
					{
						Name:               fmt.Sprintf("%s.tar.gz", assetName),
						BrowserDownloadURL: server.URL + "/download/test-worktree-utils.tar.gz",
						Size:               int64(len(archiveWorktreeUtils)),
					},
				},
			},
			wantError: false,
		},
		{
			name: "no matching asset",
			release: &Release{
				TagName: "v2.0.0",
				Assets: []Asset{
					{
						Name:               "wrong_platform.tar.gz",
						BrowserDownloadURL: server.URL + "/download/wrong.tar.gz",
						Size:               1024,
					},
				},
			},
			wantError: true,
			errorMsg:  "no release found",
		},
		{
			name: "download failure",
			release: &Release{
				TagName: "v2.0.0",
				Assets: []Asset{
					{
						Name:               fmt.Sprintf("%s.tar.gz", assetName),
						BrowserDownloadURL: server.URL + "/nonexistent.tar.gz",
						Size:               1024,
					},
				},
			},
			wantError: true,
			errorMsg:  "failed to download update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary executable
			tempDir := t.TempDir()
			exePath := filepath.Join(tempDir, "wt-bin")
			if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
				t.Fatal(err)
			}

			// Override executablePath
			oldExecutable := executablePath
			executablePath = func() (string, error) {
				return exePath, nil
			}
			defer func() { executablePath = oldExecutable }()

			// Track progress calls
			var progressCalls int
			progressFunc := func(downloaded, total int64) {
				progressCalls++
			}

			err := DownloadAndInstall(tt.release, progressFunc)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got none for test %s", tt.name)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Check binary was replaced
				content, err := os.ReadFile(exePath)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(content, testBinaryContent) {
					t.Error("Binary was not updated correctly")
				}

				// Check progress was called
				if progressCalls == 0 {
					t.Error("Progress callback was not called")
				}
			}
		})
	}
}

func TestAssetNameMatching(t *testing.T) {
	tests := []struct {
		name        string
		assetName   string
		targetName  string
		shouldMatch bool
	}{
		{
			name:        "exact match",
			assetName:   "wt_Darwin_all",
			targetName:  "wt_Darwin_all",
			shouldMatch: true,
		},
		{
			name:        "exact match with tar.gz",
			assetName:   "wt_Darwin_all.tar.gz",
			targetName:  "wt_Darwin_all",
			shouldMatch: true,
		},
		{
			name:        "versioned asset name",
			assetName:   "wt_0.4.0_Darwin_all.tar.gz",
			targetName:  "wt_Darwin_all",
			shouldMatch: true,
		},
		{
			name:        "different version",
			assetName:   "wt_1.2.3_Darwin_all.tar.gz",
			targetName:  "wt_Darwin_all",
			shouldMatch: true,
		},
		{
			name:        "linux asset",
			assetName:   "wt_0.4.0_Linux_x86_64.tar.gz",
			targetName:  "wt_Linux_x86_64",
			shouldMatch: true,
		},
		{
			name:        "wrong platform",
			assetName:   "wt_0.4.0_Linux_x86_64.tar.gz",
			targetName:  "wt_Darwin_all",
			shouldMatch: false,
		},
		{
			name:        "completely different asset",
			assetName:   "checksums.txt",
			targetName:  "wt_Darwin_all",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the matching logic from DownloadAndInstall
			matched := tt.assetName == tt.targetName ||
				tt.assetName == tt.targetName+".tar.gz" ||
				strings.HasSuffix(tt.assetName, "_"+tt.targetName[3:]+".tar.gz")

			if matched != tt.shouldMatch {
				t.Errorf("Asset matching failed: asset=%s, target=%s, expected=%v, got=%v",
					tt.assetName, tt.targetName, tt.shouldMatch, matched)
			}
		})
	}
}

func TestBinaryNameRecognition(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "wt-bin exact match",
			filename: "wt-bin",
			expected: true,
		},
		{
			name:     "worktree-utils exact match",
			filename: "worktree-utils",
			expected: true,
		},
		{
			name:     "wt-bin with path",
			filename: "some/path/wt-bin",
			expected: true,
		},
		{
			name:     "worktree-utils with path",
			filename: "directory/worktree-utils",
			expected: true,
		},
		{
			name:     "wrong binary name",
			filename: "some-other-binary",
			expected: false,
		},
		{
			name:     "README file",
			filename: "README.md",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBinaryFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isBinaryFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestExtractAndInstall(t *testing.T) {
	tests := []struct {
		name        string
		archiveFunc func(t *testing.T) string
		wantError   bool
		errorMsg    string
	}{
		{
			name: "successful extraction",
			archiveFunc: func(t *testing.T) string {
				archive := createTestArchive(t, "wt-bin", []byte("new binary"))
				tempFile := filepath.Join(t.TempDir(), "archive.tar.gz")
				_ = os.WriteFile(tempFile, archive, 0644)
				return tempFile
			},
		},
		{
			name: "binary not in archive",
			archiveFunc: func(t *testing.T) string {
				archive := createTestArchive(t, "wrong-name", []byte("content"))
				tempFile := filepath.Join(t.TempDir(), "archive.tar.gz")
				_ = os.WriteFile(tempFile, archive, 0644)
				return tempFile
			},
			wantError: true,
			errorMsg:  "binary not found in archive",
		},
		{
			name: "corrupt archive",
			archiveFunc: func(t *testing.T) string {
				tempFile := filepath.Join(t.TempDir(), "corrupt.tar.gz")
				_ = os.WriteFile(tempFile, []byte("not a tar.gz"), 0644)
				return tempFile
			},
			wantError: true,
			errorMsg:  "failed to create gzip reader",
		},
		{
			name: "archive not found",
			archiveFunc: func(t *testing.T) string {
				return "/nonexistent/archive.tar.gz"
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary executable
			tempDir := t.TempDir()
			exePath := filepath.Join(tempDir, "wt-bin")
			if err := os.WriteFile(exePath, []byte("old binary"), 0755); err != nil {
				t.Fatal(err)
			}

			// Override executablePath
			oldExecutable := executablePath
			executablePath = func() (string, error) {
				return exePath, nil
			}
			defer func() { executablePath = oldExecutable }()

			archivePath := tt.archiveFunc(t)
			err := extractAndInstall(archivePath)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				// Check binary was replaced
				content, err := os.ReadFile(exePath)
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != "new binary" {
					t.Error("Binary was not updated")
				}
			}
		})
	}
}

func TestProgressWriter(t *testing.T) {
	tests := []struct {
		name          string
		total         int64
		writes        [][]byte
		expectedCalls int
	}{
		{
			name:  "single write",
			total: 100,
			writes: [][]byte{
				[]byte("hello"),
			},
			expectedCalls: 1,
		},
		{
			name:  "multiple writes",
			total: 100,
			writes: [][]byte{
				[]byte("hello"),
				[]byte("world"),
				[]byte("test"),
			},
			expectedCalls: 3,
		},
		{
			name:  "empty write",
			total: 100,
			writes: [][]byte{
				[]byte(""),
			},
			expectedCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			var totalDownloaded int64

			pw := &ProgressWriter{
				Total: tt.total,
				OnProgress: func(downloaded, total int64) {
					calls++
					totalDownloaded = downloaded
					if total != tt.total {
						t.Errorf("Expected total %d, got %d", tt.total, total)
					}
				},
			}

			var expectedBytes int
			for _, data := range tt.writes {
				n, err := pw.Write(data)
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if n != len(data) {
					t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
				}
				expectedBytes += len(data)
			}

			if calls != tt.expectedCalls {
				t.Errorf("Expected %d progress calls, got %d", tt.expectedCalls, calls)
			}

			if totalDownloaded != int64(expectedBytes) {
				t.Errorf("Expected total downloaded %d, got %d", expectedBytes, totalDownloaded)
			}
		})
	}
}

func TestProgressWriterNilCallback(t *testing.T) {
	pw := &ProgressWriter{
		Total:      100,
		OnProgress: nil,
	}

	// Should not panic
	data := []byte("test data")
	n, err := pw.Write(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if pw.Downloaded != int64(len(data)) {
		t.Errorf("Expected downloaded %d, got %d", len(data), pw.Downloaded)
	}
}

func TestCalculateChecksum(t *testing.T) {
	tests := []struct {
		name         string
		content      []byte
		expectedHash string
		wantError    bool
	}{
		{
			name:         "simple content",
			content:      []byte("hello world"),
			expectedHash: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		{
			name:         "empty file",
			content:      []byte(""),
			expectedHash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:         "binary content",
			content:      []byte{0x00, 0xFF, 0x42, 0x13, 0x37},
			expectedHash: "9247fbe7e59629112c3950ac4b7d24d207bce105f59fa316649af04f3eaf2405",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tempFile := filepath.Join(t.TempDir(), "test.bin")
			if err := os.WriteFile(tempFile, tt.content, 0644); err != nil {
				t.Fatal(err)
			}

			hash, err := CalculateChecksum(tempFile)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if hash != tt.expectedHash {
					t.Errorf("Expected hash %s, got %s", tt.expectedHash, hash)
				}
			}
		})
	}
}

func TestCalculateChecksumErrors(t *testing.T) {
	tests := []struct {
		name     string
		filepath string
	}{
		{
			name:     "file not found",
			filepath: "/nonexistent/file.bin",
		},
		{
			name:     "directory instead of file",
			filepath: t.TempDir(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CalculateChecksum(tt.filepath)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestDownloadFile(t *testing.T) {
	testContent := []byte("test file content")

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/success":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
			_, _ = w.Write(testContent)
		case "/error":
			w.WriteHeader(500)
		case "/slow":
			// Simulate slow download
			time.Sleep(100 * time.Millisecond)
			_, _ = w.Write(testContent)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	tests := []struct {
		name           string
		url            string
		size           int64
		wantError      bool
		wantProgress   bool
		expectedOutput []byte
	}{
		{
			name:           "successful download",
			url:            server.URL + "/success",
			size:           int64(len(testContent)),
			wantProgress:   true,
			expectedOutput: testContent,
		},
		{
			name:      "server error",
			url:       server.URL + "/error",
			size:      100,
			wantError: true,
		},
		{
			name:      "not found",
			url:       server.URL + "/notfound",
			size:      100,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var progressCalled bool

			err := downloadFile(&buf, tt.url, tt.size, func(downloaded, total int64) {
				progressCalled = true
			})

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}

				if !bytes.Equal(buf.Bytes(), tt.expectedOutput) {
					t.Error("Downloaded content doesn't match expected")
				}

				if tt.wantProgress && !progressCalled {
					t.Error("Progress callback was not called")
				}
			}
		})
	}
}

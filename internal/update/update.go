package update

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const userAgent = "wt-updater"

var (
	// githubAPIURL can be overridden for testing
	githubAPIURL = "https://api.github.com/repos/tobiase/worktree-utils/releases/latest"
	// httpClient can be overridden for testing
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

// Release represents a GitHub release
type Release struct {
	TagName     string   `json:"tag_name"`
	Name        string   `json:"name"`
	Body        string   `json:"body"`
	PublishedAt string   `json:"published_at"`
	Assets      []Asset  `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// ProgressWriter wraps io.Writer to report progress
type ProgressWriter struct {
	Total      int64
	Downloaded int64
	OnProgress func(downloaded, total int64)
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Downloaded += int64(n)
	if pw.OnProgress != nil {
		pw.OnProgress(pw.Downloaded, pw.Total)
	}
	return n, nil
}

// CheckForUpdate checks if a new version is available
func CheckForUpdate(currentVersion string) (*Release, bool, error) {
	release, err := getLatestRelease()
	if err != nil {
		return nil, false, err
	}

	// Clean version strings for comparison
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(release.TagName, "v")

	// Simple version comparison (could be improved)
	hasUpdate := latest != current && current != "dev"

	return release, hasUpdate, nil
}

// getLatestRelease fetches the latest release information from GitHub
func getLatestRelease() (*Release, error) {
	client := httpClient

	req, err := http.NewRequest("GET", githubAPIURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

// DownloadAndInstall downloads and installs the update
func DownloadAndInstall(release *Release, onProgress func(downloaded, total int64)) error {
	// Find the appropriate asset for this platform
	assetName := getAssetName()
	var asset *Asset
	for _, a := range release.Assets {
		if a.Name == assetName || a.Name == assetName+".tar.gz" {
			asset = &a
			break
		}
	}

	if asset == nil {
		return fmt.Errorf("no release found for %s", assetName)
	}

	// Create temporary file for download
	tmpFile, err := os.CreateTemp("", "wt-update-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Download the asset
	if err := downloadFile(tmpFile, asset.BrowserDownloadURL, asset.Size, onProgress); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Extract and install
	if err := extractAndInstall(tmpFile.Name()); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	return nil
}

// platformInfo provides OS and architecture info (can be overridden for testing)
var platformInfo = struct {
	OS   string
	Arch string
}{
	OS:   runtime.GOOS,
	Arch: runtime.GOARCH,
}

// getAssetName returns the expected asset name for the current platform
func getAssetName() string {
	os := platformInfo.OS
	arch := platformInfo.Arch
	
	// Map to match GoReleaser output
	if os == "darwin" {
		// Universal binary for macOS
		return "wt_Darwin_all"
	}
	
	// For Linux, map architecture names
	if arch == "amd64" {
		arch = "x86_64"
	}
	
	return fmt.Sprintf("wt_%s_%s", strings.Title(os), arch)
}

// downloadFile downloads a file with progress reporting
func downloadFile(dst io.Writer, url string, size int64, onProgress func(downloaded, total int64)) error {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	pw := &ProgressWriter{
		Total:      size,
		OnProgress: onProgress,
	}

	writer := io.MultiWriter(dst, pw)
	_, err = io.Copy(writer, resp.Body)
	return err
}

// executablePath is a wrapper for os.Executable that can be overridden for testing
var executablePath = os.Executable

// extractAndInstall extracts the binary from the tarball and installs it
func extractAndInstall(tarPath string) error {
	// Get current executable path
	exePath, err := executablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Open tarball
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Find and extract the binary
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		// Look for the wt-bin file
		if header.Name == "wt-bin" || filepath.Base(header.Name) == "wt-bin" {
			// Create temporary file for new binary
			tmpBin, err := os.CreateTemp(filepath.Dir(exePath), "wt-bin-new-*")
			if err != nil {
				return fmt.Errorf("failed to create temp binary: %w", err)
			}
			tmpPath := tmpBin.Name()
			defer os.Remove(tmpPath)

			// Copy binary content
			if _, err := io.Copy(tmpBin, tr); err != nil {
				tmpBin.Close()
				return fmt.Errorf("failed to extract binary: %w", err)
			}
			tmpBin.Close()

			// Make it executable
			if err := os.Chmod(tmpPath, 0755); err != nil {
				return fmt.Errorf("failed to make binary executable: %w", err)
			}

			// Atomically replace the old binary
			if err := os.Rename(tmpPath, exePath); err != nil {
				return fmt.Errorf("failed to replace binary: %w", err)
			}

			return nil
		}
	}

	return fmt.Errorf("binary not found in archive")
}

// CalculateChecksum calculates SHA256 checksum of a file
func CalculateChecksum(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
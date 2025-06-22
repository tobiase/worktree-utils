package filesystem

import (
	"io"
	"os"
	"path/filepath"
)

// OSFilesystem implements Filesystem using the standard os package
type OSFilesystem struct{}

// NewOSFilesystem creates a new OS-based filesystem
func NewOSFilesystem() *OSFilesystem {
	return &OSFilesystem{}
}

// ReadFile reads the entire file and returns its contents
func (f *OSFilesystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file with the specified permissions
func (f *OSFilesystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// MkdirAll creates a directory and all necessary parents
func (f *OSFilesystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Remove removes a file or empty directory
func (f *OSFilesystem) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll removes a path and any children it contains
func (f *OSFilesystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Stat returns file info for a path
func (f *OSFilesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// Open opens a file for reading
func (f *OSFilesystem) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

// Create creates or truncates a file
func (f *OSFilesystem) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

// Chmod changes the mode of the file
func (f *OSFilesystem) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// UserHomeDir returns the current user's home directory
func (f *OSFilesystem) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// Getwd returns the current working directory
func (f *OSFilesystem) Getwd() (string, error) {
	return os.Getwd()
}

// Chdir changes the current working directory
func (f *OSFilesystem) Chdir(dir string) error {
	return os.Chdir(dir)
}

// Walk walks the file tree rooted at root
func (f *OSFilesystem) Walk(root string, fn func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(root, fn)
}

// Exists checks if a path exists
func (f *OSFilesystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func (f *OSFilesystem) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

package filesystem

import (
	"io"
	"os"
)

// Filesystem defines the interface for filesystem operations
type Filesystem interface {
	// ReadFile reads the entire file and returns its contents
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file with the specified permissions
	WriteFile(path string, data []byte, perm os.FileMode) error

	// MkdirAll creates a directory and all necessary parents
	MkdirAll(path string, perm os.FileMode) error

	// Remove removes a file or empty directory
	Remove(path string) error

	// RemoveAll removes a path and any children it contains
	RemoveAll(path string) error

	// Stat returns file info for a path
	Stat(path string) (os.FileInfo, error)

	// Open opens a file for reading
	Open(path string) (io.ReadCloser, error)

	// Create creates or truncates a file
	Create(path string) (io.WriteCloser, error)

	// Chmod changes the mode of the file
	Chmod(path string, mode os.FileMode) error

	// UserHomeDir returns the current user's home directory
	UserHomeDir() (string, error)

	// Getwd returns the current working directory
	Getwd() (string, error)

	// Chdir changes the current working directory
	Chdir(dir string) error

	// Walk walks the file tree rooted at root
	Walk(root string, fn func(path string, info os.FileInfo, err error) error) error

	// Exists checks if a path exists
	Exists(path string) bool

	// IsDir checks if a path is a directory
	IsDir(path string) bool
}

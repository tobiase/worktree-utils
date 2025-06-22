package worktree

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/tobiase/worktree-utils/internal/git"
)

// MockGitClient is a mock implementation of git.Client for testing
type MockGitClient struct {
	// RevParse mock
	RevParseFunc func(args ...string) (string, error)

	// WorktreeList mock
	WorktreeListFunc func() ([]git.GitWorktree, error)

	// WorktreeAdd mock
	WorktreeAddFunc func(path, branch string, options ...string) error

	// WorktreeRemove mock
	WorktreeRemoveFunc func(path string) error

	// ShowRef mock
	ShowRefFunc func(ref string) error

	// GetRemoteURL mock
	GetRemoteURLFunc func(remote string) (string, error)
}

func (m *MockGitClient) RevParse(args ...string) (string, error) {
	if m.RevParseFunc != nil {
		return m.RevParseFunc(args...)
	}
	return "", nil
}

func (m *MockGitClient) WorktreeList() ([]git.GitWorktree, error) {
	if m.WorktreeListFunc != nil {
		return m.WorktreeListFunc()
	}
	return nil, nil
}

func (m *MockGitClient) WorktreeAdd(path, branch string, options ...string) error {
	if m.WorktreeAddFunc != nil {
		return m.WorktreeAddFunc(path, branch, options...)
	}
	return nil
}

func (m *MockGitClient) WorktreeRemove(path string) error {
	if m.WorktreeRemoveFunc != nil {
		return m.WorktreeRemoveFunc(path)
	}
	return nil
}

func (m *MockGitClient) ShowRef(ref string) error {
	if m.ShowRefFunc != nil {
		return m.ShowRefFunc(ref)
	}
	return nil
}

func (m *MockGitClient) GetRemoteURL(remote string) (string, error) {
	if m.GetRemoteURLFunc != nil {
		return m.GetRemoteURLFunc(remote)
	}
	return "", nil
}

// Implement remaining interface methods with default behavior
func (m *MockGitClient) BranchList() ([]string, error)                        { return nil, nil }
func (m *MockGitClient) GetCurrentBranch() (string, error)                    { return "", nil }
func (m *MockGitClient) Log(format string, options ...string) (string, error) { return "", nil }
func (m *MockGitClient) Status(options ...string) (string, error)             { return "", nil }
func (m *MockGitClient) RevList(options ...string) (string, error)            { return "", nil }

// MockFilesystem is a mock implementation of filesystem.Filesystem
type MockFilesystem struct {
	Files     map[string][]byte
	Dirs      map[string]bool
	ExistsMap map[string]bool
}

func NewMockFilesystem() *MockFilesystem {
	return &MockFilesystem{
		Files:     make(map[string][]byte),
		Dirs:      make(map[string]bool),
		ExistsMap: make(map[string]bool),
	}
}

func (m *MockFilesystem) ReadFile(path string) ([]byte, error) {
	if content, ok := m.Files[path]; ok {
		return content, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFilesystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.Files[path] = data
	return nil
}

func (m *MockFilesystem) MkdirAll(path string, perm os.FileMode) error {
	m.Dirs[path] = true
	return nil
}

func (m *MockFilesystem) Exists(path string) bool {
	if exists, ok := m.ExistsMap[path]; ok {
		return exists
	}
	_, fileExists := m.Files[path]
	_, dirExists := m.Dirs[path]
	return fileExists || dirExists
}

func (m *MockFilesystem) Stat(path string) (os.FileInfo, error) {
	return nil, nil // Simplified for this example
}

// Implement remaining interface methods with default behavior
func (m *MockFilesystem) Remove(path string) error                   { return nil }
func (m *MockFilesystem) RemoveAll(path string) error                { return nil }
func (m *MockFilesystem) Open(path string) (io.ReadCloser, error)    { return nil, nil }
func (m *MockFilesystem) Create(path string) (io.WriteCloser, error) { return nil, nil }
func (m *MockFilesystem) Chmod(path string, mode os.FileMode) error  { return nil }
func (m *MockFilesystem) UserHomeDir() (string, error)               { return "", nil }
func (m *MockFilesystem) Getwd() (string, error)                     { return "", nil }
func (m *MockFilesystem) Chdir(dir string) error                     { return nil }
func (m *MockFilesystem) Walk(root string, fn func(string, os.FileInfo, error) error) error {
	return nil
}
func (m *MockFilesystem) IsDir(path string) bool { return false }

// MockShell is a mock implementation of shell.Shell
type MockShell struct {
	Commands []string
}

func (m *MockShell) Run(name string, args ...string) error {
	m.Commands = append(m.Commands, name+" "+joinArgs(args))
	return nil
}

func (m *MockShell) RunWithOutput(name string, args ...string) (string, error) {
	m.Commands = append(m.Commands, name+" "+joinArgs(args))
	return "", nil
}

func (m *MockShell) RunInDir(dir, name string, args ...string) error {
	m.Commands = append(m.Commands, "cd "+dir+" && "+name+" "+joinArgs(args))
	return nil
}

func (m *MockShell) RunInDirWithOutput(dir, name string, args ...string) (string, error) {
	return "", nil
}

func (m *MockShell) RunWithIO(stdin io.Reader, stdout, stderr io.Writer, name string, args ...string) error {
	return nil
}

func joinArgs(args []string) string {
	result := ""
	for i, arg := range args {
		if i > 0 {
			result += " "
		}
		result += arg
	}
	return result
}

// TestServiceGetRepoRoot demonstrates testing with mocked dependencies
func TestServiceGetRepoRoot(t *testing.T) {
	tests := []struct {
		name      string
		revParse  func(args ...string) (string, error)
		want      string
		wantError bool
	}{
		{
			name: "success",
			revParse: func(args ...string) (string, error) {
				return "/home/user/project", nil
			},
			want:      "/home/user/project",
			wantError: false,
		},
		{
			name: "not in git repo",
			revParse: func(args ...string) (string, error) {
				return "", errors.New("not a git repository")
			},
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockGit := &MockGitClient{
				RevParseFunc: tt.revParse,
			}

			service := NewService(mockGit, nil, nil)
			got, err := service.GetRepoRoot()

			if (err != nil) != tt.wantError {
				t.Errorf("GetRepoRoot() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if got != tt.want {
				t.Errorf("GetRepoRoot() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestServiceAdd demonstrates testing the Add function with mocks
func TestServiceAdd(t *testing.T) {
	mockGit := &MockGitClient{
		RevParseFunc: func(args ...string) (string, error) {
			return "/repo", nil
		},
		WorktreeAddFunc: func(path, branch string, options ...string) error {
			if branch == "error-branch" {
				return errors.New("branch already exists")
			}
			return nil
		},
	}

	mockFS := NewMockFilesystem()
	mockShell := &MockShell{}

	service := NewService(mockGit, mockFS, mockShell)

	// Test successful add
	err := service.Add("feature", nil)
	if err != nil {
		t.Errorf("Add() unexpected error: %v", err)
	}

	// Verify directory was created
	if !mockFS.Dirs["/repo-worktrees"] {
		t.Error("Expected worktree base directory to be created")
	}

	// Test error case
	err = service.Add("error-branch", nil)
	if err == nil {
		t.Error("Add() expected error for existing branch")
	}
}

// TestServiceCheckBranchExists demonstrates mocking branch existence checks
func TestServiceCheckBranchExists(t *testing.T) {
	mockGit := &MockGitClient{
		ShowRefFunc: func(ref string) error {
			if ref == "refs/heads/main" {
				return nil
			}
			return errors.New("ref not found")
		},
	}

	service := NewService(mockGit, nil, nil)

	if !service.CheckBranchExists("main") {
		t.Error("Expected 'main' branch to exist")
	}

	if service.CheckBranchExists("nonexistent") {
		t.Error("Expected 'nonexistent' branch to not exist")
	}
}

// TestServiceList demonstrates mocking worktree listing
func TestServiceList(t *testing.T) {
	mockGit := &MockGitClient{
		WorktreeListFunc: func() ([]git.GitWorktree, error) {
			return []git.GitWorktree{
				{Path: "/repo", Branch: "main"},
				{Path: "/repo-worktrees/feature", Branch: "feature"},
			}, nil
		},
	}

	service := NewService(mockGit, nil, nil)

	worktrees, err := service.List()
	if err != nil {
		t.Fatalf("List() unexpected error: %v", err)
	}

	if len(worktrees) != 2 {
		t.Errorf("List() returned %d worktrees, want 2", len(worktrees))
	}

	if worktrees[0].Branch != "main" {
		t.Errorf("First worktree branch = %s, want main", worktrees[0].Branch)
	}
}

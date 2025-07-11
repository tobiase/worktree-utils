package git

// GitWorktree represents a git worktree
type GitWorktree struct {
	Path   string
	Branch string
}

// Client defines the interface for git operations
type Client interface {
	// RevParse runs git rev-parse with the given arguments
	RevParse(args ...string) (string, error)

	// WorktreeList returns a list of all worktrees
	WorktreeList() ([]GitWorktree, error)

	// WorktreeAdd creates a new worktree
	WorktreeAdd(path, branch string, options ...string) error

	// WorktreeRemove removes a worktree
	WorktreeRemove(path string) error

	// BranchList returns a list of all branches
	BranchList() ([]string, error)

	// ShowRef checks if a reference exists
	ShowRef(ref string) error

	// GetRemoteURL returns the URL of a remote
	GetRemoteURL(remote string) (string, error)

	// GetCurrentBranch returns the current branch name
	GetCurrentBranch() (string, error)

	// Log returns git log output with specified format and options
	Log(format string, options ...string) (string, error)

	// Status returns git status output
	Status(options ...string) (string, error)

	// RevList returns git rev-list output
	RevList(options ...string) (string, error)

	// ForEachRef returns git for-each-ref output with specified format and options
	ForEachRef(format string, options ...string) (string, error)

	// GetConfigValue returns a git config value for the given key
	GetConfigValue(key string) (string, error)

	// Checkout switches to the specified branch
	Checkout(branch string) error
}

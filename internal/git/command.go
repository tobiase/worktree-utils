package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// CommandClient implements the Git interface using command line git
type CommandClient struct {
	workDir string
}

// NewCommandClient creates a new git client that executes git commands
func NewCommandClient(workDir string) *CommandClient {
	return &CommandClient{workDir: workDir}
}

// RevParse runs git rev-parse with the given arguments
func (c *CommandClient) RevParse(args ...string) (string, error) {
	cmdArgs := append([]string{"rev-parse"}, args...)
	output, err := c.runCommand(cmdArgs...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// WorktreeList returns a list of all worktrees
func (c *CommandClient) WorktreeList() ([]GitWorktree, error) {
	output, err := c.runCommand("worktree", "list", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %v", err)
	}

	var worktrees []GitWorktree
	var currentPath string

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "worktree ") {
			currentPath = strings.TrimPrefix(line, "worktree ")
		} else if strings.HasPrefix(line, "branch refs/heads/") {
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			if currentPath != "" && branch != "" {
				worktrees = append(worktrees, GitWorktree{
					Path:   currentPath,
					Branch: branch,
				})
			}
		} else if currentPath != "" && line == "" {
			// Empty line indicates end of worktree entry
			currentPath = ""
		}
	}

	return worktrees, scanner.Err()
}

// WorktreeAdd creates a new worktree
func (c *CommandClient) WorktreeAdd(path, branch string, options ...string) error {
	args := append([]string{"worktree", "add", path}, options...)
	args = append(args, branch)
	_, err := c.runCommand(args...)
	return err
}

// WorktreeRemove removes a worktree
func (c *CommandClient) WorktreeRemove(path string) error {
	_, err := c.runCommand("worktree", "remove", path)
	return err
}

// BranchList returns a list of all branches
func (c *CommandClient) BranchList() ([]string, error) {
	output, err := c.runCommand("branch", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	branches := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// ShowRef checks if a reference exists
func (c *CommandClient) ShowRef(ref string) error {
	_, err := c.runCommand("show-ref", "--verify", "--quiet", ref)
	return err
}

// GetRemoteURL returns the URL of a remote
func (c *CommandClient) GetRemoteURL(remote string) (string, error) {
	output, err := c.runCommand("remote", "get-url", remote)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch returns the current branch name
func (c *CommandClient) GetCurrentBranch() (string, error) {
	output, err := c.runCommand("branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// Log returns git log output with specified format and options
func (c *CommandClient) Log(format string, options ...string) (string, error) {
	args := append([]string{"log", "--pretty=format:" + format}, options...)
	output, err := c.runCommand(args...)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// Status returns git status output
func (c *CommandClient) Status(options ...string) (string, error) {
	args := append([]string{"status"}, options...)
	output, err := c.runCommand(args...)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// RevList returns git rev-list output
func (c *CommandClient) RevList(options ...string) (string, error) {
	args := append([]string{"rev-list"}, options...)
	output, err := c.runCommand(args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// runCommand executes a git command and returns the output
func (c *CommandClient) runCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	if c.workDir != "" {
		cmd.Dir = c.workDir
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git %s failed: %v\nstderr: %s",
				strings.Join(args, " "), err, string(exitErr.Stderr))
		}
		return nil, err
	}

	return output, nil
}

// ForEachRef returns git for-each-ref output with specified format and options
func (c *CommandClient) ForEachRef(format string, options ...string) (string, error) {
	args := []string{"for-each-ref"}
	if format != "" {
		args = append(args, "--format="+format)
	}
	args = append(args, options...)

	output, err := c.runCommand(args...)
	if err != nil {
		return "", fmt.Errorf("failed to run for-each-ref: %v", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetConfigValue returns a git config value for the given key
func (c *CommandClient) GetConfigValue(key string) (string, error) {
	output, err := c.runCommand("config", "--get", key)
	if err != nil {
		// Git returns exit code 1 when the key doesn't exist
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return "", nil // Return empty string for non-existent keys
		}
		return "", fmt.Errorf("failed to get config value for %s: %v", key, err)
	}
	return strings.TrimSpace(string(output)), nil
}

// Checkout switches to the specified branch
func (c *CommandClient) Checkout(branch string) error {
	_, err := c.runCommand("checkout", branch)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %s: %v", branch, err)
	}
	return nil
}

// GetLastNonMergeCommit returns info about the last non-merge commit on a branch
func (c *CommandClient) GetLastNonMergeCommit(branch string, format string) (string, error) {
	// Use git log to find the last non-merge commit
	// --no-merges excludes merge commits
	// -n 1 gets only the most recent commit
	args := []string{"log", "--no-merges", "-n", "1"}
	if format != "" {
		args = append(args, "--format="+format)
	}
	args = append(args, branch)

	output, err := c.runCommand(args...)
	if err != nil {
		// Branch might not exist or have no non-merge commits
		return "", fmt.Errorf("failed to get last non-merge commit for branch %s: %v", branch, err)
	}
	return strings.TrimSpace(string(output)), nil
}

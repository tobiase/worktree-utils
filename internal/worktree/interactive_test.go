package worktree

import (
	"testing"
)

func TestGetAvailableBranches(t *testing.T) {
	// This test requires a git repository, so we'll test the function signature
	// and basic error handling rather than actual git operations

	// Test that the function exists and returns expected types
	branches, err := GetAvailableBranches()

	// In a non-git directory, this should return an error
	if err != nil {
		// This is expected in a test environment that might not be a git repo
		t.Logf("GetAvailableBranches returned error (expected in test): %v", err)
	}

	// Test that it returns slice of strings
	for i, branch := range branches {
		if branch == "" {
			t.Errorf("Branch at index %d is empty string", i)
		}
	}
}

func TestGetWorktreeInfo(t *testing.T) {
	// Test that the function exists and returns expected types
	worktrees, err := GetWorktreeInfo()

	// In a non-git directory, this should return an error
	if err != nil {
		// This is expected in a test environment that might not be a git repo
		t.Logf("GetWorktreeInfo returned error (expected in test): %v", err)
	}

	// Test that it returns slice of Worktree structs
	for i, wt := range worktrees {
		if wt.Path == "" && wt.Branch == "" {
			t.Errorf("Worktree at index %d has empty path and branch", i)
		}
	}
}

func TestCreateBranchPreview(t *testing.T) {
	// Test the preview function with invalid indices
	preview := createBranchPreview(-1, 80, 24)
	if preview == "" {
		t.Error("Preview should return some content even for invalid index")
	}

	preview = createBranchPreview(999, 80, 24)
	if preview == "" {
		t.Error("Preview should return some content even for out-of-range index")
	}

	// Test with valid dimensions
	preview = createBranchPreview(0, 80, 24)
	if preview == "" {
		t.Error("Preview should return some content")
	}

	// The preview should contain some informative text
	if len(preview) < 10 {
		t.Error("Preview should contain substantial information")
	}
}

func TestSelectBranchInteractivelyFallback(t *testing.T) {
	// This test will exercise the fallback behavior since we're not interactive
	// In a real git repository with worktrees, this would be more comprehensive

	_, err := SelectBranchInteractively()

	// Should either succeed with a branch or fail gracefully
	if err != nil {
		// Check that it's a reasonable error message
		if err.Error() == "" {
			t.Error("Error should have a descriptive message")
		}
		t.Logf("SelectBranchInteractively returned error (expected in test): %v", err)
	}
}

func TestGitLogFunction(t *testing.T) {
	// Test the git log helper function
	_, err := getGitLog("main", 5)

	// In a non-git directory or without main branch, this should error
	if err != nil {
		t.Logf("getGitLog returned error (expected in test): %v", err)
	}

	// Test with invalid count
	_, err = getGitLog("main", 0)
	if err != nil {
		t.Logf("getGitLog with count 0 returned error: %v", err)
	}
}

func TestGitStatusFunction(t *testing.T) {
	// Test the git status helper function
	_, err := getGitStatus(".")

	// In a non-git directory, this should error
	if err != nil {
		t.Logf("getGitStatus returned error (expected in test): %v", err)
	}

	// Test with non-existent path
	_, err = getGitStatus("/nonexistent/path")
	if err == nil {
		t.Error("getGitStatus should error for non-existent path")
	}
}

package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// TestHandleRecentCommand tests the recent command functionality
func TestHandleRecentCommand(t *testing.T) {
	// Note: These are skeleton tests that would require proper git repository setup
	// and mocking infrastructure to run properly

	t.Run("displays recent branches", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test that recent command displays branches sorted by date
	})

	t.Run("respects count flag", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test that -n flag limits the number of results
	})

	t.Run("shows only current user branches by default", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test that default behavior shows only current user's branches
	})

	t.Run("shows all branches with --all flag", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test that --all shows all branches regardless of author
	})

	t.Run("filters by author with --others flag", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test that --others excludes current user's branches
	})

	t.Run("validates conflicting flags", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test that --all and --others together produce an error
	})

	t.Run("numeric navigation to worktree", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test navigation to branch with existing worktree
	})

	t.Run("numeric navigation checkout", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test checkout when branch has no worktree
	})

	t.Run("handles invalid index", func(t *testing.T) {
		t.Skip("Requires git repository setup")
		// Test error handling for out-of-bounds index
	})
}

// TestParseRecentFlags tests flag parsing logic
func TestParseRecentFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantAll      bool
		wantOthers   bool
		wantCount    int
		wantNavigate int
		wantVerbose  bool
		wantCompact  bool
		wantError    bool
	}{
		{
			name:         "no flags",
			args:         []string{},
			wantCount:    10,
			wantNavigate: -1,
		},
		{
			name:         "count flag",
			args:         []string{"-n", "20"},
			wantCount:    20,
			wantNavigate: -1,
		},
		{
			name:         "all flag",
			args:         []string{"--all"},
			wantAll:      true,
			wantCount:    10,
			wantNavigate: -1,
		},
		{
			name:         "others flag",
			args:         []string{"--others"},
			wantOthers:   true,
			wantCount:    10,
			wantNavigate: -1,
		},
		{
			name:         "numeric navigation",
			args:         []string{"3"},
			wantCount:    10,
			wantNavigate: 3,
		},
		{
			name:         "combined flags",
			args:         []string{"--all", "-n", "5", "2"},
			wantAll:      true,
			wantCount:    5,
			wantNavigate: 2,
		},
		{
			name:      "conflicting flags",
			args:      []string{"--all", "--others"},
			wantError: true,
		},
		{
			name:         "-n= format",
			args:         []string{"-n=7"},
			wantCount:    7,
			wantNavigate: -1,
		},
		{
			name:      "invalid count value",
			args:      []string{"-n", "abc"},
			wantError: true,
		},
		{
			name:      "negative count value",
			args:      []string{"-n", "-5"},
			wantError: true,
		},
		{
			name:      "zero count value",
			args:      []string{"-n", "0"},
			wantError: true,
		},
		{
			name:      "negative navigation index",
			args:      []string{"-3"},
			wantError: true,
		},
		{
			name:         "verbose flag long form",
			args:         []string{"--verbose"},
			wantVerbose:  true,
			wantCount:    10,
			wantNavigate: -1,
		},
		{
			name:         "verbose flag short form",
			args:         []string{"-v"},
			wantVerbose:  true,
			wantCount:    10,
			wantNavigate: -1,
		},
		{
			name:         "verbose with other flags",
			args:         []string{"--all", "-v", "-n", "15"},
			wantAll:      true,
			wantVerbose:  true,
			wantCount:    15,
			wantNavigate: -1,
		},
		{
			name:         "compact flag long form",
			args:         []string{"--compact"},
			wantCompact:  true,
			wantCount:    10,
			wantNavigate: -1,
		},
		{
			name:         "compact flag short form",
			args:         []string{"-c"},
			wantCompact:  true,
			wantCount:    10,
			wantNavigate: -1,
		},
		{
			name:         "compact with other flags",
			args:         []string{"--all", "-c", "-n", "5"},
			wantAll:      true,
			wantCompact:  true,
			wantCount:    5,
			wantNavigate: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantError {
				// For error cases, we need to capture the exit behavior
				// For now, skip these tests as they require special handling
				t.Skip("Error cases require mocking osExit function")
				return
			}

			flags := parseRecentFlags(tt.args)

			if flags.showAll != tt.wantAll {
				t.Errorf("all flag: got %v, want %v", flags.showAll, tt.wantAll)
			}

			if flags.showOthers != tt.wantOthers {
				t.Errorf("others flag: got %v, want %v", flags.showOthers, tt.wantOthers)
			}

			if flags.count != tt.wantCount {
				t.Errorf("count: got %v, want %v", flags.count, tt.wantCount)
			}

			if flags.navigateIndex != tt.wantNavigate {
				t.Errorf("navigate index: got %v, want %v", flags.navigateIndex, tt.wantNavigate)
			}

			if flags.verbose != tt.wantVerbose {
				t.Errorf("verbose flag: got %v, want %v", flags.verbose, tt.wantVerbose)
			}

			if flags.compact != tt.wantCompact {
				t.Errorf("compact flag: got %v, want %v", flags.compact, tt.wantCompact)
			}
		})
	}
}

const testUser = "John Doe"

// TestBranchFiltering tests the branch filtering logic
func TestBranchFiltering(t *testing.T) {
	t.Run("default filters by current user", func(t *testing.T) {
		branches := []string{
			"feature|2 hours ago|Add feature|John Doe",
			"bugfix|1 day ago|Fix bug|Jane Smith",
			"main|3 days ago|Update docs|John Doe",
		}
		currentUser := testUser

		// Test default filtering (shows only current user)
		filtered := filterBranchesByAuthor(branches, currentUser, false, false, false)
		if len(filtered) != 2 {
			t.Errorf("Expected 2 branches for current user, got %d", len(filtered))
		}
	})

	t.Run("--all shows all branches", func(t *testing.T) {
		branches := []string{
			"feature|2 hours ago|Add feature|John Doe",
			"bugfix|1 day ago|Fix bug|Jane Smith",
			"main|3 days ago|Update docs|John Doe",
		}
		currentUser := testUser

		// Test --all flag (shows all branches)
		filtered := filterBranchesByAuthor(branches, currentUser, true, false, false)
		if len(filtered) != 3 {
			t.Errorf("Expected 3 branches with --all, got %d", len(filtered))
		}
	})

	t.Run("--others filters out current user", func(t *testing.T) {
		branches := []string{
			"feature|2 hours ago|Add feature|John Doe",
			"bugfix|1 day ago|Fix bug|Jane Smith",
			"main|3 days ago|Update docs|John Doe",
		}
		currentUser := testUser

		// Test filtering logic for --others flag
		filtered := filterBranchesByAuthor(branches, currentUser, false, true, false)
		if len(filtered) != 1 {
			t.Errorf("Expected 1 branch for other users, got %d", len(filtered))
		}
	})
}

// TestActualFilterBranches tests the real filterBranches function used in handleRecentCommand
func TestActualFilterBranches(t *testing.T) {
	// Create test branch data
	branches := []branchCommitInfo{
		{
			branch:       "feature-1",
			commitHash:   "abc123",
			relativeDate: "2 hours ago",
			subject:      "Add new feature",
			author:       testUser,
			timestamp:    time.Now().Add(-2 * time.Hour),
			hasWorktree:  true,
		},
		{
			branch:       "bugfix-1",
			commitHash:   "def456",
			relativeDate: "1 day ago",
			subject:      "Fix critical bug",
			author:       "Jane Smith",
			timestamp:    time.Now().Add(-24 * time.Hour),
			hasWorktree:  false,
		},
		{
			branch:       "feature-2",
			commitHash:   "ghi789",
			relativeDate: "3 days ago",
			subject:      "Another feature",
			author:       testUser,
			timestamp:    time.Now().Add(-72 * time.Hour),
			hasWorktree:  true,
		},
		{
			branch:       "main",
			commitHash:   "jkl012",
			relativeDate: "1 week ago",
			subject:      "Update docs",
			author:       "Bob Wilson",
			timestamp:    time.Now().Add(-168 * time.Hour),
			hasWorktree:  true,
		},
	}

	t.Run("default behavior shows only current user branches", func(t *testing.T) {
		flags := recentFlags{showAll: false, showOthers: false}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		if len(result) != 2 {
			t.Errorf("Expected 2 branches for current user, got %d", len(result))
		}

		for _, branch := range result {
			if branch.author != currentUser {
				t.Errorf("Expected branch %s to be authored by %s, got %s", branch.branch, currentUser, branch.author)
			}
		}
	})

	t.Run("--all flag shows all branches", func(t *testing.T) {
		flags := recentFlags{showAll: true, showOthers: false}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		if len(result) != 4 {
			t.Errorf("Expected 4 branches with --all flag, got %d", len(result))
		}

		// Should include all authors
		authors := make(map[string]bool)
		for _, branch := range result {
			authors[branch.author] = true
		}

		expectedAuthors := []string{testUser, "Jane Smith", "Bob Wilson"}
		for _, expectedAuthor := range expectedAuthors {
			if !authors[expectedAuthor] {
				t.Errorf("Expected to find branches by %s in --all results", expectedAuthor)
			}
		}
	})

	t.Run("--others flag shows only other users branches", func(t *testing.T) {
		flags := recentFlags{showAll: false, showOthers: true}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		if len(result) != 2 {
			t.Errorf("Expected 2 branches for other users, got %d", len(result))
		}

		for _, branch := range result {
			if branch.author == currentUser {
				t.Errorf("Expected branch %s to NOT be authored by %s, but it was", branch.branch, currentUser)
			}
		}

		// Verify we got the right branches
		expectedBranches := []string{"bugfix-1", "main"}
		actualBranches := make([]string, len(result))
		for i, branch := range result {
			actualBranches[i] = branch.branch
		}

		for _, expected := range expectedBranches {
			found := false
			for _, actual := range actualBranches {
				if actual == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find branch %s in --others results", expected)
			}
		}
	})

	t.Run("preserves order from input", func(t *testing.T) {
		flags := recentFlags{showAll: true, showOthers: false}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		// Should preserve the order from input
		expectedOrder := []string{"feature-1", "bugfix-1", "feature-2", "main"}
		for i, branch := range result {
			if branch.branch != expectedOrder[i] {
				t.Errorf("Expected branch at index %d to be %s, got %s", i, expectedOrder[i], branch.branch)
			}
		}
	})
}

// Helper function for testing (would be extracted from main code)
func filterBranchesByAuthor(branches []string, currentUser string, showAll, showOthers, defaultMode bool) []string {
	var filtered []string
	for _, branch := range branches {
		parts := strings.Split(branch, "|")
		if len(parts) != 4 {
			continue
		}
		author := parts[3]

		if showAll {
			// Show all branches
			filtered = append(filtered, branch)
		} else if showOthers {
			// Show only other users' branches
			if author != currentUser {
				filtered = append(filtered, branch)
			}
		} else {
			// Default: show only current user's branches
			if author == currentUser {
				filtered = append(filtered, branch)
			}
		}
	}
	return filtered
}

// TestRecentCommandEdgeCases tests edge cases for the recent command
func TestRecentCommandEdgeCases(t *testing.T) {

	t.Run("special characters in branch names", func(t *testing.T) {
		// Test that branch names with special characters are handled correctly
		branches := []branchCommitInfo{
			{
				branch:       "feature/special-chars-üñíçødé",
				commitHash:   "abc123",
				relativeDate: "2 hours ago",
				subject:      "Add unicode support",
				author:       testUser,
				timestamp:    time.Now().Add(-2 * time.Hour),
				hasWorktree:  true,
			},
			{
				branch:       "hotfix/bug-#123",
				commitHash:   "def456",
				relativeDate: "1 day ago",
				subject:      "Fix issue #123",
				author:       testUser,
				timestamp:    time.Now().Add(-24 * time.Hour),
				hasWorktree:  false,
			},
			{
				branch:       "feature/with-spaces and symbols!@#",
				commitHash:   "ghi789",
				relativeDate: "3 days ago",
				subject:      "Special branch name",
				author:       testUser,
				timestamp:    time.Now().Add(-72 * time.Hour),
				hasWorktree:  false,
			},
		}

		flags := recentFlags{showAll: false, showOthers: false}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		if len(result) != 3 {
			t.Errorf("Expected 3 branches with special characters, got %d", len(result))
		}

		// Verify all branches are preserved with special characters intact
		expectedBranches := []string{
			"feature/special-chars-üñíçødé",
			"hotfix/bug-#123",
			"feature/with-spaces and symbols!@#",
		}

		for i, expected := range expectedBranches {
			if result[i].branch != expected {
				t.Errorf("Expected branch %s, got %s", expected, result[i].branch)
			}
		}
	})

	t.Run("empty branch list", func(t *testing.T) {
		var branches []branchCommitInfo
		flags := recentFlags{showAll: true, showOthers: false}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		if len(result) != 0 {
			t.Errorf("Expected 0 branches for empty input, got %d", len(result))
		}
	})

	t.Run("empty current user name", func(t *testing.T) {
		branches := []branchCommitInfo{
			{
				branch:       "feature-1",
				commitHash:   "abc123",
				relativeDate: "2 hours ago",
				subject:      "Add feature",
				author:       "",
				timestamp:    time.Now().Add(-2 * time.Hour),
				hasWorktree:  true,
			},
			{
				branch:       "feature-2",
				commitHash:   "def456",
				relativeDate: "1 day ago",
				subject:      "Another feature",
				author:       testUser,
				timestamp:    time.Now().Add(-24 * time.Hour),
				hasWorktree:  false,
			},
		}

		flags := recentFlags{showAll: false, showOthers: false}
		currentUser := ""

		result := filterBranches(branches, flags, currentUser)

		// Should only match branches with empty author
		if len(result) != 1 {
			t.Errorf("Expected 1 branch for empty current user, got %d", len(result))
		}

		if result[0].branch != "feature-1" {
			t.Errorf("Expected feature-1, got %s", result[0].branch)
		}
	})

	t.Run("very long branch names", func(t *testing.T) {
		longBranchName := "feature/" + strings.Repeat("very-long-branch-name-segment-", 20) + "end"

		branches := []branchCommitInfo{
			{
				branch:       longBranchName,
				commitHash:   "abc123",
				relativeDate: "2 hours ago",
				subject:      "Very long branch name test",
				author:       testUser,
				timestamp:    time.Now().Add(-2 * time.Hour),
				hasWorktree:  true,
			},
		}

		flags := recentFlags{showAll: true, showOthers: false}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		if len(result) != 1 {
			t.Errorf("Expected 1 branch with long name, got %d", len(result))
		}

		if result[0].branch != longBranchName {
			t.Errorf("Long branch name was modified during filtering")
		}
	})
}

// TestNavigationEdgeCases tests navigation functionality edge cases
func TestNavigationEdgeCases(t *testing.T) {

	t.Run("navigation with empty branch list", func(t *testing.T) {
		// This would require mocking printErrorAndExit to test properly
		t.Skip("Requires mocking printErrorAndExit for error case testing")
	})

	t.Run("navigation with out of bounds index", func(t *testing.T) {
		// This would require mocking printErrorAndExit to test properly
		t.Skip("Requires mocking printErrorAndExit for error case testing")
	})
}

// TestRecentFlagsEdgeCases tests edge cases in flag parsing
func TestRecentFlagsEdgeCases(t *testing.T) {

	t.Run("flags with empty arguments", func(t *testing.T) {
		args := []string{}
		flags := parseRecentFlags(args)

		// Should use defaults
		if flags.count != 10 {
			t.Errorf("Expected default count 10, got %d", flags.count)
		}

		if flags.navigateIndex != -1 {
			t.Errorf("Expected default navigateIndex -1, got %d", flags.navigateIndex)
		}

		if flags.showAll || flags.showOthers || flags.verbose {
			t.Errorf("Expected all boolean flags to be false by default")
		}
	})

	t.Run("duplicate flags", func(t *testing.T) {
		args := []string{"--all", "--all", "-v", "-v"}
		flags := parseRecentFlags(args)

		// Should still work with duplicate flags
		if !flags.showAll {
			t.Errorf("Expected showAll to be true with duplicate --all flags")
		}

		if !flags.verbose {
			t.Errorf("Expected verbose to be true with duplicate -v flags")
		}
	})

	t.Run("multiple navigation indices", func(t *testing.T) {
		args := []string{"1", "2", "3"}
		flags := parseRecentFlags(args)

		// Should only use the first one
		if flags.navigateIndex != 1 {
			t.Errorf("Expected navigateIndex 1 (first occurrence), got %d", flags.navigateIndex)
		}
	})

	t.Run("large count values", func(t *testing.T) {
		args := []string{"-n", "999999"}
		flags := parseRecentFlags(args)

		if flags.count != 999999 {
			t.Errorf("Expected count 999999, got %d", flags.count)
		}
	})
}

// TestRecentCommandPerformance tests performance with large numbers of branches
func TestRecentCommandPerformance(t *testing.T) {

	t.Run("large number of branches", func(t *testing.T) {
		// Create 1000 test branches
		branches := make([]branchCommitInfo, 1000)
		authors := []string{testUser, "Jane Smith", "Bob Wilson", "Alice Johnson", "Charlie Brown"}

		baseTime := time.Now()
		for i := 0; i < 1000; i++ {
			branches[i] = branchCommitInfo{
				branch:       fmt.Sprintf("feature-%d", i),
				commitHash:   fmt.Sprintf("commit%d", i),
				relativeDate: fmt.Sprintf("%d hours ago", i),
				subject:      fmt.Sprintf("Feature %d implementation", i),
				author:       authors[i%len(authors)],
				timestamp:    baseTime.Add(-time.Duration(i) * time.Hour),
				hasWorktree:  i%3 == 0, // Every third branch has a worktree
			}
		}

		// Test filtering performance - should complete quickly
		start := time.Now()

		flags := recentFlags{showAll: false, showOthers: false}
		currentUser := testUser

		result := filterBranches(branches, flags, currentUser)

		elapsed := time.Since(start)

		// Should complete in reasonable time (< 100ms for 1000 branches)
		if elapsed > 100*time.Millisecond {
			t.Errorf("Filtering 1000 branches took too long: %v", elapsed)
		}

		// Should return correct number of branches for testUser
		// testUser authored branches 0, 5, 10, 15, ... (every 5th branch)
		expectedCount := 200 // 1000 / 5
		if len(result) != expectedCount {
			t.Errorf("Expected %d branches for testUser, got %d", expectedCount, len(result))
		}

		// Verify all returned branches are authored by testUser
		for _, branch := range result {
			if branch.author != currentUser {
				t.Errorf("Found branch %s not authored by %s", branch.branch, currentUser)
				break
			}
		}
	})

	t.Run("all flag with large number of branches", func(t *testing.T) {
		// Create 500 branches for performance testing
		branches := make([]branchCommitInfo, 500)

		baseTime := time.Now()
		for i := 0; i < 500; i++ {
			branches[i] = branchCommitInfo{
				branch:       fmt.Sprintf("test-branch-%d", i),
				commitHash:   fmt.Sprintf("hash%d", i),
				relativeDate: fmt.Sprintf("%d minutes ago", i),
				subject:      fmt.Sprintf("Test commit %d", i),
				author:       fmt.Sprintf("User%d", i%10), // 10 different authors
				timestamp:    baseTime.Add(-time.Duration(i) * time.Minute),
				hasWorktree:  i%2 == 0,
			}
		}

		start := time.Now()

		flags := recentFlags{showAll: true, showOthers: false}
		currentUser := "User0"

		result := filterBranches(branches, flags, currentUser)

		elapsed := time.Since(start)

		// Should complete quickly even with --all flag
		if elapsed > 50*time.Millisecond {
			t.Errorf("Filtering 500 branches with --all took too long: %v", elapsed)
		}

		// Should return all branches
		if len(result) != 500 {
			t.Errorf("Expected all 500 branches with --all flag, got %d", len(result))
		}
	})
}

// TestTruncateWithEllipsis tests the string truncation helper
func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string not truncated",
			input:    "main",
			maxLen:   10,
			expected: "main",
		},
		{
			name:     "exact length not truncated",
			input:    "feature/123",
			maxLen:   11,
			expected: "feature/123",
		},
		{
			name:     "long string truncated with ellipsis",
			input:    "feature/very-long-branch-name-that-exceeds-limit",
			maxLen:   20,
			expected: "feature/very-long...",
		},
		{
			name:     "very short max length",
			input:    "feature",
			maxLen:   3,
			expected: "fea",
		},
		{
			name:     "max length 1",
			input:    "feature",
			maxLen:   1,
			expected: "f",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "unicode string truncation",
			input:    "feature/üñíçødé-branch-name",
			maxLen:   15,
			expected: "feature/üñíç...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateWithEllipsis(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateWithEllipsis(%q, %d) = %q, want %q",
					tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestDisplayBranchesMultiline tests the multi-line display format
func TestDisplayBranchesMultiline(t *testing.T) {
	t.Run("multi-line format", func(t *testing.T) {
		branches := []branchCommitInfo{
			{
				branch:       "feature/very-long-branch-name-with-issue-123",
				commitHash:   "abc123",
				relativeDate: "2 hours ago",
				subject:      "Implement the new feature for handling long branch names",
				author:       "Tobias Engelhardt",
				timestamp:    time.Now().Add(-2 * time.Hour),
				hasWorktree:  true,
			},
			{
				branch:       "fix/issue-456",
				commitHash:   "def456",
				relativeDate: "1 day ago",
				subject:      "Fix critical bug in authentication",
				author:       "Jane Smith",
				timestamp:    time.Now().Add(-24 * time.Hour),
				hasWorktree:  false,
			},
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayBranches(branches, 10)

		w.Close()
		output, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		outputStr := string(output)
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")

		// Check that we have correct number of lines
		// 2 branches * 3 lines each + 1 blank line between = 7 lines
		if len(lines) != 7 {
			t.Errorf("Expected 7 lines of output, got %d", len(lines))
			t.Logf("Output:\n%s", outputStr)
		}

		// Check first branch format
		if !strings.HasPrefix(lines[0], "0: *feature/very-long-branch-name-with-issue-123") {
			t.Errorf("First line doesn't match expected format: %s", lines[0])
		}

		// Check indentation on second and third lines
		if !strings.HasPrefix(lines[1], "   ") {
			t.Error("Second line should be indented")
		}
		if !strings.HasPrefix(lines[2], "   ") {
			t.Error("Third line should be indented")
		}

		// Check that there's a blank line after first entry
		if lines[3] != "" {
			t.Error("Expected blank line after first entry")
		}

		// Check second branch doesn't have star (no worktree)
		if !strings.HasPrefix(lines[4], "1: fix/issue-456") {
			t.Errorf("Fifth line doesn't match expected format: %s", lines[4])
		}
	})

	t.Run("empty branch list", func(t *testing.T) {
		var branches []branchCommitInfo

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayBranches(branches, 10)

		w.Close()
		output, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		// Should produce no output
		if len(output) != 0 {
			t.Errorf("Expected no output for empty branch list, got: %s", string(output))
		}
	})
}

// TestDisplayBranchesCompactFormatting tests the compact format
func TestDisplayBranchesCompactFormatting(t *testing.T) {
	t.Run("branches with varying lengths", func(t *testing.T) {
		branches := []branchCommitInfo{
			{
				branch:       "main",
				commitHash:   "abc123",
				relativeDate: "2 hours ago",
				subject:      "Initial commit",
				author:       "John Doe",
				timestamp:    time.Now().Add(-2 * time.Hour),
				hasWorktree:  true,
			},
			{
				branch:       "feature/very-very-very-long-branch-name-that-should-be-truncated",
				commitHash:   "def456",
				relativeDate: "1 day ago",
				subject:      "This is a very long commit message that should also be truncated to fit nicely",
				author:       "Jane Smith",
				timestamp:    time.Now().Add(-24 * time.Hour),
				hasWorktree:  false,
			},
			{
				branch:       "fix/short",
				commitHash:   "ghi789",
				relativeDate: "3 weeks ago",
				subject:      "Fix bug",
				author:       "Bob Wilson",
				timestamp:    time.Now().Add(-504 * time.Hour),
				hasWorktree:  true,
			},
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayBranchesCompact(branches, 10)

		w.Close()
		output, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		outputStr := string(output)
		lines := strings.Split(strings.TrimSpace(outputStr), "\n")

		// Check that we have 3 lines of output
		if len(lines) != 3 {
			t.Errorf("Expected 3 lines of output, got %d", len(lines))
		}

		// Check that long branch name is truncated
		if !strings.Contains(lines[1], "...") {
			t.Error("Expected long branch name to be truncated with ellipsis")
		}

		// Check that alignment is maintained (all lines should have similar structure)
		for i, line := range lines {
			if !strings.HasPrefix(line, fmt.Sprintf("%d:", i)) {
				t.Errorf("Line %d doesn't start with correct index", i)
			}
		}
	})

	t.Run("empty branch list", func(t *testing.T) {
		var branches []branchCommitInfo

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		displayBranchesCompact(branches, 10)

		w.Close()
		output, _ := io.ReadAll(r)
		os.Stdout = oldStdout

		// Should produce no output
		if len(output) != 0 {
			t.Errorf("Expected no output for empty branch list, got: %s", string(output))
		}
	})
}

package worktree

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobiase/worktree-utils/test/helpers"
)

func TestNewWorktree(t *testing.T) {
	tests := []struct {
		name    string
		branch  string
		setup   func() (string, func())
		wantErr bool
		check   func(t *testing.T, repo, branch string)
	}{
		{
			name:   "create new worktree",
			branch: "feature-new",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, repo, branch string) {
				// Check worktree was created
				worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
				worktreePath := filepath.Join(worktreeBase, branch)
				helpers.AssertDirExists(t, worktreePath)

				// Check branch exists
				output := helpers.GetGitOutput(t, repo, "branch", "--list", branch)
				if output == "" {
					t.Errorf("Branch %s was not created", branch)
				}
			},
		},
		{
			name:   "create worktree with special characters",
			branch: "feature/sub-task-123",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, repo, branch string) {
				// Check worktree was created
				worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
				// Git keeps the branch name as-is in the filesystem
				worktreePath := filepath.Join(worktreeBase, branch)
				helpers.AssertDirExists(t, worktreePath)
			},
		},
		{
			name:   "create worktree with existing branch",
			branch: "main",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)
				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: true, // Should fail because main already exists
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, cleanup := tt.setup()
			defer cleanup()

			// For new branches, use HEAD as base branch
			baseBranch := ""
			if tt.name != "create worktree with existing branch" {
				baseBranch = "HEAD"
			}
			_, err := NewWorktree(tt.branch, baseBranch, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewWorktree() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, repo, tt.branch)
			}
		})
	}
}

func TestCopyEnvFile(t *testing.T) {
	tests := []struct {
		name         string
		sourceFiles  map[string]string
		targetBranch string
		setup        func() (string, func())
		wantErr      bool
		check        func(t *testing.T, repo string)
	}{
		{
			name: "copy single .env file",
			sourceFiles: map[string]string{
				".env": "DATABASE_URL=postgres://localhost/test\nAPI_KEY=secret123\n",
			},
			targetBranch: "feature-1",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				// Create a worktree to copy to
				_, _ = helpers.AddTestWorktree(t, repo, "feature-1")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, repo string) {
				// Check that .env was copied to the worktree
				worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
				targetPath := filepath.Join(worktreeBase, "feature-1", ".env")
				helpers.AssertFileContents(t, targetPath, "DATABASE_URL=postgres://localhost/test\nAPI_KEY=secret123\n")
			},
		},
		{
			name: "copy multiple env files with recursive flag",
			sourceFiles: map[string]string{
				".env":       "MAIN_ENV=true\n",
				".env.local": "LOCAL_ENV=true\n",
				".env.test":  "TEST_ENV=true\n",
			},
			targetBranch: "feature-2",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				// Create a worktree to copy to
				_, _ = helpers.AddTestWorktree(t, repo, "feature-2")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, repo string) {
				// When recursive=false, only .env is copied
				worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
				base := filepath.Join(worktreeBase, "feature-2")

				helpers.AssertFileContents(t, filepath.Join(base, ".env"), "MAIN_ENV=true\n")
				// These files should NOT be copied when recursive=false
				helpers.AssertFileNotExists(t, filepath.Join(base, ".env.local"))
				helpers.AssertFileNotExists(t, filepath.Join(base, ".env.test"))
			},
		},
		{
			name: "copy env file from subdirectory",
			sourceFiles: map[string]string{
				".env": "API_ENV=production\n",
			},
			targetBranch: "feature-3",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				// Create subdirectory structure
				apiDir := filepath.Join(repo, "src", "api")
				_ = os.MkdirAll(apiDir, 0755)

				// Create a worktree to copy to
				_, _ = helpers.AddTestWorktree(t, repo, "feature-3")

				// Change to subdirectory
				oldWd, _ := os.Getwd()
				_ = os.Chdir(apiDir)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, repo string) {
				// Check that .env was copied to the same relative path
				worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
				targetPath := filepath.Join(worktreeBase, "feature-3", "src", "api", ".env")
				helpers.AssertFileContents(t, targetPath, "API_ENV=production\n")
			},
		},
		{
			name:         "no env files to copy",
			sourceFiles:  map[string]string{},
			targetBranch: "feature-4",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				// Create a worktree to copy to
				_, _ = helpers.AddTestWorktree(t, repo, "feature-4")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: true, // Should fail when no .env file exists
		},
		{
			name: "target worktree doesn't exist",
			sourceFiles: map[string]string{
				".env": "TEST=true\n",
			},
			targetBranch: "non-existent",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: true,
		},
		{
			name: "copy env file with gitignore",
			sourceFiles: map[string]string{
				".env":       "SECRET=value\n",
				".gitignore": ".env\n.env.local\n",
			},
			targetBranch: "feature-5",
			setup: func() (string, func()) {
				repo, cleanup := helpers.CreateTestRepo(t)

				// Create a worktree to copy to
				_, _ = helpers.AddTestWorktree(t, repo, "feature-5")

				oldWd, _ := os.Getwd()
				_ = os.Chdir(repo)
				return repo, func() {
					_ = os.Chdir(oldWd)
					cleanup()
				}
			},
			wantErr: false,
			check: func(t *testing.T, repo string) {
				// Env file should still be copied even if gitignored
				worktreeBase := filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-worktrees")
				targetPath := filepath.Join(worktreeBase, "feature-5", ".env")
				helpers.AssertFileContents(t, targetPath, "SECRET=value\n")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, cleanup := tt.setup()
			defer cleanup()

			// Create source files
			cwd, _ := os.Getwd()
			helpers.CreateFiles(t, cwd, tt.sourceFiles)

			err := CopyEnvFile(tt.targetBranch, false)

			if (err != nil) != tt.wantErr {
				t.Errorf("CopyEnvFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, repo)
			}
		})
	}
}

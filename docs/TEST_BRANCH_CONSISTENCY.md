# Test Branch Consistency Guidelines

## Problem

Git's default branch name varies depending on:
- Git version (older versions default to `master`, newer to `main`)
- Global git configuration (`init.defaultBranch` setting)
- CI/CD environment configuration

This causes test failures when tests expect `main` but git creates `master` or vice versa.

## Solution

### For New Test Repositories

Always use one of these approaches when creating test repositories:

1. **Use the centralized helper** (preferred):
   ```go
   import "github.com/tobiase/worktree-utils/test/helpers"

   repoPath, cleanup := helpers.CreateTestRepo(t)
   defer cleanup()
   ```

2. **Use InitTestRepoWithMain for custom scenarios**:
   ```go
   helpers.InitTestRepoWithMain(t, tempDir)
   ```

3. **For git command tests**, use the local helper:
   ```go
   helper := NewTestHelper(t)
   defer helper.Cleanup()
   ```

### Implementation Details

The helpers ensure `main` is the default branch by:
1. Using `git init --initial-branch=main` for Git 2.28+
2. Falling back to `git init` + `git config init.defaultBranch main` for older versions
3. Renaming the branch to `main` after the first commit if needed

### What NOT to Do

Never use plain `git init` in tests:
```go
// BAD - branch name is unpredictable
cmd := exec.Command("git", "init")

// GOOD - branch name is always main
helpers.InitTestRepoWithMain(t, dir)
```

### CI/CD Considerations

GitHub Actions and other CI environments may have different default configurations. Our approach handles this by explicitly setting the branch name rather than relying on defaults.

## Testing Checklist

When writing tests that use git:
- [ ] Use centralized helpers from `test/helpers` package
- [ ] Never assume the default branch name
- [ ] Test with both `main` and `master` if writing git-agnostic code
- [ ] Run tests locally AND in CI to catch environment differences

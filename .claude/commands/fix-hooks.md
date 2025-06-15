# Fix Pre-commit Hooks

Handle pre-commit hook modifications and conflicts:

1. **Check Status**: Run `git status` to see if pre-commit hooks modified files
2. **Stage Fixes**: Add the files that were automatically fixed by hooks
3. **Amend Commit**: Use `git commit --amend --no-edit` to include hook fixes in the previous commit
4. **Verify Clean State**: Ensure `git status` shows a clean working directory

This resolves the common issue where pre-commit hooks (like trailing whitespace fixes) create unstaged changes after a commit, requiring the stage → amend workflow to maintain clean commit history.

Handle this pattern automatically and ensure the final commit includes all hook modifications.

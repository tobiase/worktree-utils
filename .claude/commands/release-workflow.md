# Release Workflow

Execute the complete release workflow for worktree-utils:

1. **Push Changes**: Push current commits to main branch
2. **Monitor CI**: Get the latest GitHub Actions run ID and watch it with `gh run watch <id>`
3. **Version Decision**: Analyze recent commits to determine appropriate version bump (patch/minor/major) based on semantic versioning
4. **Create Release**:
   - Create version tag (e.g., `git tag v0.x.x`)
   - Push tag to trigger release workflow
   - Monitor release workflow completion
5. **Verify Release**: Check that release was created successfully with `gh release view <version>`

Follow the project's semantic versioning guidelines:
- PATCH: Bug fixes, docs, tests
- MINOR: New features, new commands
- MAJOR: Breaking changes to CLI interface

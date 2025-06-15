# Test and CI Workflow

Run comprehensive testing and fix any issues for worktree-utils:

1. **Run CI Tests**: Execute `make test-ci` to run the same tests as GitHub Actions
2. **Fix Test Failures**: Address any failing tests, considering edge cases and error scenarios
3. **Run Linting**: Execute `make lint` to check code quality
4. **Fix Linting Issues**: Resolve any linting warnings or errors
5. **Verify Build**: Ensure `make build` succeeds
6. **Check Coverage**: Review test coverage and add tests for new functionality if needed

Focus on the project's testing philosophy: "Edge Cases Over Coverage" - ensure robust handling of real-world scenarios rather than just achieving percentage thresholds.

Address any locale issues, Git repository edge cases, or system integration problems that arise during testing.

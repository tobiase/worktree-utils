# Contributing

Thank you for your interest! However, this project is not accepting external contributions at this time.

This repository is public to facilitate script access without authentication, but it's actively maintained as a personal utility.

Feel free to fork for your own use!

## Internal Development Reference

### Development Setup

#### Prerequisites

- Go 1.24+ installed
- Python 3.6+ (for pre-commit)
- golangci-lint installed (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)

#### Setting Up Pre-commit Hooks

```bash
# Install pre-commit hooks
make setup-hooks

# Run pre-commit on all files
make lint-all
```

#### Pre-commit Checks

The following checks run automatically on commit:

1. **File formatting**: Trailing whitespace, end of file, line endings
2. **Go specific**: gofmt, golangci-lint, go mod tidy, go build
3. **Commit messages**: Conventional commits format

#### Testing

```bash
# Run tests
make test

# Run tests with coverage
make test-coverage
```

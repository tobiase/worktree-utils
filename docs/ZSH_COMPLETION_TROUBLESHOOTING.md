# Zsh Completion Troubleshooting Guide

## Our Working Solution: Process Substitution

We use the same approach as kubectl, helm, and terraform:

```bash
# What gets added to ~/.zshrc:
source <(wt-bin completion zsh)
```

**Why this works**: No files, no fpath manipulation, works with all zsh configurations.

## Fresh Shell Testing Tools

To avoid shell corruption during development and testing, use these Make commands:

```bash
# Quick completion test in fresh shell
make test-completion

# Interactive shell with completion loaded (for manual TAB testing)
make test-completion-interactive

# Debug completion script generation
make debug-completion

# Test entire setup process in clean environment
make test-setup

# Completely fresh shell environment
make test-fresh

# Standalone test script
./scripts/test-completion.sh
```

**Why these tools are essential**: During development, the current shell gets corrupted with debugging attempts, causing false negatives. These tools ensure testing happens in clean environments.

## Useful Links and References

- [Zsh Completion System Documentation](http://zsh.sourceforge.net/Doc/Release/Completion-System.html)
- [Kubectl Completion Setup](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#enable-shell-autocompletion)
- [GitHub CLI Completion](https://cli.github.com/manual/gh_completion)
- [Docker Completion Guide](https://docs.docker.com/engine/cli-completion/)
- [Oh-My-Zsh Completion Plugin](https://github.com/zsh-users/zsh-completions)
- [Zsh fpath and autoload explained](https://github.com/zsh-users/zsh/blob/master/Etc/completion-style-guide)

## Troubleshooting

### If completion doesn't work after setup:

```bash
# 1. Check if the line was added to your shell config
grep "wt-bin completion" ~/.zshrc

# 2. Test manually in current shell
source <(wt-bin completion zsh)
type _wt    # Should show: "_wt is a shell function"

# 3. Clear completion cache if needed
rm ~/.zcompdump* && exec zsh

# 4. Use testing tools for clean verification
make test-completion
```

### Common Issues

- **"wt-bin: command not found"**: Binary not in PATH, run `which wt-bin`
- **"_comps: assignment to invalid subscript range"**: Completion cache corruption, clear with `rm ~/.zcompdump*`
- **Completion works manually but not automatically**: Shell config line missing or not being sourced

## Development Notes

- **Never test in current shell**: Use `make test-completion` for clean testing
- **Process substitution always works**: Unlike file-based approaches, this has no initialization order issues
- **Cache clearing**: `rm ~/.zcompdump*` solves most completion cache problems

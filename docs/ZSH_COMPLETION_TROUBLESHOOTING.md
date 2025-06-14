# Zsh Completion Troubleshooting Guide

This document contains the exact patterns and commands needed to debug and fix zsh completion for wt.

## Working Pattern (TESTED AND VERIFIED)

### 1. File Structure Requirements
```
~/.config/wt/completions/_wt    # ← Underscore prefix is REQUIRED
```

### 2. File Content Format
```bash
#compdef wt
# Zsh completion for wt (worktree-utils)
# Generated automatically - do not edit manually

_wt() {
    local context state line
    typeset -A opt_args

    _arguments -C \
        '1: :_wt_commands' \
        '*:: :->args'
    # ... completion logic here
}

# IMPORTANT: NO calls to compdef or _wt at the end!
```

### 3. Loading Sequence (Manual Test)
```bash
# Test this exact sequence in a fresh zsh shell:
fpath=(~/.config/wt/completions $fpath)
autoload -Uz compinit && compinit
autoload -Uz _wt

# Verify it worked:
type _wt
# Should output: "_wt is an autoload shell function from /Users/.../completions/_wt"
```

### 4. Init Script Pattern
```bash
elif [[ -n "$ZSH_VERSION" ]]; then
  # Add wt completion directory to fpath
  fpath=(~/.config/wt/completions $fpath)

  # Ensure completion system is initialized
  if ! command -v compinit >/dev/null 2>&1; then
    autoload -Uz compinit
    compinit
  fi

  # Load completion if available
  if [[ -f ~/.config/wt/completions/_wt ]]; then
    autoload -Uz _wt
  fi
fi
```

## Debugging Steps

### Step 1: Verify File Exists
```bash
ls -la ~/.config/wt/completions/_wt
# Should show: -rw-r--r--  1 user  staff  3557 date _wt
```

### Step 2: Test Manual Loading
```bash
# Run in fresh zsh:
fpath=(~/.config/wt/completions $fpath)
autoload -Uz compinit && compinit
autoload -Uz _wt
type _wt
```

### Step 3: Check fpath
```bash
echo $fpath | grep wt
# Should show: /Users/username/.config/wt/completions
```

### Step 4: Verify Completion System
```bash
# Check if compinit is available:
command -v compinit
# Should show: compinit

# Check if completion functions work:
type _describe
# Should show: _describe is a shell function
```

## Common Issues and Fixes

### Issue: "_wt not found" after autoload
**Cause**: File name wrong or not in fpath
**Fix**: Ensure file is named `_wt` (with underscore) in `completions/` directory

### Issue: "_arguments:comparguments:327" error
**Cause**: Completion file has `compdef _wt wt` or `_wt "$@"` at end
**Fix**: Remove those lines from completion file

### Issue: "command not found: _describe"
**Cause**: Completion system not initialized
**Fix**: Run `autoload -Uz compinit && compinit` first

### Issue: Completion loads manually but not automatically
**Cause**: Init script not being sourced or has logic errors
**Fix**: Check that `~/.zshrc` has: `[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh`

## Testing Commands

### Quick Manual Test
```bash
# In fresh shell, run these commands in sequence:
source ~/.config/wt/init.sh
type _wt                           # Should work
echo $fpath | grep wt             # Should show completion dir
```

### Full End-to-End Test
```bash
# Start completely fresh shell:
exec zsh

# Check automatic loading:
type _wt                          # Should work without manual steps
wt <TAB>                         # Should show completions
```

## Files Modified in This Session
- `internal/completion/zsh.go` - Fixed generation to not include problematic calls
- `internal/setup/setup.go` - Updated to create `completions/_wt` instead of `completion.zsh`
- `cmd/wt/main.go` - Added `--help` flag support
- Setup tests - Updated expectations for new file structure

## Status
- ✅ Manual loading works perfectly
- ❓ Automatic loading in fresh shells needs verification
- ✅ File structure follows zsh conventions
- ✅ Errors eliminated from completion system

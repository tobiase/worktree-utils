# Help System Integration

Add comprehensive help system integration to new commands in worktree-utils:

1. **Add Help Flag Check**: Add `if help.HasHelpFlag(args, "commandName") { return }` at the start of command handlers
2. **Create Help Content**: Add entry to `commandHelpMap` in `internal/help/help.go` including:
   - NAME: Command and brief description
   - USAGE: Syntax with optional/required parameters
   - ALIASES: Alternative command names (if any)
   - OPTIONS: Detailed flag descriptions with examples
   - EXAMPLES: Real-world usage scenarios
   - SEE ALSO: Related commands
3. **Update Constants**: Use `helpFlag` and `helpFlagShort` constants instead of hardcoded strings
4. **Test Help**: Verify `--help` and `-h` flags work correctly

Follow the established help content format and ensure all new commands support universal help integration as documented in CLAUDE.md.

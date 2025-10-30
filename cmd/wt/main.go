package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tobiase/worktree-utils/internal/completion"
	"github.com/tobiase/worktree-utils/internal/config"
	"github.com/tobiase/worktree-utils/internal/git"
	"github.com/tobiase/worktree-utils/internal/help"
	"github.com/tobiase/worktree-utils/internal/interactive"
	"github.com/tobiase/worktree-utils/internal/setup"
	"github.com/tobiase/worktree-utils/internal/update"
	"github.com/tobiase/worktree-utils/internal/worktree"
)

// Version information - set by build flags
var (
	version = "dev"
	date    = "unknown"
)

// osExit is a variable for testing - allows overriding os.Exit
var osExit = os.Exit

// Command flags
const (
	fuzzyFlag      = "--fuzzy"
	fuzzyFlagShort = "-f"
	helpFlag       = "--help"
	helpFlagShort  = "-h"
	forceFlag      = "--force"
)

// Command constants
const (
	shellInitCmd = "shell-init"
	listCmd      = "list"
	helpCmd      = "help"
)

const shellWrapper = `# Shell function to handle CD: and EXEC: prefixes
wt() {
  # Commands that need interactive terminal access (no output capture)
  if [ $# -eq 0 ] || [[ "$*" == *"--fuzzy"* ]] || [[ "$*" == *"-f"* ]]; then
    # Run interactively, then get CD path separately
    "${WT_BIN:-wt-bin}" "$@"
    exit_code=$?

    # If successful and it's a 'go' command, try to get the CD path
    if [ $exit_code -eq 0 ] && [[ "$1" == "go" || $# -eq 0 ]]; then
      # Use a separate call to get just the CD path without interaction
      cd_result=$("${WT_BIN:-wt-bin}" go "$2" 2>/dev/null)
      if [[ "$cd_result" == "CD:"* ]]; then
        cd "${cd_result#CD:}"
      fi
    fi
    return $exit_code
  fi

  # Non-interactive commands use output capture
  output=$("${WT_BIN:-wt-bin}" "$@" 2>&1)
  exit_code=$?

  if [ $exit_code -eq 0 ]; then
    # Check for CD: or EXEC: commands in the output
    cd_path=""
    exec_cmd=""
    while IFS= read -r line; do
      if [[ "$line" == "CD:"* ]]; then
        cd_path="${line#CD:}"
      elif [[ "$line" == "EXEC:"* ]]; then
        exec_cmd="${line#EXEC:}"
      else
        # Print non-command lines (including empty lines)
        echo "$line"
      fi
    done <<< "$output"

    # Execute CD or EXEC commands after printing other output
    if [ -n "$cd_path" ]; then
      cd "$cd_path"
    elif [ -n "$exec_cmd" ]; then
      # Security note: EXEC commands are only used for virtualenv activation
      # and paths are quoted by the Go binary to prevent injection
      eval "$exec_cmd"
    fi
  else
    echo "$output" >&2
    return $exit_code
  fi
}
`

func main() {
	if len(os.Args) < 2 {
		// No command specified - show usage
		showUsage()
		osExit(1)
	}

	// Handle help flags
	if len(os.Args) == 2 && (os.Args[1] == helpFlag || os.Args[1] == helpFlagShort) {
		showUsage()
		return
	}

	cmd := resolveCommandAlias(os.Args[1])
	args := os.Args[2:]

	// Special handling for numeric commands (direct index access: wt 0, wt 1, etc.)
	if isNumericCommand(cmd) {
		handleGoCommand([]string{cmd})
		return
	}

	// Special handling for setup command (doesn't need config)
	if cmd == "setup" {
		handleSetupCommand(args)
		return
	}

	// Initialize and run command
	configMgr := initializeConfig()
	loadProjectConfig(configMgr, cmd)
	runCommand(cmd, args, configMgr)
}

func resolveCommandAlias(cmd string) string {
	aliases := map[string]string{
		"ls":     "list",
		"switch": "go",
		"s":      "go",
	}
	if alias, ok := aliases[cmd]; ok {
		return alias
	}
	return cmd
}

// printErrorAndExit prints an error message and exits with status 1
// NOTE: This function is used throughout the codebase for consistent error handling.
// Future refactoring should move towards returning errors from command handlers
// and handling exits centrally in main(), but this pattern is maintained for
// backward compatibility and to minimize changes.
func printErrorAndExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "wt: "+format+"\n", args...)
	osExit(1)
}

// selectBranchInteractively handles interactive branch selection with consistent error handling
func selectBranchInteractively(useFuzzy bool, usageMsg string) string {
	branches, branchErr := worktree.GetAvailableBranches()
	if branchErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", usageMsg)
		osExit(1)
	}

	if interactive.ShouldUseFuzzy(len(branches), useFuzzy) {
		selectedBranch, selectErr := worktree.SelectBranchInteractively()
		if selectErr != nil {
			if selectErr == interactive.ErrUserCancelled {
				fmt.Fprintf(os.Stderr, "wt: selection cancelled\n")
				osExit(1)
			}
			fmt.Fprintf(os.Stderr, "wt: %v\n", selectErr)
			osExit(1)
		}
		return selectedBranch
	}

	fmt.Fprintf(os.Stderr, "%s\n", usageMsg)
	osExit(1)
	return "" // unreachable
}

func initializeConfig() *config.Manager {
	configMgr, err := config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: failed to initialize config: %v\n", err)
		osExit(1)
	}
	return configMgr
}

func loadProjectConfig(configMgr *config.Manager, cmd string) {
	if cmd != shellInitCmd {
		cwd, _ := os.Getwd()
		gitRemote, _ := worktree.GetGitRemote()
		_ = configMgr.LoadProject(cwd, gitRemote)
	}
}

func runCommand(cmd string, args []string, configMgr *config.Manager) {
	switch cmd {
	case shellInitCmd:
		fmt.Print(shellWrapper)
	case listCmd:
		handleListCommand(args)
	case "recent":
		handleRecentCommand(args)
	case "rm":
		handleRemoveCommand(args)
	case "integrate":
		handleIntegrateCommand(args)
	case "go":
		handleGoCommand(args)
	case "new":
		handleNewCommand(args, configMgr)
	case "env-copy":
		handleEnvCopyCommand(args)
	case "env":
		handleEnvCommand(args)
	case "project":
		handleProjectCommand(args, configMgr)
	case "completion":
		handleCompletionCommand(args, configMgr)
	case "version":
		handleVersionCommand(args)
	case "update":
		handleUpdateCommand(args)
	case helpCmd:
		showUsage()
	default:
		handleCustomCommand(cmd, configMgr)
	}
}

// isNumericCommand checks if the command is a numeric index for direct access
func isNumericCommand(cmd string) bool {
	_, err := strconv.Atoi(cmd)
	return err == nil
}

func handleListCommand(args []string) {
	if help.HasHelpFlag(args, "list") {
		return
	}

	if err := worktree.List(); err != nil {
		printErrorAndExit("%v", err)
	}
}

func handleRecentCommand(args []string) {
	if help.HasHelpFlag(args, "recent") {
		return
	}

	// Parse flags
	flags := parseRecentFlags(args)

	// Get git client
	gitClient := git.NewCommandClient("")

	// Get current user (we need it unless --all is used)
	var currentUserName string
	if !flags.showAll {
		var err error
		currentUserName, err = gitClient.GetConfigValue("user.name")
		if err != nil {
			printErrorAndExit("failed to get user.name: %v", err)
		}
		if currentUserName == "" {
			printErrorAndExit("user.name not configured in git. Run: git config user.name \"Your Name\"")
		}
	}

	// Collect branch information
	branchResult := collectBranchInfo(gitClient)
	if len(branchResult.branches) == 0 {
		displayNoBranchesMessage(branchResult.skipped, flags.verbose)
		return
	}

	// Update worktree information
	updateWorktreeInfo(branchResult.branches, gitClient)

	// Filter branches based on flags
	branches := filterBranches(branchResult.branches, flags, currentUserName)

	// Handle numeric navigation if requested
	if flags.navigateIndex >= 0 {
		navigateToBranch(branches, flags.navigateIndex, gitClient)
		return
	}

	// Display branches
	if flags.compact {
		displayBranchesCompact(branches, flags.count)
	} else {
		displayBranches(branches, flags.count)
	}

	// Display summary of skipped branches if verbose mode is enabled
	displaySkippedBranchesIfVerbose(branchResult.skipped, flags.verbose)
}

func handleRemoveCommand(args []string) {
	if help.HasHelpFlag(args, "rm") {
		return
	}

	// Parse flags
	var useFuzzy bool
	var target string
	var deleteBranch bool
	var force bool

	for _, arg := range args {
		if arg == fuzzyFlag || arg == fuzzyFlagShort {
			useFuzzy = true
		} else if arg == "--branch" {
			deleteBranch = true
		} else if arg == "--force" {
			force = true
		} else if arg == helpFlag || arg == helpFlagShort {
			// Skip help flags - they're handled separately
			continue
		} else {
			// First non-flag argument is the target
			target = arg
		}
	}

	if target == "" {
		target = selectBranchInteractively(useFuzzy, "Usage: wt rm <branch>")
	} else {
		// Resolve target with fuzzy matching
		branches, branchErr := worktree.GetAvailableBranches()
		if branchErr != nil {
			printErrorAndExit("%v", branchErr)
		}

		resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
		if resolveErr != nil {
			printErrorAndExit("%v", resolveErr)
		}
		target = resolvedTarget
	}

	var err error
	if deleteBranch {
		err = worktree.RemoveWithOptions(target, worktree.RemoveOptions{DeleteBranch: true, Force: force})
	} else {
		err = worktree.Remove(target)
	}

	if err != nil {
		printErrorAndExit("%v", err)
	}
}

func handleIntegrateCommand(args []string) {
	if help.HasHelpFlag(args, "integrate") {
		return
	}

	var useFuzzy bool
	var target string

	for _, arg := range args {
		switch arg {
		case fuzzyFlag, fuzzyFlagShort:
			useFuzzy = true
		case helpFlag, helpFlagShort:
			continue
		default:
			if strings.HasPrefix(arg, "-") {
				if arg == "--force" {
					printErrorAndExit("wt integrate does not support --force; delete the branch manually if you need to bypass merge checks")
				}
				continue
			}
			target = arg
		}
	}

	if target == "" {
		target = selectBranchInteractively(useFuzzy, "Usage: wt integrate <worktree>")
	} else {
		branches, branchErr := worktree.GetAvailableBranches()
		if branchErr != nil {
			printErrorAndExit("%v", branchErr)
		}

		resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
		if resolveErr != nil {
			printErrorAndExit("%v", resolveErr)
		}
		target = resolvedTarget
	}

	if err := worktree.Integrate(target); err != nil {
		printErrorAndExit("%v", err)
	}
}

func handleGoCommand(args []string) {
	if help.HasHelpFlag(args, "go") {
		return
	}

	useFuzzy, target := parseGoCommandArgs(args)
	path, err := resolveGoTarget(target, useFuzzy)

	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}
	fmt.Printf("CD:%s", path)
}

// parseGoCommandArgs parses the arguments for the go command
func parseGoCommandArgs(args []string) (useFuzzy bool, target string) {
	for _, arg := range args {
		if arg == fuzzyFlag || arg == fuzzyFlagShort {
			useFuzzy = true
		} else if arg == helpFlag || arg == helpFlagShort {
			// Skip help flags - they're handled separately
			continue
		} else {
			// First non-flag argument is the target
			target = arg
			break
		}
	}
	return useFuzzy, target
}

// resolveGoTarget resolves the target for the go command
func resolveGoTarget(target string, useFuzzy bool) (string, error) {
	if target == "" {
		return handleNoTarget(useFuzzy)
	}
	return handleTargetProvided(target)
}

// handleNoTarget handles the case where no target is specified
func handleNoTarget(useFuzzy bool) (string, error) {
	branches, branchErr := worktree.GetAvailableBranches()
	if branchErr != nil {
		// Fall back to repo root if we can't get branches
		return worktree.GetRepoRoot()
	}

	if interactive.ShouldUseFuzzy(len(branches), useFuzzy) {
		// Use interactive selection
		selectedBranch, selectErr := worktree.SelectBranchInteractively()
		if selectErr != nil {
			if selectErr == interactive.ErrUserCancelled {
				fmt.Fprintf(os.Stderr, "wt: selection cancelled\n")
				osExit(1)
			}
			fmt.Fprintf(os.Stderr, "wt: %v\n", selectErr)
			osExit(1)
		}
		return handleTargetProvided(selectedBranch)
	}

	// Fall back to repo root
	return worktree.GetRepoRoot()
}

// handleTargetProvided handles the case where a target is provided
func handleTargetProvided(target string) (string, error) {
	// Check if target is numeric (index-based access)
	if _, numErr := strconv.Atoi(target); numErr == nil {
		// Numeric target - use directly with worktree.Go
		return worktree.Go(target)
	}

	// Non-numeric target - resolve with fuzzy matching
	branches, branchErr := worktree.GetAvailableBranches()
	if branchErr != nil {
		return "", branchErr
	}

	// Try to resolve the target with fuzzy matching
	resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
	if resolveErr != nil {
		return "", resolveErr
	}

	return worktree.Go(resolvedTarget)
}

func handleNewCommand(args []string, configMgr *config.Manager) {
	if help.HasHelpFlag(args, "new") {
		return
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: wt new <branch> [--base <branch>] [--no-switch]\n")
		osExit(1)
	}

	branch, baseBranch, noSwitch := parseNewCommandArgs(args)

	// Use smart worktree creation - handles all branch states intelligently
	path, err := worktree.SmartNewWorktree(branch, baseBranch, configMgr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}

	if noSwitch {
		fmt.Printf("Created worktree at %s\n", path)
	} else {
		fmt.Printf("CD:%s", path)
	}
}

func parseNewCommandArgs(args []string) (branch, baseBranch string, noSwitch bool) {
	// Filter out help flags and find the first non-help argument as branch
	var filteredArgs []string
	for _, arg := range args {
		if arg != helpFlag && arg != helpFlagShort {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) == 0 {
		return "", "", false
	}

	branch = filteredArgs[0]
	for i := 1; i < len(filteredArgs); i++ {
		if filteredArgs[i] == "--base" && i+1 < len(filteredArgs) {
			baseBranch = filteredArgs[i+1]
			i++
		} else if filteredArgs[i] == "--no-switch" {
			noSwitch = true
		}
	}
	return
}

func handleEnvCopyCommand(args []string) {
	if help.HasHelpFlag(args, "env-copy") {
		return
	}

	// Parse flags
	var useFuzzy bool
	var target string
	var recursive bool

	for _, arg := range args {
		if arg == fuzzyFlag || arg == fuzzyFlagShort {
			useFuzzy = true
		} else if arg == "--recursive" {
			recursive = true
		} else if arg == helpFlag || arg == helpFlagShort {
			// Skip help flags - they're handled separately
			continue
		} else {
			// First non-flag argument is the target
			target = arg
			break
		}
	}

	if target == "" {
		target = selectBranchInteractively(useFuzzy, "Usage: wt env-copy <branch> [--recursive]")
	}

	if err := worktree.CopyEnvFile(target, recursive); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}
}

func handleEnvCommand(args []string) {
	if help.HasHelpFlag(args, "env") {
		return
	}

	if len(args) == 0 {
		// Interactive mode - show menu
		handleEnvInteractive()
		return
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "sync":
		handleEnvSyncCommand(subargs)
	case "diff":
		handleEnvDiffCommand(subargs)
	case "list":
		handleEnvListCommand(subargs)
	default:
		fmt.Fprintf(os.Stderr, "wt: unknown env subcommand '%s'\n", subcommand)
		fmt.Fprintf(os.Stderr, "Available subcommands: sync, diff, list\n")
		fmt.Fprintf(os.Stderr, "Use 'wt env --help' for detailed help\n")
		osExit(1)
	}
}

func handleEnvSyncCommand(args []string) {
	// Parse flags
	var useFuzzy bool
	var recursive bool
	var syncAll bool
	var target string

	for _, arg := range args {
		if arg == fuzzyFlag || arg == fuzzyFlagShort {
			useFuzzy = true
		} else if arg == "--recursive" {
			recursive = true
		} else if arg == "--all" {
			syncAll = true
		} else if arg == helpFlag || arg == helpFlagShort {
			// Skip help flags - they're handled separately
			continue
		} else {
			// First non-flag argument is the target
			target = arg
			break
		}
	}

	if syncAll {
		if err := worktree.SyncEnvFiles("", recursive, true); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			osExit(1)
		}
		return
	}

	if target == "" {
		target = selectBranchInteractively(useFuzzy, "Usage: wt env sync <branch> [--recursive] [--all]")
	} else {
		// Resolve target with fuzzy matching
		branches, branchErr := worktree.GetAvailableBranches()
		if branchErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", branchErr)
			osExit(1)
		}

		resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", resolveErr)
			osExit(1)
		}
		target = resolvedTarget
	}

	if err := worktree.SyncEnvFiles(target, recursive, false); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}
}

func handleEnvDiffCommand(args []string) {
	// Parse flags and target
	var useFuzzy bool
	var target string

	for _, arg := range args {
		if arg == fuzzyFlag || arg == fuzzyFlagShort {
			useFuzzy = true
		} else if arg == helpFlag || arg == helpFlagShort {
			// Skip help flags - they're handled separately
			continue
		} else {
			// First non-flag argument is the target
			target = arg
			break
		}
	}

	if target == "" {
		target = selectBranchInteractively(useFuzzy, "Usage: wt env diff <branch>")
	} else {
		// Resolve target with fuzzy matching
		branches, branchErr := worktree.GetAvailableBranches()
		if branchErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", branchErr)
			osExit(1)
		}

		resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", resolveErr)
			osExit(1)
		}
		target = resolvedTarget
	}

	if err := worktree.DiffEnvFiles(target); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}
}

func handleEnvListCommand(args []string) {
	if err := worktree.ListEnvFiles(); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}
}

func handleEnvInteractive() {
	if !interactive.IsInteractive() {
		fmt.Fprintf(os.Stderr, "Usage: wt env <subcommand> [options]\n")
		fmt.Fprintf(os.Stderr, "Subcommands: sync, diff, list\n")
		fmt.Fprintf(os.Stderr, "Use 'wt env --help' for detailed help\n")
		osExit(1)
	}

	// Show interactive menu for env operations
	options := []string{"sync", "diff", listCmd, helpCmd}
	selected, err := interactive.SelectString(options, "Environment operation:")
	if err != nil {
		if err == interactive.ErrUserCancelled {
			fmt.Fprintf(os.Stderr, "wt: selection cancelled\n")
			osExit(1)
		}
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}

	switch selected {
	case "sync":
		handleEnvSyncCommand([]string{})
	case "diff":
		handleEnvDiffCommand([]string{})
	case listCmd:
		handleEnvListCommand([]string{})
	case helpCmd:
		help.ShowCommandHelp("env")
	}
}

func handleVersionCommand(args []string) {
	if help.HasHelpFlag(args, "version") {
		return
	}

	versionStr := version
	if versionStr == "" {
		versionStr = "dev"
	}

	if date != "" && date != "unknown" {
		fmt.Printf("wt version %s (built %s)\n", versionStr, date)
	} else {
		fmt.Printf("wt version %s\n", versionStr)
	}
}

func handleCustomCommand(cmd string, configMgr *config.Manager) {
	if navCmd, exists := configMgr.GetCommand(cmd); exists {
		if navCmd.Type == "virtualenv" {
			handleVirtualenvCommand(navCmd, configMgr)
		} else {
			handleNavigationCommand(navCmd)
		}
	} else {
		fmt.Fprintf(os.Stderr, "wt: unknown command '%s'\n", cmd)
		showUsage()
		osExit(1)
	}
}

func handleNavigationCommand(navCmd *config.NavigationCommand) {
	repo, err := worktree.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}

	targetPath := filepath.Join(repo, navCmd.Target)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "wt: '%s' not found under repo\n", navCmd.Target)
		osExit(1)
	}

	fmt.Printf("CD:%s", targetPath)
}

func handleVirtualenvCommand(navCmd *config.NavigationCommand, configMgr *config.Manager) {
	repo, err := worktree.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}

	venvConfig := configMgr.GetVirtualenvConfig()
	if venvConfig == nil {
		fmt.Fprintf(os.Stderr, "wt: virtualenv not configured for this project\n")
		osExit(1)
		return // This will never be reached, but satisfies the linter
	}

	// Default values
	venvName := venvConfig.Name
	if venvName == "" {
		venvName = ".venv"
	}

	python := venvConfig.Python
	if python == "" {
		python = "python3"
	}

	venvPath := filepath.Join(repo, venvName)

	// Validate that venvPath is within the repository
	absRepo, _ := filepath.Abs(repo)
	absVenvPath, _ := filepath.Abs(venvPath)
	if !strings.HasPrefix(absVenvPath, absRepo) {
		printErrorAndExit("invalid virtualenv path: must be within repository")
	}

	switch navCmd.Target {
	case "activate":
		// Check if virtualenv exists
		activateScript := filepath.Join(venvPath, "bin", "activate")
		if _, err := os.Stat(activateScript); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "wt: virtualenv not found at %s\n", venvPath)
			fmt.Fprintf(os.Stderr, "Run 'wt mkvenv' to create it\n")
			osExit(1)
		}
		// Output EXEC command to activate virtualenv
		// Use printf with %q to properly quote the path for shell safety
		fmt.Printf("EXEC:source %q", activateScript)

	case "create":
		// Check if virtualenv already exists
		if _, err := os.Stat(venvPath); err == nil {
			fmt.Fprintf(os.Stderr, "wt: virtualenv already exists at %s\n", venvPath)
			osExit(1)
		}

		// Create virtualenv
		fmt.Printf("Creating virtualenv at %s...\n", venvPath)
		if err := worktree.RunCommand(python, "-m", "venv", venvPath); err != nil {
			fmt.Fprintf(os.Stderr, "wt: failed to create virtualenv: %v\n", err)
			osExit(1)
		}
		fmt.Println("Virtualenv created successfully")

	case "remove":
		// Check if virtualenv exists
		if _, err := os.Stat(venvPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "wt: virtualenv not found at %s\n", venvPath)
			osExit(1)
		}

		// Remove virtualenv
		fmt.Printf("Removing virtualenv at %s...\n", venvPath)
		if err := os.RemoveAll(venvPath); err != nil {
			fmt.Fprintf(os.Stderr, "wt: failed to remove virtualenv: %v\n", err)
			osExit(1)
		}
		fmt.Println("Virtualenv removed successfully")

	default:
		fmt.Fprintf(os.Stderr, "wt: unknown virtualenv action '%s'\n", navCmd.Target)
		osExit(1)
	}
}

func handleProjectCommand(args []string, configMgr *config.Manager) {
	if help.HasHelpFlag(args, "project") {
		return
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: wt project [init|setup]\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project --help' for detailed help\n")
		osExit(1)
		return // Needed for testing when osExit is mocked
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "init":
		if len(subargs) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt project init <project-name>\n")
			osExit(1)
		}

		projectName := subargs[0]
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: failed to get current directory: %v\n", err)
			osExit(1)
		}

		repo, err := worktree.GetRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			osExit(1)
		}

		// Create project config
		project := &config.ProjectConfig{
			Name: projectName,
			Match: config.ProjectMatch{
				Paths: []string{repo},
			},
			Commands: make(map[string]config.NavigationCommand),
		}

		// Get git remote if available
		if remote, err := worktree.GetGitRemote(); err == nil && remote != "" {
			project.Match.Remotes = []string{remote}
		}

		// Get worktree base
		if base, err := worktree.GetWorktreeBase(); err == nil {
			project.Settings.WorktreeBase = base
			// Add worktree paths
			project.Match.Paths = append(project.Match.Paths, filepath.Join(base, "*"))
		}

		if err := configMgr.SaveProjectConfig(project); err != nil {
			fmt.Fprintf(os.Stderr, "wt: failed to save project config: %v\n", err)
			osExit(1)
		}

		fmt.Printf("Project '%s' initialized at %s\n", projectName, cwd)
		fmt.Printf("Config saved to: %s/projects/%s.yaml\n", configMgr.GetConfigDir(), projectName)

	case "setup":
		handleProjectSetupCommand(subargs, configMgr)

	default:
		fmt.Fprintf(os.Stderr, "wt: unknown project subcommand '%s'\n", subcmd)
		fmt.Fprintf(os.Stderr, "Available subcommands: init, setup\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project --help' for detailed help\n")
		osExit(1)
	}
}

func handleProjectSetupCommand(args []string, configMgr *config.Manager) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "wt: project setup command requires a subcommand\n")
		fmt.Fprintf(os.Stderr, "Available subcommands: run, show\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project --help' for detailed help\n")
		osExit(1)
	}

	subcommand := args[0]
	subargs := args[1:]

	switch subcommand {
	case "run":
		handleProjectSetupRunCommand(subargs, configMgr)
	case "show":
		handleProjectSetupShowCommand(subargs, configMgr)
	default:
		fmt.Fprintf(os.Stderr, "wt: unknown project setup subcommand '%s'\n", subcommand)
		fmt.Fprintf(os.Stderr, "Available subcommands: run, show\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project --help' for detailed help\n")
		osExit(1)
	}
}

func handleProjectSetupRunCommand(args []string, configMgr *config.Manager) {
	// Get current project configuration
	currentProject := configMgr.GetCurrentProject()
	if currentProject == nil {
		fmt.Fprintf(os.Stderr, "wt: no project configuration found for current directory\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project init <name>' to configure this project\n")
		osExit(1)
		return // This will never be reached, but satisfies the linter
	}

	if currentProject.Setup == nil {
		fmt.Printf("No setup automation configured for project '%s'\n", currentProject.Name)
		return
	}

	// Get repository root and current worktree path
	repoRoot, err := worktree.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: failed to get current directory: %v\n", err)
		osExit(1)
	}

	// Determine which worktree we're in
	worktrees, err := worktree.GetWorktreeInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		osExit(1)
	}

	var currentWorktreePath string
	for _, wt := range worktrees {
		if strings.HasPrefix(currentDir, wt.Path) {
			currentWorktreePath = wt.Path
			break
		}
	}

	if currentWorktreePath == "" {
		fmt.Fprintf(os.Stderr, "wt: not currently in a worktree\n")
		osExit(1)
	}

	fmt.Printf("Running setup automation for project '%s'...\n", currentProject.Name)
	if err := worktree.RunSetup(repoRoot, currentWorktreePath, currentProject.Setup); err != nil {
		fmt.Fprintf(os.Stderr, "wt: setup failed: %v\n", err)
		osExit(1)
	}

	fmt.Println("Setup completed successfully!")
}

func handleProjectSetupShowCommand(args []string, configMgr *config.Manager) {
	// Get current project configuration
	currentProject := configMgr.GetCurrentProject()
	if currentProject == nil {
		fmt.Fprintf(os.Stderr, "wt: no project configuration found for current directory\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project init <name>' to configure this project\n")
		osExit(1)
		return // This will never be reached, but satisfies the linter
	}

	if currentProject.Setup == nil {
		fmt.Printf("No setup automation configured for project '%s'\n", currentProject.Name)
		fmt.Println("\nTo configure setup automation, edit your project config file:")
		fmt.Printf("~/.config/wt/projects/%s.yaml\n", currentProject.Name)
		return
	}

	fmt.Printf("Setup automation for project '%s':\n\n", currentProject.Name)

	if len(currentProject.Setup.CreateDirectories) > 0 {
		fmt.Println("Create directories:")
		for _, dir := range currentProject.Setup.CreateDirectories {
			fmt.Printf("  - %s\n", dir)
		}
		fmt.Println()
	}

	if len(currentProject.Setup.CopyFiles) > 0 {
		fmt.Println("Copy files:")
		for _, copyFile := range currentProject.Setup.CopyFiles {
			fmt.Printf("  - %s → %s\n", copyFile.Source, copyFile.Target)
		}
		fmt.Println()
	}

	if len(currentProject.Setup.Commands) > 0 {
		fmt.Println("Run commands:")
		for _, cmd := range currentProject.Setup.Commands {
			fmt.Printf("  - %s (in %s)\n", cmd.Command, cmd.Directory)
		}
		fmt.Println()
	}
}

const (
	completionNone = "none"
)

func handleSetupCommand(args []string) {
	if help.HasHelpFlag(args, "setup") {
		return
	}

	var completionOpts setup.CompletionOptions
	var installMode bool

	// Parse arguments and handle special modes first
	for i, arg := range args {
		switch arg {
		case "--check":
			handleSetupCheck()
			return
		case "--uninstall":
			handleSetupUninstall()
			return
		case "--completion":
			handleCompletionOption(args, i, &completionOpts, &installMode)
		case "--no-completion":
			completionOpts.Install = false
			completionOpts.Shell = completionNone
			installMode = true
		default:
			handleUnknownSetupOption(args, i, arg)
		}
	}

	// Default setup if no specific mode
	if !installMode && len(args) == 0 {
		completionOpts.Install = true
		completionOpts.Shell = "auto"
		installMode = true
	}

	if installMode {
		performSetupInstallation(completionOpts)
	}
}

func handleSetupCheck() {
	if err := setup.Check(); err != nil {
		fmt.Fprintf(os.Stderr, "Check failed: %v\n", err)
		osExit(1)
	}
}

func handleSetupUninstall() {
	if err := setup.Uninstall(); err != nil {
		fmt.Fprintf(os.Stderr, "Uninstall failed: %v\n", err)
		osExit(1)
	}
}

func handleCompletionOption(args []string, i int, opts *setup.CompletionOptions, installMode *bool) {
	if i+1 >= len(args) {
		fmt.Fprintf(os.Stderr, "Error: --completion requires a value (auto|bash|zsh|none)\n")
		showSetupUsage()
		osExit(1)
	}
	completionValue := args[i+1]
	switch completionValue {
	case "auto", "bash", "zsh", completionNone:
		opts.Install = completionValue != completionNone
		opts.Shell = completionValue
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid completion option '%s'. Use auto|bash|zsh|none\n", completionValue)
		showSetupUsage()
		osExit(1)
	}
	*installMode = true
}

func handleUnknownSetupOption(args []string, i int, arg string) {
	// Skip completion values
	if i > 0 && args[i-1] == "--completion" {
		return
	}
	fmt.Fprintf(os.Stderr, "Unknown setup option: %s\n", arg)
	showSetupUsage()
	osExit(1)
}

func performSetupInstallation(opts setup.CompletionOptions) {
	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get executable path: %v\n", err)
		osExit(1)
	}

	if err := setup.SetupWithOptions(binaryPath, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
		osExit(1)
	}
}

func showSetupUsage() {
	fmt.Fprintf(os.Stderr, "Usage: wt setup [options]\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  --completion <shell>   Install completion for specified shell (auto|bash|zsh|none)\n")
	fmt.Fprintf(os.Stderr, "  --no-completion        Skip completion installation\n")
	fmt.Fprintf(os.Stderr, "  --check                Check installation status\n")
	fmt.Fprintf(os.Stderr, "  --uninstall            Remove wt from system\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  wt setup                      # Install with auto-detected completion\n")
	fmt.Fprintf(os.Stderr, "  wt setup --completion bash    # Install with bash completion\n")
	fmt.Fprintf(os.Stderr, "  wt setup --no-completion      # Install without completion\n")
	fmt.Fprintf(os.Stderr, "  wt setup --check              # Check installation\n")
}

func showUsage() {
	fmt.Fprint(os.Stderr, getCoreUsage())
	printProjectCommands()
}

func getCoreUsage() string {
	return `Usage: wt <command> [arguments]

Quick access:
  wt 0, wt 1, wt 2    Quick switch to worktree by index (shortcut for 'wt go 0')
  --fuzzy, -f         Force interactive selection for branch/worktree arguments

Smart commands (with fuzzy branch matching):
  list, ls            List all worktrees
  recent              Show YOUR recently active branches (default: your branches only)
                      Navigate directly: 'wt recent 2' → go to your 3rd recent branch
                      Default: multi-line format for better readability
                      Options: --all, --others, -n <count>, --compact, --verbose
  new <branch>        Smart worktree creation - handles all branch states:
                      • Branch doesn't exist → Create branch + worktree + switch
                      • Branch exists, no worktree → Create worktree + switch
                      • Branch + worktree exist → Just switch
                      Options: --base <branch>, --no-switch
  go, switch, s       Switch to a worktree (no args = repo root)
                      Supports fuzzy matching: 'wt go mai' → switches to 'main'
                      Options: --fuzzy, -f (force interactive selection)
  rm <branch>         Remove a worktree (supports fuzzy matching)
                      Options: --fuzzy, -f (force interactive selection)

Utility commands:
  env <subcommand>    Unified environment file management
                      Subcommands: sync, diff, list
                      Options: --all, --fuzzy, -f, --recursive
  project init <name> Initialize project configuration

Setup commands:
  setup               Install wt to ~/.local/bin with shell completion
                      Options: --completion <shell>, --no-completion, --check, --uninstall
  update              Check and install updates
                      Options: --check, --force

Other commands:
  ` + shellInitCmd + `          Output shell initialization code
  completion <shell>  Generate shell completion scripts (bash|zsh)
  version             Show version information`
}

func printProjectCommands() {
	configMgr, err := config.NewManager()
	if err != nil {
		return
	}

	cwd, _ := os.Getwd()
	gitRemote, _ := worktree.GetGitRemote()
	_ = configMgr.LoadProject(cwd, gitRemote)

	project := configMgr.GetCurrentProject()
	if project == nil || len(project.Commands) == 0 {
		return
	}

	navCommands, venvCommands := categorizeProjectCommands(project.Commands)
	printCommandCategories(navCommands, venvCommands, project.Name)
}

func categorizeProjectCommands(commands map[string]config.NavigationCommand) ([]string, []string) {
	var navCommands, venvCommands []string

	for name, cmd := range commands {
		if cmd.Type == "virtualenv" {
			venvCommands = append(venvCommands, name)
		} else {
			navCommands = append(navCommands, name)
		}
	}

	sort.Strings(navCommands)
	sort.Strings(venvCommands)
	return navCommands, venvCommands
}

func printCommandCategories(navCommands, venvCommands []string, projectName string) {
	// We need access to project commands - get them from config manager
	configMgr, err := config.NewManager()
	if err != nil {
		return
	}

	cwd, _ := os.Getwd()
	gitRemote, _ := worktree.GetGitRemote()
	_ = configMgr.LoadProject(cwd, gitRemote)

	project := configMgr.GetCurrentProject()
	if project == nil {
		return
	}

	// Show navigation commands
	if len(navCommands) > 0 {
		fmt.Fprintf(os.Stderr, "\n\nProject '%s' navigation:", projectName)
		for _, name := range navCommands {
			cmd := project.Commands[name]
			fmt.Fprintf(os.Stderr, "\n  %-18s %s", name, cmd.Description)
		}
	}

	// Show virtualenv commands
	if len(venvCommands) > 0 {
		fmt.Fprintf(os.Stderr, "\n\nProject '%s' virtualenv:", projectName)
		for _, name := range venvCommands {
			cmd := project.Commands[name]
			fmt.Fprintf(os.Stderr, "\n  %-18s %s", name, cmd.Description)
		}
	}

	fmt.Fprintln(os.Stderr)
}

func handleCompletionCommand(args []string, configMgr *config.Manager) {
	if help.HasHelpFlag(args, "completion") {
		return
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: wt completion <bash|zsh>\n")
		fmt.Fprintf(os.Stderr, "\nGenerate shell completion scripts for wt.\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Install bash completion\n")
		fmt.Fprintf(os.Stderr, "  wt completion bash >> ~/.bashrc\n\n")
		fmt.Fprintf(os.Stderr, "  # Install zsh completion\n")
		fmt.Fprintf(os.Stderr, "  wt completion zsh >> ~/.zshrc\n\n")
		fmt.Fprintf(os.Stderr, "  # Or use with eval\n")
		fmt.Fprintf(os.Stderr, "  eval \"$(wt completion bash)\"\n")
		osExit(1)
	}

	shell := args[0]
	switch shell {
	case "bash":
		fmt.Print(completion.GenerateBashCompletion(configMgr))
	case "zsh":
		fmt.Print(completion.GenerateZshCompletion(configMgr))
	default:
		fmt.Fprintf(os.Stderr, "wt: unsupported shell '%s'\n", shell)
		fmt.Fprintf(os.Stderr, "Supported shells: bash, zsh\n")
		osExit(1)
	}
}

func handleUpdateCommand(args []string) {
	if help.HasHelpFlag(args, "update") {
		return
	}

	// Parse flags
	checkOnly := false
	force := false

	for _, arg := range args {
		switch arg {
		case "--check":
			checkOnly = true
		case forceFlag:
			force = true
		case helpFlag, helpFlagShort:
			// Skip help flags - they're handled separately
			continue
		}
	}

	fmt.Printf("Current version: %s\n", version)

	// Check for updates
	release, hasUpdate, err := update.CheckForUpdate(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: failed to check for updates: %v\n", err)
		osExit(1)
	}

	fmt.Printf("Latest version: %s\n", release.TagName)

	if !hasUpdate && !force {
		fmt.Println("\nYou are already on the latest version!")
		return
	}

	if checkOnly {
		if hasUpdate {
			fmt.Printf("\nUpdate available: %s\n", release.TagName)
			fmt.Println("\nChanges:")
			fmt.Println(release.Body)
		}
		return
	}

	// Download and install update
	fmt.Println("\nDownloading update...")

	var lastProgress int
	err = update.DownloadAndInstall(release, func(downloaded, total int64) {
		progress := int(float64(downloaded) / float64(total) * 100)
		if progress != lastProgress && progress%10 == 0 {
			fmt.Printf("\rProgress: %d%%", progress)
			lastProgress = progress
		}
	})

	fmt.Println() // New line after progress

	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: update failed: %v\n", err)
		osExit(1)
	}

	fmt.Printf("✓ Successfully updated to %s\n", release.TagName)

	if release.Body != "" {
		fmt.Println("\nChanges in this version:")
		fmt.Println(release.Body)
	}
}

// recentFlags holds parsed flags for the recent command
type recentFlags struct {
	showOthers    bool
	showAll       bool
	count         int
	navigateIndex int
	verbose       bool
	compact       bool
}

// parseRecentFlags parses command line flags for the recent command
func parseRecentFlags(args []string) recentFlags {
	flags := recentFlags{
		count:         10,
		navigateIndex: -1,
	}

	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--others":
			flags.showOthers = true
			i++
		case arg == "--all":
			flags.showAll = true
			i++
		case arg == "--verbose" || arg == "-v":
			flags.verbose = true
			i++
		case arg == "--compact" || arg == "-c":
			flags.compact = true
			i++
		case arg == "-n" && i+1 < len(args):
			flags.count = parseAndValidateCount(args[i+1])
			i += 2
		case strings.HasPrefix(arg, "-n="):
			countStr := strings.TrimPrefix(arg, "-n=")
			flags.count = parseAndValidateCount(countStr)
			i++
		default:
			// Check if it's a numeric argument for navigation
			if idx, err := strconv.Atoi(arg); err == nil && flags.navigateIndex == -1 {
				flags.navigateIndex = validateNavigationIndex(idx)
			}
			i++
		}
	}

	// Validate flags
	if flags.showOthers && flags.showAll {
		printErrorAndExit("cannot use --others and --all together")
	}

	return flags
}

// parseAndValidateCount parses and validates a count value
func parseAndValidateCount(countStr string) int {
	n, err := strconv.Atoi(countStr)
	if err != nil {
		printErrorAndExit("invalid count value: %s (must be a number)", countStr)
	}
	if n <= 0 {
		printErrorAndExit("count must be positive, got: %d", n)
	}
	return n
}

// validateNavigationIndex validates a navigation index
func validateNavigationIndex(idx int) int {
	if idx < 0 {
		printErrorAndExit("navigation index must be non-negative, got: %d", idx)
	}
	return idx
}

// branchCommitInfo holds information about a branch and its last commit
type branchCommitInfo struct {
	branch       string
	commitHash   string
	relativeDate string
	subject      string
	author       string
	timestamp    time.Time
	hasWorktree  bool
}

// skippedBranchInfo holds information about why a branch was skipped
type skippedBranchInfo struct {
	branch string
	reason string
}

// branchCollectionResult holds the result of collecting branch information
type branchCollectionResult struct {
	branches       []branchCommitInfo
	skipped        []skippedBranchInfo
	totalProcessed int
}

// collectBranchInfo collects commit information for all branches
func collectBranchInfo(gitClient git.Client) branchCollectionResult {
	// Get all branches first
	branchesOutput, err := gitClient.ForEachRef("%(refname:short)", "refs/heads/")
	if err != nil {
		printErrorAndExit("failed to get branches: %v", err)
	}

	result := branchCollectionResult{
		branches: make([]branchCommitInfo, 0),
		skipped:  make([]skippedBranchInfo, 0),
	}

	if branchesOutput == "" {
		return result
	}

	// Parse branch names
	branchNames := strings.Split(branchesOutput, "\n")
	branchInfos := make([]branchCommitInfo, 0, len(branchNames))

	// Format for the commit info - include unix timestamp for sorting
	commitFormat := "%H|%cr|%s|%an|%ct"

	for _, branch := range branchNames {
		branch = strings.TrimSpace(branch)
		if branch == "" {
			continue
		}

		result.totalProcessed++

		// Get last non-merge commit info
		commitInfo, err := gitClient.GetLastNonMergeCommit(branch, commitFormat)
		if err != nil {
			result.skipped = append(result.skipped, skippedBranchInfo{
				branch: branch,
				reason: fmt.Sprintf("git command failed: %v", err),
			})
			continue
		}

		if commitInfo == "" {
			result.skipped = append(result.skipped, skippedBranchInfo{
				branch: branch,
				reason: "no non-merge commits found",
			})
			continue
		}

		// Parse commit info
		parts := strings.Split(commitInfo, "|")
		if len(parts) != 5 {
			result.skipped = append(result.skipped, skippedBranchInfo{
				branch: branch,
				reason: fmt.Sprintf("invalid commit info format: expected 5 parts, got %d", len(parts)),
			})
			continue
		}

		// Parse unix timestamp
		unixTime, err := strconv.ParseInt(parts[4], 10, 64)
		if err != nil {
			result.skipped = append(result.skipped, skippedBranchInfo{
				branch: branch,
				reason: fmt.Sprintf("invalid timestamp: %v", err),
			})
			continue
		}

		branchInfos = append(branchInfos, branchCommitInfo{
			branch:       branch,
			commitHash:   parts[0],
			relativeDate: parts[1],
			subject:      parts[2],
			author:       parts[3],
			timestamp:    time.Unix(unixTime, 0),
		})
	}

	// Sort by commit timestamp (most recent first)
	sort.Slice(branchInfos, func(i, j int) bool {
		return branchInfos[i].timestamp.After(branchInfos[j].timestamp)
	})

	result.branches = branchInfos
	return result
}

// updateWorktreeInfo updates branch info with worktree status
func updateWorktreeInfo(branchInfos []branchCommitInfo, gitClient git.Client) {
	// Get worktree list to check which branches have worktrees
	worktrees, err := gitClient.WorktreeList()
	if err != nil {
		printErrorAndExit("failed to get worktrees: %v", err)
	}

	// Create a map of branches that have worktrees
	worktreeBranches := make(map[string]bool)
	for _, wt := range worktrees {
		worktreeBranches[wt.Branch] = true
	}

	// Update hasWorktree field
	for i := range branchInfos {
		branchInfos[i].hasWorktree = worktreeBranches[branchInfos[i].branch]
	}
}

// filterBranches filters branches based on author and flags
func filterBranches(branchInfos []branchCommitInfo, flags recentFlags, currentUserName string) []branchCommitInfo {
	branches := make([]branchCommitInfo, 0, len(branchInfos))
	for _, bi := range branchInfos {
		if flags.showAll {
			// Show all branches
			branches = append(branches, bi)
		} else if flags.showOthers {
			// Show only other users' branches
			if bi.author != currentUserName {
				branches = append(branches, bi)
			}
		} else {
			// Default: show only current user's branches
			if bi.author == currentUserName {
				branches = append(branches, bi)
			}
		}
	}
	return branches
}

// navigateToBranch handles navigation to a specific branch by index
func navigateToBranch(branches []branchCommitInfo, index int, gitClient git.Client) {
	if index >= len(branches) {
		printErrorAndExit("invalid index: %d (only %d branches available)", index, len(branches))
	}

	targetBranch := branches[index]

	// Check if branch has a worktree
	if targetBranch.hasWorktree {
		// Find the worktree path
		worktrees, err := gitClient.WorktreeList()
		if err != nil {
			printErrorAndExit("failed to get worktrees: %v", err)
		}

		for _, wt := range worktrees {
			if wt.Branch == targetBranch.branch {
				fmt.Printf("CD:%s", wt.Path)
				return
			}
		}
	} else {
		// No worktree, checkout the branch
		if err := gitClient.Checkout(targetBranch.branch); err != nil {
			printErrorAndExit("failed to checkout branch %s: %v", targetBranch.branch, err)
		}
		fmt.Printf("Switched to branch '%s'\n", targetBranch.branch)
	}
}

// displayBranches shows the list of branches in multi-line format (default)
func displayBranches(branches []branchCommitInfo, count int) {
	displayCount := len(branches)
	if displayCount > count {
		displayCount = count
	}

	if displayCount == 0 {
		return
	}

	for i := 0; i < displayCount; i++ {
		branch := branches[i]

		// First line: index with proper spacing, then branch name with star
		if branch.hasWorktree {
			fmt.Printf("%d: *%s\n", i, branch.branch)
		} else {
			fmt.Printf("%d: %s\n", i, branch.branch)
		}

		// Second line: commit subject (indented to align with branch name)
		fmt.Printf("   %s\n", branch.subject)

		// Third line: author and date (indented to align with branch name)
		fmt.Printf("   %s, %s\n", branch.author, branch.relativeDate)

		// Add blank line between entries (but not after the last one)
		if i < displayCount-1 {
			fmt.Println()
		}
	}
}

// displayBranchesCompact shows the list of branches in compact single-line format
func displayBranchesCompact(branches []branchCommitInfo, count int) {
	displayCount := len(branches)
	if displayCount > count {
		displayCount = count
	}

	if displayCount == 0 {
		return
	}

	// Calculate dynamic column widths based on actual content
	maxBranchLen := 15  // minimum width
	maxDateLen := 10    // minimum width
	maxSubjectLen := 30 // minimum width

	// Find the maximum length for each column
	for i := 0; i < displayCount; i++ {
		branch := branches[i]
		branchRuneLen := len([]rune(branch.branch))
		dateRuneLen := len([]rune(branch.relativeDate))
		subjectRuneLen := len([]rune(branch.subject))

		if branchRuneLen > maxBranchLen {
			maxBranchLen = branchRuneLen
		}
		if dateRuneLen > maxDateLen {
			maxDateLen = dateRuneLen
		}
		if subjectRuneLen > maxSubjectLen {
			maxSubjectLen = subjectRuneLen
		}
	}

	// Set reasonable maximum widths to prevent overly wide columns
	const (
		maxBranchWidth  = 40
		maxSubjectWidth = 50
		maxDateWidth    = 20
	)

	if maxBranchLen > maxBranchWidth {
		maxBranchLen = maxBranchWidth
	}
	if maxSubjectLen > maxSubjectWidth {
		maxSubjectLen = maxSubjectWidth
	}
	if maxDateLen > maxDateWidth {
		maxDateLen = maxDateWidth
	}

	// Display branches with dynamic formatting
	for i := 0; i < displayCount; i++ {
		branch := branches[i]
		worktreeIndicator := " "
		if branch.hasWorktree {
			worktreeIndicator = "*"
		}

		// Truncate fields if they exceed maximum width
		branchName := truncateWithEllipsis(branch.branch, maxBranchLen)
		subject := truncateWithEllipsis(branch.subject, maxSubjectLen)
		date := truncateWithEllipsis(branch.relativeDate, maxDateLen)

		fmt.Printf("%d: %s%-*s %-*s %-*s %s\n",
			i, worktreeIndicator,
			maxBranchLen, branchName,
			maxDateLen, date,
			maxSubjectLen, subject,
			branch.author)
	}
}

// truncateWithEllipsis truncates a string to maxLen and adds ellipsis if needed
func truncateWithEllipsis(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}

// displayNoBranchesMessage shows appropriate message when no branches are found
func displayNoBranchesMessage(skipped []skippedBranchInfo, verbose bool) {
	if len(skipped) > 0 {
		fmt.Printf("No valid branches found (%d branches skipped)\n", len(skipped))
		if verbose {
			fmt.Println("\nSkipped branches:")
			for _, s := range skipped {
				fmt.Printf("  %s: %s\n", s.branch, s.reason)
			}
		}
	} else {
		fmt.Println("No branches found")
	}
}

// displaySkippedBranchesIfVerbose shows skipped branches summary if verbose mode is enabled
func displaySkippedBranchesIfVerbose(skipped []skippedBranchInfo, verbose bool) {
	if verbose && len(skipped) > 0 {
		fmt.Printf("\n%d branches were skipped:\n", len(skipped))
		for _, s := range skipped {
			fmt.Printf("  %s: %s\n", s.branch, s.reason)
		}
	}
}

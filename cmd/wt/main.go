package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/tobiase/worktree-utils/internal/completion"
	"github.com/tobiase/worktree-utils/internal/config"
	"github.com/tobiase/worktree-utils/internal/help"
	"github.com/tobiase/worktree-utils/internal/interactive"
	"github.com/tobiase/worktree-utils/internal/setup"
	"github.com/tobiase/worktree-utils/internal/update"
	"github.com/tobiase/worktree-utils/internal/worktree"
)

// Version information - set by build flags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

// Command flags
const (
	fuzzyFlag      = "--fuzzy"
	fuzzyFlagShort = "-f"
	helpFlag       = "--help"
	helpFlagShort  = "-h"
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
    if [[ "$output" == "CD:"* ]]; then
      cd "${output#CD:}"
    elif [[ "$output" == "EXEC:"* ]]; then
      eval "${output#EXEC:}"
    else
      [ -n "$output" ] && echo "$output"
    fi
  else
    echo "$output" >&2
    return $exit_code
  fi
}
`

func main() {
	if len(os.Args) < 2 {
		// No command specified - try interactive command selection
		if interactive.IsInteractive() {
			selectedCommand, err := interactive.SelectCommandInteractively()
			if err != nil {
				if err == interactive.ErrUserCancelled {
					fmt.Fprintf(os.Stderr, "wt: selection cancelled\n")
					os.Exit(1)
				}
				// Fall back to showing usage if interactive selection fails
				showUsage()
				os.Exit(1)
			}
			// Set os.Args as if the user typed the command
			os.Args = []string{os.Args[0], selectedCommand}
		} else {
			showUsage()
			os.Exit(1)
		}
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

// selectBranchInteractively handles interactive branch selection with consistent error handling
func selectBranchInteractively(useFuzzy bool, usageMsg string) string {
	branches, branchErr := worktree.GetAvailableBranches()
	if branchErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", usageMsg)
		os.Exit(1)
	}

	if interactive.ShouldUseFuzzy(len(branches), useFuzzy) {
		selectedBranch, selectErr := worktree.SelectBranchInteractively()
		if selectErr != nil {
			if selectErr == interactive.ErrUserCancelled {
				fmt.Fprintf(os.Stderr, "wt: selection cancelled\n")
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "wt: %v\n", selectErr)
			os.Exit(1)
		}
		return selectedBranch
	}

	fmt.Fprintf(os.Stderr, "%s\n", usageMsg)
	os.Exit(1)
	return "" // unreachable
}

func initializeConfig() *config.Manager {
	configMgr, err := config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: failed to initialize config: %v\n", err)
		os.Exit(1)
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
	case "rm":
		handleRemoveCommand(args)
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
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
	}
}

func handleRemoveCommand(args []string) {
	if help.HasHelpFlag(args, "rm") {
		return
	}

	// Parse flags
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
		target = selectBranchInteractively(useFuzzy, "Usage: wt rm <branch>")
	} else {
		// Resolve target with fuzzy matching
		branches, branchErr := worktree.GetAvailableBranches()
		if branchErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", branchErr)
			os.Exit(1)
		}

		resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", resolveErr)
			os.Exit(1)
		}
		target = resolvedTarget
	}

	if err := worktree.Remove(target); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
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
		os.Exit(1)
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
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "wt: %v\n", selectErr)
			os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Usage: wt new <branch> [--base <branch>]\n")
		os.Exit(1)
	}

	branch, baseBranch := parseNewCommandArgs(args)

	// Use smart worktree creation - handles all branch states intelligently
	path, err := worktree.SmartNewWorktree(branch, baseBranch, configMgr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("CD:%s", path)
}

func parseNewCommandArgs(args []string) (branch, baseBranch string) {
	// Filter out help flags and find the first non-help argument as branch
	var filteredArgs []string
	for _, arg := range args {
		if arg != helpFlag && arg != helpFlagShort {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	if len(filteredArgs) == 0 {
		return "", ""
	}

	branch = filteredArgs[0]
	for i := 1; i < len(filteredArgs); i++ {
		if filteredArgs[i] == "--base" && i+1 < len(filteredArgs) {
			baseBranch = filteredArgs[i+1]
			i++
		}
	}
	return
}

func handleEnvCopyCommand(args []string) {
	if help.HasHelpFlag(args, "env-copy") {
		return
	}

	// Show deprecation warning
	fmt.Fprintf(os.Stderr, "Warning: 'wt env-copy' is deprecated. Use 'wt env sync' instead.\n")

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
		os.Exit(1)
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
		os.Exit(1)
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
			os.Exit(1)
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
			os.Exit(1)
		}

		resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", resolveErr)
			os.Exit(1)
		}
		target = resolvedTarget
	}

	if err := worktree.SyncEnvFiles(target, recursive, false); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
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
			os.Exit(1)
		}

		resolvedTarget, resolveErr := worktree.ResolveBranchName(target, branches)
		if resolveErr != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", resolveErr)
			os.Exit(1)
		}
		target = resolvedTarget
	}

	if err := worktree.DiffEnvFiles(target); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
	}
}

func handleEnvListCommand(args []string) {
	if err := worktree.ListEnvFiles(); err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
	}
}

func handleEnvInteractive() {
	if !interactive.IsInteractive() {
		fmt.Fprintf(os.Stderr, "Usage: wt env <subcommand> [options]\n")
		fmt.Fprintf(os.Stderr, "Subcommands: sync, diff, list\n")
		fmt.Fprintf(os.Stderr, "Use 'wt env --help' for detailed help\n")
		os.Exit(1)
	}

	// Show interactive menu for env operations
	options := []string{"sync", "diff", listCmd, helpCmd}
	selected, err := interactive.SelectString(options, "Environment operation:")
	if err != nil {
		if err == interactive.ErrUserCancelled {
			fmt.Fprintf(os.Stderr, "wt: selection cancelled\n")
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
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

	fmt.Printf("wt version %s\n", version)
	if version != "dev" {
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
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
		os.Exit(1)
	}
}

func handleNavigationCommand(navCmd *config.NavigationCommand) {
	repo, err := worktree.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
	}

	targetPath := filepath.Join(repo, navCmd.Target)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "wt: '%s' not found under repo\n", navCmd.Target)
		os.Exit(1)
	}

	fmt.Printf("CD:%s", targetPath)
}

func handleVirtualenvCommand(navCmd *config.NavigationCommand, configMgr *config.Manager) {
	repo, err := worktree.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
	}

	venvConfig := configMgr.GetVirtualenvConfig()
	if venvConfig == nil {
		fmt.Fprintf(os.Stderr, "wt: virtualenv not configured for this project\n")
		os.Exit(1)
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

	switch navCmd.Target {
	case "activate":
		// Check if virtualenv exists
		activateScript := filepath.Join(venvPath, "bin", "activate")
		if _, err := os.Stat(activateScript); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "wt: virtualenv not found at %s\n", venvPath)
			fmt.Fprintf(os.Stderr, "Run 'wt mkvenv' to create it\n")
			os.Exit(1)
		}
		// Output EXEC command to activate virtualenv
		fmt.Printf("EXEC:source %s", activateScript)

	case "create":
		// Check if virtualenv already exists
		if _, err := os.Stat(venvPath); err == nil {
			fmt.Fprintf(os.Stderr, "wt: virtualenv already exists at %s\n", venvPath)
			os.Exit(1)
		}

		// Create virtualenv
		fmt.Printf("Creating virtualenv at %s...\n", venvPath)
		if err := worktree.RunCommand(python, "-m", "venv", venvPath); err != nil {
			fmt.Fprintf(os.Stderr, "wt: failed to create virtualenv: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Virtualenv created successfully")

	case "remove":
		// Check if virtualenv exists
		if _, err := os.Stat(venvPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "wt: virtualenv not found at %s\n", venvPath)
			os.Exit(1)
		}

		// Remove virtualenv
		fmt.Printf("Removing virtualenv at %s...\n", venvPath)
		if err := os.RemoveAll(venvPath); err != nil {
			fmt.Fprintf(os.Stderr, "wt: failed to remove virtualenv: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Virtualenv removed successfully")

	default:
		fmt.Fprintf(os.Stderr, "wt: unknown virtualenv action '%s'\n", navCmd.Target)
		os.Exit(1)
	}
}

func handleProjectCommand(args []string, configMgr *config.Manager) {
	if help.HasHelpFlag(args, "project") {
		return
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: wt project [init|setup]\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project --help' for detailed help\n")
		os.Exit(1)
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "init":
		if len(subargs) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt project init <project-name>\n")
			os.Exit(1)
		}

		projectName := subargs[0]
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: failed to get current directory: %v\n", err)
			os.Exit(1)
		}

		repo, err := worktree.GetRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
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
			os.Exit(1)
		}

		fmt.Printf("Project '%s' initialized at %s\n", projectName, cwd)
		fmt.Printf("Config saved to: %s/projects/%s.yaml\n", configMgr.GetConfigDir(), projectName)

	case "setup":
		handleProjectSetupCommand(subargs, configMgr)

	default:
		fmt.Fprintf(os.Stderr, "wt: unknown project subcommand '%s'\n", subcmd)
		fmt.Fprintf(os.Stderr, "Available subcommands: init, setup\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project --help' for detailed help\n")
		os.Exit(1)
	}
}

func handleProjectSetupCommand(args []string, configMgr *config.Manager) {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "wt: project setup command requires a subcommand\n")
		fmt.Fprintf(os.Stderr, "Available subcommands: run, show\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project --help' for detailed help\n")
		os.Exit(1)
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
		os.Exit(1)
	}
}

func handleProjectSetupRunCommand(args []string, configMgr *config.Manager) {
	// Get current project configuration
	currentProject := configMgr.GetCurrentProject()
	if currentProject == nil {
		fmt.Fprintf(os.Stderr, "wt: no project configuration found for current directory\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project init <name>' to configure this project\n")
		os.Exit(1)
	}

	if currentProject.Setup == nil {
		fmt.Printf("No setup automation configured for project '%s'\n", currentProject.Name)
		return
	}

	// Get repository root and current worktree path
	repoRoot, err := worktree.GetRepoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Determine which worktree we're in
	worktrees, err := worktree.GetWorktreeInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: %v\n", err)
		os.Exit(1)
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
		os.Exit(1)
	}

	fmt.Printf("Running setup automation for project '%s'...\n", currentProject.Name)
	if err := worktree.RunSetup(repoRoot, currentWorktreePath, currentProject.Setup); err != nil {
		fmt.Fprintf(os.Stderr, "wt: setup failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Setup completed successfully!")
}

func handleProjectSetupShowCommand(args []string, configMgr *config.Manager) {
	// Get current project configuration
	currentProject := configMgr.GetCurrentProject()
	if currentProject == nil {
		fmt.Fprintf(os.Stderr, "wt: no project configuration found for current directory\n")
		fmt.Fprintf(os.Stderr, "Use 'wt project init <name>' to configure this project\n")
		os.Exit(1)
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
		os.Exit(1)
	}
}

func handleSetupUninstall() {
	if err := setup.Uninstall(); err != nil {
		fmt.Fprintf(os.Stderr, "Uninstall failed: %v\n", err)
		os.Exit(1)
	}
}

func handleCompletionOption(args []string, i int, opts *setup.CompletionOptions, installMode *bool) {
	if i+1 >= len(args) {
		fmt.Fprintf(os.Stderr, "Error: --completion requires a value (auto|bash|zsh|none)\n")
		showSetupUsage()
		os.Exit(1)
	}
	completionValue := args[i+1]
	switch completionValue {
	case "auto", "bash", "zsh", completionNone:
		opts.Install = completionValue != completionNone
		opts.Shell = completionValue
	default:
		fmt.Fprintf(os.Stderr, "Error: invalid completion option '%s'. Use auto|bash|zsh|none\n", completionValue)
		showSetupUsage()
		os.Exit(1)
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
	os.Exit(1)
}

func performSetupInstallation(opts setup.CompletionOptions) {
	binaryPath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	if err := setup.SetupWithOptions(binaryPath, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
		os.Exit(1)
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

Interactive features:
  wt                  Launch interactive command selection (when no command specified)
  wt 0, wt 1, wt 2    Quick switch to worktree by index (shortcut for 'wt go 0')
  --fuzzy, -f         Force interactive selection for branch/worktree arguments

Smart commands (with fuzzy branch matching):
  list, ls            List all worktrees
  new <branch>        Smart worktree creation - handles all branch states:
                      • Branch doesn't exist → Create branch + worktree + switch
                      • Branch exists, no worktree → Create worktree + switch
                      • Branch + worktree exist → Just switch
                      Options: --base <branch>
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
		os.Exit(1)
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
		os.Exit(1)
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
		case "--force":
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
		os.Exit(1)
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
		os.Exit(1)
	}

	fmt.Printf("✓ Successfully updated to %s\n", release.TagName)

	if release.Body != "" {
		fmt.Println("\nChanges in this version:")
		fmt.Println(release.Body)
	}
}

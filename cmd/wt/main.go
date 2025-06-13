package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tobiase/worktree-utils/internal/config"
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

const shellWrapper = `# Shell function to handle CD: and EXEC: prefixes
wt() {
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
		showUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	// Handle command aliases
	aliases := map[string]string{
		"ls":     "list",
		"switch": "go",
		"s":      "go",
	}
	if alias, ok := aliases[cmd]; ok {
		cmd = alias
	}

	// Special handling for setup command (doesn't need config)
	if cmd == "setup" {
		handleSetupCommand(args)
		return
	}

	// Initialize config manager
	configMgr, err := config.NewManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wt: failed to initialize config: %v\n", err)
		os.Exit(1)
	}

	// Load project config for current directory (except for shell-init)
	if cmd != "shell-init" {
		cwd, _ := os.Getwd()
		gitRemote, _ := worktree.GetGitRemote()
		configMgr.LoadProject(cwd, gitRemote)
	}

	switch cmd {
	case "shell-init":
		fmt.Print(shellWrapper)
		return

	case "list":
		if err := worktree.List(); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}

	case "add":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt add <branch>\n")
			os.Exit(1)
		}
		if err := worktree.Add(args[0], configMgr); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}

	case "rm":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt rm <branch>\n")
			os.Exit(1)
		}
		if err := worktree.Remove(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}

	case "go":
		var path string
		var err error
		
		if len(args) < 1 {
			// No arguments - go to repository root
			path, err = worktree.GetRepoRoot()
		} else {
			// Go to specified worktree
			path, err = worktree.Go(args[0])
		}
		
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CD:%s", path)

	case "new":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt new <branch> [--base <branch>]\n")
			os.Exit(1)
		}
		
		branch := args[0]
		baseBranch := ""
		
		// Parse flags
		for i := 1; i < len(args); i++ {
			if args[i] == "--base" && i+1 < len(args) {
				baseBranch = args[i+1]
				i++
			}
		}
		
		path, err := worktree.NewWorktree(branch, baseBranch, configMgr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CD:%s", path)

	case "env-copy":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt env-copy <branch> [--recursive]\n")
			os.Exit(1)
		}
		
		targetBranch := args[0]
		recursive := false
		
		// Check for --recursive flag
		for _, arg := range args[1:] {
			if arg == "--recursive" {
				recursive = true
			}
		}
		
		if err := worktree.CopyEnvFile(targetBranch, recursive); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}

	case "project":
		handleProjectCommand(args, configMgr)

	case "version":
		fmt.Printf("wt version %s\n", version)
		if version != "dev" {
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		}

	case "update":
		handleUpdateCommand(args)

	default:
		// Check if it's a project-specific command
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
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: wt project [init|add|list]\n")
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

	default:
		fmt.Fprintf(os.Stderr, "wt: unknown project subcommand '%s'\n", subcmd)
		os.Exit(1)
	}
}

func handleSetupCommand(args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "--check":
			if err := setup.Check(); err != nil {
				fmt.Fprintf(os.Stderr, "Check failed: %v\n", err)
				os.Exit(1)
			}
		case "--uninstall":
			if err := setup.Uninstall(); err != nil {
				fmt.Fprintf(os.Stderr, "Uninstall failed: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unknown setup option: %s\n", args[0])
			fmt.Fprintf(os.Stderr, "Usage: wt setup [--check|--uninstall]\n")
			os.Exit(1)
		}
	} else {
		// Get current binary path
		binaryPath, err := os.Executable()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get executable path: %v\n", err)
			os.Exit(1)
		}
		
		if err := setup.Setup(binaryPath); err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
	}
}

func showUsage() {
	usage := `Usage: wt <command> [arguments]

Core commands:
  list, ls            List all worktrees
  add <branch>        Add a new worktree
  rm <branch>         Remove a worktree
  go, switch, s       Switch to a worktree (no args = repo root)
  new <branch>        Create and switch to a new worktree
                      Options: --base <branch>

Utility commands:
  env-copy <branch>   Copy .env files to another worktree
                      Options: --recursive
  project init <name> Initialize project configuration

Setup commands:
  setup               Install wt to ~/.local/bin
                      Options: --check, --uninstall
  update              Check and install updates
                      Options: --check, --force

Other commands:
  shell-init          Output shell initialization code
  version             Show version information`

	// Add project-specific commands if available
	if configMgr, err := config.NewManager(); err == nil {
		cwd, _ := os.Getwd()
		gitRemote, _ := worktree.GetGitRemote()
		configMgr.LoadProject(cwd, gitRemote)
		
		if project := configMgr.GetCurrentProject(); project != nil {
			if len(project.Commands) > 0 {
				// Separate commands by type
				var navCommands []string
				var venvCommands []string
				
				for name, cmd := range project.Commands {
					if cmd.Type == "virtualenv" {
						venvCommands = append(venvCommands, name)
					} else {
						navCommands = append(navCommands, name)
					}
				}
				
				// Sort both lists
				sort.Strings(navCommands)
				sort.Strings(venvCommands)
				
				// Show navigation commands
				if len(navCommands) > 0 {
					usage += fmt.Sprintf("\n\nProject '%s' navigation:", project.Name)
					for _, name := range navCommands {
						cmd := project.Commands[name]
						usage += fmt.Sprintf("\n  %-18s %s", name, cmd.Description)
					}
				}
				
				// Show virtualenv commands
				if len(venvCommands) > 0 {
					usage += fmt.Sprintf("\n\nProject '%s' virtualenv:", project.Name)
					for _, name := range venvCommands {
						cmd := project.Commands[name]
						usage += fmt.Sprintf("\n  %-18s %s", name, cmd.Description)
					}
				}
			}
		}
	}

	fmt.Fprintln(os.Stderr, usage)
}

func handleUpdateCommand(args []string) {
	// Parse flags
	checkOnly := false
	force := false
	
	for _, arg := range args {
		switch arg {
		case "--check":
			checkOnly = true
		case "--force":
			force = true
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
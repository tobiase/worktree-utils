package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tobiase/worktree-utils/internal/config"
	"github.com/tobiase/worktree-utils/internal/worktree"
)

const shellWrapper = `# Shell function to handle CD: prefix
wt() {
  output=$("${WT_BIN:-wt-bin}" "$@" 2>&1)
  exit_code=$?
  
  if [ $exit_code -eq 0 ]; then
    if [[ "$output" == "CD:"* ]]; then
      cd "${output#CD:}"
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
		if err := worktree.Add(args[0]); err != nil {
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
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt go <index|branch>\n")
			os.Exit(1)
		}
		path, err := worktree.Go(args[0])
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
		
		path, err := worktree.NewWorktree(branch, baseBranch)
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

	default:
		// Check if it's a project-specific navigation command
		if navCmd, exists := configMgr.GetCommand(cmd); exists {
			handleNavigationCommand(navCmd)
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

func showUsage() {
	usage := `Usage: wt <command> [arguments]

Core commands:
  list                List all worktrees
  add <branch>        Add a new worktree
  rm <branch>         Remove a worktree
  go <index|branch>   Switch to a worktree
  new <branch>        Create and switch to a new worktree
                      Options: --base <branch>

Utility commands:
  env-copy <branch>   Copy .env files to another worktree
                      Options: --recursive
  project init <name> Initialize project configuration

Other commands:
  shell-init          Output shell initialization code`

	// Add project-specific commands if available
	if configMgr, err := config.NewManager(); err == nil {
		cwd, _ := os.Getwd()
		gitRemote, _ := worktree.GetGitRemote()
		configMgr.LoadProject(cwd, gitRemote)
		
		if project := configMgr.GetCurrentProject(); project != nil {
			if len(project.Commands) > 0 {
				usage += fmt.Sprintf("\n\nProject '%s' commands:", project.Name)
				for name, cmd := range project.Commands {
					usage += fmt.Sprintf("\n  %-18s %s", name, cmd.Description)
				}
			}
		}
	}

	fmt.Fprintln(os.Stderr, usage)
}
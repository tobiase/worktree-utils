package completion

import (
	"os/exec"
	"strings"

	"github.com/tobiase/worktree-utils/internal/config"
)

// CompletionData holds all data needed for generating completions
type CompletionData struct {
	Commands          []Command
	Aliases           map[string]string
	AvailableBranches []string
	ProjectCommands   []string
}

// Command represents a command with its flags and description
type Command struct {
	Name        string
	Description string
	Flags       []Flag
	Args        []Argument
}

// Flag represents a command flag
type Flag struct {
	Name        string
	Description string
	HasValue    bool
}

// Argument represents a command argument
type Argument struct {
	Name        string
	Description string
	Type        ArgumentType
}

// ArgumentType defines the type of argument for completion
type ArgumentType int

const (
	ArgString ArgumentType = iota
	ArgBranch
	ArgWorktreeBranch
	ArgProject
	ArgFile
)

// GetCompletionData returns all data needed for generating completions
func GetCompletionData(configMgr *config.Manager) *CompletionData {
	data := &CompletionData{
		Commands: getCoreCommands(),
		Aliases: map[string]string{
			"ls":     "list",
			"switch": "go",
			"s":      "go",
		},
	}

	// Get available branches
	if branches, err := getAvailableBranches(); err == nil {
		data.AvailableBranches = branches
	}

	// Get project commands
	if configMgr != nil {
		if project := configMgr.GetCurrentProject(); project != nil {
			for name := range project.Commands {
				data.ProjectCommands = append(data.ProjectCommands, name)
			}
		}
	}

	return data
}

// getCoreCommands returns all core commands with their metadata
func getCoreCommands() []Command {
	return []Command{
		{
			Name:        "list",
			Description: "List all worktrees",
			Flags:       []Flag{},
			Args:        []Argument{},
		},
		{
			Name:        "add",
			Description: "Add a new worktree",
			Flags:       []Flag{},
			Args: []Argument{
				{Name: "branch", Description: "Branch name", Type: ArgString},
			},
		},
		{
			Name:        "rm",
			Description: "Remove a worktree",
			Flags:       []Flag{},
			Args: []Argument{
				{Name: "branch", Description: "Worktree branch name", Type: ArgWorktreeBranch},
			},
		},
		{
			Name:        "go",
			Description: "Switch to a worktree",
			Flags:       []Flag{},
			Args: []Argument{
				{Name: "branch", Description: "Worktree branch name (optional)", Type: ArgWorktreeBranch},
			},
		},
		{
			Name:        "new",
			Description: "Create and switch to a new worktree",
			Flags: []Flag{
				{Name: "--base", Description: "Base branch", HasValue: true},
				{Name: "--no-switch", Description: "Create without switching", HasValue: false},
			},
			Args: []Argument{
				{Name: "branch", Description: "New branch name", Type: ArgString},
			},
		},
		{
			Name:        "env-copy",
			Description: "Copy .env files to another worktree",
			Flags: []Flag{
				{Name: "--recursive", Description: "Copy recursively", HasValue: false},
			},
			Args: []Argument{
				{Name: "branch", Description: "Target worktree branch", Type: ArgWorktreeBranch},
			},
		},
		{
			Name:        "project",
			Description: "Project configuration commands",
			Flags:       []Flag{},
			Args: []Argument{
				{Name: "subcommand", Description: "Project subcommand", Type: ArgString},
			},
		},
		{
			Name:        "setup",
			Description: "Install wt to ~/.local/bin",
			Flags: []Flag{
				{Name: "--check", Description: "Check installation", HasValue: false},
				{Name: "--uninstall", Description: "Uninstall", HasValue: false},
			},
			Args: []Argument{},
		},
		{
			Name:        "update",
			Description: "Check and install updates",
			Flags: []Flag{
				{Name: "--check", Description: "Check for updates only", HasValue: false},
				{Name: "--force", Description: "Force update", HasValue: false},
			},
			Args: []Argument{},
		},
		{
			Name:        "version",
			Description: "Show version information",
			Flags:       []Flag{},
			Args:        []Argument{},
		},
		{
			Name:        "help",
			Description: "Show help information",
			Flags:       []Flag{},
			Args:        []Argument{},
		},
		{
			Name:        "completion",
			Description: "Generate shell completion scripts",
			Flags:       []Flag{},
			Args: []Argument{
				{Name: "shell", Description: "Shell type (bash|zsh)", Type: ArgString},
			},
		},
		{
			Name:        "shell-init",
			Description: "Output shell initialization code",
			Flags:       []Flag{},
			Args:        []Argument{},
		},
		{
			Name:        "recent",
			Description: "Show and navigate to recently active branches",
			Flags: []Flag{
				{Name: "--all", Description: "Show all branches regardless of author", HasValue: false},
				{Name: "--others", Description: "Show only other users' branches", HasValue: false},
				{Name: "-n", Description: "Number of branches to show", HasValue: true},
			},
			Args: []Argument{
				{Name: "index", Description: "Branch index to navigate to (optional)", Type: ArgString},
			},
		},
	}
}

// getAvailableBranches returns a list of available branches
func getAvailableBranches() ([]string, error) {
	// We need to access the parseWorktrees function from worktree package
	// For now, let's use git directly as a fallback
	return getGitBranches()
}

// getGitBranches gets branches directly from git
func getGitBranches() ([]string, error) {
	// Try to get worktree branches first
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err == nil {
		return parseWorktreeBranches(string(output)), nil
	}

	// Fallback to regular git branches
	cmd = exec.Command("git", "branch", "--format=%(refname:short)")
	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

// parseWorktreeBranches parses git worktree list --porcelain output
func parseWorktreeBranches(output string) []string {
	var branches []string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch ")
			if branch != "" && !strings.HasPrefix(branch, "refs/heads/") {
				branches = append(branches, branch)
			} else if strings.HasPrefix(branch, "refs/heads/") {
				branch = strings.TrimPrefix(branch, "refs/heads/")
				branches = append(branches, branch)
			}
		}
	}

	return branches
}

// getWorktreeBranches returns only branches that have existing worktrees
func getWorktreeBranches() ([]string, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseWorktreeBranches(string(output)), nil
}

// GetCommandByName returns a command by its name or alias
func (d *CompletionData) GetCommandByName(name string) *Command {
	// Check aliases first
	if alias, exists := d.Aliases[name]; exists {
		name = alias
	}

	// Find command
	for _, cmd := range d.Commands {
		if cmd.Name == name {
			return &cmd
		}
	}

	return nil
}

// GetAllCommandNames returns all command names including aliases
func (d *CompletionData) GetAllCommandNames() []string {
	names := make([]string, 0, len(d.Commands)+len(d.Aliases)+len(d.ProjectCommands))

	// Add core commands
	for _, cmd := range d.Commands {
		names = append(names, cmd.Name)
	}

	// Add aliases
	for alias := range d.Aliases {
		names = append(names, alias)
	}

	// Add project commands
	names = append(names, d.ProjectCommands...)

	return names
}

// GetCompletionCandidates returns completion candidates for a given context
func (d *CompletionData) GetCompletionCandidates(args []string, argType ArgumentType) []string {
	switch argType {
	case ArgBranch:
		if d.AvailableBranches == nil {
			return []string{}
		}
		return d.AvailableBranches
	case ArgWorktreeBranch:
		// For worktree branches, we need to get only existing worktree branches
		if worktreeBranches, err := getWorktreeBranches(); err == nil {
			if worktreeBranches == nil {
				return []string{}
			}
			return worktreeBranches
		}
		return []string{}
	case ArgProject:
		if d.ProjectCommands == nil {
			return []string{}
		}
		return d.ProjectCommands
	default:
		return []string{}
	}
}

// NormalizeCommandName resolves aliases to canonical command names
func (d *CompletionData) NormalizeCommandName(name string) string {
	if alias, exists := d.Aliases[name]; exists {
		return alias
	}
	return name
}

// SplitCompletionLine splits a completion line into components
func SplitCompletionLine(line string) []string {
	// Simple split for now - could be enhanced for quoted arguments
	return strings.Fields(line)
}

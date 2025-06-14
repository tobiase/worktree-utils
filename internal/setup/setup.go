package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tobiase/worktree-utils/internal/completion"
	"github.com/tobiase/worktree-utils/internal/config"
)

const initScript = `# worktree-utils shell initialization
if command -v wt-bin &> /dev/null; then
  source <(wt-bin shell-init)

  # Load shell completion if available
  if [[ -n "$BASH_VERSION" ]] && [[ -f ~/.config/wt/completion.bash ]]; then
    source ~/.config/wt/completion.bash
  elif [[ -n "$ZSH_VERSION" ]]; then
    # Add wt completion directory to fpath
    fpath=(~/.config/wt/completions $fpath)

    # Initialize completion system (required after fpath changes)
    autoload -Uz compinit && compinit

    # Load completion if available
    if [[ -f ~/.config/wt/completions/_wt ]]; then
      autoload -Uz _wt
    fi
  fi
fi
`

// CompletionOptions controls which completion scripts to install
type CompletionOptions struct {
	Install bool
	Shell   string // "auto", "bash", "zsh", "none"
}

const (
	shellAuto = "auto"
	shellBash = "bash"
	shellZsh  = "zsh"
)

// detectUserShell determines the user's primary shell
func detectUserShell() string {
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, shellZsh) {
		return shellZsh
	} else if strings.Contains(shell, shellBash) {
		return shellBash
	}
	// Default to bash if unable to detect
	return shellBash
}

// generateCompletionFiles creates completion script files for all shells in the config directory
func generateCompletionFiles(configDir string) error {
	return generateCompletionFilesForShells(configDir, []string{"bash", "zsh"})
}

// generateCompletionFilesForShells creates completion script files for specified shells
func generateCompletionFilesForShells(configDir string, shells []string) error {
	// Create config manager for completion generation
	configMgr, err := config.NewManager()
	if err != nil {
		// If config fails, create completion without project integration
		configMgr = nil
	}

	for _, shell := range shells {
		switch shell {
		case "bash":
			bashCompletion := completion.GenerateBashCompletion(configMgr)
			bashPath := filepath.Join(configDir, "completion.bash")
			if err := os.WriteFile(bashPath, []byte(bashCompletion), 0644); err != nil {
				return fmt.Errorf("failed to write bash completion: %v", err)
			}

		case "zsh":
			zshCompletion := completion.GenerateZshCompletion(configMgr)

			// Create zsh completions directory
			zshCompletionDir := filepath.Join(configDir, "completions")
			if err := os.MkdirAll(zshCompletionDir, 0755); err != nil {
				return fmt.Errorf("failed to create zsh completion directory: %v", err)
			}

			// Write zsh completion with proper name
			zshPath := filepath.Join(zshCompletionDir, "_wt")
			if err := os.WriteFile(zshPath, []byte(zshCompletion), 0644); err != nil {
				return fmt.Errorf("failed to write zsh completion: %v", err)
			}
		}
	}

	return nil
}

// installCompletion handles completion installation based on options
func installCompletion(configDir string, opts CompletionOptions) error {
	if !opts.Install {
		return nil
	}

	if opts.Shell == "none" {
		return nil
	}

	// Get shell config files to modify
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	var configFiles []string
	var completionLines []string

	// Determine which shells to install for and their config files
	if opts.Shell == shellAuto || opts.Shell == shellZsh {
		zshrc := filepath.Join(homeDir, ".zshrc")
		if _, err := os.Stat(zshrc); err == nil {
			configFiles = append(configFiles, zshrc)
			completionLines = append(completionLines, "source <(wt-bin completion zsh)")
		}
	}

	if opts.Shell == shellAuto || opts.Shell == shellBash {
		bashrc := filepath.Join(homeDir, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			configFiles = append(configFiles, bashrc)
			completionLines = append(completionLines, "source <(wt-bin completion bash)")
		}
	}

	if opts.Shell != "auto" && opts.Shell != shellBash && opts.Shell != shellZsh {
		return fmt.Errorf("unsupported shell: %s", opts.Shell)
	}

	// Add completion lines to shell configs
	var installedShells []string
	for i, configFile := range configFiles {
		completionLine := completionLines[i]

		// Check if completion is already configured
		if hasCompletionConfig(configFile) {
			continue
		}

		// Add completion line
		if err := addToShellConfig(configFile, completionLine); err != nil {
			return fmt.Errorf("failed to add completion to %s: %v", configFile, err)
		}

		if strings.Contains(configFile, ".zshrc") {
			installedShells = append(installedShells, "zsh")
		} else if strings.Contains(configFile, ".bashrc") {
			installedShells = append(installedShells, "bash")
		}
	}

	if len(installedShells) > 0 {
		fmt.Printf("✓ Installed completion for: %s\n", strings.Join(installedShells, ", "))
	}

	return nil
}

// hasCompletionConfig checks if shell config already has wt completion
func hasCompletionConfig(configFile string) bool {
	file, err := os.Open(configFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "wt-bin completion") ||
			strings.Contains(line, "source <(wt completion") {
			return true
		}
	}

	return false
}

// Setup installs wt to the user's system with default completion options
func Setup(currentBinaryPath string) error {
	return SetupWithOptions(currentBinaryPath, CompletionOptions{
		Install: true,
		Shell:   "auto",
	})
}

// SetupWithOptions installs wt to the user's system with specified completion options
func SetupWithOptions(currentBinaryPath string, completionOpts CompletionOptions) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Ensure directories exist
	binDir := filepath.Join(homeDir, ".local", "bin")
	configDir := filepath.Join(homeDir, ".config", "wt")

	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %v", err)
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Copy binary to ~/.local/bin/wt-bin
	targetBinary := filepath.Join(binDir, "wt-bin")
	if err := copyBinary(currentBinaryPath, targetBinary); err != nil {
		return fmt.Errorf("failed to copy binary: %v", err)
	}

	// Add shell function to shell configs
	shellConfigs := detectShellConfigs(homeDir)
	if len(shellConfigs) == 0 {
		return fmt.Errorf("no shell configuration files found")
	}

	shellFunctionLine := "source <(wt-bin shell-init)"

	for _, configFile := range shellConfigs {
		// Check if shell function already configured
		if hasWtInit(configFile) {
			fmt.Printf("✓ %s already configured\n", configFile)
			continue
		}

		if err := addToShellConfig(configFile, shellFunctionLine); err != nil {
			fmt.Printf("Warning: failed to update %s: %v\n", configFile, err)
		} else {
			fmt.Printf("✓ Updated %s\n", configFile)
		}
	}

	// Install completion via process substitution
	if err := installCompletion(configDir, completionOpts); err != nil {
		fmt.Printf("Warning: failed to install completion: %v\n", err)
	}

	// Validate and repair installation
	if err := validateAndRepairInstallation(targetBinary, configDir, completionOpts); err != nil {
		fmt.Printf("Warning: installation validation failed: %v\n", err)
	}

	// Check if ~/.local/bin is in PATH
	if !isInPath(binDir) {
		fmt.Printf("\nWarning: %s is not in your PATH\n", binDir)
		fmt.Printf("Add this to your shell config:\n")
		fmt.Printf("  export PATH=\"$PATH:%s\"\n", binDir)
	}

	fmt.Printf("\n✓ Setup complete!\n")
	fmt.Printf("\nTo start using wt:\n")
	fmt.Printf("  1. Restart your shell or run: source ~/.config/wt/init.sh\n")
	fmt.Printf("  2. Run: wt list\n")

	return nil
}

// Uninstall removes wt from the system
func Uninstall() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	// Remove binary
	binPath := filepath.Join(homeDir, ".local", "bin", "wt-bin")
	if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove binary: %v\n", err)
	} else {
		fmt.Printf("✓ Removed binary\n")
	}

	// Remove config directory
	configDir := filepath.Join(homeDir, ".config", "wt")
	if err := os.RemoveAll(configDir); err != nil {
		fmt.Printf("Warning: failed to remove config directory: %v\n", err)
	} else {
		fmt.Printf("✓ Removed config directory\n")
	}

	fmt.Printf("\n✓ Uninstall complete!\n")
	fmt.Printf("\nPlease manually remove the wt initialization line from your shell config files:\n")
	fmt.Printf("  ~/.bashrc, ~/.zshrc, etc.\n")

	return nil
}

// Check verifies the installation
func Check() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	fmt.Println("Checking wt installation...")

	// Check binary
	binPath := filepath.Join(homeDir, ".local", "bin", "wt-bin")
	if _, err := os.Stat(binPath); err != nil {
		fmt.Printf("✗ Binary not found at %s\n", binPath)
	} else {
		fmt.Printf("✓ Binary found at %s\n", binPath)

		// Check if executable
		if err := exec.Command(binPath, "shell-init").Run(); err != nil {
			fmt.Printf("✗ Binary is not executable\n")
		} else {
			fmt.Printf("✓ Binary is executable\n")
		}
	}

	// Check config directory
	configDir := filepath.Join(homeDir, ".config", "wt")
	if _, err := os.Stat(configDir); err != nil {
		fmt.Printf("✗ Config directory not found at %s\n", configDir)
	} else {
		fmt.Printf("✓ Config directory found at %s\n", configDir)
	}

	// Check init script
	initPath := filepath.Join(configDir, "init.sh")
	if _, err := os.Stat(initPath); err != nil {
		fmt.Printf("✗ Init script not found at %s\n", initPath)
	} else {
		fmt.Printf("✓ Init script found at %s\n", initPath)
	}

	// Check completion files
	bashCompletionPath := filepath.Join(configDir, "completion.bash")
	if _, err := os.Stat(bashCompletionPath); err != nil {
		fmt.Printf("✗ Bash completion not found at %s\n", bashCompletionPath)
	} else {
		fmt.Printf("✓ Bash completion found at %s\n", bashCompletionPath)
	}

	zshCompletionPath := filepath.Join(configDir, "completions", "_wt")
	if _, err := os.Stat(zshCompletionPath); err != nil {
		fmt.Printf("✗ Zsh completion not found at %s\n", zshCompletionPath)
	} else {
		fmt.Printf("✓ Zsh completion found at %s\n", zshCompletionPath)
	}

	// Check PATH
	binDir := filepath.Join(homeDir, ".local", "bin")
	if !isInPath(binDir) {
		fmt.Printf("✗ %s is not in PATH\n", binDir)
	} else {
		fmt.Printf("✓ %s is in PATH\n", binDir)
	}

	// Check shell configs
	shellConfigs := detectShellConfigs(homeDir)
	for _, config := range shellConfigs {
		if hasWtInit(config) {
			fmt.Printf("✓ wt configured in %s\n", config)
		} else {
			fmt.Printf("✗ wt not configured in %s\n", config)
		}
	}

	return nil
}

// copyBinary copies the current binary to the target location
func copyBinary(source, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	// Copy content
	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return err
	}

	// Make executable
	return os.Chmod(target, 0755)
}

// detectShellConfigs finds shell configuration files
func detectShellConfigs(homeDir string) []string {
	var configs []string

	// Check for common shell configs
	possibleConfigs := []string{
		".bashrc",
		".zshrc",
		".bash_profile",
		".profile",
	}

	// Detect current shell
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		// Prioritize zsh config
		if _, err := os.Stat(filepath.Join(homeDir, ".zshrc")); err == nil {
			configs = append(configs, filepath.Join(homeDir, ".zshrc"))
		}
	} else if strings.Contains(shell, "bash") {
		// Prioritize bash config
		if _, err := os.Stat(filepath.Join(homeDir, ".bashrc")); err == nil {
			configs = append(configs, filepath.Join(homeDir, ".bashrc"))
		}
	}

	// Add other existing configs
	for _, config := range possibleConfigs {
		configPath := filepath.Join(homeDir, config)
		if _, err := os.Stat(configPath); err == nil {
			// Avoid duplicates
			found := false
			for _, existing := range configs {
				if existing == configPath {
					found = true
					break
				}
			}
			if !found {
				configs = append(configs, configPath)
			}
		}
	}

	return configs
}

// addToShellConfig adds a line to shell config if not already present
func addToShellConfig(configFile, line string) error {
	// Check if already configured
	if hasWtInit(configFile) {
		return nil // Already configured
	}

	// Append to file
	file, err := os.OpenFile(configFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Add newline before our config
	if _, err := file.WriteString("\n# worktree-utils\n"); err != nil {
		return err
	}

	if _, err := file.WriteString(line + "\n"); err != nil {
		return err
	}

	return nil
}

// hasWtInit checks if shell config already has wt initialization
func hasWtInit(configFile string) bool {
	file, err := os.Open(configFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "worktree-utils") ||
			strings.Contains(line, "~/.config/wt/init.sh") ||
			strings.Contains(line, "wt-bin shell-init") {
			return true
		}
	}

	return false
}

// isInPath checks if a directory is in the PATH
func isInPath(dir string) bool {
	pathEnv := os.Getenv("PATH")
	paths := filepath.SplitList(pathEnv)

	for _, p := range paths {
		if p == dir {
			return true
		}
	}

	return false
}

// validateAndRepairInstallation checks the installation and fixes common issues
func validateAndRepairInstallation(binaryPath, configDir string, completionOpts CompletionOptions) error {
	var issues []string
	var repaired []string

	// Check binary issues
	binaryIssues, binaryRepairs := checkBinary(binaryPath)
	issues = append(issues, binaryIssues...)
	repaired = append(repaired, binaryRepairs...)

	// Check configuration files
	configIssues, configRepairs := checkConfigFiles(configDir, completionOpts)
	issues = append(issues, configIssues...)
	repaired = append(repaired, configRepairs...)

	// Check shell configurations
	shellIssues := checkShellConfigs()
	issues = append(issues, shellIssues...)

	// Report results
	reportValidationResults(issues, repaired)
	return nil
}

// checkBinary validates the binary installation
func checkBinary(binaryPath string) ([]string, []string) {
	var issues []string
	var repaired []string

	// Check if binary exists and is not empty
	if stat, err := os.Stat(binaryPath); err != nil {
		issues = append(issues, "binary not found")
	} else if stat.Size() == 0 {
		issues = append(issues, "binary is empty")
	}

	// Check if binary is executable
	if _, err := exec.LookPath(binaryPath); err != nil {
		issues = append(issues, "binary not executable")
		// Try to fix permissions
		if err := os.Chmod(binaryPath, 0755); err == nil {
			repaired = append(repaired, "fixed binary permissions")
		}
	}

	// Test if binary works
	cmd := exec.Command(binaryPath, "version")
	if err := cmd.Run(); err != nil {
		issues = append(issues, "binary does not run correctly")
	}

	return issues, repaired
}

// checkConfigFiles validates configuration files
func checkConfigFiles(configDir string, completionOpts CompletionOptions) ([]string, []string) {
	var issues []string
	var repaired []string

	// Check init script
	initIssues, initRepairs := checkInitScript(configDir)
	issues = append(issues, initIssues...)
	repaired = append(repaired, initRepairs...)

	// Check completion files if enabled
	if completionOpts.Install {
		completionIssues, completionRepairs := checkCompletionFiles(configDir, completionOpts)
		issues = append(issues, completionIssues...)
		repaired = append(repaired, completionRepairs...)
	}

	return issues, repaired
}

// checkInitScript validates and repairs the init script
func checkInitScript(configDir string) ([]string, []string) {
	var issues []string
	var repaired []string

	initPath := filepath.Join(configDir, "init.sh")
	if _, err := os.Stat(initPath); err != nil {
		issues = append(issues, "init script missing")
		// Regenerate init script
		if err := os.WriteFile(initPath, []byte(initScript), 0644); err == nil {
			repaired = append(repaired, "regenerated init script")
		}
	}

	return issues, repaired
}

// checkCompletionFiles validates and repairs completion files
func checkCompletionFiles(configDir string, completionOpts CompletionOptions) ([]string, []string) {
	var issues []string
	var repaired []string

	// Check completion files exist and are not corrupted
	bashCompletionPath := filepath.Join(configDir, "completion.bash")
	zshCompletionPath := filepath.Join(configDir, "completions", "_wt")

	completionIssues := false
	if stat, err := os.Stat(bashCompletionPath); err != nil || stat.Size() < 100 {
		completionIssues = true
	}
	if stat, err := os.Stat(zshCompletionPath); err != nil || stat.Size() < 100 {
		completionIssues = true
	}

	if completionIssues {
		issues = append(issues, "completion files missing or corrupted")

		// Determine shells to regenerate for
		shells := getShellsForCompletion(completionOpts)
		if len(shells) > 0 {
			if err := generateCompletionFilesForShells(configDir, shells); err == nil {
				repaired = append(repaired, "regenerated completion files")
			}
		}
	}

	return issues, repaired
}

// getShellsForCompletion determines which shells to generate completion for
func getShellsForCompletion(completionOpts CompletionOptions) []string {
	var shells []string

	if completionOpts.Shell == "auto" {
		// Auto means install for all supported shells
		shells = []string{shellBash, shellZsh}
	} else if completionOpts.Shell == shellBash || completionOpts.Shell == shellZsh {
		shells = append(shells, completionOpts.Shell)
	}

	return shells
}

// checkShellConfigs validates shell configuration files
func checkShellConfigs() []string {
	var issues []string

	homeDir, _ := os.UserHomeDir()
	shellConfigs := detectShellConfigs(homeDir)
	for _, configFile := range shellConfigs {
		if hasOldWtConfig(configFile) {
			issues = append(issues, fmt.Sprintf("old wt configuration in %s", configFile))
		}
	}

	return issues
}

// reportValidationResults prints validation and repair results
func reportValidationResults(issues, repaired []string) {
	if len(issues) > 0 {
		fmt.Printf("Installation issues detected:\n")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	if len(repaired) > 0 {
		fmt.Printf("Automatically repaired:\n")
		for _, repair := range repaired {
			fmt.Printf("  ✓ %s\n", repair)
		}
	}

	if len(issues) == 0 {
		fmt.Printf("✓ Installation validated successfully\n")
	}
}

// hasOldWtConfig checks if a shell config has old/problematic wt configuration
func hasOldWtConfig(configFile string) bool {
	file, err := os.Open(configFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Look for old shell function definitions or problematic patterns
		if strings.Contains(line, "wt() {") ||
			strings.Contains(line, "function wt") ||
			strings.Contains(line, "_arguments:comparguments") {
			return true
		}
	}

	return false
}

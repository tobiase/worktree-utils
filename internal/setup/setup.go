package setup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const initScript = `# worktree-utils shell initialization
if command -v wt-bin &> /dev/null; then
  source <(wt-bin shell-init)
fi
`

// Setup installs wt to the user's system
func Setup(currentBinaryPath string) error {
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

	// Create init script
	initScriptPath := filepath.Join(configDir, "init.sh")
	if err := os.WriteFile(initScriptPath, []byte(initScript), 0644); err != nil {
		return fmt.Errorf("failed to create init script: %v", err)
	}

	// Add to shell configs
	shellConfigs := detectShellConfigs(homeDir)
	if len(shellConfigs) == 0 {
		return fmt.Errorf("no shell configuration files found")
	}

	initLine := fmt.Sprintf("[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh")
	
	for _, configFile := range shellConfigs {
		if err := addToShellConfig(configFile, initLine); err != nil {
			fmt.Printf("Warning: failed to update %s: %v\n", configFile, err)
		} else {
			fmt.Printf("✓ Updated %s\n", configFile)
		}
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
	if err := os.Chmod(target, 0755); err != nil {
		return err
	}

	return nil
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
package setup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Test helpers for mocking file system operations
type testEnv struct {
	homeDir     string
	shell       string
	path        string
	files       map[string][]byte
	dirs        []string
	executables map[string]bool
	execErrors  map[string]error
}

func newTestEnv(t *testing.T) *testEnv {
	tempDir := t.TempDir()
	return &testEnv{
		homeDir:     tempDir,
		shell:       "/bin/bash",
		path:        "/usr/bin:/usr/local/bin",
		files:       make(map[string][]byte),
		dirs:        []string{},
		executables: make(map[string]bool),
		execErrors:  make(map[string]error),
	}
}

func (e *testEnv) setup(t *testing.T) func() {
	// Save original env
	origHome := os.Getenv("HOME")
	origShell := os.Getenv("SHELL")
	origPath := os.Getenv("PATH")

	// Set test env
	os.Setenv("HOME", e.homeDir)
	os.Setenv("SHELL", e.shell)
	os.Setenv("PATH", e.path)

	// Create directories
	for _, dir := range e.dirs {
		fullPath := filepath.Join(e.homeDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create files
	for path, content := range e.files {
		fullPath := filepath.Join(e.homeDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Return cleanup function
	return func() {
		os.Setenv("HOME", origHome)
		os.Setenv("SHELL", origShell)
		os.Setenv("PATH", origPath)
	}
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name        string
		env         *testEnv
		binaryPath  string
		wantError   bool
		errorMsg    string
		checkResult func(t *testing.T, env *testEnv)
	}{
		{
			name: "successful installation with bash",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.files[".bashrc"] = []byte("# My bashrc\n")
				// Create a mock binary to copy
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				return env
			}(),
			binaryPath: "mock-wt-bin",
			wantError:  false,
			checkResult: func(t *testing.T, env *testEnv) {
				// Check binary was copied
				binPath := filepath.Join(env.homeDir, ".local", "bin", "wt-bin")
				if _, err := os.Stat(binPath); err != nil {
					t.Errorf("Binary not found at %s", binPath)
				}

				// Check binary is executable
				info, _ := os.Stat(binPath)
				if info.Mode()&0111 == 0 {
					t.Error("Binary is not executable")
				}

				// Check init script created
				initPath := filepath.Join(env.homeDir, ".config", "wt", "init.sh")
				if _, err := os.Stat(initPath); err != nil {
					t.Errorf("Init script not found at %s", initPath)
				}

				// Check bashrc was updated
				bashrcContent, _ := os.ReadFile(filepath.Join(env.homeDir, ".bashrc"))
				if !strings.Contains(string(bashrcContent), "worktree-utils") {
					t.Error("Bashrc was not updated with wt initialization")
				}
			},
		},
		{
			name: "successful installation with zsh",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.shell = "/bin/zsh"
				env.files[".zshrc"] = []byte("# My zshrc\n")
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				return env
			}(),
			binaryPath: "mock-wt-bin",
			wantError:  false,
			checkResult: func(t *testing.T, env *testEnv) {
				// Check zshrc was updated
				zshrcContent, _ := os.ReadFile(filepath.Join(env.homeDir, ".zshrc"))
				if !strings.Contains(string(zshrcContent), "worktree-utils") {
					t.Error("Zshrc was not updated with wt initialization")
				}
			},
		},
		{
			name: "installation with existing configuration",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.files[".bashrc"] = []byte("# My bashrc\n# worktree-utils\n[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh\n")
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				return env
			}(),
			binaryPath: "mock-wt-bin",
			wantError:  false,
			checkResult: func(t *testing.T, env *testEnv) {
				// Check bashrc wasn't duplicated
				bashrcContent, _ := os.ReadFile(filepath.Join(env.homeDir, ".bashrc"))
				count := strings.Count(string(bashrcContent), "worktree-utils")
				if count != 1 {
					t.Errorf("Expected 1 occurrence of worktree-utils, got %d", count)
				}
			},
		},
		{
			name: "installation with PATH warning",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.path = "/usr/bin:/usr/local/bin" // Does not include ~/.local/bin
				env.files[".bashrc"] = []byte("# My bashrc\n")
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				return env
			}(),
			binaryPath: "mock-wt-bin",
			wantError:  false,
			checkResult: func(t *testing.T, env *testEnv) {
				// Just verify installation succeeded
				binPath := filepath.Join(env.homeDir, ".local", "bin", "wt-bin")
				if _, err := os.Stat(binPath); err != nil {
					t.Errorf("Binary not found despite PATH warning")
				}
			},
		},
		{
			name: "no shell configs found",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				// No shell config files
				return env
			}(),
			binaryPath: "mock-wt-bin",
			wantError:  true,
			errorMsg:   "no shell configuration files found",
		},
		{
			name: "binary not found",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.files[".bashrc"] = []byte("# My bashrc\n")
				return env
			}(),
			binaryPath: "non-existent-binary",
			wantError:  true,
			errorMsg:   "failed to copy binary",
		},
		{
			name: "multiple shell configs",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.files[".bashrc"] = []byte("# bashrc\n")
				env.files[".zshrc"] = []byte("# zshrc\n")
				env.files[".profile"] = []byte("# profile\n")
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				return env
			}(),
			binaryPath: "mock-wt-bin",
			wantError:  false,
			checkResult: func(t *testing.T, env *testEnv) {
				// Check all configs were updated
				for _, config := range []string{".bashrc", ".zshrc", ".profile"} {
					content, _ := os.ReadFile(filepath.Join(env.homeDir, config))
					if !strings.Contains(string(content), "worktree-utils") {
						t.Errorf("%s was not updated", config)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.env.setup(t)
			defer cleanup()

			// Adjust binary path to be relative to test home
			if tt.binaryPath != "" && !filepath.IsAbs(tt.binaryPath) {
				tt.binaryPath = filepath.Join(tt.env.homeDir, tt.binaryPath)
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := Setup(tt.binaryPath)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check error
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Check output contains expected messages
			if !tt.wantError {
				if !strings.Contains(output, "Setup complete") {
					t.Error("Output missing 'Setup complete' message")
				}
			}

			// Run custom checks
			if tt.checkResult != nil {
				tt.checkResult(t, tt.env)
			}
		})
	}
}

func TestUninstall(t *testing.T) {
	tests := []struct {
		name        string
		env         *testEnv
		setupFirst  bool
		checkResult func(t *testing.T, env *testEnv)
	}{
		{
			name: "uninstall after setup",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.files[".bashrc"] = []byte("# My bashrc\n")
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				return env
			}(),
			setupFirst: true,
			checkResult: func(t *testing.T, env *testEnv) {
				// Check binary was removed
				binPath := filepath.Join(env.homeDir, ".local", "bin", "wt-bin")
				if _, err := os.Stat(binPath); !os.IsNotExist(err) {
					t.Error("Binary still exists after uninstall")
				}

				// Check config dir was removed
				configDir := filepath.Join(env.homeDir, ".config", "wt")
				if _, err := os.Stat(configDir); !os.IsNotExist(err) {
					t.Error("Config directory still exists after uninstall")
				}
			},
		},
		{
			name: "uninstall with no installation",
			env:  newTestEnv(t),
			checkResult: func(t *testing.T, env *testEnv) {
				// Should complete without errors
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.env.setup(t)
			defer cleanup()

			// Setup first if requested
			if tt.setupFirst {
				binPath := filepath.Join(tt.env.homeDir, "mock-wt-bin")
				if err := Setup(binPath); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := Uninstall()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output
			if !strings.Contains(output, "Uninstall complete") {
				t.Error("Output missing 'Uninstall complete' message")
			}

			// Run custom checks
			if tt.checkResult != nil {
				tt.checkResult(t, tt.env)
			}
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name       string
		env        *testEnv
		setupFirst bool
		wantChecks map[string]bool // Expected check results
	}{
		{
			name: "check complete installation",
			env: func() *testEnv {
				env := newTestEnv(t)
				env.path = fmt.Sprintf("%s:%s", env.path, filepath.Join(env.homeDir, ".local", "bin"))
				env.files[".bashrc"] = []byte("# My bashrc\n")
				env.files["mock-wt-bin"] = []byte("#!/bin/sh\necho mock")
				return env
			}(),
			setupFirst: true,
			wantChecks: map[string]bool{
				"Binary found":       true,
				"Binary is executable": true,
				"Config directory found": true,
				"Init script found":  true,
				"is in PATH":        true,
				"wt configured":     true,
			},
		},
		{
			name: "check no installation",
			env:  newTestEnv(t),
			wantChecks: map[string]bool{
				"Binary not found":       true,
				"Config directory not found": true,
				"Init script not found":  true,
				"is not in PATH":        true,
			},
		},
		{
			name: "check partial installation",
			env: func() *testEnv {
				env := newTestEnv(t)
				// Create only binary, no config
				env.dirs = append(env.dirs, ".local/bin")
				env.files[".local/bin/wt-bin"] = []byte("#!/bin/sh\necho mock")
				env.files[".bashrc"] = []byte("# My bashrc\n")
				return env
			}(),
			wantChecks: map[string]bool{
				"Binary found":       true,
				"Config directory not found": true,
				"Init script not found":  true,
				"is not in PATH":    true,
				"wt not configured": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.env.setup(t)
			defer cleanup()

			// Setup first if requested
			if tt.setupFirst {
				binPath := filepath.Join(tt.env.homeDir, "mock-wt-bin")
				if err := Setup(binPath); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			// Make binary executable if it exists
			binPath := filepath.Join(tt.env.homeDir, ".local", "bin", "wt-bin")
			if _, err := os.Stat(binPath); err == nil {
				os.Chmod(binPath, 0755)
			}

			// Capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := Check()

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check output contains expected checks
			for check, shouldFind := range tt.wantChecks {
				if shouldFind && !strings.Contains(output, check) {
					t.Errorf("Expected to find %q in output", check)
				}
			}
		})
	}
}

func TestCopyBinary(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (source, target string)
		wantError bool
		errorMsg  string
	}{
		{
			name: "successful copy",
			setup: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				source := filepath.Join(tempDir, "source")
				target := filepath.Join(tempDir, "target")
				
				// Create source file
				content := []byte("#!/bin/sh\necho test")
				if err := os.WriteFile(source, content, 0644); err != nil {
					t.Fatal(err)
				}
				
				return source, target
			},
		},
		{
			name: "source not found",
			setup: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				return filepath.Join(tempDir, "nonexistent"), filepath.Join(tempDir, "target")
			},
			wantError: true,
		},
		{
			name: "target directory doesn't exist",
			setup: func(t *testing.T) (string, string) {
				tempDir := t.TempDir()
				source := filepath.Join(tempDir, "source")
				target := filepath.Join(tempDir, "nonexistent", "dir", "target")
				
				// Create source file
				if err := os.WriteFile(source, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
				
				return source, target
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, target := tt.setup(t)
			
			err := copyBinary(source, target)
			
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				
				// Check target exists and is executable
				info, err := os.Stat(target)
				if err != nil {
					t.Errorf("Target file not created: %v", err)
				} else if info.Mode()&0111 == 0 {
					t.Error("Target file is not executable")
				}
				
				// Check content matches
				sourceContent, _ := os.ReadFile(source)
				targetContent, _ := os.ReadFile(target)
				if !bytes.Equal(sourceContent, targetContent) {
					t.Error("Target content doesn't match source")
				}
			}
		})
	}
}

func TestDetectShellConfigs(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		files    map[string]bool
		expected []string
	}{
		{
			name:  "bash shell with bashrc",
			shell: "/bin/bash",
			files: map[string]bool{
				".bashrc": true,
				".zshrc":  true,
			},
			expected: []string{".bashrc", ".zshrc"},
		},
		{
			name:  "zsh shell with zshrc",
			shell: "/usr/local/bin/zsh",
			files: map[string]bool{
				".bashrc": true,
				".zshrc":  true,
			},
			expected: []string{".zshrc", ".bashrc"},
		},
		{
			name:  "all config files present",
			shell: "/bin/sh",
			files: map[string]bool{
				".bashrc":       true,
				".zshrc":        true,
				".bash_profile": true,
				".profile":      true,
			},
			expected: []string{".bashrc", ".zshrc", ".bash_profile", ".profile"},
		},
		{
			name:     "no config files",
			shell:    "/bin/bash",
			files:    map[string]bool{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			
			// Create files
			for file := range tt.files {
				path := filepath.Join(tempDir, file)
				if err := os.WriteFile(path, []byte("# config"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			
			// Set shell env
			oldShell := os.Getenv("SHELL")
			os.Setenv("SHELL", tt.shell)
			defer os.Setenv("SHELL", oldShell)
			
			configs := detectShellConfigs(tempDir)
			
			// Check we got expected number of configs
			if len(configs) != len(tt.expected) {
				t.Errorf("Expected %d configs, got %d", len(tt.expected), len(configs))
			}
			
			// For zsh, check prioritization
			if strings.Contains(tt.shell, "zsh") && len(configs) > 0 {
				if !strings.HasSuffix(configs[0], ".zshrc") {
					t.Error("Expected .zshrc to be prioritized for zsh shell")
				}
			}
			
			// For bash, check prioritization
			if strings.Contains(tt.shell, "bash") && len(configs) > 0 {
				if !strings.HasSuffix(configs[0], ".bashrc") {
					t.Error("Expected .bashrc to be prioritized for bash shell")
				}
			}
		})
	}
}

func TestAddToShellConfig(t *testing.T) {
	tests := []struct {
		name         string
		initialContent string
		line         string
		wantError    bool
		checkContent func(t *testing.T, content string)
	}{
		{
			name:         "add to empty file",
			initialContent: "",
			line:         "[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh",
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "worktree-utils") {
					t.Error("Missing worktree-utils comment")
				}
				if !strings.Contains(content, "~/.config/wt/init.sh") {
					t.Error("Missing init line")
				}
			},
		},
		{
			name:         "add to existing file",
			initialContent: "# My bashrc\nexport PATH=$PATH:/usr/local/bin\n",
			line:         "[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh",
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "My bashrc") {
					t.Error("Original content was lost")
				}
				if !strings.Contains(content, "worktree-utils") {
					t.Error("Missing worktree-utils comment")
				}
			},
		},
		{
			name:         "already configured",
			initialContent: "# My bashrc\n# worktree-utils\n[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh\n",
			line:         "[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh",
			checkContent: func(t *testing.T, content string) {
				// Should not duplicate
				count := strings.Count(content, "worktree-utils")
				if count != 1 {
					t.Errorf("Expected 1 occurrence of worktree-utils, got %d", count)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := filepath.Join(t.TempDir(), ".bashrc")
			
			// Create initial file
			if err := os.WriteFile(tempFile, []byte(tt.initialContent), 0644); err != nil {
				t.Fatal(err)
			}
			
			err := addToShellConfig(tempFile, tt.line)
			
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				
				// Check final content
				content, err := os.ReadFile(tempFile)
				if err != nil {
					t.Fatal(err)
				}
				
				tt.checkContent(t, string(content))
			}
		})
	}
}

func TestHasWtInit(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "has worktree-utils comment",
			content:  "# worktree-utils\nsome config",
			expected: true,
		},
		{
			name:     "has init.sh source",
			content:  "[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh",
			expected: true,
		},
		{
			name:     "has shell-init",
			content:  "source <(wt-bin shell-init)",
			expected: true,
		},
		{
			name:     "no wt config",
			content:  "# My bashrc\nexport PATH=$PATH:/usr/local/bin",
			expected: false,
		},
		{
			name:     "empty file",
			content:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := filepath.Join(t.TempDir(), ".bashrc")
			
			if err := os.WriteFile(tempFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			
			result := hasWtInit(tempFile)
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsInPath(t *testing.T) {
	tests := []struct {
		name     string
		pathEnv  string
		dir      string
		expected bool
	}{
		{
			name:     "directory in PATH",
			pathEnv:  "/usr/bin:/usr/local/bin:/home/user/.local/bin",
			dir:      "/home/user/.local/bin",
			expected: true,
		},
		{
			name:     "directory not in PATH",
			pathEnv:  "/usr/bin:/usr/local/bin",
			dir:      "/home/user/.local/bin",
			expected: false,
		},
		{
			name:     "empty PATH",
			pathEnv:  "",
			dir:      "/home/user/.local/bin",
			expected: false,
		},
		{
			name:     "PATH with trailing colon",
			pathEnv:  "/usr/bin:/usr/local/bin:",
			dir:      "/usr/local/bin",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldPath := os.Getenv("PATH")
			os.Setenv("PATH", tt.pathEnv)
			defer os.Setenv("PATH", oldPath)
			
			result := isInPath(tt.dir)
			
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Test for exec.Command mocking in Check function
func TestCheckBinaryExecutable(t *testing.T) {
	// This test verifies the binary executable check
	// We'll create a real binary that can be executed
	tempDir := t.TempDir()
	binDir := filepath.Join(tempDir, ".local", "bin")
	os.MkdirAll(binDir, 0755)
	
	// Create a simple executable script
	binPath := filepath.Join(binDir, "wt-bin")
	script := `#!/bin/sh
if [ "$1" = "shell-init" ]; then
    echo "shell wrapper"
    exit 0
fi
exit 1`
	
	if err := os.WriteFile(binPath, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	
	// Test that our mock executable works
	cmd := exec.Command(binPath, "shell-init")
	if err := cmd.Run(); err != nil {
		t.Errorf("Mock binary failed to execute: %v", err)
	}
}
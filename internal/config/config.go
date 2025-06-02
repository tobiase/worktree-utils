package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// NavigationCommand represents a simple directory navigation command
type NavigationCommand struct {
	Description string `yaml:"description"`
	Target      string `yaml:"target"`
	Type        string `yaml:"type,omitempty"`
}

// ProjectConfig represents configuration for a specific project
type ProjectConfig struct {
	Name     string                       `yaml:"name"`
	Match    ProjectMatch                 `yaml:"match"`
	Commands map[string]NavigationCommand `yaml:"commands"`
	Settings ProjectSettings              `yaml:"settings"`
}

// ProjectMatch defines how to match a project
type ProjectMatch struct {
	Paths   []string `yaml:"paths"`
	Remotes []string `yaml:"remotes"`
}

// ProjectSettings contains project-specific settings
type ProjectSettings struct {
	WorktreeBase string `yaml:"worktree_base"`
}

// Config represents the global wt configuration
type Config struct {
	Projects map[string]string `yaml:"projects"` // name -> path to project config
}

// Manager handles configuration loading and project detection
type Manager struct {
	configDir     string
	globalConfig  *Config
	currentProject *ProjectConfig
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(homeDir, ".config", "wt")
	return &Manager{
		configDir: configDir,
	}, nil
}

// LoadProject loads configuration for the current directory
func (m *Manager) LoadProject(currentPath string, gitRemote string) error {
	// Ensure config directory exists
	projectsDir := filepath.Join(m.configDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Load all project configs and find a match
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No projects configured yet
		}
		return err
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		configPath := filepath.Join(projectsDir, entry.Name())
		project, err := m.loadProjectConfig(configPath)
		if err != nil {
			continue // Skip invalid configs
		}

		if m.matchesProject(project, currentPath, gitRemote) {
			m.currentProject = project
			return nil
		}
	}

	return nil // No matching project found
}

// loadProjectConfig loads a single project configuration file
func (m *Manager) loadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// matchesProject checks if current directory matches a project configuration
func (m *Manager) matchesProject(project *ProjectConfig, currentPath, gitRemote string) bool {
	// Check path matches
	for _, pattern := range project.Match.Paths {
		// Handle wildcards in paths
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(currentPath, prefix) {
				return true
			}
		} else if currentPath == pattern || strings.HasPrefix(currentPath, pattern+"/") {
			return true
		}
	}

	// Check remote matches
	if gitRemote != "" {
		for _, remote := range project.Match.Remotes {
			if gitRemote == remote {
				return true
			}
		}
	}

	return false
}

// GetCurrentProject returns the currently loaded project config
func (m *Manager) GetCurrentProject() *ProjectConfig {
	return m.currentProject
}

// GetCommand returns a command from the current project
func (m *Manager) GetCommand(name string) (*NavigationCommand, bool) {
	if m.currentProject == nil {
		return nil, false
	}

	cmd, exists := m.currentProject.Commands[name]
	return &cmd, exists
}

// SaveProjectConfig saves a project configuration
func (m *Manager) SaveProjectConfig(project *ProjectConfig) error {
	projectsDir := filepath.Join(m.configDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(projectsDir, project.Name+".yaml")
	
	data, err := yaml.Marshal(project)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetConfigDir returns the configuration directory
func (m *Manager) GetConfigDir() string {
	return m.configDir
}
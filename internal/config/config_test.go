package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tobiase/worktree-utils/test/helpers"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() failed: %v", err)
	}

	homeDir, _ := os.UserHomeDir()
	expectedConfigDir := filepath.Join(homeDir, ".config", "wt")
	
	if manager.configDir != expectedConfigDir {
		t.Errorf("Expected config dir %s, got %s", expectedConfigDir, manager.configDir)
	}
}

func TestLoadProjectConfig(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		check       func(t *testing.T, pc *ProjectConfig)
	}{
		{
			name: "valid project config",
			yamlContent: `name: myproject
match:
  paths:
    - /home/user/myproject
    - /home/user/myproject-worktrees/*
  remotes:
    - git@github.com:user/myproject.git
commands:
  api:
    description: "Go to API"
    target: "services/api"
  web:
    description: "Go to web"
    target: "apps/web"
settings:
  worktree_base: /home/user/custom-worktrees
`,
			wantErr: false,
			check: func(t *testing.T, pc *ProjectConfig) {
				if pc.Name != "myproject" {
					t.Errorf("Expected name 'myproject', got '%s'", pc.Name)
				}
				
				if len(pc.Match.Paths) != 2 {
					t.Errorf("Expected 2 paths, got %d", len(pc.Match.Paths))
				}
				
				if len(pc.Commands) != 2 {
					t.Errorf("Expected 2 commands, got %d", len(pc.Commands))
				}
				
				if pc.Settings.WorktreeBase != "/home/user/custom-worktrees" {
					t.Errorf("Expected worktree_base '/home/user/custom-worktrees', got '%s'", pc.Settings.WorktreeBase)
				}
			},
		},
		{
			name: "project with virtualenv config",
			yamlContent: `name: python-project
match:
  paths:
    - /home/user/python-project
virtualenv:
  name: .venv
  python: python3.11
  auto_commands: true
commands:
  src:
    description: "Go to source"
    target: "src"
`,
			wantErr: false,
			check: func(t *testing.T, pc *ProjectConfig) {
				if pc.Virtualenv == nil {
					t.Fatal("Expected virtualenv config to be present")
				}
				
				if pc.Virtualenv.Name != ".venv" {
					t.Errorf("Expected virtualenv name '.venv', got '%s'", pc.Virtualenv.Name)
				}
				
				if pc.Virtualenv.Python != "python3.11" {
					t.Errorf("Expected python 'python3.11', got '%s'", pc.Virtualenv.Python)
				}
				
				if !pc.Virtualenv.AutoCommands {
					t.Error("Expected auto_commands to be true")
				}
			},
		},
		{
			name: "invalid yaml",
			yamlContent: `name: invalid
match:
  paths: [unclosed
`,
			wantErr: true,
		},
		{
			name: "empty config",
			yamlContent: ``,
			wantErr: false,
			check: func(t *testing.T, pc *ProjectConfig) {
				if pc.Name != "" {
					t.Errorf("Expected empty name, got '%s'", pc.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.WithTempDir(t, func(dir string) {
				// Create test config file
				configPath := filepath.Join(dir, "test.yaml")
				if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644); err != nil {
					t.Fatalf("Failed to write test config: %v", err)
				}

				// Create manager and load config
				manager := &Manager{configDir: dir}
				pc, err := manager.loadProjectConfig(configPath)

				if (err != nil) != tt.wantErr {
					t.Errorf("loadProjectConfig() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr && tt.check != nil {
					tt.check(t, pc)
				}
			})
		})
	}
}

func TestMatchesProject(t *testing.T) {
	tests := []struct {
		name        string
		project     *ProjectConfig
		currentPath string
		gitRemote   string
		want        bool
	}{
		{
			name: "exact path match",
			project: &ProjectConfig{
				Match: ProjectMatch{
					Paths: []string{"/home/user/myproject"},
				},
			},
			currentPath: "/home/user/myproject",
			want:        true,
		},
		{
			name: "wildcard path match",
			project: &ProjectConfig{
				Match: ProjectMatch{
					Paths: []string{"/home/user/myproject-worktrees/*"},
				},
			},
			currentPath: "/home/user/myproject-worktrees/feature-branch",
			want:        true,
		},
		{
			name: "remote match",
			project: &ProjectConfig{
				Match: ProjectMatch{
					Remotes: []string{"git@github.com:user/repo.git"},
				},
			},
			currentPath: "/any/path",
			gitRemote:   "git@github.com:user/repo.git",
			want:        true,
		},
		{
			name: "no match",
			project: &ProjectConfig{
				Match: ProjectMatch{
					Paths:   []string{"/home/user/other"},
					Remotes: []string{"git@github.com:other/repo.git"},
				},
			},
			currentPath: "/home/user/myproject",
			gitRemote:   "git@github.com:user/repo.git",
			want:        false,
		},
		{
			name: "multiple paths with one match",
			project: &ProjectConfig{
				Match: ProjectMatch{
					Paths: []string{
						"/home/user/old-location",
						"/home/user/new-location",
						"/home/user/new-location-worktrees/*",
					},
				},
			},
			currentPath: "/home/user/new-location-worktrees/feature",
			want:        true,
		},
		{
			name: "case sensitive path match",
			project: &ProjectConfig{
				Match: ProjectMatch{
					Paths: []string{"/home/User/Project"},
				},
			},
			currentPath: "/home/user/project",
			want:        false,
		},
		{
			name: "subdirectory of matched path",
			project: &ProjectConfig{
				Match: ProjectMatch{
					Paths: []string{"/home/user/myproject"},
				},
			},
			currentPath: "/home/user/myproject/src/api",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{}
			got := manager.matchesProject(tt.project, tt.currentPath, tt.gitRemote)
			
			if got != tt.want {
				t.Errorf("matchesProject() = %v, want %v", got, tt.want)
			}
		})
	}
}


func TestRegisterVirtualenvCommands(t *testing.T) {
	tests := []struct {
		name    string
		project *ProjectConfig
		check   func(t *testing.T, project *ProjectConfig)
	}{
		{
			name: "register virtualenv commands",
			project: &ProjectConfig{
				Virtualenv: &VirtualenvConfig{
					Name:         ".venv",
					Python:       "python3.11",
					AutoCommands: true,
				},
			},
			check: func(t *testing.T, project *ProjectConfig) {
				expectedCommands := []string{"venv", "mkvenv", "rmvenv"}
				
				for _, cmdName := range expectedCommands {
					if cmd, ok := project.Commands[cmdName]; !ok {
						t.Errorf("Expected virtualenv command '%s' to be registered", cmdName)
					} else {
						if cmd.Type != "virtualenv" {
							t.Errorf("Expected command '%s' to have type 'virtualenv', got '%s'", cmdName, cmd.Type)
						}
					}
				}
				
				// Check specific command targets
				if venv, ok := project.Commands["venv"]; ok && venv.Target != "activate" {
					t.Errorf("Expected venv target 'activate', got '%s'", venv.Target)
				}
				
				if mkvenv, ok := project.Commands["mkvenv"]; ok && mkvenv.Target != "create" {
					t.Errorf("Expected mkvenv target 'create', got '%s'", mkvenv.Target)
				}
				
				if rmvenv, ok := project.Commands["rmvenv"]; ok && rmvenv.Target != "remove" {
					t.Errorf("Expected rmvenv target 'remove', got '%s'", rmvenv.Target)
				}
			},
		},
		{
			name: "register commands with existing commands",
			project: &ProjectConfig{
				Commands: map[string]NavigationCommand{
					"api": {
						Description: "Go to API",
						Target:      "api",
					},
				},
				Virtualenv: &VirtualenvConfig{
					Name:         ".venv",
					AutoCommands: true,
				},
			},
			check: func(t *testing.T, project *ProjectConfig) {
				// Should have both original and virtualenv commands
				if len(project.Commands) != 4 { // 1 original + 3 virtualenv
					t.Errorf("Expected 4 commands, got %d", len(project.Commands))
				}
				
				// Original command should still exist
				if _, ok := project.Commands["api"]; !ok {
					t.Error("Original 'api' command was lost")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{}
			manager.registerVirtualenvCommands(tt.project)
			
			if tt.check != nil {
				tt.check(t, tt.project)
			}
		})
	}
}

func TestLoadProject(t *testing.T) {
	tests := []struct {
		name         string
		projectFiles map[string]string
		currentPath  string
		gitRemote    string
		wantProject  bool
		wantName     string
	}{
		{
			name: "load matching project",
			projectFiles: map[string]string{
				"projects/myproject.yaml": `name: myproject
match:
  paths:
    - /home/user/myproject
  remotes:
    - git@github.com:user/myproject.git
commands:
  api:
    description: "Go to API"
    target: "api"
`,
			},
			currentPath: "/home/user/myproject",
			gitRemote:   "git@github.com:user/myproject.git",
			wantProject: true,
			wantName:    "myproject",
		},
		{
			name: "no matching project",
			projectFiles: map[string]string{
				"projects/other.yaml": `name: other
match:
  paths:
    - /home/user/other
`,
			},
			currentPath: "/home/user/myproject",
			wantProject: false,
		},
		{
			name:         "no project files",
			projectFiles: map[string]string{},
			currentPath:  "/home/user/myproject",
			wantProject:  false,
		},
		{
			name: "invalid project file skipped",
			projectFiles: map[string]string{
				"projects/invalid.yaml": `invalid yaml content
[[[`,
				"projects/valid.yaml": `name: valid
match:
  paths:
    - /home/user/myproject
`,
			},
			currentPath: "/home/user/myproject",
			wantProject: true,
			wantName:    "valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.WithTempDir(t, func(dir string) {
				// Create test project files
				for path, content := range tt.projectFiles {
					helpers.CreateFiles(t, dir, map[string]string{path: content})
				}

				// Create manager with test config dir
				manager := &Manager{configDir: dir}
				
				// Load project
				err := manager.LoadProject(tt.currentPath, tt.gitRemote)
				if err != nil {
					t.Fatalf("LoadProject() error = %v", err)
				}

				// Check results
				if tt.wantProject {
					if manager.currentProject == nil {
						t.Error("Expected project to be loaded, but got nil")
					} else if manager.currentProject.Name != tt.wantName {
						t.Errorf("Expected project name '%s', got '%s'", tt.wantName, manager.currentProject.Name)
					}
				} else {
					if manager.currentProject != nil {
						t.Errorf("Expected no project, but got '%s'", manager.currentProject.Name)
					}
				}
			})
		})
	}
}

func TestGetCommand(t *testing.T) {
	manager := &Manager{
		currentProject: &ProjectConfig{
			Commands: map[string]NavigationCommand{
				"api": {
					Description: "Go to API",
					Target:      "services/api",
				},
			},
		},
	}

	tests := []struct {
		name     string
		cmdName  string
		wantCmd  bool
		wantDesc string
	}{
		{
			name:     "existing command",
			cmdName:  "api",
			wantCmd:  true,
			wantDesc: "Go to API",
		},
		{
			name:    "non-existing command",
			cmdName: "web",
			wantCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, exists := manager.GetCommand(tt.cmdName)
			
			if exists != tt.wantCmd {
				t.Errorf("GetCommand() exists = %v, want %v", exists, tt.wantCmd)
			}
			
			if tt.wantCmd && cmd != nil {
				if cmd.Description != tt.wantDesc {
					t.Errorf("GetCommand() description = %q, want %q", cmd.Description, tt.wantDesc)
				}
			}
		})
	}
}

func TestGetCommandNoProject(t *testing.T) {
	manager := &Manager{
		currentProject: nil,
	}

	cmd, exists := manager.GetCommand("any")
	if exists {
		t.Error("Expected GetCommand to return false when no project is loaded")
	}
	if cmd != nil {
		t.Error("Expected GetCommand to return nil when no project is loaded")
	}
}

func TestSaveProjectConfig(t *testing.T) {
	helpers.WithTempDir(t, func(dir string) {
		manager := &Manager{configDir: dir}
		
		project := &ProjectConfig{
			Name: "testproject",
			Match: ProjectMatch{
				Paths:   []string{"/home/user/test"},
				Remotes: []string{"git@github.com:user/test.git"},
			},
			Commands: map[string]NavigationCommand{
				"src": {
					Description: "Go to source",
					Target:      "src",
				},
			},
			Settings: ProjectSettings{
				WorktreeBase: "/home/user/test-worktrees",
			},
		}
		
		// Save project
		err := manager.SaveProjectConfig(project)
		if err != nil {
			t.Fatalf("SaveProjectConfig() error = %v", err)
		}
		
		// Verify file was created
		configPath := filepath.Join(dir, "projects", "testproject.yaml")
		helpers.AssertFileExists(t, configPath)
		
		// Load and verify content
		loaded, err := manager.loadProjectConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load saved config: %v", err)
		}
		
		if loaded.Name != project.Name {
			t.Errorf("Loaded project name = %q, want %q", loaded.Name, project.Name)
		}
		
		if len(loaded.Commands) != len(project.Commands) {
			t.Errorf("Loaded project has %d commands, want %d", len(loaded.Commands), len(project.Commands))
		}
		
		if loaded.Settings.WorktreeBase != project.Settings.WorktreeBase {
			t.Errorf("Loaded worktree_base = %q, want %q", loaded.Settings.WorktreeBase, project.Settings.WorktreeBase)
		}
	})
}

func TestGetVirtualenvConfig(t *testing.T) {
	tests := []struct {
		name    string
		project *ProjectConfig
		want    *VirtualenvConfig
	}{
		{
			name: "project with virtualenv",
			project: &ProjectConfig{
				Virtualenv: &VirtualenvConfig{
					Name:   ".venv",
					Python: "python3.11",
				},
			},
			want: &VirtualenvConfig{
				Name:   ".venv",
				Python: "python3.11",
			},
		},
		{
			name:    "project without virtualenv",
			project: &ProjectConfig{},
			want:    nil,
		},
		{
			name:    "no project loaded",
			project: nil,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &Manager{
				currentProject: tt.project,
			}
			
			got := manager.GetVirtualenvConfig()
			
			if tt.want == nil {
				if got != nil {
					t.Error("Expected nil virtualenv config")
				}
			} else {
				if got == nil {
					t.Error("Expected non-nil virtualenv config")
				} else {
					if got.Name != tt.want.Name {
						t.Errorf("Virtualenv name = %q, want %q", got.Name, tt.want.Name)
					}
					if got.Python != tt.want.Python {
						t.Errorf("Virtualenv python = %q, want %q", got.Python, tt.want.Python)
					}
				}
			}
		})
	}
}
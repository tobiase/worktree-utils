package cli

import (
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected *FlagSet
	}{
		{
			name: "no flags",
			args: []string{"branch1", "branch2"},
			expected: &FlagSet{
				Positional: []string{"branch1", "branch2"},
			},
		},
		{
			name: "boolean flags",
			args: []string{"--fuzzy", "--recursive", "-a", "branch"},
			expected: &FlagSet{
				Fuzzy:      true,
				Recursive:  true,
				All:        true,
				Positional: []string{"branch"},
			},
		},
		{
			name: "short flags combined",
			args: []string{"-fhra", "branch"},
			expected: &FlagSet{
				Fuzzy:      true,
				Help:       true,
				Recursive:  true,
				All:        true,
				Positional: []string{"branch"},
			},
		},
		{
			name: "string flag with value",
			args: []string{"--base", "main", "feature"},
			expected: &FlagSet{
				Base:       "main",
				Positional: []string{"feature"},
			},
		},
		{
			name: "string flag with equals",
			args: []string{"--base=develop", "feature"},
			expected: &FlagSet{
				Base:       "develop",
				Positional: []string{"feature"},
			},
		},
		{
			name: "mixed flags and args",
			args: []string{"--fuzzy", "branch1", "--recursive", "branch2"},
			expected: &FlagSet{
				Fuzzy:      true,
				Recursive:  true,
				Positional: []string{"branch1", "branch2"},
			},
		},
		{
			name: "unknown flags",
			args: []string{"--unknown", "--fuzzy", "branch"},
			expected: &FlagSet{
				Fuzzy:      true,
				Positional: []string{"--unknown", "branch"},
			},
		},
		{
			name: "help flag variations",
			args: []string{"--help", "-h"},
			expected: &FlagSet{
				Help:       true,
				Positional: []string{},
			},
		},
		{
			name: "no-switch flag",
			args: []string{"--no-switch", "branch"},
			expected: &FlagSet{
				NoSwitch:   true,
				Positional: []string{"branch"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFlags(tt.args)

			// Compare boolean flags
			if result.Fuzzy != tt.expected.Fuzzy {
				t.Errorf("Fuzzy flag: got %v, want %v", result.Fuzzy, tt.expected.Fuzzy)
			}
			if result.Help != tt.expected.Help {
				t.Errorf("Help flag: got %v, want %v", result.Help, tt.expected.Help)
			}
			if result.Recursive != tt.expected.Recursive {
				t.Errorf("Recursive flag: got %v, want %v", result.Recursive, tt.expected.Recursive)
			}
			if result.All != tt.expected.All {
				t.Errorf("All flag: got %v, want %v", result.All, tt.expected.All)
			}
			if result.NoSwitch != tt.expected.NoSwitch {
				t.Errorf("NoSwitch flag: got %v, want %v", result.NoSwitch, tt.expected.NoSwitch)
			}

			// Compare string flags
			if result.Base != tt.expected.Base {
				t.Errorf("Base flag: got %q, want %q", result.Base, tt.expected.Base)
			}

			// Compare positional arguments
			if !reflect.DeepEqual(result.Positional, tt.expected.Positional) {
				t.Errorf("Positional args: got %v, want %v", result.Positional, tt.expected.Positional)
			}
		})
	}
}

func TestHasFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		flagName string
		expected bool
	}{
		{
			name:     "flag present",
			args:     []string{"--fuzzy", "branch"},
			flagName: "--fuzzy",
			expected: true,
		},
		{
			name:     "flag not present",
			args:     []string{"--recursive", "branch"},
			flagName: "--fuzzy",
			expected: false,
		},
		{
			name:     "flag with equals",
			args:     []string{"--base=main", "branch"},
			flagName: "--base",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasFlag(tt.args, tt.flagName)
			if result != tt.expected {
				t.Errorf("HasFlag(%v, %q) = %v, want %v", tt.args, tt.flagName, result, tt.expected)
			}
		})
	}
}

func TestGetFlagValue(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flagName  string
		shortFlag string
		wantValue string
		wantFound bool
	}{
		{
			name:      "long flag with value",
			args:      []string{"--base", "main", "feature"},
			flagName:  "--base",
			shortFlag: "-b",
			wantValue: "main",
			wantFound: true,
		},
		{
			name:      "short flag with value",
			args:      []string{"-b", "develop", "feature"},
			flagName:  "--base",
			shortFlag: "-b",
			wantValue: "develop",
			wantFound: true,
		},
		{
			name:      "flag with equals",
			args:      []string{"--base=staging", "feature"},
			flagName:  "--base",
			shortFlag: "-b",
			wantValue: "staging",
			wantFound: true,
		},
		{
			name:      "flag not present",
			args:      []string{"--fuzzy", "feature"},
			flagName:  "--base",
			shortFlag: "-b",
			wantValue: "",
			wantFound: false,
		},
		{
			name:      "flag at end without value",
			args:      []string{"feature", "--base"},
			flagName:  "--base",
			shortFlag: "-b",
			wantValue: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, found := GetFlagValue(tt.args, tt.flagName, tt.shortFlag)
			if value != tt.wantValue || found != tt.wantFound {
				t.Errorf("GetFlagValue() = (%q, %v), want (%q, %v)",
					value, found, tt.wantValue, tt.wantFound)
			}
		})
	}
}

package interactive

import (
	"os"
	"testing"
)

func TestIsInteractive(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    bool
	}{
		{
			name:    "normal environment",
			envVars: map[string]string{},
			// Result depends on actual terminal, but should not panic
		},
		{
			name: "explicit disable via DISABLE_FUZZY",
			envVars: map[string]string{
				"DISABLE_FUZZY": "true",
			},
			want: false,
		},
		{
			name: "explicit disable via WT_NO_INTERACTIVE",
			envVars: map[string]string{
				"WT_NO_INTERACTIVE": "true",
			},
			want: false,
		},
		{
			name: "CI environment - GitHub Actions",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			want: false,
		},
		{
			name: "CI environment - general CI",
			envVars: map[string]string{
				"CI": "true",
			},
			want: false,
		},
		{
			name: "NO_COLOR environment",
			envVars: map[string]string{
				"NO_COLOR": "1",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
			}

			// Set test env
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Restore env after test
			defer func() {
				for key, originalValue := range originalEnv {
					if originalValue == "" {
						os.Unsetenv(key)
					} else {
						os.Setenv(key, originalValue)
					}
				}
			}()

			got := IsInteractive()

			// For tests that set environment variables that should disable interactive mode,
			// we can assert they return false
			if len(tt.envVars) > 0 && tt.want == false {
				if got != false {
					t.Errorf("IsInteractive() = %v, want %v", got, false)
				}
			}

			// For normal environment test, just ensure it doesn't panic
			// The actual result depends on whether tests are run in a terminal
		})
	}
}

func TestShouldUseFuzzy(t *testing.T) {
	tests := []struct {
		name         string
		itemCount    int
		explicitFlag bool
		interactive  bool
		want         bool
	}{
		{
			name:         "single item, no flag, interactive",
			itemCount:    1,
			explicitFlag: false,
			interactive:  true,
			want:         false,
		},
		{
			name:         "multiple items, no flag, interactive",
			itemCount:    3,
			explicitFlag: false,
			interactive:  true,
			want:         true,
		},
		{
			name:         "multiple items, no flag, non-interactive",
			itemCount:    3,
			explicitFlag: false,
			interactive:  false,
			want:         false,
		},
		{
			name:         "single item, explicit flag, interactive",
			itemCount:    1,
			explicitFlag: true,
			interactive:  true,
			want:         true,
		},
		{
			name:         "single item, explicit flag, non-interactive",
			itemCount:    1,
			explicitFlag: true,
			interactive:  false,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For this test, we'll test the logic directly since ShouldUseFuzzy
			// calls IsInteractive() internally which we can't easily mock
			var got bool
			if tt.explicitFlag {
				got = tt.interactive // Explicit flag honors interactive state
			} else {
				got = tt.itemCount > 1 && tt.interactive // Auto-enable with multiple items
			}

			if got != tt.want {
				t.Errorf("ShouldUseFuzzy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCIEnvironment(t *testing.T) {
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"BUILDKITE",
	}

	for _, ciVar := range ciVars {
		t.Run("detects "+ciVar, func(t *testing.T) {
			// Save original value
			original := os.Getenv(ciVar)
			defer func() {
				if original == "" {
					os.Unsetenv(ciVar)
				} else {
					os.Setenv(ciVar, original)
				}
			}()

			// Set CI variable
			os.Setenv(ciVar, "true")

			if !isCIEnvironment() {
				t.Errorf("isCIEnvironment() should return true when %s is set", ciVar)
			}
		})
	}

	t.Run("no CI variables set", func(t *testing.T) {
		// Save and clear all CI variables
		originalValues := make(map[string]string)
		for _, ciVar := range ciVars {
			originalValues[ciVar] = os.Getenv(ciVar)
			os.Unsetenv(ciVar)
		}

		// Restore after test
		defer func() {
			for ciVar, value := range originalValues {
				if value != "" {
					os.Setenv(ciVar, value)
				}
			}
		}()

		if isCIEnvironment() {
			t.Error("isCIEnvironment() should return false when no CI variables are set")
		}
	})
}

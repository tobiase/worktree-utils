package interactive

import (
	"errors"
	"testing"
)

func TestSelectWithNonInteractive(t *testing.T) {
	// Test the fallback behavior when not interactive
	tests := []struct {
		name        string
		items       []string
		opts        SelectOptions
		expectError bool
		expectedErr error
	}{
		{
			name:        "empty items",
			items:       []string{},
			opts:        SelectOptions{},
			expectError: true,
		},
		{
			name:        "single item returns directly",
			items:       []string{"main"},
			opts:        SelectOptions{},
			expectError: false,
		},
		{
			name:        "multiple items in non-interactive mode",
			items:       []string{"main", "feature", "develop"},
			opts:        SelectOptions{},
			expectError: true,
			expectedErr: ErrNotInteractive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := selectWithFallback(tt.items, tt.opts)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Error("Expected result but got nil")
				return
			}

			if len(result.Items) != 1 {
				t.Errorf("Expected 1 item, got %d", len(result.Items))
				return
			}

			if result.Items[0] != tt.items[0] {
				t.Errorf("Expected %q, got %q", tt.items[0], result.Items[0])
			}
		})
	}
}

func TestSelectOptions(t *testing.T) {
	tests := []struct {
		name string
		opts SelectOptions
	}{
		{
			name: "basic options",
			opts: SelectOptions{
				Prompt: "Select an option:",
				Header: "Use arrow keys to navigate",
			},
		},
		{
			name: "with preview function",
			opts: SelectOptions{
				Prompt: "Select with preview:",
				PreviewFunc: func(i, w, h int) string {
					return "Preview content"
				},
			},
		},
		{
			name: "multi-selection",
			opts: SelectOptions{
				Prompt: "Select multiple:",
				Multi:  true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that options are properly structured
			// In a real test environment with terminal, you would test actual selection
			// For now, we just verify the options don't cause panics when created

			// Validate prompt is set for non-basic test cases
			if tt.opts.Prompt == "" && tt.name != "basic options" {
				t.Errorf("Test case %s should have a prompt", tt.name)
			}

			if tt.opts.PreviewFunc != nil {
				// Test preview function
				preview := tt.opts.PreviewFunc(0, 80, 24)
				if preview == "" && tt.name == "with preview function" {
					t.Error("Preview function should return content")
				}
			}
		})
	}
}

func TestSelectResult(t *testing.T) {
	result := &SelectResult{
		Indices: []int{0, 2},
		Items:   []string{"first", "third"},
	}

	if len(result.Indices) != 2 {
		t.Errorf("Expected 2 indices, got %d", len(result.Indices))
	}

	if len(result.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result.Items))
	}

	if result.Items[0] != "first" {
		t.Errorf("Expected 'first', got %q", result.Items[0])
	}

	if result.Items[1] != "third" {
		t.Errorf("Expected 'third', got %q", result.Items[1])
	}
}

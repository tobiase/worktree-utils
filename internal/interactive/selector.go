package interactive

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

var (
	// ErrNotInteractive is returned when interactive features are not available
	ErrNotInteractive = errors.New("interactive features not available in this environment")
	// ErrUserCancelled is returned when the user cancels the selection
	ErrUserCancelled = errors.New("selection cancelled by user")
)

// SelectOptions configures the selection behavior
type SelectOptions struct {
	Prompt      string
	Header      string
	PreviewFunc func(int, int, int) string // index, width, height -> preview content
	Multi       bool                       // Allow multiple selections
}

// SelectResult contains the result of a selection operation
type SelectResult struct {
	Indices []int
	Items   []string
}

// Select presents an interactive selection interface for the given items
func Select(items []string, opts SelectOptions) (*SelectResult, error) {
	if len(items) == 0 {
		return nil, errors.New("no items to select from")
	}

	if !IsInteractive() {
		return selectWithFallback(items, opts)
	}

	return selectWithFuzzyFinder(items, opts)
}

// selectWithFuzzyFinder uses go-fuzzyfinder for interactive selection
func selectWithFuzzyFinder(items []string, opts SelectOptions) (*SelectResult, error) {
	var findOpts []fuzzyfinder.Option

	// Set prompt
	if opts.Prompt != "" {
		findOpts = append(findOpts, fuzzyfinder.WithPromptString(opts.Prompt))
	}

	// Set header
	if opts.Header != "" {
		findOpts = append(findOpts, fuzzyfinder.WithHeader(opts.Header))
	}

	// Set preview function
	if opts.PreviewFunc != nil {
		findOpts = append(findOpts, fuzzyfinder.WithPreviewWindow(opts.PreviewFunc))
	}

	if opts.Multi {
		// Multi-selection mode
		indices, err := fuzzyfinder.FindMulti(
			items,
			func(i int) string { return items[i] },
			findOpts...,
		)
		if err != nil {
			if err == fuzzyfinder.ErrAbort {
				return nil, ErrUserCancelled
			}
			return nil, fmt.Errorf("selection failed: %w", err)
		}

		selectedItems := make([]string, len(indices))
		for i, idx := range indices {
			selectedItems[i] = items[idx]
		}

		return &SelectResult{
			Indices: indices,
			Items:   selectedItems,
		}, nil
	}

	// Single selection mode
	idx, err := fuzzyfinder.Find(
		items,
		func(i int) string { return items[i] },
		findOpts...,
	)
	if err != nil {
		if err == fuzzyfinder.ErrAbort {
			return nil, ErrUserCancelled
		}
		return nil, fmt.Errorf("selection failed: %w", err)
	}

	return &SelectResult{
		Indices: []int{idx},
		Items:   []string{items[idx]},
	}, nil
}

// selectWithFallback provides a numbered list fallback for non-interactive environments
func selectWithFallback(items []string, opts SelectOptions) (*SelectResult, error) {
	if len(items) == 1 {
		// Only one item, return it directly
		return &SelectResult{
			Indices: []int{0},
			Items:   []string{items[0]},
		}, nil
	}

	// In non-interactive mode with multiple items, we can't prompt
	// Return an error with helpful information
	return nil, fmt.Errorf("%w: multiple options available: %s",
		ErrNotInteractive,
		strings.Join(items, ", "))
}

// SelectString is a convenience function for single string selection
func SelectString(items []string, prompt string) (string, error) {
	result, err := Select(items, SelectOptions{
		Prompt: prompt,
		Header: "Use arrow keys to navigate, Enter to select, Esc to cancel",
	})
	if err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", errors.New("no item selected")
	}

	return result.Items[0], nil
}

// SelectStringWithPreview is a convenience function for single string selection with preview
func SelectStringWithPreview(items []string, prompt string, previewFunc func(int, int, int) string) (string, error) {
	result, err := Select(items, SelectOptions{
		Prompt:      prompt,
		Header:      "Use arrow keys to navigate, Enter to select, Esc to cancel",
		PreviewFunc: previewFunc,
	})
	if err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return "", errors.New("no item selected")
	}

	return result.Items[0], nil
}

// PromptNumberedSelection provides a simple numbered list for environments where fuzzy finding isn't available
func PromptNumberedSelection(items []string, prompt string) (string, error) {
	if len(items) == 0 {
		return "", errors.New("no items to select from")
	}

	if len(items) == 1 {
		return items[0], nil
	}

	fmt.Printf("%s\n", prompt)
	for i, item := range items {
		fmt.Printf("  %d) %s\n", i+1, item)
	}

	fmt.Print("Enter your choice (1-" + strconv.Itoa(len(items)) + "): ")

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	choice, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || choice < 1 || choice > len(items) {
		return "", fmt.Errorf("invalid selection: %s", input)
	}

	return items[choice-1], nil
}

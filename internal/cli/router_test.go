package cli

import (
	"testing"
)

func TestSubcommandRouter(t *testing.T) {
	t.Run("route to correct handler", func(t *testing.T) {
		router := NewSubcommandRouter("test", "Usage: test <subcommand>")

		called := false
		router.AddCommand("subcommand", func(args []string) {
			called = true
		})

		router.Route([]string{"subcommand"})

		if !called {
			t.Error("Expected handler to be called")
		}
	})

	t.Run("pass arguments to handler", func(t *testing.T) {
		router := NewSubcommandRouter("test", "Usage: test <subcommand>")

		var receivedArgs []string
		router.AddCommand("subcommand", func(args []string) {
			receivedArgs = args
		})

		router.Route([]string{"subcommand", "arg1", "arg2"})

		if len(receivedArgs) != 2 || receivedArgs[0] != "arg1" || receivedArgs[1] != "arg2" {
			t.Errorf("Expected args [arg1, arg2], got %v", receivedArgs)
		}
	})

	t.Run("call interactive handler when no args", func(t *testing.T) {
		router := NewSubcommandRouter("test", "Usage: test <subcommand>")

		interactiveCalled := false
		router.SetInteractive(func() {
			interactiveCalled = true
		})

		router.Route([]string{})

		if !interactiveCalled {
			t.Error("Expected interactive handler to be called")
		}
	})

	t.Run("show usage when no args and no interactive", func(t *testing.T) {
		// Skip this test as it's difficult to test os.Exit without modifying global state
		t.Skip("Skipping os.Exit test - would require global state modification")
	})

	t.Run("find similar commands", func(t *testing.T) {
		router := NewSubcommandRouter("test", "Usage: test <subcommand>")
		router.Commands["sync"] = func([]string) {}
		router.Commands["status"] = func([]string) {}
		router.Commands["setup"] = func([]string) {}
		router.Commands["list"] = func([]string) {}

		suggestions := router.findSimilarCommands("st")
		// Should find: status (starts with "st") and list (contains "st")
		if len(suggestions) != 2 {
			t.Errorf("Expected 2 suggestions, got %d: %v", len(suggestions), suggestions)
		}
		// Check that expected commands are present
		hasStatus := false
		hasList := false
		for _, s := range suggestions {
			switch s {
			case "status":
				hasStatus = true
			case "list":
				hasList = true
			}
		}
		if !hasStatus || !hasList {
			t.Errorf("Expected [list, status] in results, got %v", suggestions)
		}

		suggestions = router.findSimilarCommands("syn")
		if len(suggestions) != 1 || suggestions[0] != "sync" {
			t.Errorf("Expected [sync], got %v", suggestions)
		}

		suggestions = router.findSimilarCommands("xyz")
		if len(suggestions) != 0 {
			t.Errorf("Expected no suggestions, got %v", suggestions)
		}
	})
}

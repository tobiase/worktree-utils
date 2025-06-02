package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tobiase/worktree-utils/internal/worktree"
)

const shellWrapper = `# Shell function to handle CD: prefix
wt() {
  output=$("${WT_BIN:-wt-bin}" "$@" 2>&1)
  exit_code=$?
  
  if [ $exit_code -eq 0 ]; then
    if [[ "$output" == "CD:"* ]]; then
      cd "${output#CD:}"
    else
      [ -n "$output" ] && echo "$output"
    fi
  else
    echo "$output" >&2
    return $exit_code
  fi
}
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: wt [list|add <branch>|rm <branch>|go <index|branch>|dash]\n")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "shell-init":
		fmt.Print(shellWrapper)
		return

	case "list":
		if err := worktree.List(); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}

	case "add":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt add <branch>\n")
			os.Exit(1)
		}
		if err := worktree.Add(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}

	case "rm":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt rm <branch>\n")
			os.Exit(1)
		}
		if err := worktree.Remove(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}

	case "go":
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: wt go <index|branch>\n")
			os.Exit(1)
		}
		path, err := worktree.Go(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CD:%s", path)

	case "dash":
		repo, err := worktree.GetRepoRoot()
		if err != nil {
			fmt.Fprintf(os.Stderr, "wt: %v\n", err)
			os.Exit(1)
		}
		dashPath := filepath.Join(repo, "applications", "dashboard-app")
		if _, err := os.Stat(dashPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "wt: 'applications/dashboard-app' not found under repo\n")
			os.Exit(1)
		}
		fmt.Printf("CD:%s", dashPath)

	default:
		fmt.Fprintf(os.Stderr, "Usage: wt [list|add <branch>|rm <branch>|go <index|branch>|dash]\n")
		os.Exit(1)
	}
}
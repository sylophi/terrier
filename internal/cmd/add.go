package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sylophi/terrier/internal/git"
	"github.com/sylophi/terrier/internal/store"
)

const addUsage = "usage: terrier add [<path>]"

// Add registers the repository containing path, defaulting to the working
// directory.
func Add(args []string) error {
	rest, err := flags{}.parse(args, addUsage)
	if err != nil {
		return err
	}
	if len(rest) > 1 {
		return fmt.Errorf("expected at most one path\n%s", addUsage)
	}

	dir := "."
	if len(rest) == 1 {
		dir = expandTilde(rest[0])
	}

	// Registering the main worktree rather than the directory given means
	// `add` from inside a linked worktree registers the project it belongs
	// to, which is the only thing that could be meant.
	root, err := git.Root(dir)
	if err != nil {
		return err
	}

	added := false
	if err := store.Update(func(s *store.Store) error {
		added = s.Add(root)
		return nil
	}); err != nil {
		return err
	}

	label := describe(root).Slug
	if label == "" {
		label = filepath.Base(root)
	}
	if added {
		fmt.Printf("Registered %s (%s)\n", label, tilde(root))
	} else {
		fmt.Printf("Already registered: %s (%s)\n", label, tilde(root))
	}
	return nil
}

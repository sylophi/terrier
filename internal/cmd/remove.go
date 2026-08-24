package cmd

import (
	"fmt"
	"slices"

	"github.com/sylophi/terrier/internal/store"
)

const removeUsage = "usage: terrier rm <project>..."

// Remove unregisters projects. Nothing on disk is touched: the registry
// holds paths, so dropping one only makes terrier forget the repository
// exists.
func Remove(args []string) error {
	rest, err := flags{}.parse(args, removeUsage)
	if err != nil {
		return err
	}
	if len(rest) == 0 {
		return fmt.Errorf("no project given\n%s", removeUsage)
	}

	var removed []string
	// Everything is resolved before anything is removed, so naming one
	// project twice resolves twice rather than failing the second time
	// against a store the first pass already changed. Update saves nothing
	// if this returns an error, so a typo in the second argument cannot
	// half-apply the command either.
	if err := store.Update(func(s *store.Store) error {
		removed = removed[:0]
		paths := make([]string, 0, len(rest))
		for _, ref := range rest {
			path, err := resolveRef(s, ref)
			if err != nil {
				return err
			}
			if !slices.Contains(paths, path) {
				paths = append(paths, path)
			}
		}
		for _, path := range paths {
			if s.Remove(path) {
				removed = append(removed, path)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	for _, path := range removed {
		fmt.Printf("Unregistered %s (files untouched)\n", tilde(path))
	}
	return nil
}

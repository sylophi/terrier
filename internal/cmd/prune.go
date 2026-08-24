package cmd

import (
	"fmt"

	"github.com/sylophi/terrier/internal/store"
)

const pruneUsage = "usage: terrier prune [--yes]"

// Prune drops registered projects whose directory is gone.
//
// It confirms first, because "gone" and "not mounted right now" look
// identical from here, and an unplugged drive should not quietly empty
// the registry.
func Prune(args []string) error {
	var yes bool
	rest, err := yesFlag(&yes).parse(args, pruneUsage)
	if err != nil {
		return err
	}
	if len(rest) > 0 {
		return fmt.Errorf("unexpected argument: %s\n%s", rest[0], pruneUsage)
	}

	s, err := store.Load()
	if err != nil {
		return err
	}
	var gone []string
	for _, path := range s.Projects {
		if missing(path) {
			gone = append(gone, path)
		}
	}
	if len(gone) == 0 {
		fmt.Println("Nothing to prune.")
		return nil
	}

	verb, them := "no longer exist", "them"
	if len(gone) == 1 {
		verb, them = "no longer exists", "it"
	}
	fmt.Printf("%s %s:\n", plural(len(gone), "registered path"), verb)
	for _, path := range gone {
		fmt.Printf("  %s\n", tilde(path))
	}
	if !yes {
		ok, err := confirm("Unregister " + them + "?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Aborted.")
			return nil
		}
	}

	// Checked again under the lock, because the prompt has no time limit
	// and the drive that made these look missing may have come back during
	// it. Resolving what to change inside the closure is how every other
	// mutating command works.
	dropped := 0
	if err := store.Update(func(s *store.Store) error {
		dropped = 0
		for _, path := range gone {
			if missing(path) && s.Remove(path) {
				dropped++
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if back := len(gone) - dropped; back > 0 {
		fmt.Printf("%s came back and %s kept.\n", plural(back, "path"), wasWere(back))
	}
	fmt.Printf("Unregistered %s.\n", plural(dropped, "project"))
	return nil
}

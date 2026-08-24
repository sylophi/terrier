package cmd

import (
	"fmt"

	"github.com/sylophi/terrier/internal/store"
)

const pathUsage = "usage: terrier path [<project>] [--json]"

// Path prints where a project lives, defaulting to the one the working
// directory belongs to.
//
// The bare output is the path alone, so `cd $(terrier path)` works and
// `terrier path >/dev/null` is a complete "am I in a registered project?"
// check: an unregistered directory is a non-zero exit.
func Path(args []string) error {
	var asJSON bool
	rest, err := flags{"json": &asJSON}.parse(args, pathUsage)
	if err != nil {
		return err
	}
	if len(rest) > 1 {
		return fmt.Errorf("expected at most one project\n%s", pathUsage)
	}

	s, err := store.Load()
	if err != nil {
		return err
	}

	// The current project is resolved either way, so that a project named
	// on the command line is only reported as current when the user is
	// actually standing in it. Not being in one is only an error when no
	// name was given and it was the whole question.
	current, currentErr := currentPath(s)
	path := current
	if len(rest) == 1 {
		if path, err = resolveRef(s, rest[0]); err != nil {
			return err
		}
	} else if currentErr != nil {
		return currentErr
	}

	if asJSON {
		return writeJSON(describe(path, current))
	}
	fmt.Println(path)
	return nil
}

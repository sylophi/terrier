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

	var path string
	if len(rest) == 1 {
		path, err = resolveRef(s, rest[0])
	} else {
		path, err = currentPath(s)
	}
	if err != nil {
		return err
	}

	if asJSON {
		return writeJSON(describe(path))
	}
	fmt.Println(path)
	return nil
}

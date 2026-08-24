package cmd

import (
	"fmt"

	"github.com/sylophi/terrier/internal/store"
)

const listUsage = "usage: terrier ls [--json]"

// List prints the registry. This is terrier's hottest command, the one
// other tools call to discover projects, so it does no work it can avoid.
// Resolving the current project runs no subprocess, and --json costs a
// config-file read per project rather than a subprocess per project.
func List(args []string) error {
	var asJSON bool
	rest, err := flags{"json": &asJSON}.parse(args, listUsage)
	if err != nil {
		return err
	}
	if len(rest) > 0 {
		return fmt.Errorf("unexpected argument: %s\n%s", rest[0], listUsage)
	}

	s, err := store.Load()
	if err != nil {
		return err
	}
	// Being outside a repository is not an error here, it just means
	// nothing is marked as current.
	current, _ := currentPath(s)

	if asJSON {
		projects := make([]Project, 0, len(s.Projects))
		for _, path := range s.Projects {
			projects = append(projects, describe(path, current))
		}
		return writeJSON(struct {
			Projects []Project `json:"projects"`
		}{projects})
	}

	if len(s.Projects) == 0 {
		fmt.Println("No projects yet. Register one with `terrier add`.")
		return nil
	}
	for _, path := range s.Projects {
		marker := " "
		if path == current {
			marker = "*"
		}
		line := fmt.Sprintf("%s %s", marker, tilde(path))
		if missing(path) {
			line += "  (missing)"
		}
		fmt.Println(line)
	}
	return nil
}

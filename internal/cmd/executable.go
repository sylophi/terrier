package cmd

import (
	"os"
	"path/filepath"
)

// resolveExecutable returns the path of the running binary, resolved
// through any symlinks. That matters here because `ter` is one: running
// `ter update` has to replace terrier itself, not the alias pointing at
// it. EvalSymlinks failures fall back to the unresolved path silently.
func resolveExecutable() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return path, nil
}

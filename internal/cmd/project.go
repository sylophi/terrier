package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sylophi/terrier/internal/git"
	"github.com/sylophi/terrier/internal/store"
)

// Project is the record terrier reports. Only Path comes from the
// registry. Slug is derived from the repository on the spot, which is why
// it cannot be stale, and Missing from whether the path is still there.
//
// Nothing here restates something the caller already has. A name is the
// base of the path, and which project the caller is standing in is what
// `terrier path` answers, so neither is worth a field.
type Project struct {
	Path string `json:"path"`
	// Slug is "owner/name" for a GitHub origin, absent otherwise.
	Slug string `json:"slug,omitempty"`
	// Missing marks a project whose directory is gone. It is reported
	// rather than hidden, so `prune` has something to explain.
	Missing bool `json:"missing,omitempty"`
}

// missing reports whether a registered path has gone away. It is the one
// thing terrier can get out of step with the world, since a path is the
// only thing it stores, so every command decides it here rather than each
// spelling out its own stat.
func missing(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}

// describe builds the full record for a registered path. It reads the
// repository's config file directly rather than running git, so listing
// every project costs no subprocesses at all.
func describe(path string) Project {
	p := Project{Path: path}
	if p.Missing = missing(path); p.Missing {
		return p
	}
	p.Slug = git.Slug(git.OriginURL(path))
	return p
}

// currentPath returns the registered project the working directory
// belongs to.
//
// The registry lookup comes first because it is free: a string comparison
// per project, no subprocess. Git is only worth asking when the working
// directory sits under no registered path. That is what a linked worktree
// living elsewhere looks like.
func currentPath(s *store.Store) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir = git.Canonical(dir)

	if path := s.Containing(dir); path != "" {
		return path, nil
	}

	root, err := git.Root(dir)
	if err != nil {
		return "", err
	}
	if !s.Has(root) {
		return "", fmt.Errorf("%s is not registered\nRegister it with `terrier add`", tilde(root))
	}
	return root, nil
}

// resolveRef turns what a user typed into a registered path.
//
// Anything that looks like a path is matched as one. Anything else is
// matched against the trailing segments of every registered path, so
// `whatagain` finds .../dittofleet/whatagain and typing more of the path
// settles it when two repos share a name. Nothing is stored to make this
// work, so there are no aliases to go stale.
func resolveRef(s *store.Store, ref string) (string, error) {
	ref = strings.TrimRight(strings.TrimSpace(ref), string(filepath.Separator))
	if ref == "" {
		return "", fmt.Errorf("empty project reference")
	}

	if looksLikePath(ref) {
		path := git.Canonical(expandTilde(ref))
		if s.Has(path) {
			return path, nil
		}
		// Being handed a path inside a project is a good enough answer.
		if found := s.Containing(path); found != "" {
			return found, nil
		}
		return "", fmt.Errorf("no registered project at %s", tilde(path))
	}

	matches := suffixMatches(s.Projects, ref, true)
	if len(matches) == 0 {
		matches = suffixMatches(s.Projects, ref, false)
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("no project matches %q\nSee the registered ones with `terrier ls`", ref)
	case 1:
		return matches[0], nil
	default:
		var b strings.Builder
		fmt.Fprintf(&b, "%q matches %s:", ref, plural(len(matches), "project"))
		for _, m := range matches {
			fmt.Fprintf(&b, "\n  %s", tilde(m))
		}
		b.WriteString("\nAdd more of the path to say which one.")
		return "", fmt.Errorf("%s", b.String())
	}
}

// looksLikePath distinguishes a path the user typed from a bare name.
func looksLikePath(ref string) bool {
	return strings.HasPrefix(ref, "/") ||
		strings.HasPrefix(ref, "~") ||
		strings.HasPrefix(ref, ".")
}

// suffixMatches returns the paths ending in ref at a segment boundary, so
// "again" does not match ".../whatagain". The insensitive pass is a
// fallback for a name typed in the wrong case, tried only when the exact
// one found nothing.
func suffixMatches(paths []string, ref string, sensitive bool) []string {
	suffix := string(filepath.Separator) + filepath.Clean(ref)
	if !sensitive {
		suffix = strings.ToLower(suffix)
	}
	var matches []string
	for _, p := range paths {
		candidate := p
		if !sensitive {
			candidate = strings.ToLower(candidate)
		}
		if strings.HasSuffix(candidate, suffix) {
			matches = append(matches, p)
		}
	}
	return matches
}

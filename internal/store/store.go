// Package store reads and writes the registry: one JSON file holding the
// absolute path of every registered project, and nothing else.
//
// Nothing but paths is recorded on purpose. A name, a remote, or a
// default branch kept here would be a copy of something git already
// knows, and a copy is a thing that can go stale. Every other fact
// terrier reports is derived at read time instead.
//
// The file lives in the config directory, but unlike a remote-keyed store
// it is machine-local by nature: paths differ between machines, so
// syncing it is not expected to be useful.
package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sylophi/terrier/internal/xdg"
)

// SchemaVersion is what this build writes. A newer file is refused:
// dropping fields this build does not know about would quietly delete
// them.
const SchemaVersion = 1

// Store is the whole registry.
type Store struct {
	SchemaVersion int `json:"schemaVersion"`
	// Projects holds absolute, symlink-resolved paths to main worktree
	// roots, sorted. Sorting is what lets Has, Add, and Remove binary
	// search, and what keeps the file's diffs small.
	Projects []string `json:"projects"`
}

// Path returns the location of the registry file.
func Path() string {
	return filepath.Join(xdg.ConfigDir(xdg.App), "projects.json")
}

// Load reads the registry. A missing file is not an error: it yields an
// empty registry, so the first write is what creates the file. Neither is
// an empty one, which is what a `touch` leaves behind and holds nothing
// to lose.
func Load() (*Store, error) {
	path := Path()
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &Store{SchemaVersion: SchemaVersion}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return &Store{SchemaVersion: SchemaVersion}, nil
	}

	s := &Store{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", path, err)
	}
	switch {
	case s.SchemaVersion > SchemaVersion:
		return nil, fmt.Errorf("%s was written by a newer terrier (schemaVersion %d, this build understands %d)\nUpdate with `terrier update`", path, s.SchemaVersion, SchemaVersion)
	case s.SchemaVersion < 1:
		return nil, fmt.Errorf("invalid %s:\n  - schemaVersion: expected %d, got %d", path, SchemaVersion, s.SchemaVersion)
	}
	slices.Sort(s.Projects)
	return s, nil
}

// Save writes the registry back atomically, so a reader never observes a
// half-written file.
func (s *Store) Save() error {
	s.SchemaVersion = SchemaVersion
	slices.Sort(s.Projects)
	s.Projects = slices.Compact(s.Projects)
	if s.Projects == nil {
		s.Projects = []string{}
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	path := Path()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".projects-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	_, writeErr := tmp.Write(data)
	// Synced before the rename: the rename is atomic against a reader, but
	// not against power loss, and a zero-length projects.json is one Load
	// would read as a registry with nothing in it.
	syncErr := tmp.Sync()
	closeErr := tmp.Close()
	if writeErr != nil || syncErr != nil || closeErr != nil {
		_ = os.Remove(tmpPath)
		return errors.Join(writeErr, syncErr, closeErr)
	}
	if err := os.Chmod(tmpPath, 0o600); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

// Has reports whether path is registered.
func (s *Store) Has(path string) bool {
	_, found := slices.BinarySearch(s.Projects, path)
	return found
}

// Add registers path, reporting whether it was new. Registering something
// already registered is not an error anywhere in terrier: `add` run twice
// should succeed twice.
func (s *Store) Add(path string) bool {
	i, found := slices.BinarySearch(s.Projects, path)
	if found {
		return false
	}
	s.Projects = slices.Insert(s.Projects, i, path)
	return true
}

// Remove unregisters path, reporting whether it was there.
func (s *Store) Remove(path string) bool {
	i, found := slices.BinarySearch(s.Projects, path)
	if !found {
		return false
	}
	s.Projects = slices.Delete(s.Projects, i, i+1)
	return true
}

// Containing returns the registered project that dir sits inside, or "".
//
// This is the fast path behind every lookup: a string comparison per
// project and no subprocess at all, which matters because listing and
// resolving are what other tools call on every invocation. The longest
// match wins, so a repo nested inside another registered one resolves to
// itself rather than its parent.
//
// It only sees projects dir is literally underneath. A linked worktree
// living elsewhere is not, and callers fall back to asking git for those.
func (s *Store) Containing(dir string) string {
	best := ""
	for _, p := range s.Projects {
		if len(p) > len(best) && under(dir, p) {
			best = p
		}
	}
	return best
}

// under reports whether dir is root or sits beneath it, comparing whole
// path segments so /a/bc is not treated as living inside /a/b.
//
// The comparison is done by index rather than by building the prefix to
// match against, because Containing calls this once per registered project
// on terrier's hottest path and the string it would build is garbage every
// time.
func under(dir, root string) bool {
	if dir == root {
		return true
	}
	if root == string(filepath.Separator) {
		return strings.HasPrefix(dir, root)
	}
	return len(dir) > len(root) && dir[len(root)] == filepath.Separator && dir[:len(root)] == root
}

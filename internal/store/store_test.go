package store

import (
	"slices"
	"testing"
)

func TestAddKeepsProjectsSorted(t *testing.T) {
	s := &Store{}
	for _, p := range []string{"/b", "/a", "/c"} {
		if !s.Add(p) {
			t.Fatalf("Add(%q) reported nothing added", p)
		}
	}
	if !slices.IsSorted(s.Projects) {
		t.Errorf("projects not sorted: %v", s.Projects)
	}
}

func TestAddIsIdempotent(t *testing.T) {
	s := &Store{}
	s.Add("/a")
	if s.Add("/a") {
		t.Error("re-adding reported a new project")
	}
	if len(s.Projects) != 1 {
		t.Errorf("got %d projects, want 1", len(s.Projects))
	}
}

func TestRemove(t *testing.T) {
	s := &Store{Projects: []string{"/a", "/b"}}
	if !s.Remove("/a") {
		t.Error("Remove reported nothing removed")
	}
	if s.Has("/a") {
		t.Error("/a still registered")
	}
	if s.Remove("/nope") {
		t.Error("removing an unregistered path reported a removal")
	}
}

func TestContainingPrefersTheLongestMatch(t *testing.T) {
	// A repo nested inside another registered one resolves to itself.
	s := &Store{Projects: []string{"/repos/outer", "/repos/outer/vendor/inner"}}
	if got := s.Containing("/repos/outer/vendor/inner/src"); got != "/repos/outer/vendor/inner" {
		t.Errorf("got %q", got)
	}
	if got := s.Containing("/repos/outer/src"); got != "/repos/outer" {
		t.Errorf("got %q", got)
	}
}

func TestContainingRespectsSegmentBoundaries(t *testing.T) {
	s := &Store{Projects: []string{"/repos/api"}}
	if got := s.Containing("/repos/api-client"); got != "" {
		t.Errorf("/repos/api-client resolved to %q, want no match", got)
	}
	if got := s.Containing("/repos/api"); got != "/repos/api" {
		t.Errorf("got %q, want the project itself", got)
	}
}

func TestContainingEmptyStore(t *testing.T) {
	s := &Store{}
	if got := s.Containing("/anywhere"); got != "" {
		t.Errorf("got %q, want no match", got)
	}
}

func TestContainingAtTheFilesystemRoot(t *testing.T) {
	// Absurd to register, but the index comparison in under() must not
	// read past the end of a one-character root.
	s := &Store{Projects: []string{"/"}}
	if got := s.Containing("/anywhere/at/all"); got != "/" {
		t.Errorf("got %q, want %q", got, "/")
	}
}

func TestContainingIgnoresAShorterUnrelatedPath(t *testing.T) {
	s := &Store{Projects: []string{"/repos/api"}}
	if got := s.Containing("/repos"); got != "" {
		t.Errorf("a parent of a project resolved to it: %q", got)
	}
}

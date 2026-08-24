package cmd

import (
	"strings"
	"testing"

	"github.com/sylophi/terrier/internal/store"
)

func TestResolveRefByName(t *testing.T) {
	s := &store.Store{Projects: []string{"/s/dittofleet/whatagain", "/s/cli/terrier"}}
	got, err := resolveRef(s, "whatagain")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/s/dittofleet/whatagain" {
		t.Errorf("got %q", got)
	}
}

func TestResolveRefNeedsAWholeSegment(t *testing.T) {
	s := &store.Store{Projects: []string{"/s/dittofleet/whatagain"}}
	if _, err := resolveRef(s, "again"); err == nil {
		t.Error("a partial segment matched")
	}
}

func TestResolveRefDisambiguatesByPath(t *testing.T) {
	s := &store.Store{Projects: []string{"/s/alpha/api", "/s/beta/api"}}
	if _, err := resolveRef(s, "api"); err == nil {
		t.Fatal("an ambiguous ref resolved")
	} else if !strings.Contains(err.Error(), "alpha/api") || !strings.Contains(err.Error(), "beta/api") {
		t.Errorf("the error does not name the candidates: %v", err)
	}
	got, err := resolveRef(s, "beta/api")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/s/beta/api" {
		t.Errorf("got %q", got)
	}
}

func TestResolveRefFallsBackToCaseInsensitive(t *testing.T) {
	s := &store.Store{Projects: []string{"/s/cli/PReviewer"}}
	got, err := resolveRef(s, "previewer")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/s/cli/PReviewer" {
		t.Errorf("got %q", got)
	}
}

func TestResolveRefPrefersAnExactCaseMatch(t *testing.T) {
	s := &store.Store{Projects: []string{"/s/one/api", "/s/two/API"}}
	got, err := resolveRef(s, "api")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/s/one/api" {
		t.Errorf("got %q, want the exactly-cased match", got)
	}
}

func TestResolveRefUnknown(t *testing.T) {
	s := &store.Store{Projects: []string{"/s/cli/terrier"}}
	if _, err := resolveRef(s, "nope"); err == nil {
		t.Error("an unregistered name resolved")
	}
}

func TestLooksLikePath(t *testing.T) {
	cases := map[string]bool{
		"/abs/path": true,
		"~/repos/x": true,
		"./here":    true,
		"../there":  true,
		"whatagain": false,
		"a/b":       false,
	}
	for ref, want := range cases {
		if got := looksLikePath(ref); got != want {
			t.Errorf("looksLikePath(%q) = %v, want %v", ref, got, want)
		}
	}
}

// describe takes the current project's path rather than a bool so that no
// caller can assert currentness on its own. `terrier path <name> --json`
// run from a different repo used to report the named project as current.
func TestDescribeMarksOnlyTheCurrentProject(t *testing.T) {
	here, there := t.TempDir(), t.TempDir()
	if p := describe(there, here); p.Current {
		t.Error("a project the user is not standing in was marked current")
	}
	if p := describe(here, here); !p.Current {
		t.Error("the current project was not marked current")
	}
	if p := describe(here, ""); p.Current {
		t.Error("a project was marked current with no current project resolved")
	}
}

func TestDescribeReportsAMissingPath(t *testing.T) {
	p := describe("/no/such/path/anywhere", "")
	if !p.Missing {
		t.Error("a path that does not exist was not reported missing")
	}
	if p.Name != "anywhere" {
		t.Errorf("name = %q, want %q", p.Name, "anywhere")
	}
}

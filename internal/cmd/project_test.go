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

func TestDescribeReportsAMissingPath(t *testing.T) {
	p := describe("/no/such/path/anywhere")
	if !p.Missing {
		t.Error("a path that does not exist was not reported missing")
	}
	if p.Slug != "" {
		t.Errorf("slug = %q, want empty for a path that is not there", p.Slug)
	}
}

// A repo with no GitHub origin reports a path and nothing else, rather
// than a slug that would mean nothing.
func TestDescribeOmitsSlugWithoutAGitHubOrigin(t *testing.T) {
	dir := t.TempDir()
	p := describe(dir)
	if p.Missing {
		t.Error("an existing directory was reported missing")
	}
	if p.Slug != "" {
		t.Errorf("slug = %q, want empty", p.Slug)
	}
	if p.Path != dir {
		t.Errorf("path = %q, want %q", p.Path, dir)
	}
}

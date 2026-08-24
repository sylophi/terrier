package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"git@github.com:dittofleet/whatagain.git":   "dittofleet/whatagain",
		"https://github.com/sylophi/terrier.git":    "sylophi/terrier",
		"https://github.com/sylophi/terrier":        "sylophi/terrier",
		"ssh://git@github.com/sylophi/terrier.git":  "sylophi/terrier",
		"git://github.com/sylophi/terrier.git":      "sylophi/terrier",
		"https://GitHub.com/sylophi/terrier.git":    "sylophi/terrier",
		"https://github.com/sylophi/terrier/":       "sylophi/terrier",
		"git@gitlab.com:sylophi/terrier.git":        "",
		"https://notgithub.com/sylophi/terrier.git": "",
		"https://github.com/sylophi":                "",
		"https://github.com/sylophi/terrier/extra":  "",
		"":           "",
		"github.com": "",
	}
	for url, want := range cases {
		if got := Slug(url); got != want {
			t.Errorf("Slug(%q) = %q, want %q", url, got, want)
		}
	}
}

func TestParseOriginURL(t *testing.T) {
	config := `[core]
	repositoryformatversion = 0
[remote "upstream"]
	url = git@github.com:someone/else.git
[remote "origin"]
	url = git@github.com:sylophi/terrier.git
	fetch = +refs/heads/*:refs/remotes/origin/*
[branch "main"]
	remote = origin
`
	if got := parseOriginURL(config); got != "git@github.com:sylophi/terrier.git" {
		t.Errorf("got %q", got)
	}
}

func TestParseOriginURLLastWins(t *testing.T) {
	// git takes the last value, so a url added after the clone is the one
	// in effect.
	config := "[remote \"origin\"]\n\turl = first\n\turl = second\n"
	if got := parseOriginURL(config); got != "second" {
		t.Errorf("got %q, want %q", got, "second")
	}
}

func TestParseOriginURLNoOrigin(t *testing.T) {
	config := "[core]\n\tbare = false\n"
	if got := parseOriginURL(config); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestParseOriginURLIgnoresComments(t *testing.T) {
	config := "[remote \"origin\"]\n# url = commented\n; url = also commented\n\turl = real\n"
	if got := parseOriginURL(config); got != "real" {
		t.Errorf("got %q, want %q", got, "real")
	}
}

func TestIsOriginHeader(t *testing.T) {
	cases := map[string]bool{
		`[remote "origin"]`:   true,
		`[Remote "origin"]`:   true,
		`[remote "Origin"]`:   false, // subsection names are case-sensitive
		`[remote "upstream"]`: false,
		`[core]`:              false,
		`[branch "origin"]`:   false,
	}
	for line, want := range cases {
		if got := isOriginHeader(line); got != want {
			t.Errorf("isOriginHeader(%q) = %v, want %v", line, got, want)
		}
	}
}

func TestFindRootAtTheRepoItself(t *testing.T) {
	dir := t.TempDir()
	mkdirAll(t, filepath.Join(dir, ".git"))
	root, decided := findRoot(dir)
	if !decided || root != dir {
		t.Errorf("findRoot = (%q, %v), want (%q, true)", root, decided, dir)
	}
}

func TestFindRootFromASubdirectory(t *testing.T) {
	dir := t.TempDir()
	mkdirAll(t, filepath.Join(dir, ".git"))
	deep := filepath.Join(dir, "internal", "cmd")
	mkdirAll(t, deep)
	root, decided := findRoot(deep)
	if !decided || root != dir {
		t.Errorf("findRoot = (%q, %v), want (%q, true)", root, decided, dir)
	}
}

// A linked worktree has to resolve to the project it belongs to, which is
// the whole reason terrier can be run from inside one.
func TestFindRootFromALinkedWorktree(t *testing.T) {
	base := t.TempDir()
	main := filepath.Join(base, "project")
	wt := filepath.Join(base, "somewhere-else")
	mkdirAll(t, filepath.Join(main, ".git", "worktrees", "wt"))
	mkdirAll(t, wt)
	writeFile(t, filepath.Join(main, ".git", "worktrees", "wt", "commondir"), "../..\n")
	writeFile(t, filepath.Join(wt, ".git"), "gitdir: "+filepath.Join(main, ".git", "worktrees", "wt")+"\n")

	root, decided := findRoot(wt)
	if !decided || root != main {
		t.Errorf("findRoot = (%q, %v), want (%q, true)", root, decided, main)
	}
}

// Deciding this without git is what keeps `ls` from forking when it runs
// somewhere like $HOME, which is most of the time for a tool shelling out.
func TestFindRootDecidesThereIsNoRepo(t *testing.T) {
	root, decided := findRoot(t.TempDir())
	if !decided || root != "" {
		t.Errorf("findRoot = (%q, %v), want (\"\", true)", root, decided)
	}
}

// A .git this does not understand (a submodule, say) is the one case worth
// handing to git rather than guessing at.
func TestFindRootDefersOnAnUnrecognizedGitFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".git"), "gitdir: /nowhere/in/particular\n")
	if _, decided := findRoot(dir); decided {
		t.Error("an unreadable worktree layout was decided rather than deferred")
	}
}

func TestDefersElsewhere(t *testing.T) {
	cases := map[string]bool{
		"[remote \"origin\"]\n\turl = x\n":               false,
		"[include]\n\tpath = ../shared\n":                true,
		"[includeIf \"gitdir:~/work/\"]\n\tpath = w\n":   true,
		"[url \"git@github.com:\"]\n\tinsteadOf = gh:\n": true,
		"": false,
	}
	for config, want := range cases {
		if got := defersElsewhere(config); got != want {
			t.Errorf("defersElsewhere(%q) = %v, want %v", config, got, want)
		}
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

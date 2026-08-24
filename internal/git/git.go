// Package git derives everything terrier knows about a project from the
// repository itself. The registry stores nothing but paths, so every
// other fact (name, origin URL, GitHub slug) is read fresh here and can
// never drift from what git actually says.
package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ErrNoRepo means the directory is not inside a git work tree, for any of
// the reasons diagnose spells out.
var ErrNoRepo = errors.New("not a git repository")

// Root returns the absolute path of the main worktree of the repository
// containing dir.
//
// A linked worktree reports the project it belongs to rather than itself,
// which is what lets a tool ask "which project is this?" from inside an
// sm worktree and get the same answer as from the primary checkout.
//
// findRoot answers without running anything. Only a .git it does not
// recognize falls through to git itself, which costs a subprocess: this
// is on the path other tools call on every invocation, and paying ~6ms
// for a fork here was more than the rest of the program put together.
func Root(dir string) (string, error) {
	root, decided := findRoot(Canonical(dir))
	switch {
	case decided && root != "":
		return Canonical(root), nil
	case decided:
		return "", fmt.Errorf("%w: %s", ErrNoRepo, describeFailure(dir))
	}
	return rootViaGit(dir)
}

// findRoot walks up from dir looking for a .git entry.
//
// It reports decided=false only when it finds a .git it cannot interpret,
// which is the single case worth handing to git. Reaching the filesystem
// root without finding one is a decision, not a failure, and answering it
// here is what keeps `ls` from forking when it runs somewhere like $HOME.
func findRoot(dir string) (root string, decided bool) {
	for d := dir; ; {
		gitPath := filepath.Join(d, ".git")
		if info, err := os.Lstat(gitPath); err == nil {
			if info.IsDir() {
				return d, true
			}
			if r, ok := rootFromWorktreeFile(gitPath); ok {
				return r, true
			}
			return "", false
		}
		parent := filepath.Dir(d)
		if parent == d {
			return "", true
		}
		d = parent
	}
}

// rootFromWorktreeFile resolves a linked worktree's .git file, which holds
// "gitdir: <path>" pointing into the main repository's .git/worktrees. The
// commondir beside it names the main .git, whose parent is the root.
//
// Anything laid out differently, a submodule for instance, reports false
// so the caller asks git instead of guessing.
func rootFromWorktreeFile(gitFile string) (string, bool) {
	data, err := os.ReadFile(gitFile)
	if err != nil {
		return "", false
	}
	gitdir, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "gitdir:")
	if !ok {
		return "", false
	}
	gitdir = resolveAgainst(filepath.Dir(gitFile), strings.TrimSpace(gitdir))

	data, err = os.ReadFile(filepath.Join(gitdir, "commondir"))
	if err != nil {
		return "", false
	}
	common := resolveAgainst(gitdir, strings.TrimSpace(string(data)))
	if filepath.Base(common) != ".git" {
		return "", false
	}
	return filepath.Dir(common), true
}

// resolveAgainst joins a possibly-relative path git wrote onto the
// directory it was written relative to.
func resolveAgainst(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(base, path))
}

// rootViaGit is the fallback for a repository laid out in some way
// findRoot does not anticipate. It goes through --git-common-dir rather
// than --show-toplevel so a linked worktree still resolves to its project.
func rootViaGit(dir string) (string, error) {
	out, err := run(dir, "rev-parse", "--path-format=absolute", "--git-common-dir", "--is-bare-repository")
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return "", fmt.Errorf("%w: git is not installed", ErrNoRepo)
		}
		return "", fmt.Errorf("%w: %s", ErrNoRepo, describeFailure(dir))
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		return "", fmt.Errorf("%w: unexpected output from git rev-parse", ErrNoRepo)
	}
	if lines[1] == "true" {
		return "", fmt.Errorf("%w: this is a bare repository", ErrNoRepo)
	}
	// The common dir is the main worktree's .git, and its parent is the
	// root. A common dir named anything else belongs to a repository with
	// no main worktree, a bare one with worktrees hung off it, where the
	// parent is some unrelated directory. Registering that would be worse
	// than refusing: every repo beside it would resolve as living inside
	// the project.
	common := filepath.Clean(lines[0])
	if filepath.Base(common) != ".git" {
		return "", fmt.Errorf("%w: this repository has no main worktree (its git directory is %s)", ErrNoRepo, common)
	}
	return Canonical(filepath.Dir(common)), nil
}

// Canonical resolves symlinks so the same repository reached by two paths
// registers once. A path that cannot be resolved (it may not exist yet) is
// returned cleaned and absolute, which is still comparable.
func Canonical(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	return filepath.Clean(path)
}

// describeFailure explains why no repository was found.
func describeFailure(dir string) string {
	if _, err := os.Stat(dir); err != nil {
		return "no such directory: " + dir
	}
	return "no repository at or above " + dir
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// OriginURL returns the `origin` remote URL of the repository rooted at
// root, or "" when it has none.
//
// It reads .git/config directly rather than shelling out to `git remote
// get-url`. Listing every project is terrier's hottest path, the one other
// tools call, and a subprocess per project would dominate it. The registry
// only ever holds main worktree roots, so .git is a real directory here
// and its config is the same file git would read.
//
// Scraping one key out of that file is only safe while the file is the
// whole story. When it defers elsewhere, git is asked, so an answer this
// reports can never quietly disagree with the repository.
func OriginURL(root string) string {
	if path := gitConfigPath(root); path != "" {
		if data, err := os.ReadFile(path); err == nil {
			// Checked before the scraped url is trusted, not after: an
			// insteadOf rule rewrites a url that is present, so returning
			// early on a non-empty result would skip the very case this
			// guards. An empty result here means the repo has no origin.
			if config := string(data); !defersElsewhere(config) {
				return parseOriginURL(config)
			}
		}
	}
	url, err := run(root, "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	return url
}

// gitConfigPath locates the config file for a main worktree, returning ""
// when .git is anything other than the expected directory and the answer
// has to come from git.
func gitConfigPath(root string) string {
	dot := filepath.Join(root, ".git")
	if info, err := os.Stat(dot); err != nil || !info.IsDir() {
		return ""
	}
	return filepath.Join(dot, "config")
}

// defersElsewhere reports whether a config pulls in settings this scraper
// cannot see: an include pointing at another file, or an insteadOf rule
// rewriting the URL that is there. Either means "no origin here" is not
// the same as "no origin", and only git can tell them apart.
func defersElsewhere(config string) bool {
	lower := strings.ToLower(config)
	return strings.Contains(lower, "[include") || strings.Contains(lower, "insteadof")
}

// parseOriginURL pulls url out of [remote "origin"] in git config's INI
// dialect. Later entries win, matching git, which is what makes a url
// added after a clone take effect.
func parseOriginURL(config string) string {
	inOrigin := false
	url := ""
	for _, line := range strings.Split(config, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inOrigin = isOriginHeader(line)
			continue
		}
		if !inOrigin {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found || !strings.EqualFold(strings.TrimSpace(key), "url") {
			continue
		}
		url = strings.Trim(strings.TrimSpace(value), `"`)
	}
	return url
}

// isOriginHeader recognizes [remote "origin"] in the spellings git
// accepts, where the section name is case-insensitive but the subsection
// name is not.
func isOriginHeader(line string) bool {
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]"))
	section, sub, found := strings.Cut(inner, " ")
	if !found || !strings.EqualFold(section, "remote") {
		return false
	}
	return strings.Trim(strings.TrimSpace(sub), `"`) == "origin"
}

// Slug extracts "owner/name" from a GitHub remote URL, in any of the
// forms git accepts: scp-style (git@github.com:owner/name.git), https,
// ssh://, and git://. It returns "" for anything else, so a GitLab remote
// reports a url and no slug rather than a slug that means nothing.
func Slug(url string) string {
	rest := url
	if _, after, found := strings.Cut(rest, "://"); found {
		rest = after
	}
	// Whichever form the URL took, the host ends at the first ':' or '/'.
	// Isolating it rather than searching the whole URL for "github.com"
	// keeps a host like notgithub.com from passing as the real one.
	sep := strings.IndexAny(rest, ":/")
	if sep < 0 {
		return ""
	}
	host, path := rest[:sep], rest[sep+1:]
	if _, after, found := strings.Cut(host, "@"); found {
		host = after
	}
	if !strings.EqualFold(host, "github.com") {
		return ""
	}

	path = strings.TrimSuffix(strings.TrimRight(path, "/"), ".git")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return parts[0] + "/" + parts[1]
}

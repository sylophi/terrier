package update

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sylophi/terrier/internal/release"
)

const (
	checkInterval = 24 * time.Hour
	fetchTimeout  = 2 * time.Second
)

func shouldCheck(version string) bool {
	if version == "dev" {
		return false
	}
	if os.Getenv("CI") != "" {
		return false
	}
	if os.Getenv("TERRIER_NO_UPDATE_CHECK") != "" {
		return false
	}
	// Other tools shell out to terrier constantly, and this check can cost
	// a network round trip that blocks the command from exiting. Stdout is
	// what says whether anyone is reading: the usual shell-out form,
	// `out=$(terrier ls --json)`, captures stdout while leaving stderr
	// attached to the terminal, so testing stderr alone would have let
	// exactly the case this guards against through.
	return isTerminal(os.Stdout) && isTerminal(os.Stderr)
}

// isTerminal reports whether f looks like a terminal rather than a pipe or
// a file. A character device is not proof of one, since /dev/null passes
// too, but nothing is harmed by skipping or running the check on those.
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func parseVersion(s string) []int {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, ".")
	out := make([]int, len(parts))
	for i, p := range parts {
		n, _ := strconv.Atoi(p)
		out[i] = n
	}
	return out
}

func isNewer(latest, current string) bool {
	a := parseVersion(latest)
	b := parseVersion(current)
	for i := 0; i < max(len(a), len(b)); i++ {
		var ai, bi int
		if i < len(a) {
			ai = a[i]
		}
		if i < len(b) {
			bi = b[i]
		}
		if ai > bi {
			return true
		}
		if ai < bi {
			return false
		}
	}
	return false
}

func printHint(latest, current string) {
	fmt.Fprintf(os.Stderr,
		"\nterrier %s is available (current: %s). Run `terrier update` to install.\n",
		latest, current,
	)
}

// MaybeCheck does a daily best-effort GitHub API ping for a newer release
// and prints a stderr hint when one exists. Errors are silently swallowed,
// because this is non-blocking informational output that never fails a
// command.
func MaybeCheck(version string) {
	if !shouldCheck(version) {
		return
	}

	cache := LoadCache()
	now := time.Now().UnixMilli()

	if cache != nil && isNewer(cache.LatestVersion, version) {
		printHint(cache.LatestVersion, version)
	}

	if cache != nil && now-cache.LastCheck < checkInterval.Milliseconds() {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	latest, err := release.FetchLatestTag(ctx)
	if err != nil {
		// Record the attempt anyway, so an unreachable network costs one
		// timeout a day rather than one on every command. Standing in for
		// the unknown latest with the running version keeps the entry
		// valid without ever claiming an update exists.
		latest = version
		if cache != nil {
			latest = cache.LatestVersion
		}
	}
	_ = SaveCache(&Cache{LastCheck: now, LatestVersion: latest})
}

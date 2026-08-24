package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sylophi/terrier/internal/xdg"
)

// writeJSON prints v as indented JSON, which is what every --json flag
// emits. Machines parse it either way, and a human running the same
// command by hand can read it.
func writeJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// tilde shortens a path under the home directory for display. Only
// display: everything terrier stores and prints as a path for scripts to
// consume stays absolute.
func tilde(path string) string {
	home := xdg.Home()
	if home == "/" {
		return path
	}
	if path == home {
		return "~"
	}
	if rest, found := strings.CutPrefix(path, home+string(filepath.Separator)); found {
		return "~" + string(filepath.Separator) + rest
	}
	return path
}

// expandTilde is the inverse, applied to paths a user types.
func expandTilde(path string) string {
	if path == "~" {
		return xdg.Home()
	}
	if rest, found := strings.CutPrefix(path, "~"+string(filepath.Separator)); found {
		return filepath.Join(xdg.Home(), rest)
	}
	return path
}

// confirm asks a yes/no question, defaulting to no.
//
// Nothing here checks for a terminal: stat cannot tell one from
// /dev/null, and an ioctl to find out is more machinery than the question
// deserves. Reaching end of input without an answer says the same thing
// more plainly, and says it as an error, so a script that forgot --yes
// exits non-zero instead of looking like someone declined.
func confirm(prompt string) (bool, error) {
	fmt.Printf("%s [y/N]: ", prompt)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && strings.TrimSpace(line) == "" {
		fmt.Println()
		return false, errors.New("no answer on stdin; pass --yes to skip the confirmation")
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

// plural takes a singular noun with a regular plural, which covers the
// few it is ever called with.
func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// wasWere agrees the verb with a count, for the one message that needs it.
func wasWere(n int) string {
	if n == 1 {
		return "was"
	}
	return "were"
}

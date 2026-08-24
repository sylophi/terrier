package main

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/sylophi/terrier/internal/cmd"
	"github.com/sylophi/terrier/internal/update"
)

var errUnknownCommand = errors.New("unknown command")

var version = "dev"

const usage = `Usage: terrier <command>

Commands:
  add [<path>]          Register a repo, defaulting to the current one
  rm <project>...       Unregister a repo, leaving every file alone
  ls                    List registered projects
  path [<project>]      Print where a project lives, defaulting to this one
  prune                 Unregister projects whose directory is gone
  update                Download and install the latest version
  uninstall [--yes]     Remove the binary, config, and cache
  version               Print the installed version
  help                  Print this help message

Flags:
      --json            (ls, path) Machine-readable output
  -y, --yes             (prune, uninstall) Skip the confirmation
      --                Everything after this is an argument, not a flag

A project is a git repository. Name one by any trailing part of its path:
"whatagain" finds ~/Software/dittofleet/whatagain, and "dittofleet/whatagain"
settles it when two repos share a name.

Terrier records paths and nothing else. Names, origin URLs, and GitHub slugs
are read out of git each time they are asked for, so nothing it reports can
disagree with the repository. Other tools read the registry by running
` + "`terrier ls --json`" + `, and ask which project they are standing in with
` + "`terrier path`" + `, which prints one path and exits non-zero when the
current directory is not a registered project.

` + "`ter`" + ` is installed as a shorter name for the same tool.
`

func printUsage() {
	fmt.Print(usage)
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(0)
	}

	if err := dispatch(args); err != nil {
		if errors.Is(err, errUnknownCommand) {
			// Naming it catches the common slip of putting a flag before
			// the command, where a bare usage dump explains nothing.
			fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n\n", args[0])
			printUsage()
		} else {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
		os.Exit(1)
	}

	// `update` has just talked to the release API, and `uninstall` has
	// deleted the cache directory this would recreate.
	if !slices.Contains([]string{"update", "uninstall"}, args[0]) {
		update.MaybeCheck(version)
	}
}

func dispatch(args []string) error {
	switch args[0] {
	case "add":
		return cmd.Add(args[1:])
	case "rm", "remove":
		return cmd.Remove(args[1:])
	case "ls", "list":
		return cmd.List(args[1:])
	case "path":
		return cmd.Path(args[1:])
	case "prune":
		return cmd.Prune(args[1:])
	case "update":
		return cmd.SelfUpdate(version)
	case "uninstall":
		return cmd.Uninstall(args[1:], version)
	case "version", "--version", "-v":
		fmt.Println(version)
		return nil
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return errUnknownCommand
	}
}

# terrier

One place to register your repos, so every tool already knows them.

Named after the register of land holdings. You add a project once, and anything
that speaks terrier picks it up: no second registration, no per-tool config, no
list of paths to keep in step.

```sh
$ cd ~/Software/dittofleet/whatagain
$ terrier add
Registered dittofleet/whatagain (~/Software/dittofleet/whatagain)

$ terrier ls
  ~/Software/cli/terrier
  ~/Software/dittofleet/port-pool
* ~/Software/dittofleet/whatagain
```

Name a project by any trailing part of its path. Type more of the path when two
repos share a name:

```sh
$ terrier path whatagain
/Users/you/Software/dittofleet/whatagain

$ cd $(terrier path port-pool)
```

`ter` is installed as a shorter name for the same tool.

## What it stores

Paths. That is the whole file:

```json
{
  "schemaVersion": 1,
  "projects": [
    "/Users/you/Software/dittofleet/whatagain"
  ]
}
```

A name, an origin URL, or a default branch kept here would be a copy of
something git already knows, and a copy is a thing that goes stale. So terrier
records the path and reads everything else out of the repository each time it
is asked. Nothing it reports can disagree with the repo.

The one thing it can get out of step with is a repo that moved or was deleted.
Those are listed as `(missing)` rather than hidden, and `terrier prune` drops
them after confirming.

Projects are registered by the path of the **main worktree**, so `terrier add`
and `terrier path` inside a linked worktree resolve to the project it belongs
to, not to the worktree.

## Commands

```
terrier add [<path>]          Register a repo, defaulting to the current one
terrier rm <project>...       Unregister a repo, leaving every file alone
terrier ls                    List registered projects
terrier path [<project>]      Print where a project lives, defaulting to this one
terrier prune                 Unregister projects whose directory is gone
terrier update                Download and install the latest version
terrier uninstall [--yes]     Remove the binary, config, and cache
```

Run `terrier help` for the flags.

## For tools

The CLI is the contract. Read the registry with one command:

```sh
$ terrier ls --json
{
  "projects": [
    {
      "path": "/Users/you/Software/dittofleet/whatagain",
      "name": "whatagain",
      "remote": "https://github.com/dittofleet/whatagain.git",
      "slug": "dittofleet/whatagain",
      "current": true
    }
  ]
}
```

`path` and `name` are always present. `remote` is absent when the repo has no
`origin`, and `slug` is absent unless that origin is a GitHub URL, so a remote
hosted elsewhere reports a URL rather than a slug that would mean nothing. A
repo whose config defers elsewhere, through an `include` or an `insteadOf`
rewrite, is resolved by asking git, so the shortcut never reports something the
repository would disagree with. A
project whose directory is gone carries `"missing": true`. At most one carries
`"current": true`.

To ask which project the user is standing in:

```sh
terrier path              # prints one absolute path, nothing else
terrier path --json       # the same record ls emits
```

`path` exits non-zero when the working directory is not inside a registered
project, so `terrier path >/dev/null || exit` is a complete membership check
and needs no JSON parsing at all.

Do not read the registry file. It is private and its layout can change. The
JSON the CLI prints is what stays stable: fields are only ever added.

`terrier ls --json` is the command tools call most, so it runs no subprocesses
at all, no matter where it is called from. Across 60 registered projects it
takes under 5ms, most of which is process startup. A tool can call it on every
invocation without thinking about it.

Nothing stops a tool from working without terrier, and terrier holds no
per-tool configuration. It answers which projects exist and where they are.
What a tool stores about them stays with that tool.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/sylophi/terrier/main/install.sh | sh
```

Installs the latest release to `~/.local/bin/terrier`, with `ter` alongside it
(override with `TERRIER_INSTALL_DIR`). Supported platforms: macOS (arm64, x64),
Linux (arm64, x64).

## Updating

`terrier` checks once per day for new releases and prints a hint to stderr when
one is available. Run `terrier update` to upgrade in place. The check is
skipped when `CI` or `TERRIER_NO_UPDATE_CHECK` is set, when stdout or stderr is
not a terminal (which covers every time another tool calls it), or when the
binary was built locally.

## Agent skill

`skills/terrier/SKILL.md` tells a coding agent what the registry is and when
registering a repo is the right move, so "add this to terrier" lands without
you spelling out the command.

## License

terrier is licensed under the MIT License. See [LICENSE](LICENSE).

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

Terrier records the path of a repo and nothing else. A name, an origin URL, or
a default branch kept alongside it would be a copy of something git already
knows, and a copy is a thing that goes stale, so everything else is read out of
the repository each time it is asked for. A repo is registered by the path of
its main worktree, so `add` and `path` inside a linked worktree resolve to the
project it belongs to rather than to the worktree. The one thing terrier can be
wrong about is a repo that moved or was deleted, which `ls` shows as
`(missing)` and `terrier prune` drops after confirming.

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
      "slug": "dittofleet/whatagain"
    }
  ]
}
```

`path` is always there. `slug` is the `owner/name` of the repo's GitHub origin,
absent when it has neither. A project whose directory has gone carries
`"missing": true` and nothing else, so skip those or leave them to `prune`.

Nothing else is reported, because nothing else would be worth reading. A name
is the base of the path you already have, and which project the caller is
standing in is a separate question with its own command.

A repo whose config defers elsewhere, through an `include` or an `insteadOf`
rewrite, is resolved by asking git, so the slug never disagrees with what the
repository would say.

### Depending on terrier

A tool that requires terrier should install it, keep it current, and refuse to
run against one it does not understand.

**Install it from your own installer.** Terrier is a single binary with no
runtime, so this is one line, and it is safe to run when terrier is already
there:

```sh
curl -fsSL https://raw.githubusercontent.com/sylophi/terrier/main/install.sh | sh
```

**Update it when you update.** Run the same line from your tool's update path.
An existing registry is picked up as it is, so nothing is lost by reinstalling.

**Check the minor version.** A minor bump means something a tool could be
relying on has changed: a field gone from `--json`, an exit code with new
meaning, a command that answers differently. A patch bump never does. So the
number a tool cares about is the minor one, and a higher one than it was
written for is worth stopping over:

```sh
# The terrier this tool was written against.
want_major=0
want_minor=1

v=$(terrier version) || {
  echo "mytool needs terrier: https://github.com/sylophi/terrier" >&2
  exit 1
}
v=${v#v}
major=${v%%.*}
minor=$(printf %s "$v" | cut -d. -f2)

if [ "$major" -gt "$want_major" ] ||
   { [ "$major" -eq "$want_major" ] && [ "$minor" -gt "$want_minor" ]; }; then
  echo "mytool supports terrier $want_major.$want_minor, but $v is installed." >&2
  echo "Update mytool." >&2
  exit 1
fi
```

A locally built terrier reports `dev` rather than a number. Skip the check when
that is what you get, or the comparison will fail on a build that is fine.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/sylophi/terrier/main/install.sh | sh
```

Installs the latest release to `~/.local/bin/terrier`, with `ter` alongside it
(override with `TERRIER_INSTALL_DIR`). Supported platforms: macOS (arm64, x64),
Linux (arm64, x64).

`terrier update` upgrades in place when you want a newer one. Nothing checks
for updates on its own, and terrier reaches the network only when you run that
command.

## Agent skill

`skills/terrier/SKILL.md` tells a coding agent what the registry is and when
registering a repo is the right move, so "add this to terrier" lands without
you spelling out the command.

## License

terrier is licensed under the MIT License. See [LICENSE](LICENSE).

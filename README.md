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
  "contract": 1,
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

**Check the contract before you rely on it.** `contract` is a number that says
what terrier promises: the shape of what `--json` prints, and the exit codes
commands answer with. It is not the release version. Fields are only ever added
to the JSON, so it does not move when terrier gains a feature or a fix. It
moves only when something a tool could already be relying on changes or goes
away.

That makes a higher number than yours the one case worth stopping for:

```sh
want=1                                              # the contract this tool was written for
have=$(terrier version --json | jq -r .contract) || {
  echo "mytool needs terrier: https://github.com/sylophi/terrier" >&2
  exit 1
}
if [ "$have" -gt "$want" ]; then
  echo "mytool understands terrier contract $want, but terrier reports $have." >&2
  echo "Update mytool." >&2
  exit 1
fi
```

The same number is on every `terrier ls --json`, so a tool that reads the
registry on each run can check it there instead and never spend a second
invocation. If your tool needs something from a contract newer than the
installed terrier, compare the other way for the same reason.

To ask which project the user is standing in:

```sh
terrier path              # prints one absolute path, nothing else
terrier path --json       # the same record ls emits, for that one project
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

`terrier update` upgrades in place when you want a newer one. Nothing checks
for updates on its own, and terrier reaches the network only when you run that
command.

## Agent skill

`skills/terrier/SKILL.md` tells a coding agent what the registry is and when
registering a repo is the right move, so "add this to terrier" lands without
you spelling out the command.

## License

terrier is licensed under the MIT License. See [LICENSE](LICENSE).

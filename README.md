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

### Depending on terrier (tools)

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

It is not recommended to call any other command (apart from management commands like
`update`, etc.) as a dependent tool.

`path` is always there. `slug` is the `owner/name` of the repo's GitHub origin,
absent when it has neither. A project whose directory has gone carries
`"missing": true` and nothing else, so skip those or leave them to `prune`.

Do not read the registry file. It is private and its layout can change. The
JSON the CLI prints is what stays stable: fields are only ever added.

Nothing stops a tool from working without terrier, and terrier holds no
per-tool configuration. It answers which projects exist and where they are.
What a tool stores about them stays with that tool.

#### Tools that are fully dependent on terrier:

A tool that only uses terrier for project management should install it, keep
it current, and refuse to run against one it does not understand.

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
meaning, a command that answers differently. A patch bump never does.

```sh
$ terrier version
v0.1.0
```

Record the major and minor your tool was written against, read them back from
`terrier version`, and stop with a message naming both versions when what you
find is higher. Compare the two components as numbers rather than as text, or
0.10 will read as older than 0.9. A terrier built from source reports `dev`
instead of a version, which has nothing to compare, so decide on purpose
whether that passes or fails in your tool.

## Commands (users)

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

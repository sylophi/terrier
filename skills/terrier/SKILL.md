---
name: terrier
description: The user's registry of repos, shared by their other tools. Use when asked to register a project, or to find out which repos they track and where those live.
---

# terrier: the user's project registry

One list of the repos the user works on. Other tools of theirs read it, so a
repo registered here shows up in all of them at once.

```sh
terrier help
```

## Notes

- Whether a repo belongs on the list is the user's call, and registering one
  changes what their other tools see. Ask first, and do not register a repo
  just because you happen to be working in it.
- `terrier prune` asks for a confirmation you cannot give, so it needs `--yes`
  from you. Only pass it once the user has asked to prune: a path also reads as
  missing when a drive is simply not mounted.
- Terrier holds no per-tool configuration. It answers which repos exist and
  where they are, and nothing about how any tool should treat them.
- `command not found`: tell the user, and do not try to install it yourself.

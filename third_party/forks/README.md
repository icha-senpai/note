# Scribli Go Forks

This directory contains local Scribli-owned Go module forks used by the kernel.

## Modules

| Directory | Module path |
| --- | --- |
| `dataparser` | `github.com/icha-senpai/note/third_party/forks/dataparser` |
| `dejavu` | `github.com/icha-senpai/note/third_party/forks/dejavu` |
| `encryption` | `github.com/icha-senpai/note/third_party/forks/encryption` |
| `eventbus` | `github.com/icha-senpai/note/third_party/forks/eventbus` |
| `filelock` | `github.com/icha-senpai/note/third_party/forks/filelock` |
| `go-sqlite3` | `github.com/icha-senpai/note/third_party/forks/go-sqlite3` |
| `httpclient` | `github.com/icha-senpai/note/third_party/forks/httpclient` |
| `logging` | `github.com/icha-senpai/note/third_party/forks/logging` |
| `riff` | `github.com/icha-senpai/note/third_party/forks/riff` |

## Rules

- Keep module paths under `github.com/icha-senpai/note/third_party/forks/<name>`.
- Keep committed `replace` rules local to this repository.
- Preserve inherited copyright and license notices in copied source files.
- Pull upstream fixes intentionally; do not silently overwrite local Scribli changes.
- Audit network behavior before importing new code into a fork.
- Run `go mod tidy` in the touched fork and in `kernel/` after dependency changes.

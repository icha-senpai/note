# Scribli Internal Naming Cleanup

Scribli is a local-first fork of an upstream AGPL project. The target state is Scribli naming throughout the app, codebase, packaging, docs, and runtime behavior. Some inherited identifiers are easy private implementation names; others cross plugin, API, clipboard, workspace, or module boundaries and need explicit migrations instead of blind replacement.

## Do Not Blindly Rename

These families are not permanent exceptions. They need staged migrations because they can cross process, plugin, storage, or tooling boundaries:

| Identifier family | Why it is sensitive |
| --- | --- |
| Legacy frontend global | Removed from Scribli-owned frontend startup and export HTML. Plugins/snippets should use `window.scribli`; do not reintroduce the old global without a deliberate compatibility decision. |
| Scribli MIME and clipboard markers | MIME, clipboard, export, and paste formats can affect copied content and external integrations. The app writes Scribli clipboard formats now and keeps read fallback only where deliberately tested. |
| `github.com/icha-senpai/note/kernel` | Current Scribli Go module path and import root. Keep new kernel imports under this path. |
| Inherited workspace metadata paths | Workspace and notebook metadata folders. Migrate only with backups, upgrade code, downgrade expectations, and tests. |
| `/api/*` endpoint names inherited from upstream | Public API and plugin compatibility surface. Add Scribli endpoints or payload aliases before removing old names. |
| Plugin event/API names | Third-party plugins can depend on them. Scribli-owned plugin APIs should use Scribli names; old names require an explicit compatibility reason and tests before being kept. |

## Classification

Classify every legacy name before changing it.

| Class | Default action | Examples |
| --- | --- | --- |
| User-visible branding | Rename to Scribli now. | README text, app window titles, installer labels, product metadata, visible settings text, public docs that describe the fork. |
| Network or service identifiers | Remove or disable upstream official service behavior. | Official cloud endpoints, upstream update URLs, upstream marketplace/bazaar URLs, payment/subscription endpoints. |
| Safe private implementation names | Rename only when local and well-tested. | Private helper names, local variables, comments, test names, private filenames that are not imported or serialized. |
| Public plugin/API compatibility names | Use Scribli names by default; keep an old name only for a tested, deliberate compatibility reason. | public API payload keys, plugin APIs, command names. |
| Stored-data compatibility names | Migrate with backup-safe upgrade code and tests. | Workspace metadata paths, `conf.json` keys, notebook metadata paths, snapshot metadata, MIME strings in saved content. |

## Rename Rules

1. Rename user-visible branding and private implementation names immediately when it does not change file formats, APIs, or stored configuration.
2. For internal Electron IPC, renderer constants, and private cross-module names, rename both sides together and verify the frontend build.
3. For public API or plugin identifiers, add a Scribli alias, add tests showing both names work during migration, then remove the old name in a later breaking pass.
4. For stored-data names, keep reading the old name until a migration is designed, implemented, tested, and documented. New writes should move to Scribli names only after the migration exists.
5. For network or official-service names, remove the service call or make provider `0`/default startup do nothing. Do not leave disabled UI that can still trigger upstream requests.
6. Never infer safety from a name alone. Check reads/writes, JSON serialization, route registration, IPC registration, and plugin/docs exposure.
7. When adding an alias, choose explicit names such as `scribli` or `Scribli`, not ambiguous compatibility wrappers.

## Go Module Path

The kernel module path is detached from upstream and owned by this repository:

```text
github.com/icha-senpai/note/kernel
```

New kernel imports should use `github.com/icha-senpai/note/kernel/...`. Do not reintroduce the inherited upstream kernel import root.

## Local Go Forks

The following upstream-owned Go modules have local Scribli forks under `third_party/forks/`:

| Fork | Module path |
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

The kernel resolves these forks with committed local `replace` rules. Do not point them back to inherited upstream module paths unless Boss explicitly asks for an upstream comparison or temporary diagnostic. The sqlite fork owns the `scribli` FTS tokenizer used by kernel search indexes.

## Remaining Upstream-Owned Go Dependencies

Do not fork every remaining inherited upstream dependency solely because of its namespace. For each dependency, classify it first:

| Question | Action |
| --- | --- |
| Does it perform network requests or call official services? | Audit call sites; fork or replace only if Scribli needs different behavior. |
| Is it a general local library? | Keep upstream dependency if maintained and secure. |
| Is it stale, unmaintained, or vulnerable? | Upgrade, replace, or fork based on the specific risk. |
| Does Scribli need independent behavior? | Fork with a clear reason and document the local divergence. |

Temporary local dependency testing for modules that are not already under `third_party/forks/` can use `replace` in `kernel/go.mod`, but do not commit that temporary replace.

## Review Checklist

Before renaming a legacy identifier:

- Search for reads and writes.
- Check whether it crosses process boundaries, network boundaries, storage boundaries, plugin boundaries, or exported document boundaries.
- Check whether it appears in `docs/WORKSPACE.md`, `docs/SY-FORMAT.md`, or `docs/API.md`.
- Add tests for aliases or migrations.
- Update user-visible docs only after behavior matches the wording.
- Run the relevant frontend, kernel, or packaging verification for the touched surface.

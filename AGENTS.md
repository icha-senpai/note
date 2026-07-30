# AGENTS.md

Scribli repository guide. Scribli is a local-first fork of SiYuan. The owned Go module path is `github.com/icha-senpai/note/kernel`. License: AGPL-3.0.

Scribli must present itself as Scribli in public project surfaces and application metadata while preserving inherited upstream copyright notices in source files.

---

## 1. Current Project Shape

Scribli is a Windows desktop knowledge workspace built from:

| Path | Role |
| --- | --- |
| `kernel/` | Go backend, local data engine, HTTP API, sync, search, export, history, AI, MCP, WebDAV/CalDAV/CardDAV, and desktop kernel entrypoint |
| `app/` | TypeScript/Electron/web frontend built by webpack |
| `app/appearance/` | Themes, icons, and i18n resources |
| `app/stage/` | Generated frontend output served by the kernel |
| `scripts/` | Packaging and helper scripts |
| `third_party/forks/` | Local Scribli-owned Go module forks used by the kernel |
| `.github/` | Repository metadata and disabled publishing placeholders |

Native mobile, macOS, and Linux build/package targets are intentionally removed from this repository.

The main README is now the public Scribli surface. Keep it focused on Scribli's local-first behavior, user-controlled sync, AGPL-3.0 license, upstream attribution, and known limitations.

---

## 2. Scribli Product Rules

1. Scribli is local-first. A fresh workspace must not contact official upstream hosted services.
2. Official upstream cloud, account login, account checks, subscriptions, payments, points, quotas, marketplace/bazaar, upstream update checks, and upstream release install flows are removed or disabled.
3. Keep user-controlled features working: S3 sync, WebDAV sync, local-folder sync, local snapshots, repository history, manual backups, import/export, API, MCP, AI provider configuration, plugins, and local server features.
4. Sync provider `0` means disabled/no sync provider and must not perform network sync activity. Do not shift provider IDs unless a migration is implemented and tested.
5. `scribli://` is the only active application protocol. Do not reintroduce the old upstream protocol.
6. Build-time mirrors and upstream publishing are disabled. Do not add `npmmirror`, Aliyun mirrors, upstream Docker publishing, upstream AUR publishing, upstream GitHub release publishing, or auto-update publishing until Scribli owns that pipeline.
7. Electron builder config must keep `publish: null`, and package scripts must keep `--publish=never` unless a Scribli-owned signed release process exists.
8. Do not claim Scribli LLC unless that legal entity exists. Use `Scribli contributors` or another accurate author value.
9. Phase out inherited upstream identifiers in favor of Scribli. Do not use blind global replacement; follow `docs/INTERNAL-NAMING.md`, classify each name, and migrate risky compatibility or stored-data surfaces with tests.
10. Do not reintroduce native mobile, macOS, or Linux build/package targets unless Boss explicitly restarts those platforms as a separate project.

---

## 3. Data, Workspace, And Network Boundaries

Default Windows workspace:

```text
%USERPROFILE%\Documents\Scribli
```

Default Windows app config root:

```text
%USERPROFILE%\.config\scribli
```

Use `util.UserHomeConfDir()`-style helpers for home config paths. Mixed Scribli and legacy upstream config roots can break startup before Electron obtains the kernel port.

Network-capable behavior must be explicit and user-controlled. Acceptable examples include S3/WebDAV endpoints configured by the user, user-supplied URLs for import/fetch tools, user-configured AI providers, and local API/MCP/plugin/server features. Any new network behavior must be documented and must not call upstream official services by default.

---

## 4. Required Toolchain

| Tool | Version source |
| --- | --- |
| Go | `kernel/go.mod` |
| Node and pnpm | `app/package.json` |
| Windows C compiler | Required for CGO kernel tests/builds; this checkout has used repo-local `..\.tools\w64devkit\bin\gcc.exe` |

Prefer repo-local Go caches for Windows work:

```powershell
$env:GOPATH=(Resolve-Path "..\.tools").Path + "\go-path"
$env:GOCACHE=(Resolve-Path "..\.tools").Path + "\go-build-cache"
$env:GOMODCACHE=(Resolve-Path "..\.tools").Path + "\go-mod-cache"
```

---

## 5. Verification Commands

Frontend checks:

```powershell
cd app
pnpm run lint
pnpm test
```

Frontend production build:

```powershell
cd app
pnpm build
```

Windows package:

```powershell
cd app
pnpm run dist
```

Kernel tests on Windows with `w64devkit`:

```powershell
cd kernel
$env:PATH=(Resolve-Path "..\.tools\w64devkit\bin").Path + ";" + $env:PATH
$env:GOPATH=(Resolve-Path "..\.tools").Path + "\go-path"
$env:GOCACHE=(Resolve-Path "..\.tools").Path + "\go-build-cache"
$env:GOMODCACHE=(Resolve-Path "..\.tools").Path + "\go-mod-cache"
$env:CGO_ENABLED="1"
$env:CC=(Resolve-Path "..\.tools\w64devkit\bin\gcc.exe").Path
go test ./...
```

Run the checks that match the change. For docs-only changes, grep/status review is enough. For frontend code, run `pnpm run lint`; run `pnpm build` when production bundles or packaging readiness matter. For Go changes, run focused or full `go test ./...` as appropriate.

Do not call work verified if a build or test command failed because of a local conflict. Report the conflict and the exact command that failed.

---

## 6. Do Not Hand-Edit

- `app/stage/protyle/js/lute/lute.min.js`
- `app/stage/build/**`
- `app/src/types/dist/**`
- `app/changelogs/**`
- `app/kernel/Scribli-Kernel*`
- `*.syso`
- `kernel/kernel.aar`
- `app/pandoc/pandoc-windows-amd64.zip`
- downloaded dependency caches under `.tools/` or `app/node_modules/`

Generated build outputs may change when running build/package commands, but do not manually edit them.

---

## 7. Coding Conventions

1. Prefer existing repo patterns over new abstractions.
2. Keep changes scoped to the requested behavior.
3. Use structured APIs/parsers instead of ad hoc string edits when practical.
4. TypeScript/JavaScript: semicolons, double quotes, spaces.
5. Go: run `gofmt` after editing.
6. Markdown: keep paragraphs and table rows on single lines; do not hand-wrap prose.
7. Comments should be useful, short, and written in English unless surrounding code clearly requires another language.
8. Do not add visible in-app marketing copy to explain controls or features. Build the actual usable surface.
9. Do not hand-write SVG icons for UI work; use existing icons from `app/appearance/icons/litheness/icon.js` when possible.

---

## 8. i18n And Public Text

1. New user-facing strings belong in i18n resources when they appear in the app.
2. Keep language keys complete across active language files when adding keys.
3. Use three ASCII periods (`...`) for ellipses in localized strings.
4. Do not add upstream hosted domains, upstream social links, upstream pricing links, or claims that Scribli provides official cloud services.
5. Public README/docs must clearly state that Scribli is based on SiYuan and remains AGPL-3.0 licensed.
6. Do not remove inherited upstream copyright notices from source files.

---

## 9. Git And Release Safety

1. Never run `git commit` or `git push` unless Boss explicitly asks.
2. Do not revert user changes. If the worktree is dirty, inspect enough to avoid overwriting unrelated edits.
3. Do not add upstream release, Docker, AUR, app-store, signing, payment, subscription, or cloud-service publishing.
4. Keep `.github/workflows/*` publishing workflows disabled until Scribli has its own release pipeline.
5. Keep runtime update checks disabled unless Scribli has a signed release channel and Boss asks to enable it.

---

## 10. Upstream And Dependencies

Scribli still inherits core architecture and dependencies from SiYuan. The kernel module has been detached to `github.com/icha-senpai/note/kernel`.

The following upstream-owned Go modules are copied into this repository as local Scribli fork modules under `third_party/forks/` and resolved by committed local `replace` rules:

- `dataparser`
- `dejavu`
- `encryption`
- `eventbus`
- `filelock`
- `go-sqlite3`
- `httpclient`
- `logging`
- `riff`

Do not point these modules back to upstream module paths. If a fork needs upstream fixes, pull or cherry-pick intentionally into the matching `third_party/forks/<name>` directory and preserve upstream copyright notices. The sqlite fork owns the `scribli` FTS tokenizer used by kernel search indexes.

Rebuilding `lute.min.js` requires the upstream Lute build process; do not edit the generated file in this repo.

Do not fork every remaining upstream dependency just because of its namespace. First determine whether it performs network requests, whether it is a general local library, whether it is maintained and secure, and whether Scribli needs independent behavior. Fork only for a concrete behavior, security, or service-coupling reason.

---

## 11. Internal Naming Cleanup

Use `docs/INTERNAL-NAMING.md` as the policy for phasing out legacy names. In short:

- Rename user-visible branding and private implementation names to Scribli.
- Remove or disable upstream official service identifiers and calls.
- Migrate Electron IPC channels, internal constants, and private helpers to Scribli names when both sides live in this repository.
- Stored workspace names such as the inherited workspace metadata directory, database filenames, and encrypted database names must be migrated with backup-safe, tested data migrations.
- Public compatibility identifiers such as inherited MIME types, plugin APIs, and public API fields are temporary legacy surfaces. Replace them with Scribli names only with aliases, deprecation notes, and tests first, then remove the old names in a later breaking pass when Boss approves.
- Keep new Go imports under `github.com/icha-senpai/note/kernel` or `github.com/icha-senpai/note/third_party/forks/<name>`.

---

## 12. Response Style

Match Boss's language and keep answers direct. Explain root cause, what changed, what was verified, and what remains uncertain. Do not pretend a build, install, package, or runtime smoke test happened unless it actually did.

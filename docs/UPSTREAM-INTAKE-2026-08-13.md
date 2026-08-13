# Upstream Intake Audit - 2026-08-13

This note audits the gap between Scribli `master` and upstream SiYuan `master` as of 2026-08-13.

## Snapshot

| Item | Value |
| --- | --- |
| Scribli branch | `master` |
| Scribli commit | `2979979a82aa7f9100bd5a980b40b5eea8f238f8` |
| Upstream ref checked | `siyuan/master` |
| Upstream commit | `251596fc0de2f9528c00c224252fd073a99973f4` |
| Merge base | `eef10568384e2e7cf547adb029ae46a72e43c287` |
| Upstream commits ahead | `1404` |
| Scribli commits ahead | `31` |
| Diff size | `2989 files changed, 677702 insertions(+), 868271 deletions(-)` |
| Upstream release range seen | `v3.7.4-alpha.1` through `v3.8.0` |

The upstream range is too large and too different from Scribli's product boundary for a raw merge. Most useful changes should be intake-reviewed as small slices, with local-first behavior and Scribli-owned forks preserved.

## Non-Negotiable Guardrails

- Do not raw merge upstream `master`.
- Do not reintroduce upstream official cloud, accounts, subscriptions, payments, points, quotas, marketplace/bazaar, or hosted update behavior.
- Do not reintroduce native mobile, macOS, Linux packaging, upstream Docker publishing, AUR publishing, or upstream GitHub release publishing.
- Keep Electron publishing disabled with `publish: null` and package scripts using `--publish=never`.
- Keep `scribli://` as the active application protocol.
- Preserve Scribli's local Go module path and local forks under `third_party/forks/`.
- Treat dependency changes involving `dejavu`, `go-sqlite3`, `httpclient`, `lute`, and other Scribli-owned forks as hand-port work, not direct `go.mod` replacement.
- Keep new network behavior explicit and user-controlled.

## What Looks Useful

### P0 Security Hardening

These are the highest-value intake candidates because they protect local workspaces, APIs, MCP tools, publish surfaces, templates, assets, and auth paths.

| Commit | Why It Matters | Intake Style |
| --- | --- | --- |
| `a8d2cc4e4` | Fixes SSRF bypass via IPv6 transition addresses in host checks. | Cherry-pick or hand-port with tests around `kernel/util/httprequest.go` and `kernel/util/net.go`. |
| `9dd4b4a1f` | Restricts secret interpolation to allowed hosts. | Cherry-pick if the surrounding AI/provider code still matches; otherwise hand-port policy. |
| `3542530c5` | Blocks sensitive workspace files in the MCP file tool. | Include in the MCP hardening slice. |
| `88c74d9ae` | Validates content template paths. | Cherry-pick or hand-port; local-first safe. |
| `83877148f` | Restricts template removal to the templates directory. | Cherry-pick or hand-port; local-first safe. |
| `991d693c9` | Escapes file names and AI provider messages in UI messages. | Cherry-pick frontend pieces carefully and keep English i18n complete. |
| `2c3b7af3f` | Restricts `/api/lute/spinBlockDOM` to admin and adds input size limits. | Cherry-pick or hand-port; good API boundary fix. |
| `90f74c8d7` | Requires explicit pprof enablement and admin auth. | Cherry-pick if Scribli still exposes pprof. |
| `8e85058a3` | Prevents stored XSS through script-capable assets. | Cherry-pick or hand-port; test file serving behavior. |
| `39ed33479` | Prevents elevated command injection in Windows Defender exclusion flow. | Cherry-pick if that Windows flow exists in Scribli. |
| `cb67e0b4f` | Validates WebSocket origin. | Cherry-pick or hand-port; local app should still allow expected desktop origins. |
| `50b3ee7c7` | Throttles API token authentication attempts. | Cherry-pick kernel logic; ignore unrelated localization churn except active English strings. |
| `0aba2a831` | Cleans up asset path validation. | Cherry-pick or hand-port with file-serving tests. |
| `0a176345e` | Restricts SQL execution in live templates. | Cherry-pick if live templates are present. |
| `9c16e9851` | Prevents localhost auth bypass through fixed-port proxy. | Cherry-pick or hand-port; important for desktop-local threat model. |
| `1a5b3431d` | Parameterizes backlink mention search. | Cherry-pick or hand-port; database safety. |
| `7d273c271` | Secures backlink refresh API. | Cherry-pick or hand-port; API boundary fix. |

There are also multiple encrypted-notebook boundary commits, asset-viewer stored-XSS commits, and publish-auth hardening commits in the range. Publish-specific fixes should be ported only if Scribli keeps a local publish/share mode.

### P1 MCP And Agent Bridge

This is the most strategic feature area for Scribli if the goal is a ChatGPT/Codex-style bridge into notes. Upstream has meaningful MCP and Agent work, but it should be ported as a coherent bundle rather than one-off commits.

| Commit | Why It Matters | Intake Style |
| --- | --- | --- |
| `3083e4106` | Adds support for the MCP `2026-07-28` spec. | Port as part of a full MCP compatibility slice. |
| `943e5eb03` | Strengthens MCP tool schema validation. | Port with execution tests. |
| `d992ed1cd` | Fixes MCP tool validation behavior. | Port with `943e5eb03`. |
| `f5e9c0c56` | Hardens MCP tool execution/sync behavior. | Port with MCP compatibility slice. |
| `e6d87fe46` | Adds stdio MCP environment variable support. | Port if Scribli wants local tool servers. |
| `525ba6076` | Fixes MCP environment variable editing. | Port with `e6d87fe46`. |
| `e0a6ea804` | Preserves MCP request scopes for legacy protocols. | Port if older clients remain supported. |
| `a891150ec` | Fixes MCP output schema serialization. | Port with MCP validation work. |
| `86a2269f4` | Returns complete database data in MCP and CLI surfaces. | Port only after checking privacy and output-size boundaries. |
| `a798512dc`, `313be336a`, `6953bd7af`, `ace2790ed`, `1e3c436a0`, `9841bb087` | Add Agent/MCP capability exposure and approval controls. | Recreate in Scribli terms if upstream UI brings cloud/account assumptions. |

Scribli should prefer a local explicit approval model: the app exposes local MCP tools, the user grants scopes, and ChatGPT/Codex connects through an MCP client/server boundary. Avoid any path that depends on an upstream hosted account.

### P1 Encrypted Notebook Safety

If encrypted notebooks remain a supported Scribli feature, this set deserves a focused safety pass.

| Commit | Why It Matters | Intake Style |
| --- | --- | --- |
| `b501a3893` | Makes encrypted notebook `.sy` archives portable. | Hand-port/cherry-pick with export/import regression tests. |
| `2cd475c67` | Exports encrypted notebook assets. | Port with archive portability. |
| `b8e0af8dc` | Restores encrypted notebook config after sync. | Port if encrypted sync is supported. |
| `fcf1ab069`, `0c25b9c91`, `8e9c46082`, `9bbabe012`, `9726f5d23`, `e398518f4`, `b3ccfa8b7`, `37820739f`, `77850d5bc`, `b3bc35d6d`, `3068095cb` | Tighten encrypted notebook boundaries, recovery, formats, and device-local mount state. | Port as a tested bundle, not piecemeal. |
| `a25c2dd06` | Prevents publishing encrypted notebooks. | Port if any publish/share mode remains. |
| `8fb1b5766`, `3bc014c7d` | Restrict encrypted notebook status and key material surfaces. | Port with API tests. |
| `cb1a23fa5` | Prevents rollback from deleting data. | Port with rollback/snapshot tests. |

### P2 Editor, Search, Database, And Graph Wins

These are likely user-visible quality improvements. They need more conflict review because upstream frontend moved heavily.

| Commit Or Area | Why It Matters | Intake Style |
| --- | --- | --- |
| `eba40a7ee` | Exact alias matches in search. | Small, likely safe cherry-pick with search test. |
| `15da28cd1` | Dynamic document loading improvements. | Recreate or cherry-pick as a feature bundle after frontend inspection. |
| `f2814ce27`, `544555b5f` | Database table paste performance. | Port with AV/database tests and large-paste manual check. |
| Table editing bundle around issue `18489` | Selection, resizing, header, alignment, and table-control polish. | Recreate/cherry-pick in slices; high UX value but broad frontend touch. |
| Database selection bundle around issues `11975`, `11742`, `14978` | Range selection, multi-selection toolbar, drag selection. | Bundle with database UI tests where practical. |
| `f08548531`, `cab0ab994`, `c775528bb` | Graph display/performance/capacity improvements. | Port if Scribli graph behavior matches upstream. |
| `b9a100d01` | Lute preserves whitespace in table cells. | Hand-port through Scribli's local `lute` fork and regenerate generated JS through the approved script. |
| `6420c28ca` | Lute support for Vditor Callout. | Hand-port through local `lute` fork if wanted. |

### P2 Sync And Storage

Scribli should keep user-controlled S3, WebDAV, local-folder sync, snapshots, and history working. Upstream has sync performance and correctness fixes, but anything tied to upstream cloud must be skipped.

| Commit | Why It Matters | Intake Style |
| --- | --- | --- |
| `31d615ce2` | DejaVu upgrade for S3 reference pagination and tag index uploads. | Hand-port into `third_party/forks/dejavu`; do not replace Scribli fork through `go.mod`. |
| `5dcf82db1` | Reindexes startup sync data before merge. | Cherry-pick or hand-port if sync code matches. |
| `3ead6eeb9` | Reduces unchanged startup sync requests. | Cherry-pick or hand-port after confirming no upstream cloud dependency. |
| `3042f4e95` | Improves data synchronization performance. | Inspect carefully; likely useful if provider-neutral. |
| `f08016571` | Prevents LAN sync preflight panic. | Port only if Scribli keeps LAN sync. |
| `5752c1b95`, `7c40eecb4` | DejaVu LAN sync improvements. | Skip unless LAN sync is explicitly in scope. |

### P3 Import, Export, And Polish

These can be valuable after the safety and bridge work.

| Commit Or Area | Why It Matters | Intake Style |
| --- | --- | --- |
| `54c5a27ce` | PDF rectangle annotation interaction improvements. | Port if current PDF annotation UX is kept. |
| `e2d33f29e` | Export preview copying improvements. | Inspect for hosted publishing assumptions before porting. |
| PDF navigation/history changes | Improves reading and review workflows. | Port as a focused PDF UX slice. |
| Icon, cover, and UI polish | Can improve feel. | Recreate with Scribli assets and branding, not upstream public service links. |

## What To Skip Or Quarantine

- Upstream account login, OIDC, subscription, payment, points, quota, and official cloud flows.
- Remote Bazaar/marketplace discovery, rankings, downloads, package promotion, or hosted package management.
- Upstream update channel and release installer flow unless Scribli later owns a signed release channel.
- `.github/workflows/*` publishing workflows, Docker publishing, AUR publishing, app-store style packaging, and release automation.
- Native mobile, macOS, Linux, Android, iOS, and HarmonyOS targets.
- Bulk i18n churn for languages Scribli does not currently ship.
- Upstream branding, logos, protocol restoration, public social links, and official service URLs.

One exception: local package installation from user-supplied `.zip` files may be kept or improved, but it must remain local/user-controlled and should not restore the remote Bazaar.

## Dependency And Fork Warning

The upstream diff removes Scribli's local `third_party/forks/` shape and changes `go.work`/`go.mod` behavior heavily. Do not accept that mechanically.

For dependencies already forked by Scribli:

- Inspect the upstream dependency change.
- Port only the concrete fix into `third_party/forks/<name>`.
- Preserve Scribli module paths and copyright notices.
- Run focused tests against the fork and then kernel tests.
- Regenerate generated frontend assets only through approved scripts such as `scripts/regenerate-lute-js.ps1`.

This matters especially for `dejavu`, `go-sqlite3`, `httpclient`, and `lute`.

## Suggested Intake Order

### Slice 1: Security Foundation

Port the smallest security fixes first:

- `a8d2cc4e4`
- `cb67e0b4f`
- `50b3ee7c7`
- `2c3b7af3f`
- `0aba2a831`
- `88c74d9ae`
- `83877148f`
- `991d693c9`
- publish-access fixes only if Scribli keeps publish/share mode

Verification:

```powershell
cd kernel
$env:PATH=(Resolve-Path "..\.tools\w64devkit\bin").Path + ";" + $env:PATH
$env:GOPATH=(Resolve-Path "..\.tools").Path + "\go-path"
$env:GOCACHE=(Resolve-Path "..\.tools").Path + "\go-build-cache"
$env:GOMODCACHE=(Resolve-Path "..\.tools").Path + "\go-mod-cache"
$env:APPDATA=(Resolve-Path "..\.tools\appdata").Path
go telemetry off
$env:CGO_ENABLED="1"
$env:CC=(Resolve-Path "..\.tools\w64devkit\bin\gcc.exe").Path
go test ./...
```

Run `cd app; pnpm run lint` and `pnpm build` if frontend escaping or API/UI strings are touched.

### Slice 2: Local MCP And Agent Bridge

Port/recreate MCP compatibility and hardening:

- `3083e4106`
- `943e5eb03`
- `d992ed1cd`
- `f5e9c0c56`
- `3542530c5`
- `e0a6ea804`
- `a891150ec`
- `e6d87fe46`
- `525ba6076`
- capability/approval commits around `a798512dc`, `313be336a`, `6953bd7af`, `ace2790ed`, `1e3c436a0`, `9841bb087`

Design rule: Scribli's bridge should be local, explicit, and scoped. The safe product shape is a local MCP server or plugin surface that lets a user connect ChatGPT/Codex-style clients to selected notebooks/tools without requiring an upstream account.

### Slice 3: Encrypted Notebook Safety

Port encrypted notebook boundary, export/import, and recovery fixes as one tested feature slice:

- `b501a3893`
- `2cd475c67`
- `b8e0af8dc`
- `fcf1ab069`
- `9726f5d23`
- `37820739f`
- `77850d5bc`
- `3068095cb`
- `cb1a23fa5`
- related status/key material restrictions

Verification should include export/import of encrypted archives, sync restoration if supported, and rollback/history behavior.

### Slice 4: Search And Editor Wins

Start with low-risk, high-signal changes:

- `eba40a7ee` exact alias search.
- Clipboard/caret fixes that are small and isolated.
- Database paste performance only after AV code inspection.
- Table-selection and dynamic document-loading bundles after frontend conflict review.

### Slice 5: Sync Provider Improvements

Port provider-neutral S3/WebDAV/local-folder fixes. Hand-port DejaVu changes into the local fork. Leave LAN sync and upstream cloud paths out unless they become explicit Scribli scope.

## Mechanical Conflict Probe Notes

The following commits were probed with `git merge-tree <commit>^ master <commit>` and did not show textual conflicts in that command:

- `a8d2cc4e4`
- `3542530c5`
- `50b3ee7c7`
- `cb67e0b4f`
- `3083e4106`
- `eba40a7ee`
- `b501a3893`
- `f2814ce27`
- `15da28cd1`
- `31d615ce2`
- `7da1ea3e9`
- `4fd2945bc`

This is not semantic verification. It only means these commits may be mechanically easy to apply. Every intake still needs diff review, Scribli-scope cleanup, and tests.

## Safe Cherry-Pick Pattern

Use this pattern for a focused commit:

```powershell
git switch -c intake/security-ssrf
git cherry-pick -n <commit>
git diff --check
git diff --stat
git diff
```

Then remove unrelated upstream service, platform, branding, or bulk localization changes before testing. Commit only after the slice passes review and verification.

For broad features, prefer recreating the behavior in Scribli's current architecture or cherry-picking a small topological bundle with `--no-commit`, then inspecting the combined diff before keeping it.

## Open Product Questions

- Does Scribli still intend to ship any publish/share mode? If yes, the publish security fixes matter even if official hosted publishing stays removed.
- Are encrypted notebooks part of the release promise? If yes, the encrypted notebook slice should be early.
- Should local plugin/theme package installation remain supported without remote Bazaar discovery?
- Is LAN sync in scope, or should sync work stay limited to S3, WebDAV, and local-folder flows?
- For ChatGPT/Codex integration, should Scribli expose a built-in MCP server, generate client config snippets, or both?

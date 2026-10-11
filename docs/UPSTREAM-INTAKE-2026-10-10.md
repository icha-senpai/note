# Upstream Intake Review - 2026-10-10

This review compares Scribli 4.0.9 with its verified GitHub parent, [siyuan-note/siyuan](https://github.com/siyuan-note/siyuan). It updates the August shortlist using current source rather than treating older recommendations as unimplemented work. Scribli is based on SiYuan and remains AGPL-3.0 licensed.

## Evidence and limits

| Snapshot | Commit |
| --- | --- |
| Scribli reviewed | `769eef999543a6d823cc3a9108369e84e3a88232` |
| Upstream master fetched | `fb3355f07d22bff36f1d785999fd5ded6a90d468` |
| Upstream dev fetched | `3f143e335023594f0bd6d369655977b157c5c034` |
| Shared ancestor | `eef10568384e2e7cf547adb029ae46a72e43c287` |

GitHub's repository API confirms both the fork parent and source are SiYuan. The stable release page at review time identifies [v3.8.6](https://github.com/siyuan-note/siyuan/releases/tag/v3.8.6); dev contains subsequent 3.8.7 alpha work. The fetched dev tip is alpha.8. Master and dev were fetched as remote-tracking refs. The approved intake was subsequently adapted into Scribli as described below; neither upstream branch was merged.

The branches have 45 Scribli-only and 3,899 upstream-only commits relative to their common ancestor. Those numbers describe history divergence, not missing features. A whole-branch merge would bring unrelated product behavior and extensive structural changes.

The review tables record the original Scribli 4.0.9 baseline, not the final implementation status. This is a review of selected candidates, not an exhaustive security audit or a runtime playtest of upstream. Effort labels below are relative assessments rather than estimates in hours.

## Implemented intake

Boss approved seven items: the two data safeguards, frequent slash commands, workspace storage statistics, attachment metadata search, calendar views, and conditional colors. List views and the remaining medium-priority or larger projects stay deferred.

| Item | Result |
| --- | --- |
| Document-read safety | Document property reads distinguish actual invalid document properties from I/O failures. Temporary sharing violations and disappearing files do not move documents into corruption backups. Genuine corruption handling remains. |
| Database cleanup safety | Document, directory, template, and relation scan failures abort classification and deletion. Both single and bulk cleanup refuse incomplete scans; API, MCP, and CLI callers receive the error. Template reference scanning is streamed, and relation traversal tracks candidates independently. |
| Frequent slash commands | The five most-used eligible commands appear first. Usage stores command IDs and counts in workspace-scoped local storage. The Editor settings switch uses the existing settings save queue. |
| Workspace storage | App settings has a manual-refresh breakdown of data, attachments, repository, history, temporary files, configuration, and other files. The endpoint reads file metadata locally, coalesces concurrent scans, and reports calculation time. Attachments are a subset of data, not an additional total. |
| Attachment search | Filename/path ranking now also includes image alt/title metadata and existing local OCR text. No new OCR engine, model download, or hosted service was added. |
| Calendar | A month view uses existing database fields, filters, and ordered rows. It includes later table pages, navigation, a date field selector, undated entries, drag scheduling, a schedule dialog, date-range moves, row preview, search, new entries, and undo. |
| Conditional colors | Per-view ordered rules color table rows, gallery cards, Kanban cards, and calendar entries. The first matching enabled rule wins. The editor supports field conditions and relative today conditions, enable/disable, color selection, reordering, removal, save, and undo. Advanced API-created conditions are preserved rather than simplified by the editor. |

Calendar and colors were implemented through Scribli's existing table-backed database architecture, transaction dispatch, and renderers. The implementation does not import upstream's newer settings/API framework or database spec 10/13. Existing Kanban controls, local-first behavior, owned module paths, and platform restrictions remain.

### Stored-format compatibility

Ordinary databases remain spec 5. A database using a calendar layout, conditional colors, or custom category colors is saved as Scribli spec 6, and it is not automatically downgraded when those features are removed. Existing spec 5 databases are supported by this build. Older Scribli 4.0.9 builds cannot open spec 6 databases; use this build or a later compatible build after using these features. The barrier prevents older builds from silently dropping the new fields. No bulk workspace migration is performed. The later shared color picker and interface overrides are documented in [COLOR-SUPPORT.md](COLOR-SUPPORT.md).

### Scope and limits

Calendar is a month grid, with horizontal scrolling in narrow windows. It uses all filtered rows rather than table pagination, so very large calendars can cost more to render. Grouping configuration is retained for other layouts but is not applied to the month grid. Created/updated fields are usable as read-only calendar dates; manual scheduling uses editable date fields. The rule editor creates single-field conditions; grouped and other advanced conditions supplied through transactions remain intact and can still be enabled, recolored, reordered, or removed. Attachment metadata search follows the existing global index and does not create a decrypted index for encrypted notebooks.

### Validation

- `pnpm run lint`: passed with zero errors and 16 existing unused-variable warnings.
- `pnpm test`: all 77 tests passed, including frequent-command ranking and calendar date/range behavior across daylight saving time.
- `pnpm build`: application, Electron desktop, and export bundles passed; existing webpack bundle-size warnings remain.
- `go build -tags 'fts5 sqlcipher' -o ../app/kernel/Scribli-Kernel.exe .`: Windows kernel build passed.
- `go test -tags 'fts5 sqlcipher' ./...`: all kernel packages passed using the Windows CGO compiler. New coverage includes a real Windows sharing violation and successful read after unlocking, incomplete cleanup scans, corrupt relation metadata, streamed template boundaries, storage measurement, attachment metadata, spec compatibility, cloned configuration, conditional color precedence/validation, and later-page calendar entries.
- Isolated Electron checks exercised calendar navigation, drag scheduling, scheduling undo, undated/new entries, row preview, view switching, search retention, conditional-color editing/undo, narrow-window scrolling, storage refresh, attachment alt/title search, and frequent slash ranking. Table, gallery, Kanban, and calendar tints were each checked; an API-created grouped condition survived opening and saving the rule editor unchanged.

Runtime fixtures were created only in an isolated repository QA workspace. The installed Scribli application and the user's normal workspace were not changed. This intake does not include a version bump, installer package, installation, commit, or push.

The follow-up formatting review fixed invisible color-rule move/delete buttons by using the existing always-visible icon style, and hid misleading field/operator controls for preserved advanced conditions. Desktop and narrow-window Electron screenshots confirmed the controls and row layout, followed by successful lint and production builds. The one-pixel image-search fixture paragraphs were removed from the isolated demonstration document after their metadata search checks because their captions appeared as narrow vertical text strips.

## Recommended order

1. Port the document-read and cleanup safeguards with regression tests.
2. Ship smaller daily-use improvements: frequent slash items, storage statistics, and attachment metadata search.
3. Add calendar/list views and conditional colors as a dedicated database release, preserving Scribli's new Kanban controls and deletion undo.
4. Improve snapshot filtering and large Markdown table handling.
5. Consider local OCR provider selection, independent settings windows, richer mind maps, and automations as separate feature projects.

## First: preserve workspace data

| Candidate | Source evidence in Scribli | Upstream source | Intake |
| --- | --- | --- | --- |
| Abort database cleanup when reference scanning is incomplete | `kernel/model/attribute_view.go`: `UnusedAttributeViews` continues after `loadTree` failures, and bulk removal consumes that incomplete result. Definitions referenced only by unreadable documents can therefore be classified as unused. | [5a9b724443](https://github.com/siyuan-note/siyuan/commit/5a9b7244435841a256511ec16712983ebd7ee78f) | Hand-port error propagation through scan, API, MCP, CLI, and cleanup callers. Test unreadable documents/directories, locked encrypted notebooks, and no removal on failure. |
| Distinguish a temporary file-read error from document corruption | `kernel/model/file.go`: `docIAL` treats empty properties as corruption and moves normal-notebook data; `filesys.DocIAL` does not provide the error distinction used by the new fix. | [ce2e76ee4b](https://github.com/siyuan-note/siyuan/commit/ce2e76ee4b12009d6a868ffb430cb7a2ce6bff32) | Port read-error classification and the earlier concurrent-deletion safeguard. Keep genuine corruption backups. Test Windows sharing violations and successful reads after unlocking. |

These are identifiable source-level gaps. Their failure cases were not reproduced against the user's workspace during this review.

Also review [15339acf8f: attachment copying into encrypted notebooks](https://github.com/siyuan-note/siyuan/commit/15339acf8ff2b424126a8ac717fd8db1ecac1f2e). The upstream bundle copies normal attachments into the encrypted notebook, encrypts their content and names, rewrites references, includes PDF annotations, and cleans up failed batches. Scribli does not contain its new `PrepareClipboardPasteAssets` implementation; a complete tracing of Scribli's current paste paths is required before declaring the same behavior defective.

## Best user-facing candidates

| Feature | What it adds | Verified Scribli baseline | Upstream evidence | Priority / effort |
| --- | --- | --- | --- | --- |
| Calendar database view | Plan dated entries in a calendar, schedule undated entries, and navigate weeks. | `kernel/av/av.go` supports table/gallery/Kanban layouts; calendar layout and renderer are absent. | [Calendar](https://github.com/siyuan-note/siyuan/issues/13740), initial `dbc265d7fc`; subsequent scheduling and persistence fixes are in master. | High / large |
| List database view | A compact alternative for tasks and note collections. | No list layout in the current AV types. | [List view](https://github.com/siyuan-note/siyuan/issues/17743), `59e7109ae6`. | Medium / medium-large |
| Conditional colors | Rules can color entries according to field values; an overdue task rule would be a useful Scribli application. | No `conditionalColors` field or rule evaluator in Scribli's AV model. | [Conditional colors](https://github.com/siyuan-note/siyuan/issues/19889), `052483d1de` plus follow-up fixes, in master. | High / medium-large |
| Frequent slash commands | Put the five most-used eligible insert commands at the top, recording command IDs and counts locally. | Scribli lacks the `frequentSlash` ranking/storage modules. | [Frequent items](https://github.com/siyuan-note/siyuan/issues/20240), `2031e1637d`, dev. | High / small-medium |
| Workspace storage breakdown | See data, attachments, snapshots/repository, history, and temporary storage sizes, with manual refresh. | No workspace storage statistics endpoint or settings panel found. | [Storage statistics](https://github.com/siyuan-note/siyuan/issues/19790), `2f8f32fb92`, in master. | High / small-medium |
| Attachment metadata search | Find images using their titles, descriptions/tooltip text, and existing OCR text. | `SearchAssetsByName` currently ranks filename/path hits; this picker search does not merge metadata or OCR hits. OCR itself already exists. | [Asset search](https://github.com/siyuan-note/siyuan/issues/19976), `de3de000af`, dev. | High / medium |
| Better snapshot browsing | Filter snapshots by date, find short snapshot IDs, and show tags in history. | Current history UI has snapshots and tags but lacks the new date-range handling. | [Date filters](https://github.com/siyuan-note/siyuan/issues/19943), `e6552a4722` in dev; ID/tag improvements `4b35d1778c`, `03968cff66`, `5e869bf3c8` in master. | Medium / medium |
| Large Markdown table performance | Keep offscreen rows virtualized while editing and selecting; reduce work during sidebar/pane resizing. | Scribli's AV/database views already use virtual scrolling. The reviewed upstream bundle instead targets ordinary editor tables and adds `tableVirtualization.ts`, which Scribli lacks. | [Table performance](https://github.com/siyuan-note/siyuan/issues/19911), `f34a0ac89e`, `293c2be8f5`, `7d24a9518c`, `6c472e002e`, `07df6c6402`, `364f2ffd0b`, in master. | Medium / medium-large |

The smaller candidates still have integration dependencies. Upstream storage statistics uses its newer typed API contracts and settings tabs; Scribli should expose the same behavior through its current routing/settings structure rather than import the entire framework. Frequent slash storage likewise depends on newer entry-visibility and settings persistence modules; adapt it to Scribli's existing save queue.

## Larger features worth considering later

- **Database automations:** [issue 10863](https://github.com/siyuan-note/siyuan/issues/10863), initial `4140e17e76`, dev. The current source supports entry-added/field-changed triggers, conditions, editing fields, and adding entries into current/related/filtered targets. A completion-timestamp rule is an example use. This is a transaction feature: review bounded execution, rollback, undo/redo, missing references, and encrypted notebook boundaries as a coherent bundle. Include `ba2a683a0a` and later fixes rather than importing the initial commit alone.
- **Local OCR provider selection:** [issue 19984](https://github.com/siyuan-note/siyuan/issues/19984), `1f00701552`, dev. Scribli already supports Tesseract; the new value is selecting Tesseract or PaddleOCR plus better OCR controls. Paddle introduces ONNX Runtime native code, a worker, model configuration, and packaging work. Keep models and recognition local and user-controlled; do not confuse this with the separate upstream AI OCR feature.
- **Independent settings window:** [issue 20002](https://github.com/siyuan-note/siyuan/issues/20002), `8bdac9df19` and follow-ups, dev. Useful for changing settings while keeping the note visible, but it adds Electron window lifecycle, authentication state, and settings synchronization work.
- **Editable mind maps:** [issue 19750](https://github.com/siyuan-note/siyuan/issues/19750), `94f2b57396`, master, with additional dev refinements. Scribli already renders mind-map code blocks. Upstream's dedicated node/block types and interactive editing are a larger upgrade involving content migrations and Lute support; preserve the local Lute fork and use the approved regeneration script.

## Already present or deliberately excluded

Exact alias search is already ported in Scribli (`7f24d26a1d`). MCP schema validation, structured outputs, tool synchronization/filtering, diagnostic actions, aliases, and image retrieval have also received Scribli work. The August MCP wishlist should be reconciled against that implementation before making another bridge bundle.

Keep Scribli's Canvas, Kanban column management, move menu, settings save feedback, and deletion undo when integrating upstream UI code. Upstream database schema is now spec 10 on master and spec 13 on the fetched dev tip; Scribli is spec 5. Database additions require an intentional migration plan and regression fixtures, not just copied UI files or an arbitrary version bump.

Do not import upstream accounts, subscriptions, official cloud, marketplace discovery, auto-update/publishing pipelines, or removed platform targets. Defer map views: the current dev implementation uses a built-in OpenFreeMap basemap and needs separate review for external tile/resource traffic before it fits Scribli's network boundaries. Preserve Scribli-owned Go modules, local forks, protocol, branding, and inherited upstream copyright notices in any imported code.

## Verification for eventual ports

- Data safeguards: focused kernel tests, Windows file-lock fixtures, cleanup failure fixtures, and encrypted-notebook coverage using isolated workspaces.
- Database features: migration fixtures from spec 5, table/gallery/Kanban preservation, undo/redo, field/option deletion, import/export, MCP/CLI output, and browser/Electron checks.
- Frontend features: typecheck/lint, relevant existing tests, production build, and actual narrow-window/keyboard interaction checks.
- Native OCR or Electron windows: Windows packaging and installed-app smoke tests in addition to unit checks.

The original review did not alter application source. The approved source changes and validation above were completed afterward; packaging and installation remain outside this intake.

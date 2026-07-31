# Third-Party Fork Audit

Date: 2026-07-30

This note tracks what is in `third_party/forks`, what Scribli actually imports at runtime, and what can realistically be pruned or moved back to normal Go dependency resolution.

## Summary

`third_party/forks` currently contains 580 files after replacing copied public dependency source with normal Go modules and adding a normal `third_party/forks/go.sum`. The remaining local fork folders are top-level Scribli-owned or Scribli-modified packages; no `third_party/forks/github/**` or `third_party/forks/external/**` source tree remains.

No non-source executable or binary artifacts remain in the fork tree according to the scan for `.exe`, `.dll`, `.so`, `.dylib`, `.jar`, `.wasm`, `.cmd`, `.bat`, `.ps1`, and `.msi`.

No Google Drive sourced blobs were found in this scan. AWS `.cn` and other regional endpoint strings now come from normal upstream AWS modules used for user-configured S3 sync; they are not SiYuan, B3log, LiuYun, or LianDi services.

## Completed Cleanup

The first cleanup pass removed:

```text
fatih/set
gigawattio/window
go-resty/resty
JalfResi/justext
jaytaylor/html2text
jolestar/go-commons-pool
klippa-app/go-pdfium
levigross/exp-html
olekukonko/tablewriter
richardlehane/mscfb
richardlehane/msoleps
ssor/bom
tetratelabs/wazero
tiendc/go-deepcopy
```

It also removed the large docs/examples/packaging folders identified in the first scan. PDF asset search now uses Scribli-owned best-effort text extraction in `kernel/model/pdf_text.go`, so the PDFium WASM runtime and Wazero runtime are no longer required.

The second cleanup pass removed the copied public dependency trees:

```text
third_party/forks/github
third_party/forks/external
```

Kernel imports and remaining local fork imports now use normal upstream paths for public libraries. `kernel/go.mod` and `kernel/go.sum` record the upstream module versions and checksums.

## Current Runtime Evidence

The dependency graph was checked from `kernel/` with the repo-local Go caches and `w64devkit` C compiler:

```powershell
$env:PATH=(Resolve-Path "..\.tools\w64devkit\bin").Path + ";" + $env:PATH
$env:GOPATH=(Resolve-Path "..\.tools").Path + "\go-path"
$env:GOCACHE=(Resolve-Path "..\.tools").Path + "\go-build-cache"
$env:GOMODCACHE=(Resolve-Path "..\.tools").Path + "\go-mod-cache"
$env:APPDATA=(Resolve-Path "..\.tools\appdata").Path
$env:CGO_ENABLED="1"
$env:CC=(Resolve-Path "..\.tools\w64devkit\bin\gcc.exe").Path
go telemetry off
go list -deps -f '{{.ImportPath}}' ./...
```

That command completed successfully. Runtime imports currently touch:

| Area | Runtime role |
| --- | --- |
| `go-sqlite3` | SQLite/SQLCipher and Scribli FTS tokenizer |
| `dejavu` | Repository sync backend for local, S3, and WebDAV sync |
| `lute` | Markdown/block parser and renderer |
| `pdfcpu` | PDF export/post-processing |
| `github.com/dop251/goja` and `github.com/dop251/goja_nodejs` | Plugin JavaScript runtime support from normal Go modules |
| AWS SDK and `github.com/studio-b12/gowebdav` | User-controlled S3 and WebDAV sync from normal Go modules |
| `github.com/imroc/req/v3`, `httpclient`, websocket and HTTP packages | User-triggered HTTP/API/plugin/MCP/AI paths |
| `github.com/denisbrodbeck/machineid`, `github.com/shirou/gopsutil/v4` | Local machine identity and disk/system information reads from normal Go modules |
| `kernel/model/pdf_text.go` | Scribli-owned best-effort PDF asset text extraction |

## Keep Local Because Scribli Modified It

These should stay local unless we intentionally reimplement the Scribli-specific behavior elsewhere.

| Fork | Why keep it |
| --- | --- |
| `go-sqlite3` | Contains the `scribli` FTS tokenizer inside the SQLite/SQLCipher C amalgamation. This is core search/indexing code and should be treated as high-risk. |
| `dejavu` | Contains Scribli sync/repository path behavior, including `.scribli` handling and removed official-cloud repair behavior. This protects local, S3, and WebDAV sync. |
| `httpclient` | Carries Scribli user-agent behavior and should remain local until callers are reviewed. |
| `lute` | Carries inherited parser behavior plus Scribli renaming. Keep local until the generated frontend `lute.min.js` rebuild path is restored. |
| `dataparser`, `eventbus`, `filelock`, `encryption`, `logging`, `riff`, `gulu`, `go-humanize`, `clipboard`, `epub` | Small forks that are cheap to either keep or internalize. Review individually before deleting because some are direct kernel imports. |

## High-Risk Audit Before Touching

These are not automatically bad, but they deserve real provenance/security review before we bless them.

| Area | Why it matters | Practical path |
| --- | --- | --- |
| `go-sqlite3` | CGO, C amalgamation, SQLCipher, custom tokenizer | Keep local, diff against upstream, document exact Scribli patch, and add focused search/index tests. |
| `pdfcpu` | Pure Go PDF parser/writer for export and PDF manipulation | Keep if PDF export needs it; otherwise remove the dependent export features later. |
| `goja` and `goja_nodejs` | Executes plugin JavaScript in the kernel process | Keep only if kernel-side plugin execution remains. Otherwise move plugin execution out of the kernel in a separate project. |
| AWS SDK credential loading | Supports S3 sync, including SDK credential providers | Keep S3 user-controlled, but review whether Scribli should allow `credential_process` execution from AWS config. |
| `machineid` and `gopsutil` | Reads machine identity and local system/disk information | Keep only where local features need it; document what is read and do not send it anywhere by default. |
| Scribli PDF text extractor | Parses untrusted PDF content streams for asset search | Keep small, best-effort, and covered by focused tests. It intentionally does not run OCR or embedded JavaScript. |

## Delete Candidates

There are no remaining nested copied GitHub/external package directories under `third_party/forks`.

The remaining empty docs/example directory stubs found by name were removed:

```text
third_party/forks/github/gin-gonic/gin/docs
third_party/forks/github/go-playground/validator/v10/_examples
third_party/forks/github/gopherjs/gopherjs/doc
third_party/forks/github/PuerkitoBio/goquery/doc
```

## Returned To Normal Modules

Copied public packages under `third_party/forks/github` and `third_party/forks/external` were moved back to normal Go module provenance through upstream import paths, `kernel/go.mod`, and `kernel/go.sum`.

This includes general-purpose libraries such as `gin`, `cobra`, `fsnotify`, `go-openai`, `samber/lo`, `mimetype`, `goquery`, `validator`, `go-yaml`, `toml`, `jwt`, `websocket`, `oauth2`, `protobuf`, `golang.org/x/*`, AWS SDK/Smithy, WebDAV, and similar packages.

## Internalize Candidates

These small wrappers may be simpler as Scribli-owned kernel packages instead of separate fork packages:

```text
clipboard
dataparser
encryption
eventbus
filelock
go-humanize
gulu
httpclient
logging
riff
```

Do this only when the code is tiny and clearly local. Internalizing `go-sqlite3`, `dejavu`, `lute`, `pdfcpu`, `goja`, Wazero, or AWS SDK would be a mess and would make updates harder.

## Recommended Cleanup Order

1. Completed: delete unused nested GitHub groups and docs/examples/packaging bloat.
2. Completed: replace PDFium/Wazero text extraction with Scribli-owned best-effort PDF text extraction.
3. Completed: run `go list -deps ./...` and focused PDF extractor tests.
4. Completed: add and expand `docs/THIRD-PARTY-PROVENANCE.md` for the remaining vendored code, including network-capable and process/system-capable paths.
5. Next: decide whether S3 should continue using the broad AWS credential chain or only explicit Scribli sync configuration.
6. Next: decide whether plugin JavaScript should keep running inside the kernel. If not, plan a plugin-runtime simplification around Electron/browser execution.
7. Completed: move general public libraries back to normal Go module provenance and delete copied `third_party/forks/github` and `third_party/forks/external` source trees.

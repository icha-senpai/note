# Third-Party Provenance

Date: 2026-07-30

Scribli keeps only selected Go dependency source under `third_party/forks` so Scribli-modified or intentionally-owned packages resolve through the local module `github.com/icha-senpai/note/third_party/forks`. General public libraries now resolve from their normal upstream Go module paths through `kernel/go.mod` and `kernel/go.sum`.

The fork tree currently contains 580 files after removing copied public `github/**` and `external/**` package trees and adding a normal `third_party/forks/go.sum`.

## Current Local Fork Inventory

Top-level Scribli fork modules:

```text
clipboard
dataparser
dejavu
encryption
epub
eventbus
filelock
go-humanize
go-sqlite3
gulu
httpclient
logging
lute
pdfcpu
riff
vitess-sqlparser
```

Copied `third_party/forks/github/**` and `third_party/forks/external/**` package trees were removed. Their imports now use normal upstream module paths such as `github.com/gin-gonic/gin`, `golang.org/x/net`, `google.golang.org/protobuf`, `github.com/aws/aws-sdk-go-v2`, and similar entries in `kernel/go.mod`.

## Scribli-Modified Forks

| Path | Kept because |
| --- | --- |
| `third_party/forks/go-sqlite3` | Contains the Scribli FTS tokenizer inside the SQLite/SQLCipher C amalgamation. This is core search/indexing code. |
| `third_party/forks/dejavu` | Owns Scribli repository/sync behavior for local, S3, and WebDAV providers. |
| `third_party/forks/httpclient` | Carries Scribli HTTP client identity behavior. |
| `third_party/forks/lute` | Parser/rendering behavior is coupled to the generated frontend Lute bundle. |

These are the forks most worth keeping local. They either contain Scribli-specific behavior or are tightly coupled to generated/runtime state.

## Small Local Forks To Review For Internalization

These are small enough to consider moving into `kernel/internal` or another Scribli-owned package later:

```text
clipboard
dataparser
encryption
epub
eventbus
filelock
go-humanize
gulu
logging
riff
vitess-sqlparser
```

Do not internalize them in bulk. Move one at a time, run the kernel tests, and remove the old package only when imports are clean.

## Large General-Purpose Forks

These are no longer local source forks. They are kept as normal Go modules in `kernel/go.mod`:

| Area | Notes |
| --- | --- |
| AWS SDK and Smithy | Required for user-configured S3 sync. Review credential-provider behavior before enabling broader AWS config loading. |
| `studio-b12/gowebdav` | Required for user-configured WebDAV sync. |
| `goja` and `goja_nodejs` | Required for current kernel-side plugin JavaScript support. |
| `gin`, `cobra`, `fsnotify`, `protobuf`, `golang.org/x/*`, and similar libraries | General dependency stack resolved through normal Go module provenance. |

## Network-Capable Paths

These are the remaining local-fork/runtime areas that can perform network I/O when reached by a Scribli feature:

| Area | Why it exists | Default risk posture |
| --- | --- | --- |
| AWS SDK and Smithy | User-controlled S3 sync provider. | Keep S3, but review credential providers. The AWS SDK may read local AWS config and can honor configured `credential_process` providers unless Scribli limits that path. |
| `studio-b12/gowebdav` and `dejavu` WebDAV code | User-controlled WebDAV sync provider. | Keep. It should only use endpoints configured by the user. |
| `httpclient`, `imroc/req`, standard `net/http`, `gorilla/websocket`, `r3labs/sse`, `lxzan/gws` | API, plugin proxy, import/fetch, AI, MCP, local server, and explicit user-triggered network features. | Keep, but document caller intent. Fresh launch must not call official upstream services. |
| `sashabaranov/go-openai` | User-configured AI provider support. | Keep only as a user-configured provider path. No default upstream API calls. |
| `modelcontextprotocol/go-sdk`, OAuth2, SSE, WebSocket packages | User-configured MCP HTTP/SSE/streamable transports and OAuth. | Keep. Remote MCP should require user configuration/authorization. |
| `kernel/api/network.go` | HTTP, SSE, and WebSocket proxy helpers used through local API/plugin paths. | Keep with scrutiny. This is powerful and should remain user/API-triggered, not startup-triggered. |
| `pdfcpu` link validation code | The library contains optional URL validation paths. | Avoid invoking URL validation for local export/search unless explicitly requested. |

Current upstream-service scan result: after cleanup, the kernel and remaining fork Go/JSON/Markdown scan has no matches for official SiYuan/B3log/LiuYun/LianDi runtime service domains. AWS regional endpoint tables and public suffix data now live in the normal upstream modules instead of copied local source.

## Process And System Access Paths

These are local-only capabilities that can read system state or start a process:

| Area | Why it exists | Cleanup decision |
| --- | --- | --- |
| `kernel/mcp/client/mcp.go` | Starts user-configured MCP stdio servers with `exec.Command(server.Command, server.Args...)`. | Keep as a user-controlled MCP feature. Consider requiring explicit enablement and clear UI labeling. |
| `kernel/util/pandoc.go` and `kernel/model/export.go` | Runs Pandoc for document export. | Keep if export remains. Ensure binary path is local/user-visible. |
| `kernel/util/ebook.go` and ebook import/export model paths | Runs a user-configured local `ebook-convert` executable for MOBI/AZW/AZW3 conversion. | Keep local-only and user-visible. Do not download converters or call hosted conversion services. |
| `kernel/util/ocr.go` | Runs Tesseract for OCR. | Keep only as explicit OCR functionality. No startup execution. |
| `kernel/model/elevator_windows.go` | Uses PowerShell/Windows ShellExecute for Windows Defender exclusion support. | Optional. If Boss wants zero Defender/elevation integration, this can be disabled or removed separately. |
| `kernel/model/flushdns_windows.go` | Runs `ipconfig /flushdns` for DNS repair. | Optional convenience. Remove if Scribli should avoid system commands entirely. |
| `clipboard` and `gulu` command helpers | Clipboard and shell helpers, including non-Windows files retained in copied source. | Windows build only reaches Windows-specific code where applicable. Prune non-Windows source only if we are ready to make the fork tree Windows-only. |
| `machineid` and `gopsutil` | Reads local machine/disk/host information. | Keep only for local features. Do not send this data over the network by default. |

## Official-Service Cleanup Status

The official upstream cloud/account/payment/update/bazaar work already removed the app-level service behavior. The provenance scan now checks that the remaining kernel/fork code does not reintroduce official service calls.

The latest kernel/fork scan cleaned the last source-only leftovers:

```text
third_party/forks/dataparser/sy.go
third_party/forks/dataparser/json_unmarshal.go
third_party/forks/eventbus/eventbus.go
kernel/plugin/server_test.go
```

Those were branding comments and a fake test user-agent only; they were not runtime callbacks.

## Removed In This Cleanup

Scribli no longer vendors PDFium or Wazero for PDF asset text indexing. The old PDFium WASM blob was removed, and PDF asset search now uses the local best-effort extractor in `kernel/model/pdf_text.go`.

Removed package groups:

```text
third_party/forks/external
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
third_party/forks/github
tiendc/go-deepcopy
```

The copied public dependency trees were replaced by normal upstream Go module requirements in `kernel/go.mod`. `github.com/ConradIrwin/font` is pinned to upstream commit `44ae4cf5fb22` because the tagged release selected by `go mod tidy` did not include `sfnt.ParseCollection`, which Scribli already uses.

## Remaining Risk Register

| Area | Risk | Follow-up |
| --- | --- | --- |
| `go-sqlite3` | Native C/CGO and search tokenizer changes | Diff against upstream and add focused tokenizer/index tests. |
| PDF parsing | Untrusted PDF content streams are parsed by Scribli code | Keep extractor small and best-effort; avoid JavaScript, OCR, external processes, and embedded binary runtimes. |
| AWS credential loading | Some SDK providers can read local AWS config and run configured credential processes | Review S3 sync setup and disable credential providers Scribli does not need. |
| `goja` plugin runtime | Plugin code executes inside the kernel process | Revisit plugin runtime architecture if Scribli stops caring about inherited plugin compatibility. |
| `machineid` and `gopsutil` | Reads local machine/disk/system information | Document what is read and keep it local-only unless the user explicitly configures a network feature. |
| MCP/plugin/network proxy APIs | User/plugin/API triggered network access can be broad | Keep but treat as an explicit power-user surface; consider per-feature toggles if Scribli tightens further. |

## Recommended Next Cleanup

1. Decide whether to keep S3's default AWS credential chain. If not, wire S3 sync to explicit Scribli config only and disable `credential_process`.
2. Decide whether kernel-side JavaScript plugins are still a goal. If not, remove the `goja` and `goja_nodejs` runtime and move plugin behavior to the Electron/browser side later.
3. Decide whether Windows Defender/elevation helpers should stay. They are not evidence of spyware, but they are privileged/system-facing and should be visible.
4. Internalize one small wrapper at a time: `eventbus`, `filelock`, `logging`, `dataparser`, then `httpclient`.
5. Re-run the official-service and local-fork import scans after each cleanup:

```powershell
rg -n "api\.b3log|b3log\.org|liuyun|liandi|npmmirror|aliyun|tencent|baidu|SiYuan|siyuan" third_party\forks kernel -g "*.go" -g "*.json" -g "*.md"
rg -n "github.com/icha-senpai/note/third_party/forks/(github|external)" kernel third_party\forks -g "*.go" -g "*.mod" -g "*.md"
```

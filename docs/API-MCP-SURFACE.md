# Scribli API and MCP Surface

<!-- markdownlint-disable MD013 -->

This note records the intended split between Scribli's HTTP API and MCP tool surface as of 2026-08-14.

## HTTP API

The HTTP API is the broad runtime surface used by the desktop frontend, plugins, local integrations, and authenticated automation. The registered route table in `kernel/api/router.go` currently contains 498 routes across `/api/*`, `/ws/*`, and `/es/*`.

The largest route families are:

| Family | Routes | Coverage role |
| --- | ---: | --- |
| `/api/block` | 56 | Block CRUD, movement, folding, references, DOM/Kramdown access, and block metadata |
| `/api/system` | 46 | Boot, auth, workspace, config, TLS, logging, indexing, and process controls |
| `/api/av` | 39 | Attribute view/database rendering, keys, rows, layout, grouping, and cleanup |
| `/api/filetree` | 34 | Document tree, document creation, movement, lookup, publishing access, and tree utilities |
| `/api/export` | 33 | Document, notebook, workspace, asset, preview, and package export flows |
| `/api/ai` | 27 | AI calls, embedding, MCP OAuth, agent sessions, confirmations, questions, and skills |
| `/api/notebook` | 23 | Notebook lifecycle, encryption, icons, status, and crypto backups |
| `/api/sync` | 21 | User-configured sync, sync status, cloud file listing, and sync provider operations |
| `/api/setting` | 20 | User configuration for editor, appearance, AI, publish, snippets, and virtual refs |
| `/api/repo` | 19 | Repository keys, snapshots, tags, checkout, diff, retention, and file rollback/export |
| `/api/storage` | 19 | Local UI/session storage, recent documents, outline storage, and criteria |
| `/api/asset` | 18 | Asset upload, browse, duplicate detection, cleanup, and metadata |
| `/api/riff` | 17 | Flashcard decks, cards, due queues, review, reset, and scheduling |
| `/api/search` | 16 | Tag/template/widget/ref/embed/block/asset search and replace |
| `/api/import` | 13 | Markdown, `.sy`, Data, Pandoc, Word, and standard Markdown imports |
| `/api/history` | 11 | Document, notebook, asset, database, and workspace history |

`docs/API.md` remains a useful human-facing compatibility reference, but it is not a complete generated contract for the live router. Treat `kernel/api/router.go` as the source of truth for route existence until generated API reference work is added.

## MCP

The MCP surface is the agent-friendly subset of the app, not a one-for-one mirror of the HTTP API. It currently exposes native tools for common agent workflows: notebooks, documents, blocks, attributes, databases, assets, search, SQL read access, import/export, history, repository snapshots, sync, templates, tags, bookmarks, workspace/system info, session todos, executable blocks, local JSON Canvas files, web fetch/search, HTTP requests, image operations, skills, archive extraction, API route discovery, route detail, and guarded local API fallback calls. The `frontend` and `question` tools are in-app-agent-only bridges and are intentionally hidden from direct `/mcp` clients because they require the Scribli browser agent loop to intercept them.

MCP tools must be safe to reason about from an agent loop. Every native MCP tool declares an `EffectScope`, and every enum action declares `ActionEffects` metadata. The supported effects are:

| Effect | Meaning |
| --- | --- |
| `LocalRead` | Reads local Scribli workspace, config, index, or app state |
| `LocalWrite` | Mutates local workspace data and should create a repository snapshot before agent execution |
| `LocalStateWrite` | Mutates local app/session/export/repository state but should not create a repository snapshot |
| `DataEgress` | Sends local data or user-provided request data outside the machine |
| `ExternalCost` | May call a configured external paid or quota-bearing service |

The interactive agent confirmation gate in `kernel/agent/agent.go` uses this metadata first. `LocalWrite`, `LocalStateWrite`, `DataEgress`, and `ExternalCost` require confirmation unless the user has already allowed that tool/action; only `LocalWrite` creates a repository snapshot. Direct `/mcp` callers are authenticated local API clients and must apply their own confirmation policy if they expose these tools to autonomous agents.

Read tools that can return user content redact likely credentials and tokens by default. This includes search snippets, asset indexed content, document/block reads, database render cell values, and inline Markdown/HTML/preview export responses. Purpose-built direct read tools expose `redactSecrets`, which defaults to `true`; callers must explicitly pass `false` when they need raw content.

`api_catalog` exposes the live route table registered in Gin so a fresh MCP client can discover available local API paths without repository access. `api_route` returns details for a single route, including inferred effect metadata, a confidence-tagged JSON request schema, and ready-to-fill `api_call` arguments. Exact request schemas come from route-specific source/docs hints; fallback schemas are labeled as inferred so callers can treat them as guidance rather than a hard contract. `api_call` is a local-only escape hatch for authenticated `/api/...` calls when no purpose-built MCP tool fits; it attaches the configured API token internally, rejects full URLs, checks the route catalog when available, and derives confirmation/snapshot behavior from the selected route.

`executable_block` is the first active-note primitive. It supports `validate_js` and `run_js` for local JavaScript through Goja only, `run_sql` through the same read-only SQL guard as the `sql` tool, `run_api` through the same guarded local API path as `api_call`, and `chart` for producing Scribli-renderable `echarts` Markdown code blocks from JSON chart specs. Version 1 exposes user-provided JSON input and `console.log`, but does not expose filesystem, network, process, `fetch`, or `require`. JavaScript execution is manual, timeout-limited, structured, and marked as `LocalStateWrite` so agent hosts can require confirmation.

The desktop editor exposes the same executable primitive through explicit code block languages: `scribli-js`, `scribli-sql`, `scribli-api`, and `scribli-chart`. These blocks render a manual run action and call the authenticated local route `POST /api/executableBlock/call`; they do not auto-run while opening a document. `scribli-api` accepts either a plain `/api/...` path or a JSON object with `path`, optional `method`, optional `query`, and optional `body`.

`canvas` is the first visual-workspace primitive. It stores JSON Canvas-compatible `.canvas` files under local workspace storage and supports `create`, `create_and_embed`, `list`, `get`, `update`, `delete`, `add_node`, `add_scribli_node`, and `add_edge`. Scribli-specific metadata is stored under the `scribli` key so the base node/edge model stays portable. `add_scribli_node` gives agents stable node metadata for documents, blocks, assets, databases, queries, and executable outputs without requiring them to hand-roll raw JSON Canvas fields.

`create_and_embed` is the agent-friendly one-shot workflow for requests such as "make a canvas with this document." It creates the stored canvas, optionally adds a raw JSON Canvas node or one Scribli-aware node from the same `kind`/`refID`/`query`/`assetPath`/`databaseID` arguments used by `add_scribli_node`, and inserts a `scribli-canvas` code block into `embedParentID` or next to `embedPreviousID`/`embedNextID`. The result includes both the canvas ID and the created render block ID, so agents do not need to hand the ID back to the user for manual pasting.

The authenticated local route `POST /api/canvas/call` is available for the desktop frontend and plugins to use the same canvas handler without going through MCP.

The desktop editor can render a stored or inline visual workspace with the `scribli-canvas` code block language. Users can insert one from the slash menu with `/canvas`; an empty Canvas block renders `Create Canvas`, template, library, and import actions. Stored canvases expose a visual toolbar for text/document/block/asset/query/database cards, connector creation, duplicate/delete, local canvas undo, library switching, templates, import/export, zoom, and pan. Text cards edit inline, cards drag and resize, and persisted actions write through `POST /api/canvas/call` with the `update` action. If the block contains a JSON object, it renders that inline JSON Canvas payload without writing it back. Scribli document/block cards open their referenced block, asset cards open through the asset viewer, query cards open search, and database cards open their referenced block when available or search by database ID.

## Parity Boundaries

MCP intentionally does not expose every application operation. Settings administration, publishing configuration, flashcards, graph operations, plugin RPC, low-level network proxying, most storage internals, and full route-level API administration remain HTTP/API or frontend responsibilities unless a specific agent workflow justifies a carefully scoped MCP tool.

When adding MCP tools or actions, update effect metadata with the schema change and keep `TestNativeToolsDeclareEffectMetadata` passing. When adding HTTP API routes, update this file if the route changes the supported automation boundary.

# Scribli API and MCP Surface

<!-- markdownlint-disable MD013 -->

This note records the intended split between Scribli's HTTP API and MCP tool surface as of 2026-07-31.

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

The MCP surface is the agent-friendly subset of the app, not a one-for-one mirror of the HTTP API. It currently exposes 34 native tools for common agent workflows: notebooks, documents, blocks, attributes, databases, assets, search, SQL read access, import/export, history, repository snapshots, sync, templates, tags, bookmarks, workspace/system info, frontend handoff, questions, session todos, web fetch/search, HTTP requests, image operations, skills, archive extraction, API route discovery, route detail, and guarded local API fallback calls.

MCP tools must be safe to reason about from an agent loop. Every native MCP tool declares an `EffectScope`, and every enum action declares `ActionEffects` metadata. The supported effects are:

| Effect | Meaning |
| --- | --- |
| `LocalRead` | Reads local Scribli workspace, config, index, or app state |
| `LocalWrite` | Mutates local workspace data and should create a repository snapshot before agent execution |
| `LocalStateWrite` | Mutates local app/session/export/repository state but should not create a repository snapshot |
| `DataEgress` | Sends local data or user-provided request data outside the machine |
| `ExternalCost` | May call a configured external paid or quota-bearing service |

The interactive agent confirmation gate in `kernel/agent/agent.go` uses this metadata first. `LocalWrite`, `LocalStateWrite`, `DataEgress`, and `ExternalCost` require confirmation unless the user has already allowed that tool/action; only `LocalWrite` creates a repository snapshot. Direct `/mcp` callers are authenticated local API clients and must apply their own confirmation policy if they expose these tools to autonomous agents.

`api_catalog` exposes the live route table registered in Gin so a fresh MCP client can discover available local API paths without repository access. `api_route` returns details for a single route, including inferred effect metadata and `api_call` guidance. `api_call` is a local-only escape hatch for authenticated `/api/...` calls when no purpose-built MCP tool fits; it attaches the configured API token internally, rejects full URLs, checks the route catalog when available, and derives confirmation/snapshot behavior from the selected route.

## Parity Boundaries

MCP intentionally does not expose every application operation. Settings administration, publishing configuration, flashcards, graph operations, plugin RPC, low-level network proxying, most storage internals, and full route-level API administration remain HTTP/API or frontend responsibilities unless a specific agent workflow justifies a carefully scoped MCP tool.

When adding MCP tools or actions, update effect metadata with the schema change and keep `TestNativeToolsDeclareEffectMetadata` passing. When adding HTTP API routes, update this file if the route changes the supported automation boundary.

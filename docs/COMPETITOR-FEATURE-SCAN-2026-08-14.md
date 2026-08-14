# Competitor Feature Scan - 2026-08-14

This is a product research note for Scribli. It compares current public features from Notion, Znote, Mem, AppFlowy, Joplin, Logseq, Outline, BookStack, Obsidian, Anytype, Capacities, and AFFiNE. The goal is not to clone everything. The goal is to identify ideas worth adapting to Scribli's local-first, user-controlled direction.

## Scribli Baseline

Scribli already has a strong foundation: local-first Windows desktop use, a document tree, block editing, backlinks/search, attribute views/databases, import/export, history/snapshots/backups, plugins, API/MCP access, user-configured AI providers, and user-controlled sync such as S3, WebDAV, and local-folder sync.

Scribli should keep avoiding default hosted accounts, subscriptions, upstream official cloud services, upstream marketplace coupling, and automatic release/update flows until Scribli owns those systems. The best borrowed features are the ones that strengthen local ownership instead of replacing it with another cloud dependency.

## Highest-Value Ideas To Borrow

| Priority | Idea | Seen In | Why It Matters | Scribli Shape |
| --- | --- | --- | --- | --- |
| P0 | Inline AI blocks with explicit context | Notion, Znote, Mem, Capacities | The user can ask questions or transform notes without leaving the document. | Add an AI block or slash command that can reference selected blocks, pages, databases, files, and MCP tools. Require explicit provider configuration and make source context visible. |
| P0 | Agent-safe note and database tools | Notion, Mem, Capacities | Chat clients become useful when they can create, update, query, and summarize real workspace objects. | Keep improving MCP with stable schemas, output schemas, read/write permissions, secret redaction, and clear tool families for documents, blocks, notebooks, databases, tasks, search, import, and export. |
| P0 | Runnable blocks and local workflows | Znote, AFFiNE | Notes become active workspaces, not just static text. | Add sandboxed executable blocks for JavaScript, local API calls, SQL/database queries, and charts. Keep execution disabled or permission-gated for untrusted documents. |
| P0 | Better database views | Notion, AppFlowy, Capacities, AFFiNE | Databases become useful for planning, tracking, research, and dashboards. | Expand attribute views with polished table, kanban, calendar, gallery, chart, and saved-filter views. |
| P1 | Forms into databases | Notion | Forms are a low-friction way to capture structured data. | Add optional local/server forms that write into a selected database or create documents from templates. Keep public exposure explicitly configured. |
| P1 | Capture inbox | Mem, Obsidian, Joplin, Capacities | Fast capture is one of the biggest daily-use features. | Add web clipping, daily inbox capture, quick capture, and later voice/transcription. Save source URL, timestamp, selected text, and extraction metadata. |
| P1 | Canvas and visual thinking | Logseq, Obsidian, AFFiNE | A canvas helps with maps, research boards, planning, and connecting documents visually. | Add a canvas surface that can embed documents, blocks, images, PDFs, queries, and database cards while saving in an open format. |
| P1 | PDF and image understanding | Mem, Logseq, AFFiNE, Joplin | Real workspaces contain scans, screenshots, slides, and papers. | Add local-first OCR/indexing and attach AI summarization only when a provider is configured. Connect highlights back to source files and blocks. |
| P2 | Review metadata and verified pages | Notion, BookStack, Outline | Team or long-lived knowledge needs freshness indicators. | Add optional page status, owner, review date, verified marker, stale marker, and review queries. |
| P2 | Workflow templates and recipes | Znote, BookStack, Outline, Obsidian | Templates turn blank pages into repeatable systems. | Ship local template packs for meetings, research, CRM, projects, recipes, study, publishing, and MCP-driven workflows. |
| P2 | Optional team/self-host mode | Outline, BookStack, Anytype, AFFiNE | Some users need shared docs, roles, and admin controls. | Treat this as a later optional mode, not the default product. Build on self-host/local server boundaries, explicit auth, roles, audit logs, and exportability. |

## Competitor Notes

### Notion

Notion is strongest as a cloud team workspace. Its public product surface combines docs, wikis, projects, databases, forms, sites, charts, calendar/mail integrations, automations, comments, sharing, granular permissions, AI Agent, Enterprise Search, AI Meeting Notes, and database-aware AI operations.

Useful for Scribli: forms into databases, polished database views, chart dashboards, inline AI editing, agent workflows, meeting-note capture, verified pages, and permissions if Scribli later grows a team mode.

Avoid copying directly: cloud-first assumptions, account-centered collaboration, hosted search over third-party apps by default, and feature paths that require Scribli to become a SaaS before the local app is excellent.

### Znote

Znote is the closest direct inspiration for "notes that can do work." It is a local Markdown editor with no required account, local `.md` files, BYO AI keys, OpenAI/Ollama/OpenRouter support, inline AI generation, `@` references to notes/templates/code blocks, runnable JavaScript and Node blocks, NPM package support, charts/tables from code, workflow recipes, plugins, meeting transcription, web search from AI blocks, password-protected notes, tabs, folders, tags, and PDF/HTML export.

Useful for Scribli: executable blocks, inline AI with explicit references, reusable AI/code workflows, output renderers, and BYO-provider positioning.

Scribli advantage: Scribli already has a deeper block/document/workspace engine, API/MCP surface, databases, sync options, history, and larger document-management foundation. The opportunity is to add Znote-like active blocks without giving up Scribli's richer workspace model.

### Mem

Mem is centered on AI memory and capture. Its public surface emphasizes Voice Mode, meeting recording/transcription, web clipping, chat over workspace knowledge, related notes, automatic organization, "Heads Up" reminders, Deep Search, shared knowledge, collections, and PDF/image understanding with OCR-style indexing.

Useful for Scribli: a fast capture inbox, voice-to-note workflows, related-context panels, follow-up reminders, and attachment understanding.

Avoid copying directly: invisible automatic organization that makes local data feel less inspectable. Scribli should make inferred links, sources, and AI actions auditable.

### AppFlowy

AppFlowy is an open-source Notion alternative focused on user data control, native/offline use, self-hosting options, extensibility, documents, wikis, databases, tasks, projects, Kanban, Calendar, and opt-in OpenAI writing features.

Useful for Scribli: clean database view UX, native/offline positioning, self-host language, and extensibility without requiring hosted accounts.

The main lesson is product clarity: make local ownership feel like a feature, not a limitation.

### Joplin

Joplin is an offline-first open-source Markdown note and to-do app. It has notebooks, tags, full-text search, external editor support, Evernote and Markdown imports, mobile and desktop apps, web clipper, plugins, themes, multiple sync targets such as Nextcloud/Dropbox/OneDrive/Joplin Cloud, and end-to-end encryption for sync.

Useful for Scribli: clear E2EE sync UX, importers, web clipper, external editor escape hatch, and simple to-do/note reliability.

Scribli can be richer than Joplin, but should envy Joplin's trust posture: predictable local files, explicit sync, clear encryption story.

### Logseq

Logseq is a privacy-first outliner and knowledge graph. It emphasizes local Markdown/Org files, daily notes, backlinks, page and block references, block embeds, queries/query builder, tasks, templates, PDF highlights, whiteboards, media embeds, Zotero, flashcards, slides, calculator, tables, publishing, graph view, plugins, themes, mobile apps, and encrypted sync beta.

Useful for Scribli: stronger block references, saved queries, flashcards, PDF annotation, review/study flows, and whiteboard/canvas thinking.

The lesson is that blocks should be reusable objects, not only content inside a page.

### Outline

Outline is a team wiki and documentation platform. It uses workspaces, collections, nested documents, templates with placeholders, roles such as Admin/Editor/Viewer/Guest, groups, collection-level access, private collections, SSO options, revision history, Markdown/JSON export, API access, TLS, encryption at rest, and backups.

Useful for Scribli: collection-level permission concepts, document governance, exports, API clarity, and templates if Scribli adds a team/self-host mode.

Avoid copying directly into the desktop app too early. Outline's strengths make sense when multiple people need permission boundaries.

### BookStack

BookStack is built around an explicit hierarchy: shelves, books, chapters, and pages. It includes a default WYSIWYG editor, Markdown editor option, attachments, drawings/diagrams, tags with values, page templates, advanced search, export/import, page permalinks, role permissions, and cascading content-level permissions.

Useful for Scribli: review metadata, tag values, page templates, diagrams, clearer long-form knowledge-base hierarchy, and predictable permission inheritance for a future shared mode.

BookStack's lesson is that boring structure can be powerful when a workspace gets large.

### Obsidian

Obsidian is a local vault app with Markdown files, links, backlinks, graph view, properties, plugins, themes, CSS snippets, Canvas, Web Clipper, Sync, Publish, and a very large ecosystem. The Web Clipper can save pages and highlights to a local vault with templates, variables, Schema.org data, meta tags, and CSS selector extraction. Canvas provides an infinite 2D space for notes, attachments, and web pages.

Useful for Scribli: plugin ecosystem expectations, local vault trust, Web Clipper templates, typed properties, Canvas, and open formats.

Scribli should borrow the openness and clipper power, not the "everything is a loose file" constraint.

### Anytype

Anytype is an encrypted local-first object graph. It uses spaces, objects, types, relations, sets, collections, local/offline operation, peer-to-peer local sync, user-controlled keys, and an open-source sync protocol. Its business product adds SSO, central admin, self-hosting, team chat, knowledge bases, projects, roadmaps, CRM/contact workflows, assets, devices, and employee directory use cases.

Useful for Scribli: typed objects, relation-first thinking, local-first security language, and optional self-host/team architecture.

Risk: the object model can become abstract fast. Scribli should expose object-like power through understandable databases, templates, and document links.

### Capacities

Capacities is object-based note-taking rather than file/folder note-taking. It uses object types, properties, templates, multiple views such as lists/galleries/tables/cards, bidirectional links, backlinks, graph view, daily notes as a calendar/inbox anchor, and MCP connectors that let external clients search/read objects, add content to daily notes, link objects, create pages/tasks/custom objects, update properties, and append content.

Useful for Scribli: daily note inbox, object-type UX, source-aware research workflows, and MCP connector patterns for creating/updating structured workspace objects.

Scribli's matching direction should be: make documents, blocks, and databases feel like programmable local objects without hiding where the data lives.

### AFFiNE

AFFiNE combines docs, whiteboards, databases, and AI in a local-first open-source workspace. It positions itself as an alternative to Notion and Miro, with a block editor, edgeless whiteboard, multi-view databases, local/offline operation, self-hosting, GraphQL API, templates, presentations, asset management, and AI features such as writing help, summaries, mind maps, PDF/video summarization, and presentation generation.

Useful for Scribli: document-canvas bridge, edgeless visual planning, multi-view databases, AI mind maps, and self-host positioning.

The main lesson is that documents and whiteboards should not feel like separate worlds.

## Proposed Scribli Roadmap Slices

| Slice | Scope | Why First |
| --- | --- | --- |
| 1 | Agent-ready writing surface | Build inline AI blocks, `@` source picking, safe search snippets, and apply/insert actions. This turns the existing MCP and AI work into something users can feel inside the editor. |
| 2 | Safer MCP/database/document coverage | Finish stable create/read/update/delete tools for notebooks, documents, blocks, databases, tasks, search, import, export, and workspace metadata. Add output schemas and permission-aware writes. |
| 3 | Executable blocks | Add permission-gated JavaScript, local API, SQL/query, and chart blocks. Start local-only and make outputs reproducible. |
| 4 | Database view polish | Add or improve kanban, calendar, gallery, saved filters, charts, form capture, templates, and dashboard blocks. |
| 5 | Capture layer | Add web clipper, daily inbox, quick capture, attachment OCR, and later voice transcription. |
| 6 | Canvas layer | Add a visual workspace that embeds blocks, pages, assets, queries, and database cards. Prefer an open save format. |
| 7 | Optional team/self-host mode | Only after the local app is stable: roles, permissions, audit logs, authenticated local server, backups, and deployment docs. |

## Selected Build Tracks

These are the three tracks that should move from research into implementation first. They depend on each other: safer MCP coverage gives agents and executable blocks a trusted control surface; executable blocks prove notes can produce live outputs; canvas can then embed those outputs alongside normal documents and databases.

### Track 1: Safer MCP, Database, And Document Coverage

Goal: make Scribli's MCP feel like a stable local workspace API instead of a loose collection of convenience calls.

Initial tool families:

| Family | Needed Tools |
| --- | --- |
| System/workspace | `system.version`, `workspace.info`, `api_catalog`, `api_route`, `api_call` |
| Notebooks | list, create, rename, open, close, delete, tree |
| Documents | list/search, create, get, update, rename, move, duplicate, delete, export |
| Blocks | get, append, prepend, update, delete, move, query refs/backlinks |
| Databases | list, create, get schema, update schema, insert row, update row, delete row, render/search |
| Tasks/todos | create, list, update status, delete, optionally bind to document/block |
| Import/export | import Markdown/DOCX/HTML/ebook where supported, export Markdown/PDF/DOCX/HTML/ebook where supported |

Acceptance bar:

| Requirement | Reason |
| --- | --- |
| Stable schemas regardless of filtered discovery | ChatGPT and other clients should not see different tool contracts depending on search terms. |
| Output schemas for every stable tool | Clients can reason about results instead of parsing prose. |
| Correct IDs and names in every result | Agents need identity continuity for follow-up edits. |
| Permission-aware write metadata | Local writes, state writes, data egress, and external-cost actions should be visible before a client invokes them. |
| Secret-safe search results | Search must not leak tokens, passwords, private keys, or credentials from code blocks or attachments by default. |
| Purpose-built tools | Prioritize specific tools over generic `api_call` where possible. |
| Focused tests per tool family | Every create/update/delete path needs a read-back assertion and an error-path assertion. |

First implementation pass should fix the known tool bugs before adding more surface area: discovery consistency, `api_call` TLS/local-certificate behavior, `document.duplicate` returning the wrong ID, `document.search_docs` losing IDs/names, noisy `database.render`, DOCX output contract mismatch, incomplete `document.get`, incomplete block timestamps, poor closed-notebook rename errors, and token-heavy `api_catalog` filtering.

### Track 2: Executable Blocks

Goal: let notes produce reproducible local outputs while keeping the user in control of code execution.

Block types to consider:

| Block Type | Output |
| --- | --- |
| JavaScript | Text, JSON, table, chart, generated Markdown |
| Local API/MCP call | Structured result preview and optional inserted output |
| SQL/query | Table, count, chart, saved query result |
| Chart | Static rendered chart from inline data, database rows, or query output |
| Shell command | Later only, if permission boundaries are strong enough |

Required guardrails:

| Guardrail | Requirement |
| --- | --- |
| Manual execution | Executable blocks should not auto-run when a document opens. |
| Per-block trust | A copied/imported block must be untrusted until the user allows it. |
| Scoped permissions | Code should request access such as workspace read, document write, network, filesystem, shell, or external AI. |
| Reproducible outputs | Store code, inputs, run timestamp, environment label, and output snapshot. |
| Local-first default | No hosted execution service. |
| Clear failure state | Errors should render inside the block without crashing the editor. |

Best first version: JavaScript blocks with no network and no filesystem, plus local Scribli API/MCP calls through a small approved bridge. Outputs can start with text, JSON, Markdown, and simple tables before charts.

### Track 3: Canvas Layer

Goal: add a visual workspace for thinking across documents, blocks, assets, database cards, queries, and executable outputs.

Canvas objects:

| Object | Behavior |
| --- | --- |
| Document card | Shows title, excerpt, backlinks/counts, open/edit actions |
| Block card | Embeds a specific block with source link and update awareness |
| Asset card | Displays images, PDFs, audio, or attachments |
| Database card | Shows a table/query/list/kanban/chart slice |
| Query card | Renders saved search or database query results |
| Executable output card | Pins the output from a runnable block |
| Free text/card | Lightweight notes for layout and planning |
| Connector edge | Links cards with labels and optional direction |

Open format preference:

| Candidate | Fit |
| --- | --- |
| JSON Canvas | Good fit for Obsidian-style interoperability and simple node/edge storage. |
| Scribli JSON extension | Useful if Scribli needs block/database-specific fields that JSON Canvas does not support. |
| Markdown front matter plus JSON payload | Useful for exportability, but weaker for live canvas editing. |

Best first version: a JSON Canvas-compatible file with Scribli-specific metadata namespaced inside each node. Start with document cards, block cards, text cards, image cards, and connector edges. Add live database/query/executable cards after the runtime and permission model are stable.

## Guardrails

Do not add default hosted services. Do not call official upstream services on fresh workspaces. Do not send AI requests unless the user configures a provider and invokes the feature. Do not make executable blocks run automatically in untrusted documents. Do not expose secrets through search, MCP, logs, exports, or AI context. Keep imports, exports, backups, and local ownership visible.

Scribli's strongest product lane is not "Notion clone." It is a local-first knowledge workspace that can become active: notes, databases, search, documents, AI, and MCP all working together under the user's control.

## Sources Checked

Checked on 2026-08-14.

| Product | Sources |
| --- | --- |
| Notion | [Product](https://www.notion.com/product/notion), [AI](https://www.notion.com/product/ai?locale=en), [Forms](https://www.notion.com/product/forms), [Search](https://www.notion.com/help/search), [Notion AI FAQ](https://www.notion.com/help/notion-ai-faqs), [Enterprise Search security](https://www.notion.com/help/enterprise-search-security-and-privacy-practices) |
| Znote | [Home](https://znote.io/index.html), [Feature page](https://znote.io/?from=AppAgg.com), [Znote 1.0 release](https://blog.znote.io/2021/znote-10-release/) |
| Mem | [Home](https://mem.ai/), [Team guide](https://help.mem.ai/guides/using-mem-as-a-team), [PDF and image understanding](https://help.mem.ai/features/pdf-and-image-understanding), [Voice Mode](https://help.mem.ai/features/voice-mode) |
| AppFlowy | [Docs](https://docs.appflowy.io/docs), [Welcome](https://docs.appflowy.io/docs/appflowy/readme/welcome-to-appflowy), [Database view architecture](https://docs.appflowy.io/docs/documentation/software-contributions/architecture/frontend/database-view), [Kanban board](https://docs.appflowy.io/docs/documentation/software-contributions/architecture/frontend/database-view/kanban-board), [Calendar view](https://docs.appflowy.io/docs/documentation/software-contributions/architecture/frontend/database-view/calendar) |
| Joplin | [Help](https://joplinapp.org/help/), [E2EE](https://joplinapp.org/help/apps/sync/e2ee/), [Privacy](https://joplinapp.org/privacy/), [Release 3.6](https://joplinapp.org/news/20260505-release-3-6/) |
| Logseq | [Home](https://logseq.com/?nfs=true), [Docs](https://docs.logseq.com/) |
| Outline | [Users and groups](https://docs.getoutline.com/s/guide/doc/users-groups-cwCxXP8R3V), [Collections](https://docs.getoutline.com/s/guide/doc/collections-l9o3LD22sV), [Security](https://docs.getoutline.com/s/guide/doc/security-DlJBglbImQ), [Terminology](https://docs.getoutline.com/s/guide/doc/terminology-fKoXA2YGzH) |
| BookStack | [Docs](https://www.bookstackapp.com/docs/), [Content overview](https://www.bookstackapp.com/docs/user/content-overview/), [Roles and permissions](https://www.bookstackapp.com/docs/user/roles-and-permissions/), [Organising content](https://www.bookstackapp.com/docs/user/organising-content/), [Tags](https://www.bookstackapp.com/docs/user/tags/), [Searching](https://www.bookstackapp.com/docs/user/searching/) |
| Obsidian | [Help](https://obsidian.md/help/), [Canvas](https://obsidian.md/help/Plugins/Canvas), [Web Clipper](https://obsidian.md/clipper), [Web Clipper help](https://obsidian.md/help/web-clipper), [Properties](https://obsidian.md/help/properties) |
| Anytype | [Docs](https://doc.anytype.io/anytype-docs), [Business](https://business.anytype.io/), [Download/product overview](https://download.anytype.io/?platform=desktop) |
| Capacities | [Product](https://capacities.io/product), [Getting started](https://docs.capacities.io/tutorials/getting-started), [MCP connector release 62](https://capacities.io/whats-new/release-62), [MCP connector release 66](https://capacities.io/whats-new/release-66), [Research](https://capacities.io/research) |
| AFFiNE | [Introduction](https://toeverything-affine.mintlify.app/introduction), [Knowledge base solution](https://affine.pro/solutions/knowledge-base), [AI](https://affine.pro/ai), [Self-host](https://getaffineapp.com/self-host), [Home](https://affine.pro/?trk=public_post_comment-text) |

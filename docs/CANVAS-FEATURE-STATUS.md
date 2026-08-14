# Scribli Canvas Feature Status

This document captures the current state of the Scribli Canvas feature after the first working implementation and live Electron workflow testing.

## Short Version

Canvas is no longer a toy prototype. It is a real usable MVP that can live inside Scribli documents, save and reopen, interact with mixed document content, and support basic visual board workflows.

It is not finished product quality yet. The foundation is working, but the user experience still needs proper pickers, richer node types, better library management, stronger visual organization tools, and larger-scale stress testing.

## What Works Now

Canvas blocks can be inserted into documents using the `scribli-canvas` code block language.

Blank canvas blocks render as interactive canvas placeholders with actions for creating a canvas, using templates, opening the canvas library, and importing canvas files.

Created canvases render as interactive visual boards with a visible toolbar.

Canvas state persists through the canvas MCP/API storage path and can be reopened from a bound canvas block.

Existing or older canvases can be loaded into a document by binding a canvas ID to a `scribli-canvas` block.

The canvas library can open, close, list canvases, and bind an existing canvas into a blank canvas block.

Text nodes can be added, edited, duplicated, selected, deleted, dragged, resized, and connected.

Canvas zoom controls work, including zoom in, zoom out, and reset.

Template insertion works, and undo can restore the pre-template canvas state.

Canvas export and import work with `.canvas` JSON files.

Canvas blocks coexist with normal Scribli document content. Live testing verified a page containing a top database, text blocks, an HTML block, a canvas in the middle, more text, and a bottom database.

The canvas toolbar visibility issue has been fixed. The toolbar buttons were present but inherited `opacity: 0` from normal Scribli block icon styling; Canvas now overrides that within canvas surfaces.

## Verified Workflows

The live Electron workflow matrix currently covers:

- Slash/manual canvas creation
- Blank canvas placeholder rendering
- Create canvas from placeholder
- Save and reopen
- Add text node
- Edit text node
- Duplicate node
- Delete node
- Connect nodes
- Drag node
- Resize node
- Zoom in, zoom out, and reset
- Apply template
- Undo template changes
- Open and close canvas library
- Bind an old canvas from the library
- Export canvas
- Import canvas
- Load an existing old canvas
- Render canvas inside a mixed-content document with databases, text, and HTML blocks
- Verify canvas toolbar buttons are visible, not just clickable

## Still Bare-Bones Areas

Canvas is currently strongest as a basic visual board for text nodes and references. It is not yet a polished visual thinking workspace.

Node creation is still rough. Document, block, asset, query, and database nodes need proper pickers and search flows instead of prompt-style input.

Canvas library UX is minimal. It needs clearer names, search, previews, rename, delete, duplicate, sort/filter controls, and a clear distinction between opening an existing canvas and inserting/binding one into the current block.

Connections are functional but basic. They need visible handles, labels, relationship types, easier reconnection, and better line routing.

Visual organization tools are missing. There are no frames, groups, colors, node styles, auto-layout, alignment tools, minimap, or fit-to-content behavior yet.

Database and document nodes are still reference cards, not rich live embedded views.

Large-canvas behavior needs more stress testing with many nodes, many edges, and multiple canvas blocks in the same document.

Keyboard and context-menu support is early. Users will expect shortcuts, right-click actions, selection boxes, copy/paste, multi-select, and better undo/redo behavior.

The visual design is functional but not yet refined. It should feel like a quiet, capable Scribli-native workspace rather than a technical block renderer with controls attached.

## Product Quality Assessment

Current status: usable MVP.

The feature is good enough to keep building on. The core render, persistence, interaction, library binding, import/export, and mixed-page coexistence are working.

It is not yet good enough to call finished. The next phase should focus on user-facing ergonomics, richer node types, and better navigation/organization rather than more storage plumbing.

## Recommended Next Phase

Priority one should be improving node creation. Add proper pickers for documents, blocks, assets, databases, and saved searches. A canvas should feel like it can pull from the user’s real Scribli workspace without asking them to paste IDs.

Priority two should be improving the canvas library. Users need to find, rename, preview, delete, duplicate, and intentionally bind canvases.

Priority three should be improving visual organization. Add frames, colors, grouping, basic alignment, and fit-to-content.

Priority four should be deeper Scribli integration. Document and database nodes should become richer cards with live metadata, previews, and better open/focus actions.

Priority five should be scale and polish. Test big canvases, multiple canvases per document, scroll behavior, keyboard workflows, and long-session stability.

## Definition Of Done For A Finished Canvas

Canvas should feel finished when a normal user can create a board inside a document, add existing Scribli objects through pickers, visually organize them, connect ideas clearly, save/reopen without surprises, manage canvases from a library, and understand the controls without needing an explanation from the developer.

Until then, Canvas should be treated as a strong foundation and active product surface, not a completed feature.

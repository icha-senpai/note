# Scribli Database Feature Status

This document captures the current state of Scribli database blocks after the mixed document workflow testing that placed databases above and below a live Canvas block.

## Short Version

Databases are present and functional as document blocks. Scribli can create and render database blocks, keep them in the document flow, and coexist with text, HTML, and Canvas content.

They do not yet feel like a complete Scribli-native database product. The foundation exists, but the user experience still needs better creation flows, templates, field setup, views, relations, import/export polish, and deeper Canvas integration.

## What Works Now

Database blocks can exist as real Scribli document blocks using the inherited attribute view block type.

Multiple databases can be placed in the same document without breaking the page layout.

Live testing verified a document with one database near the top of the page, a Canvas block in the middle, and another database near the bottom.

Database blocks can coexist with normal text blocks.

Database blocks can coexist with HTML blocks.

Database blocks can coexist with live Canvas blocks.

Each tested database block was created as a distinct database block with its own attribute view identifier.

The editor recognized the database blocks as real `NodeAttributeView` blocks rather than plain text placeholders.

The current behavior is good enough to use databases as structured document content in mixed pages.

## Verified Workflows

The live Electron mixed-document workflow currently covers:

- Create a document containing database blocks
- Place a database near the top of a document
- Place a second database near the bottom of the same document
- Insert text content between database blocks
- Insert an HTML block between database blocks
- Insert a live Canvas block between database blocks
- Verify both databases remain separate blocks
- Verify both databases have distinct attribute view IDs
- Verify the page keeps the expected vertical order
- Verify Canvas controls remain visible and usable while databases exist on the same page
- Verify database blocks do not prevent the mixed document from loading

## Still Bare-Bones Areas

Database creation is too raw. A user should be able to ask for a project tracker, reading list, inventory, lore index, task board, or research table and get a useful database with sensible fields.

Database templates are missing from the product flow. Scribli needs practical starter databases for common personal knowledge workflows.

Field setup needs a friendlier surface. Users should be able to add, rename, reorder, hide, and configure fields without feeling like they are touching inherited internal machinery.

Field types need stronger user-facing controls. Important types include text, number, date, checkbox, select, multi-select, URL, file, document reference, block reference, and database relation.

Views need to become a first-class concept. Table view is only the starting point; users will expect kanban, gallery, list, calendar, and possibly timeline views.

Relations are the biggest missing product layer. Databases become much more powerful when rows can reference rows in other databases, such as tasks linked to projects, characters linked to factions, or notes linked to sources.

Rollups and formulas are not yet part of the visible workflow. These are not required for the first polished version, but they matter if databases become a serious planning and knowledge system.

Database rows do not yet feel deeply connected to Canvas. Rows should be insertable into Canvas as live cards, and Canvas database nodes should be able to show useful previews instead of acting like simple reference cards.

Import and export need to be easier to discover. CSV import/export should be visible, predictable, and safe enough for normal users.

Empty states need polish. A blank database should invite the next meaningful action, not look like a technical object waiting for a developer.

MCP/API control needs to become higher level. The goal should be that a user can ask Codex to create a document with databases, fields, rows, relations, and a Canvas in the middle, and Codex can do the full setup without asking the user to paste internal IDs.

## Product Quality Assessment

Current status: functional foundation.

Databases are not broken, and they are usable as structured blocks inside a Scribli document. They passed the important mixed-content coexistence test with Canvas, text, and HTML.

They are not yet a polished Scribli feature. The current experience feels more like an inherited structured block capability than a purpose-built local-first database workspace.

The next phase should focus on product ergonomics rather than only backend plumbing. The key work is making database creation, field setup, templates, views, and relations feel understandable to a normal user.

## Recommended Next Phase

Priority one should be database creation from intent. Add flows where Scribli or Codex can create a useful database from a plain request, including fields, starter rows, and an appropriate default view.

Priority two should be practical templates. Start with databases users can immediately understand: tasks, projects, reading list, bookmarks, inventory, contacts, research sources, characters, locations, factions, and glossary terms.

Priority three should be field management. Add clear controls for adding, editing, reordering, hiding, and configuring fields.

Priority four should be views. Make table view solid first, then add kanban and gallery as the next likely high-value views.

Priority five should be relations. Let one database reference another in a way that is visible, searchable, and useful inside documents and Canvas.

Priority six should be Canvas integration. Database rows should become draggable or insertable live cards on Canvas, and Canvas should be able to create or reference databases through proper pickers.

Priority seven should be import/export. CSV import/export should be smooth enough that users trust it with real data.

## Definition Of Done For A Finished Database Feature

Databases should feel finished when a normal user can create a database from a plain-language need, choose or adjust fields, add and edit rows comfortably, switch between useful views, relate records across databases, import/export common data, place databases anywhere in a document, and use database rows naturally inside Canvas without touching internal IDs.

Until then, databases should be treated as a working foundation and important future product surface, not a completed feature.

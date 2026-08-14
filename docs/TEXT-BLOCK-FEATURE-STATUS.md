# Scribli Text Block Feature Status

This document captures the current state of Scribli text blocks in relation to the newer Canvas and database work.

## Short Version

Text blocks are the most mature part of the document experience. They already function as the core writing surface and the stable connective tissue between richer block types such as Canvas, databases, HTML blocks, embeds, and code blocks.

They are not bare bones in the same way Canvas and databases are. The missing work is mostly about higher-level composition: better Codex control, richer transforms, stronger references, long-document navigation, review workflows, and template-driven document assembly.

## What Works Now

Text blocks are the normal document authoring baseline.

Text can be inserted, edited, moved, and arranged inside documents.

Text blocks coexist cleanly with other block types.

Live mixed-document testing verified text blocks sitting between a top database, an HTML block, a live Canvas block, and a bottom database.

Text blocks are reliable enough to act as connective narrative between structured blocks.

Text blocks are currently the safest block type for Codex to generate because they do not require hidden IDs, field schemas, bindings, or special view configuration.

Text is flexible enough to support notes, outlines, explanations, lists, planning docs, research writeups, and mixed-content pages.

The current behavior is good enough for ordinary writing and document scaffolding.

## Verified Workflows

The live Electron mixed-document workflow currently covers:

- Create a document containing text blocks
- Place text above and below richer blocks
- Place text between two database blocks
- Place text near an HTML block
- Place text near a live Canvas block
- Preserve expected vertical document order
- Verify text does not interfere with Canvas controls
- Verify text does not interfere with database block placement
- Use text as explanatory connective content in a mixed document

## Still Bare-Bones Areas

Codex targeting needs to become more precise. A user should be able to ask Codex to add, rewrite, move, split, merge, or delete a specific text block without needing to know internal block IDs.

Structured transforms need stronger user-facing workflows. Scribli should make it easy to turn selected text into a checklist, table, database rows, Canvas cards, callout, heading structure, summary, or task list.

Text relationships are still mostly plain. A text block that references a document, database row, Canvas node, source, asset, or block should be able to feel live rather than becoming dead plain text.

Long-document navigation needs more power. Users will expect outline mode, section folding, quick jump, section summaries, and easier movement across large documents.

Review workflows are not yet a complete product surface. Comments, suggestions, diff review, accepted/rejected changes, and version comparison would make text blocks much better for serious writing and editing.

Templates and snippets should be easier to apply. Meeting notes, bug reports, project briefs, design specs, research notes, lore entries, character sheets, and planning docs should be quick to create from structured starting points.

Document assembly needs to become more intelligent. Codex should be able to create a polished document with headings, text sections, databases, Canvas blocks, HTML blocks, and references in the correct order without the user manually babysitting every block.

Text-to-structure workflows need deeper integration with databases and Canvas. Users should be able to highlight notes and turn them into database rows or Canvas nodes with useful metadata preserved.

Block-level history and authorship could be clearer. Scribli has broader history support, but text editing would benefit from user-facing clarity around what changed, when, and how to restore or compare it.

## Product Quality Assessment

Current status: mature foundation.

Text blocks are stable enough to be treated as the core writing layer. They are not the weak point of the current document experience.

The opportunity is not basic editing; it is composition power. Text should become the place where Scribli connects plain writing, structured data, visual thinking, review, and Codex-driven document generation.

Compared with Canvas and databases, text needs fewer foundational fixes and more workflow intelligence.

## Recommended Next Phase

Priority one should be better Codex control over text blocks. Codex needs reliable ways to target nearby headings, selected text, current cursor context, and semantic sections without exposing internal IDs to the user.

Priority two should be structured transforms. Add polished flows for text to checklist, text to database rows, text to Canvas cards, text to summary, and text to outline.

Priority three should be stronger references. Text should be able to link to documents, blocks, database rows, Canvas nodes, files, and sources in a way that remains visible and useful.

Priority four should be long-document ergonomics. Add or improve outline navigation, folding, jump controls, section movement, and section-level operations.

Priority five should be review and revision tools. Comments, suggestions, compare, accept/reject, and clearer block-level restore flows would make Scribli better for serious writing.

Priority six should be template-driven authoring. Text blocks should be easy to generate as part of reusable document patterns rather than only as loose paragraphs.

## Definition Of Done For A Finished Text Block Experience

Text blocks should feel finished when a normal user can write naturally, reorganize large documents comfortably, transform text into useful structures, reference live Scribli objects, review and restore changes confidently, and ask Codex to assemble or revise a mixed document without needing to touch internal IDs.

Until then, text blocks should be treated as the mature foundation of Scribli's document system, with the next work focused on smarter composition rather than basic editing.

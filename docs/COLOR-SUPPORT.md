# Scribli color support

Scribli uses one custom color picker for text, highlights, database categories, conditional color rules, Canvas cards and connectors, and interface colors. Colors accept six-digit HEX, shorthand HEX, or RGB channels from 0 to 255. Stored custom colors use opaque six-digit HEX; alpha channels and gradients are outside this picker.

## Controls

| Surface | Control | Persistence and undo |
| --- | --- | --- |
| Selected text or blocks | Appearance menu → Custom text color / Custom highlight color | Existing document formatting transactions, undo and redo |
| Database categories | Category editor → Custom color; Kanban column settings → Color → Custom color | Existing category transactions and undo; chips render across database layouts |
| Conditional colors | Click the color swatch in the rule editor, then save the rule editor | Per-view rules, existing transactions and undo |
| Canvas cards | Paint icon in the card header | Standard JSON Canvas `color` property, existing Canvas undo |
| Canvas connectors | Click the connector or focus it and press Enter/Space | Standard JSON Canvas `color` property, existing Canvas undo |
| Interface | Settings → Appearance → Interface colors | Accent, background, panels and text saved separately for light and dark modes |

The picker includes theme presets, the 16 most recent colors, and up to eight named palettes with 16 colors each. Add the current color to a palette or delete a palette using its buttons. Palettes and recent colors live in the workspace's existing `data/storage/local.json` under `local-scribli-colors`, so they survive restarts and changing Electron ports. No external service is involved.

Text, highlight, Canvas and interface pickers include a Default button. Interface settings also have a reset for the current light/dark mode; resetting one mode preserves the other. Category presets retain their original IDs from 1 to 14. Custom chips use a contrasting black or white label, and filled Kanban columns use a light tint of the category color. Canvas recognizes the six preset colors from JSON Canvas and preserves unmodified imported values.

Interface overrides use existing theme variables and update immediately. Built-in semantic colors, syntax highlighting, images, and plugin/theme styles remain their own surfaces; overrides do not recolor arbitrary third-party CSS. The user can still choose a text color with low contrast. Canvas color controls respect document read-only and history views.

## Compatibility

Databases using custom category colors require Scribli database spec 6, as do Calendar views and conditional color rules. Ordinary databases using inherited category IDs remain spec 5. The kernel validates custom category and interface colors before saving. Renderers validate stored colors before inserting CSS. Custom document text colors use the existing formatting representation; Canvas uses the existing JSON Canvas format.

## Verification

Validation uses a repository-owned isolated Electron workspace, leaving the installed application and normal workspace untouched. Checks cover invalid HEX/RGB input, saved palettes and recent colors, category undo/redo, custom conditional colors in Table/Card/Calendar/Kanban, selected text formatting and highlight RGB input, Canvas card/connector save/reload/undo, and separate light/dark interface overrides and reset. Automated checks include frontend typecheck/lint/tests, the full Windows kernel suite with `fts5 sqlcipher`, production frontend bundles, and a Windows kernel build. Packaging and installation are separate release steps.

The frontend suite passed all 81 tests, with zero lint errors and 16 existing warnings. Full kernel tests and the Windows kernel build passed. All three production frontend builds passed; desktop/export report the existing webpack bundle-size warnings. A full Electron restart on a different kernel port preserved the palette and recent colors; palette deletion and Markdown export of custom text/highlight formatting also passed. Canvas lock/unlock, connector keyboard access, and Default/reset undo passed. Evidence is retained under `tmp/color-support-2026-10-10/`.

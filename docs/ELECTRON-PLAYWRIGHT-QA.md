# Electron Playwright QA

Scribli has a focused Electron Playwright audit for document creation, mixed document layout, database blocks, HTML blocks, and Canvas card workflows.

Run from `app/`:

```powershell
npx --yes --package playwright playwright test --config tests\electron\playwright.config.js --reporter=line
```

The suite is intentionally capped at 90 seconds in `app/tests/electron/playwright.config.js`. It launches one isolated Electron workspace and uses one app window. The audit exercises the ordinary development startup path without passing `--workspace`; the harness seeds an isolated `workspace.json`, sets `SCRIBLI_CONFIG_DIR`, and sets `SCRIBLI_WORKSPACE_PATH` so the run does not write temporary QA workspaces into the user's real Scribli config.

The audit creates a normal document, edits and saves a text block, creates and deletes a throwaway document, reopens an older document, creates a mixed document with a top database block, text block, HTML block, Canvas block, bottom text block, and bottom database block, then verifies the visible layout order. It also verifies Canvas toolbar visibility, card duplicate/delete, text edit/add persistence, dialog teardown, and card drag persistence.

Known runner note: `npm run`/repo-local npx cache paths currently fail on this Windows checkout because npm/pnpm cache access and Playwright worker spawning hit `EPERM`. The direct `npx --package playwright` command above is the verified runner path.

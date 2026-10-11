# Scribli Future Features

This shortlist records features discussed on 2026-10-10. These are future candidates, not implemented features or a committed release schedule. Scribli is based on SiYuan and remains AGPL-3.0 licensed.

## Recommended Order

1. Database forms.
2. Reusable action buttons.
3. Database charts.
4. Recurring tasks and templates.
5. Timeline view.
6. Simple database automations, building on the action system.
7. Typed notes as a separate, larger project.

Forms, buttons, and charts are the preferred next batch because they make existing databases easier to use. Typed notes are the larger-project recommendation, especially for research and worldbuilding.

## Shortlist

### Database Forms

- [ ] Add a configurable form for creating entries in an existing database.
- [ ] Support required fields, default values, field ordering, and validation.
- [ ] Provide an "Add another" action for repeated entry without returning to the table.
- [ ] Keep the initial form workflow inside the local app; shared or server forms can be considered separately.

Examples: log a bug, add a book, capture a recipe, or create a task without editing individual table cells.

Inspiration: [Notion forms](https://www.notion.com/help/forms).

### Reusable Action Buttons

- [ ] Let users configure buttons that update fields or create database entries.
- [ ] Support multiple actions in a defined order.
- [ ] Use existing transactions and provide undo for the complete action.
- [ ] Make missing fields, unavailable targets, and failed actions understandable.

Examples: "Complete task" sets status and completion date; "Start project" creates an entry with predefined fields; "Log reading session" adds a dated record.

Inspiration: [Notion database buttons](https://www.notion.com/help/database-buttons).

### Database Charts

- [ ] Add a native chart view configured from database fields and filters.
- [ ] Start with counts and numeric aggregates rendered as bar, line, and donut charts.
- [ ] Allow clicking a chart group to inspect its underlying entries.
- [ ] Keep chart configuration with the database view and update charts when data changes.

Examples: task counts by status, spending totals by category, or reading progress over time. Scribli already renders chart code blocks; this feature adds a database-driven configuration surface.

Inspiration: [Notion charts](https://www.notion.com/help/charts).

### Recurring Tasks and Templates

- [ ] Extend existing templates with daily, weekly, monthly, and custom recurrence schedules.
- [ ] Create new entries for each occurrence rather than overwriting completed records.
- [ ] Define timezone, missed-occurrence behavior, and duplicate prevention.
- [ ] Provide pause, resume, and schedule editing controls.

Examples: weekly chores, monthly reviews, or recurring meeting notes. Since Scribli runs locally, define what happens when the app is closed and how missed occurrences are handled on reopening.

Inspiration: [Notion repeating database templates](https://www.notion.com/en-gb/help/guides/automate-work-repeating-database-templates).

### Timeline View

- [ ] Display database entries as horizontal date ranges.
- [ ] Provide useful time scales, navigation, and a Today action.
- [ ] Support moving and resizing editable date ranges with undo.
- [ ] Retain existing filters and support inspecting the associated entry.

Examples: project plans, publishing schedules, or fictional histories. Build on the existing date fields and calendar work.

Inspiration: [Notion timelines](https://www.notion.com/help/timelines).

### Simple Database Automations

- [ ] Start with entry-added and field-changed triggers.
- [ ] Support conditions and local field-update or entry-creation actions.
- [ ] Reuse the action-button execution system where practical.
- [ ] Bound execution, prevent loops, and define failure, rollback, and undo behavior.
- [ ] Provide enable/disable controls and clear execution feedback.

Examples: when status becomes Done, record the completion date; when an entry is added, apply selected defaults. Keep the first version focused on local database changes.

Inspiration: [Notion database automations](https://www.notion.com/help/database-automations). Related upstream work is recorded in [the SiYuan intake review](UPSTREAM-INTAKE-2026-10-10.md).

### Typed Notes

- [ ] Let users define reusable note types with fields, templates, and overview views.
- [ ] Offer types such as Person, Book, Character, Location, or Research Source.
- [ ] Allow existing notes to adopt or change a type without losing their content.
- [ ] Integrate types with existing documents, database membership, relations, and links.
- [ ] Design backup-safe stored-data changes before implementation.

Examples: a Character note starts with a consistent character sheet and appears in a Characters overview; a Research Source note includes author, URL, publication date, and linked notes. This should extend Scribli's existing model rather than introduce a disconnected object system.

Inspiration: [Capacities object types](https://docs.capacities.io/reference/content-types) and [templates](https://docs.capacities.io/reference/templates).

## Shared Requirements

All candidates should work locally, preserve encrypted-notebook boundaries, use existing transaction and undo mechanisms where applicable, and keep user-facing configuration understandable. Any optional network behavior must be explicitly configured by the user. Future stored-format changes need compatibility handling and regression coverage.

The current calendar and conditional-color implementation is recorded separately in [the upstream intake document](UPSTREAM-INTAKE-2026-10-10.md); those features are not part of this future shortlist.

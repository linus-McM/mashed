# Story 03: Toolbar Builder — markdownToolbarBuilder.ts

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** Story 02 (imports `MarkdownMenuSettings` type from the store or Wails models)
**Status:** done
**UI-facing:** no (pure module; UI effect is visible via the editor in stories 06+)

## Description

Implement the pure function `buildToolbarFromSettings(settings)` that returns a Crepe `buildToolbar` callback. The callback constructs a `GroupBuilder<ToolbarItem>` with one group containing exactly the enabled items from the six supported selection-toolbar entries (bold, italic, strikethrough, code, link, latex). This is the key piece of logic that lets users curate the toolbar — without it, the settings are disconnected from the editor.

## Developer Notes

### Architecture
- New file: `/Users/linus/Development/mashed/frontend/src/components/markdownToolbarBuilder.ts`
- Export a single named function: `export function buildToolbarFromSettings(s: MarkdownMenuSettings): (builder: GroupBuilder<ToolbarItem>) => void`
- The returned function seeds a single group named `'toolbar'` and conditionally adds items using `group.addItem(key, {icon, active, onRun})`.

### Required imports (primary strategy)
Copy this block verbatim from the plan (verified against Milkdown/Crepe `packages/crepe/src/feature/toolbar/config.ts`):

```ts
import type { GroupBuilder } from '@milkdown/crepe/utils/group-builder';
import type { ToolbarItem } from '@milkdown/crepe/feature/toolbar/config';
import { toggleLinkCommand } from '@milkdown/kit/component/link-tooltip';
import { commandsCtx } from '@milkdown/kit/core';
import {
  emphasisSchema, inlineCodeSchema,
  isMarkSelectedCommand, isNodeSelectedCommand,
  linkSchema, strongSchema,
  toggleEmphasisCommand, toggleInlineCodeCommand, toggleStrongCommand,
} from '@milkdown/kit/preset/commonmark';
import {
  strikethroughSchema, toggleStrikethroughCommand,
} from '@milkdown/kit/preset/gfm';
import {
  boldIcon, codeIcon, functionsIcon,
  italicIcon, linkIcon, strikethroughIcon,
} from '@milkdown/crepe/icons';
import { toggleLatexCommand } from '@milkdown/crepe/feature/latex/command';
import { mathInlineSchema } from '@milkdown/crepe/feature/latex/inline-latex';
```

Verify subpaths against `node_modules/@milkdown/crepe/package.json#exports` before assuming they resolve. If any subpath is not exported, check the bundle entry (`@milkdown/crepe`) for the named symbol and adjust imports.

### Item table (source of truth)

| Key             | Icon                  | active                                                              | onRun                                                    |
| --------------- | --------------------- | ------------------------------------------------------------------- | -------------------------------------------------------- |
| `bold`          | `boldIcon`            | `call(isMarkSelectedCommand.key, strongSchema.type(ctx))`           | `call(toggleStrongCommand.key)`                          |
| `italic`        | `italicIcon`          | `call(isMarkSelectedCommand.key, emphasisSchema.type(ctx))`         | `call(toggleEmphasisCommand.key)`                        |
| `strikethrough` | `strikethroughIcon`   | `call(isMarkSelectedCommand.key, strikethroughSchema.type(ctx))`    | `call(toggleStrikethroughCommand.key)`                   |
| `code`          | `codeIcon`            | `call(isMarkSelectedCommand.key, inlineCodeSchema.type(ctx))`       | `call(toggleInlineCodeCommand.key)`                      |
| `link`          | `linkIcon`            | `call(isMarkSelectedCommand.key, linkSchema.type(ctx))`             | `call(toggleLinkCommand.key)`                            |
| `latex`         | `functionsIcon`       | `call(isNodeSelectedCommand.key, mathInlineSchema.type(ctx))`       | `call(toggleLatexCommand.key)`                           |

`call = ctx.get(commandsCtx).call` — factor this out if it improves readability.

### Fallback strategy — CSS data-attribute hiding
If ANY of the primary-strategy imports cannot be resolved (the subpath is not listed in `@milkdown/crepe`'s `exports` map), fall back to:

1. `buildToolbarFromSettings` returns a callback that calls Crepe's default `getGroups` directly (re-exported or copied via the bundle import) without filtering.
2. Set container-level `data-toolbar-<key>="off"` attributes on `.milkdown` based on settings (from `MarkdownEditor.svelte` — covered in story 05).
3. Add CSS rules to `frontend/src/styles/crepe-mashed.css`:

```css
.milkdown[data-toolbar-bold="off"] .toolbar-item[data-key="bold"] { display: none; }
/* ...one rule per toggle... */
```

If going the fallback route, confirm Crepe emits `data-key` on `.toolbar-item` elements. If it doesn't, add per-item class names via `buildToolbar` (e.g., tag the item's element by injecting a class into `icon` SVG wrapper).

**Primary strategy is preferred — do not ship the fallback without first attempting the primary.** Document which strategy shipped in a top-of-file comment.

### Behavior contract
- The callback must add enabled items in the fixed order: bold, italic, strikethrough, code, link, latex. This order is the same as Crepe's `getGroups` default, so the toolbar renders identically to Crepe's default when all six are on.
- If ALL six settings are false, the callback still creates the empty `'toolbar'` group without calling any `addItem`. The group presence is required by Crepe's builder contract.
- The callback must NOT call `addItem` for disabled items — it must not add them then hide them. That's the fallback strategy's job.

### Technical Considerations
- **Pure function** — `buildToolbarFromSettings(s)` must not capture external state. It takes settings and returns a callback; the callback takes the Crepe builder and mutates it. No side effects.
- **TypeScript** — `GroupBuilder<ToolbarItem>` is the generic; ensure the type import resolves. If Milkdown's type export path differs, use `any` ONLY as a last resort and flag as a deviation.
- **Tree-shaking** — each icon/command is imported individually. Do NOT default-import entire packages; bundle size matters.

### Risks & Edge Cases
- **Milkdown API drift** — the plan cites specific subpath exports; if the installed Crepe version differs, imports may fail at build time. Pin the Crepe version in `package.json` if it's not already, and surface a clear build error.
- **Active-state command signature** — `isMarkSelectedCommand.key` may require a specific `call` shape. If `ctx.get(commandsCtx).call(key, payload)` is wrong, read Crepe's own `config.ts` source for the correct invocation.
- **Latex not installed** — if the Milkdown LaTeX feature isn't enabled in Crepe construction, `toggleLatexCommand` may be undefined at runtime. Guard by only importing latex symbols and defer the `group.addItem('latex', ...)` call — if the runtime command is absent, the button click will noop. This is acceptable; LaTeX default is `off` anyway.

### Reference Files
- `/Users/linus/Development/mashed/frontend/src/components/MarkdownEditor.svelte` — existing Crepe integration (no `featureConfigs` yet; that's story 05).
- `/Users/linus/Development/mashed/frontend/src/lib/stores/markdownMenuSettings.ts` — source of the `MarkdownMenuSettings` type.
- Milkdown source (via `node_modules/@milkdown/crepe/src/feature/toolbar/config.ts` or the bundled dist) — the canonical shape to mirror.

## Acceptance Criteria

AC-1: Pure function contract
- Given any `MarkdownMenuSettings` value
- When `buildToolbarFromSettings(s)` is called twice with the same input
- Then both calls return distinct callback functions
- And neither call mutates `s`

AC-2: All-enabled builds six items in fixed order
- Given `s = {bold:true, italic:true, strikethrough:true, code:true, link:true, latex:true}`
- When the returned callback runs against a fake `GroupBuilder`
- Then `addGroup('toolbar')` is called once
- And `addItem` is called with keys `['bold','italic','strikethrough','code','link','latex']` in that exact order

AC-3: All-disabled creates empty group
- Given `s = {bold:false, italic:false, strikethrough:false, code:false, link:false, latex:false}`
- When the callback runs against a fake builder
- Then `addGroup('toolbar')` is called once
- And `addItem` is never called

AC-4: Partial enable adds only enabled items
- Given `s = {bold:true, italic:false, strikethrough:true, code:false, link:false, latex:true}`
- When the callback runs
- Then `addItem` is called with keys exactly `['bold','strikethrough','latex']` in that order

AC-5: Each item carries required fields
- Given any enabled item
- When `addItem` is called
- Then the item config has `icon`, `active`, and `onRun` fields
- And `icon` is a truthy value (the imported icon reference)
- And `active` and `onRun` are functions

AC-6: Chosen strategy is documented
- Given the source file
- When inspected
- Then the first comment block states "Strategy: primary (reconstruction)" or "Strategy: fallback (CSS masking)"

## BDD Test Scenarios

```gherkin
Feature: Conditional toolbar builder

  Scenario: All six items enabled — full Crepe default order
    Given settings with every toggle on
    When the builder callback is invoked with a fake GroupBuilder
    Then it calls addGroup once with "toolbar"
    And addItem is called for bold, italic, strikethrough, code, link, latex in that order

  Scenario: Only bold and latex enabled
    Given settings {bold:true, italic:false, strikethrough:false, code:false, link:false, latex:true}
    When the builder callback runs
    Then addItem is called exactly twice
    And the keys are ["bold","latex"] in order

  Scenario: All disabled
    Given every setting is false
    When the builder callback runs
    Then addGroup is called once with "toolbar"
    And addItem is never called

  Scenario: Builder does not mutate settings
    Given a frozen settings object
    When buildToolbarFromSettings is called and its callback runs
    Then no TypeError is thrown
    And the settings object is unchanged

  Scenario: Each enabled item has icon, active, onRun
    Given settings with bold enabled
    When the callback runs and addItem is captured
    Then the item config has a non-null icon
    And active is a function
    And onRun is a function
```

## Tasks / Subtasks

- [ ] Task 1: Verify Milkdown subpath exports (AC-6) — frontend
  - [ ] Inspect `node_modules/@milkdown/crepe/package.json` `exports` map.
  - [ ] Confirm each import subpath resolves; note deviations.
  - [ ] If all resolve, proceed with primary; else pivot to fallback (see Technical Considerations).
- [ ] Task 2: Implement builder — primary strategy (AC-1, AC-2, AC-3, AC-4, AC-5) — frontend
  - [ ] Create `frontend/src/components/markdownToolbarBuilder.ts`.
  - [ ] Add the top-of-file comment declaring strategy.
  - [ ] Import the required commands/icons/schemas/types.
  - [ ] Implement `buildToolbarFromSettings` with the six conditional `addItem` blocks.
- [ ] Task 3: Implement fallback (only if Task 1 requires) — frontend
  - [ ] Re-export default `getGroups` wrapper, relying on `MarkdownEditor` to set `data-toolbar-*` attributes.
  - [ ] Add CSS rules to `frontend/src/styles/crepe-mashed.css` for each item key.
- [ ] Task 4: Unit tests (all ACs) — frontend
  - [ ] Create `frontend/src/components/__tests__/markdownToolbarBuilder.test.ts`.
  - [ ] Build a fake `GroupBuilder` with spyable `addGroup` and `addItem`.
  - [ ] Cover all six BDD scenarios.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ line coverage on `markdownToolbarBuilder.ts`
- [ ] `svelte-check` passes with 0 warnings, 0 errors
- [ ] Vitest unit tests pass
- [ ] `/simplify` run on the new file
- [ ] Top-of-file comment declares primary vs fallback strategy
- [ ] Code review: no CRITICAL/HIGH issues

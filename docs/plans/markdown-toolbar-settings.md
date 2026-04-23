# Markdown Toolbar Settings

User-configurable selection toolbar for the Crepe-based markdown editor
(`MarkdownEditor.svelte`). Users toggle individual toolbar buttons in the
Settings page; changes persist via Wails config and apply to the editor
when the user leaves the Settings view.

## Motivation

Crepe's selection toolbar ships with a fixed button set. We want users
to curate which items appear, without re-architecting the editor. The
current default also renders extra floating affordances (block handle,
slash menu, link popover, placeholder) due to unfiltered feature flags —
this plan keeps all of those at Crepe defaults and only targets the
selection toolbar as the first scoped slice.

Followups (out of scope here):

- Slash menu item toggles (BlockEdit group/item nulls)
- Link tooltip + placeholder feature toggles
- Group-level master toggles
- Drag reorder of toolbar buttons

## Scope

**In**

- Six selection-toolbar items: Bold, Italic, Strikethrough, Code, Link, LaTeX.
- On/off toggle per item.
- Persistence in Wails config (`Config.MarkdownMenu`).
- Re-apply to live editor only when the user closes the Settings view
  (avoids destroying cursor state mid-edit).

**Out**

- Slash menu, link tooltip, placeholder, image block, reorder.

## Architecture

### Backend — `app.go`

```go
type MarkdownMenuSettings struct {
  Bold          bool `json:"bold"`
  Italic        bool `json:"italic"`
  Strikethrough bool `json:"strikethrough"`
  Code          bool `json:"code"`
  Link          bool `json:"link"`
  Latex         bool `json:"latex"`
}

// Config embed
type Config struct {
  // ... existing fields ...
  MarkdownMenu *MarkdownMenuSettings `json:"markdownMenu,omitempty"`
}

func (a *App) DefaultMarkdownMenuSettings() MarkdownMenuSettings {
  return MarkdownMenuSettings{
    Bold: true, Italic: true, Strikethrough: true,
    Code: true, Link: true, Latex: false,
  }
}

func (a *App) GetMarkdownMenuSettings() MarkdownMenuSettings { ... }
func (a *App) SetMarkdownMenuSettings(s MarkdownMenuSettings) error { ... }
```

No validation needed — all fields are bools.

Regenerate TS bindings: `wails generate module` (produces
`frontend/wailsjs/go/main/App.d.ts` entries for the new methods).

### Frontend store — `frontend/src/lib/stores/markdownMenuSettings.ts`

Mirrors `editorSettings.js`:

```ts
export const markdownMenuSettings = writable<MarkdownMenuSettings>({ ...defaults });
export const markdownMenuDirty    = writable<boolean>(false);

export function initMarkdownMenuSettings(s: Partial<MarkdownMenuSettings> | null): void { ... }

export async function updateMarkdownMenuItem<K extends keyof MarkdownMenuSettings>(
  key: K,
  value: MarkdownMenuSettings[K],
): Promise<void> {
  markdownMenuSettings.update(c => ({ ...c, [key]: value }));
  markdownMenuDirty.set(true);
  await SetMarkdownMenuSettings(get(markdownMenuSettings));
}

export function clearMarkdownMenuDirty(): void {
  markdownMenuDirty.set(false);
}
```

### Toolbar builder — `frontend/src/components/markdownToolbarBuilder.ts`

Pure function returning a Crepe `buildToolbar` callback that seeds the
default six groups and conditionally adds each item based on settings.
When `buildToolbar` is provided, Crepe skips its own default grouping
and calls the callback with a fresh `GroupBuilder<ToolbarItem>` — so
the callback must construct everything itself.

#### Research — reconstructing Crepe defaults

Crepe's default grouping lives in
`packages/crepe/src/feature/toolbar/config.ts` (function `getGroups`).
It adds six items to a single group, each using:

| Key             | Icon import             | Toggle command             | Active checker (schema)                              |
| --------------- | ----------------------- | -------------------------- | ---------------------------------------------------- |
| `bold`          | `boldIcon`              | `toggleStrongCommand`      | `isMarkSelectedCommand(strongSchema.type(ctx))`      |
| `italic`        | `italicIcon`            | `toggleEmphasisCommand`    | `isMarkSelectedCommand(emphasisSchema.type(ctx))`    |
| `strikethrough` | `strikethroughIcon`     | `toggleStrikethroughCommand` | `isMarkSelectedCommand(strikethroughSchema.type(ctx))` |
| `code`          | `codeIcon`              | `toggleInlineCodeCommand`  | `isMarkSelectedCommand(inlineCodeSchema.type(ctx))`  |
| `link`          | `linkIcon`              | `toggleLinkCommand`        | `isMarkSelectedCommand(linkSchema.type(ctx))`        |
| `latex`         | `functionsIcon`         | `toggleLatexCommand`       | `isNodeSelectedCommand(mathInlineSchema.type(ctx))`  |

Imports (verified from `packages/crepe/src/feature/toolbar/config.ts`
header, repomix lines 14826–14858):

```ts
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
import { GroupBuilder } from '@milkdown/crepe/utils/group-builder';
import {
  boldIcon, codeIcon, functionsIcon,
  italicIcon, linkIcon, strikethroughIcon,
} from '@milkdown/crepe/icons';
```

Latex paths (from same file, lines 14857–14858):

```ts
import { toggleLatexCommand } from '@milkdown/crepe/feature/latex/command';
import { mathInlineSchema } from '@milkdown/crepe/feature/latex/inline-latex';
```

Export paths above may need adjustment to Crepe's actual subpath
exports — verify via `package.json#exports` during implementation.
Fallback if a subpath is not exported: import from the built bundle
(`@milkdown/crepe`) and extract named symbols.

#### Builder implementation

```ts
import type { GroupBuilder } from '@milkdown/crepe/utils/group-builder';
import type { ToolbarItem } from '@milkdown/crepe/feature/toolbar/config';
// imports as listed above

export function buildToolbarFromSettings(s: MarkdownMenuSettings) {
  return (builder: GroupBuilder<ToolbarItem>) => {
    const group = builder.addGroup('toolbar');

    if (s.bold) {
      group.addItem('bold', {
        icon: boldIcon,
        active: (ctx) => ctx.get(commandsCtx).call(isMarkSelectedCommand.key, strongSchema.type(ctx)),
        onRun: (ctx) => ctx.get(commandsCtx).call(toggleStrongCommand.key),
      });
    }
    if (s.italic) { /* same shape */ }
    if (s.strikethrough) { /* ... */ }
    if (s.code) { /* ... */ }
    if (s.link) {
      group.addItem('link', {
        icon: linkIcon,
        active: (ctx) => ctx.get(commandsCtx).call(isMarkSelectedCommand.key, linkSchema.type(ctx)),
        onRun: (ctx) => ctx.get(commandsCtx).call(toggleLinkCommand.key),
      });
    }
    if (s.latex) {
      group.addItem('latex', {
        icon: functionsIcon,
        active: (ctx) => ctx.get(commandsCtx).call(isNodeSelectedCommand.key, mathInlineSchema.type(ctx)),
        onRun: (ctx) => ctx.get(commandsCtx).call(toggleLatexCommand.key),
      });
    }
  };
}
```

#### Fallback strategy — CSS data attributes

If the upstream subpath imports above cannot be resolved cleanly, fall
back to CSS-only hiding. Crepe renders each item with class
`.toolbar-item` (confirmed at repomix line 15946). Add per-item keys
via `buildToolbar` that just tag the element, then hide via
container-level `data-*`:

```css
.milkdown[data-toolbar-bold="off"] .toolbar-item[data-key="bold"] { display: none; }
```

This requires less reconstruction (keep Crepe defaults, just mask the
DOM) but needs confirmation that Crepe emits `data-key` or a
per-item class. Prefer the full reconstruction above as primary.

### Settings page — `frontend/src/views/Settings.svelte`

#### Layout refactor

Current layout stacks every settings section vertically inside
`.col-settings`. This change splits that column into a **2-column grid
of panels**, keeping the theme list as-is on the far left. Each type of
setting (Font, Sidebar Width, Editor, Theme Extensions, UI AST adapter,
Markdown Editor) becomes its own visually separated panel.

New structure:

```
.settings-body (flex row)
├── .col-themes      (existing, 280px, theme list)
└── .col-settings    (flex 1, NEW 2-col grid of panels)
    ├── .settings-col-1
    │   ├── <panel> Font
    │   ├── <panel> Sidebar Width
    │   └── <panel> Editor
    └── .settings-col-2
        ├── <panel> Theme Extensions
        ├── <panel> UI AST adapter
        └── <panel> Markdown Editor   ← new
```

#### Panel component

Extract the repeated `.settings-section` markup into an inline wrapper
(or small `Panel.svelte` if preferred) so every type of setting gets a
consistent card treatment:

- Background: `var(--bg-surface)`
- Border: `1px solid var(--border-subtle)`
- Radius: `var(--radius-md)`
- Padding: `var(--sp-lg)`
- Gap between panels within a column: `var(--sp-lg)`

Existing `.section-title`, `.section-desc`, `.subsection-title`,
`.setting-row`, `.setting-toggle`, `.setting-select`, `.setting-number`
classes are reused inside each panel — no changes to them.

#### CSS additions

```css
.col-settings {
  /* override existing single-column layout */
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--sp-lg);
  align-items: start;
}

.settings-col {
  display: flex;
  flex-direction: column;
  gap: var(--sp-lg);
  min-width: 0;
}

.settings-panel {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--sp-lg);
}

.settings-panel .settings-section {
  margin-bottom: 0; /* panel provides spacing */
}

/* Narrow viewport fallback — collapse to single column */
@media (max-width: 1100px) {
  .col-settings { grid-template-columns: 1fr; }
}
```

#### Panel assignments

| Column 1                     | Column 2                      |
| ---------------------------- | ----------------------------- |
| Font                         | Theme Extensions              |
| Sidebar Width                | UI AST adapter                |
| Editor                       | **Markdown Editor** (new)     |

Column 1 = per-document display/edit settings. Column 2 = integrations
and editor feature config. This grouping is a soft rule — reorder if
testing shows the Editor panel is tall enough to leave column 2
noticeably shorter.

#### Markdown Editor panel

```svelte
<div class="settings-panel">
  <section class="settings-section" data-testid="markdown-menu-section">
    <h2 class="section-title">Markdown Editor</h2>
    <p class="section-desc">
      Selection toolbar items. Changes apply when you close Settings.
    </p>

    {#each TOOLBAR_ITEMS as item}
      <div class="setting-row">
        <span class="setting-label">{item.label}</span>
        <button
          class="setting-toggle"
          class:active={$markdownMenuSettings[item.key]}
          data-testid={`toolbar-toggle-${item.key}`}
          on:click={() => updateMarkdownMenuItem(item.key, !$markdownMenuSettings[item.key])}
        >
          {$markdownMenuSettings[item.key] ? 'On' : 'Off'}
        </button>
      </div>
    {/each}
  </section>
</div>
```

`TOOLBAR_ITEMS` is a constant array `[{ key: 'bold', label: 'Bold' }, ...]`.

On `dispatch('back')`: call `clearMarkdownMenuDirty()` before dispatching.

### Editor wiring — `frontend/src/components/MarkdownEditor.svelte`

1. Import store + builder.
2. Pass `featureConfigs` into constructor:

```js
import { get } from 'svelte/store';
import { markdownMenuSettings, markdownMenuDirty } from '../lib/stores/markdownMenuSettings';
import { buildToolbarFromSettings } from './markdownToolbarBuilder';

// inside initEditor, replace constructor:
crepe = new Crepe({
  root: container,
  defaultValue: markdown,
  featureConfigs: {
    [Crepe.Feature.Toolbar]: {
      buildToolbar: buildToolbarFromSettings(get(markdownMenuSettings)),
    },
  },
});
```

3. Re-init trigger (deferred until Settings closes):

```js
let lastAppliedSettings = JSON.stringify(get(markdownMenuSettings));

$: {
  if (!$markdownMenuDirty && crepe && !loading) {
    const current = JSON.stringify($markdownMenuSettings);
    if (current !== lastAppliedSettings) {
      lastAppliedSettings = current;
      destroyEditor().then(initEditor);
    }
  }
}
```

Re-init only fires when `dirty === false` AND the applied settings differ
from the current store value. Closing Settings clears the dirty flag,
which drops into this branch and re-mounts Crepe once.

### App hydration — `frontend/src/App.svelte`

On mount, alongside other config hydrations:

```js
const cfg = await GetConfig();
initMarkdownMenuSettings(cfg.markdownMenu);
```

## Tests

### Go — `app_test.go`

- `TestDefaultMarkdownMenuSettings` — returns expected defaults.
- `TestSetGetMarkdownMenuSettings` — round-trip persists and loads.
- `TestConfigMarkdownMenuOmitempty` — nil `MarkdownMenu` serialises without the key.

### Svelte — new test files

- `markdownToolbarBuilder.test.ts` — given settings `X`, builder calls
  `addItem` only for enabled items, in expected order.
- `markdownMenuSettings.test.ts` — `updateMarkdownMenuItem` flips dirty,
  persists via Wails mock; `clearMarkdownMenuDirty` resets.
- `Settings.test.ts` — toggling Bold calls `SetMarkdownMenuSettings`;
  clicking Back clears the dirty flag before `back` event fires.
- `MarkdownEditor.test.ts` — editor does not re-init while dirty=true;
  re-inits once when dirty flips to false with changed settings.

## Verification

1. Run `npm run dev` from `/frontend`, launch Mashed.
2. Open Settings. Confirm body renders 3 regions: theme list (left),
   settings column 1 (Font / Sidebar Width / Editor panels), settings
   column 2 (Theme Extensions / UI AST adapter / Markdown Editor panels).
   Each panel has a visible card border.
3. Resize window narrow (&lt; 1100px). Columns collapse to one, panels
   still cards.
4. Settings → Markdown Editor panel → toggle off Bold and Italic.
5. Click Back.
6. Open a `.md` file. Select text. Toolbar shows only Strike / Code / Link.
7. Return to Settings. Toggle LaTeX on.
8. Click Back. Select text. Toolbar now shows Strike / Code / Link / LaTeX.
9. Quit and relaunch app. Toggles persist.
10. While in Settings with dirty toggles, confirm the open markdown editor
    behind Settings has NOT re-initialised (no cursor jump, scroll position
    retained).

## Resolved decisions

- **Crepe toolbar item defaults** — researched in the Milkdown source
  (repomix `packages/crepe/src/feature/toolbar/config.ts`). Reconstructed
  item table, imports, and builder shape above under
  "Toolbar builder — Research". Primary strategy reconstructs
  defaults via Crepe-exported icons and commands; fallback is
  CSS data-attribute hiding.

- **Cursor preservation on re-init** — `MarkdownEditor.svelte` re-init
  branch must call `saver.flush()` before `destroyEditor()`, mirroring
  the existing filePath-change handler (line 135). Reload reads from
  disk, so unsaved in-memory edits would otherwise be lost.

- **First-run config migration** — `DefaultMarkdownMenuSettings()` on
  the Go side is the single source of defaults. `GetMarkdownMenuSettings()`
  returns those defaults when `cfg.MarkdownMenu == nil`. The TS store
  does not need its own defaults object — it reads whatever the
  backend returned through `GetConfig()`. Chosen defaults:

  | Item          | Default |
  | ------------- | ------- |
  | Bold          | on      |
  | Italic        | on      |
  | Strikethrough | on      |
  | Code          | on      |
  | Link          | on      |
  | LaTeX         | off     |

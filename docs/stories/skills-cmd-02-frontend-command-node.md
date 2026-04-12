# skills-cmd-02: Frontend `CommandNode.svelte` + drop handler

**Status:** ready
**Domain:** frontend
**Size:** M
**Depends on:** skills-cmd-01
**Phase:** 2

## Description

Add a new Svelte-Flow node type `command` backed by `CommandNode.svelte`, and teach `CanvasPane.svelte`'s drop handler to accept the `application/mashed-asset` MIME type and instantiate a command node when the dropped asset's `role === "command"`. Skills must be rejected (the sidebar already blocks the drag — this is belt-and-braces on the receiving side).

This is the user-visible half of Phase 2. After this story, a user can drag a mashed-ready command from the sidebar onto the canvas and see a new node render. The node remains inert (fails on run) until Phase 3 ships.

**This is a UI-facing story — route through ui-architect review before implementation.**

## Developer Notes

- **Files to create/modify:**
  - `frontend/src/components/bmad/CommandNode.svelte` — NEW. Mirror shape of `ProcessNode.svelte` for visual consistency; different icon (`Terminal` or `Slash` from `lucide-svelte`) to signal "slash command".
  - `frontend/src/components/bmad/CanvasPane.svelte` — extend the existing drop handler to branch on `application/mashed-asset`.
  - `frontend/src/views/WorkflowBuilder.svelte` — register `command: CommandNode` in the `nodeTypes` map (around line 63).
- **Types/symbols introduced:** Svelte component `CommandNode`. Node-data shape:
  ```js
  {
    id: `cmd-${Date.now()}`,
    type: 'command',
    position: { x, y },
    data: {
      label: asset.name,
      nodeType: 'command',
      config: {
        commandName: asset.name,
        commandPath: asset.path,
        commandDescription: asset.description || '',
      },
      status: 'pending',
    },
  }
  ```
- **Drop handler snippet (from plan §Phase 2 "Accept the drop"):**
  ```js
  const mashedAssetRaw = e.dataTransfer.getData('application/mashed-asset');
  if (mashedAssetRaw) {
    const asset = JSON.parse(mashedAssetRaw);
    if (asset.role !== 'command') return; // belt-and-braces
    // compute position via project()
    // append to $nodes
  }
  ```
- **Visual requirements:**
  - Body: command name (bold) + truncated description (1 line, ellipsis).
  - NO `in:` / `out:` artifact rows — commands do not declare I/O the way `ProcessDef` does.
  - Status badge (pending/running/complete/failed) re-uses the same badge component `ProcessNode.svelte` already renders.
  - Use design-system tokens only — no hardcoded colors. Reference `var(--accent-*)` for the left border or icon tint.
- **Risks / gotchas:**
  - **Do NOT use `#39ff14`.** Cerebrum Do-Not-Repeat 2026-04-10: the design system green is `var(--accent-green: #00e57a)`. Any green accent on CommandNode must use the token.
  - `ProcessSidebar.svelte`'s `onMashedAssetDragStart` sets `effectAllowed = 'none'` on skill rows, so skills will never trigger `dragover`. The drop handler's `role !== 'command'` check is there purely as a defensive guard.
  - The drag payload is a JSON string (per Phase 1 plan §"Phase 1 — SHIPPED" row "Drag payload"); `JSON.parse` it defensively — wrap in try/catch and bail silently on parse failure.
  - Position via `useSvelteFlow()`'s `project()` — match the pattern the existing `application/bmad-process` drop branch uses.
- **Prerequisites already in place:**
  - Phase 1 drag source: `ProcessSidebar.svelte.onMashedAssetDragStart` already serialises `{name, path, kind, source, role, description}` as JSON under `application/mashed-asset`.
  - `ProcessNode.svelte` is the visual template to copy.
  - The Svelte-Flow `nodeTypes` map already exists in `WorkflowBuilder.svelte`.

## Acceptance Criteria

**AC-1: Dropping a command asset creates a command node**
- Given the canvas is empty and a mashed-ready command with `name: "simplify"` is dragged from the sidebar
- When the user drops it on a blank canvas spot
- Then a new node of `type: 'command'` appears at the drop position
- And the node's `data.config.commandName` is `"simplify"`
- And `CommandNode.svelte` renders it with the command name and description visible

**AC-2: Dropping a skill asset is rejected**
- Given a skill asset is somehow serialised into a drop payload with `role: "skill"` (simulating a bug bypassing the sidebar guard)
- When the drop handler receives it
- Then no new node is added to `$nodes`
- And no error is thrown

**AC-3: Registered node type renders correctly on load**
- Given a saved workflow containing one node with `type: 'command'` and valid config
- When `WorkflowBuilder.svelte` loads the workflow
- Then the node renders via `CommandNode.svelte`, not the fallback `bmadProcess` component
- And the displayed label matches `data.config.commandName`

**AC-4: Design-system tokens only**
- Given the `CommandNode.svelte` stylesheet
- When any color value appears
- Then it uses a CSS custom property (`var(--…)`) — no hex, no rgb/rgba literals, no `#39ff14`

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Command node drag-and-drop

  Scenario: Drop a command onto the canvas
    Given the sidebar shows a mashed-ready command named "simplify"
    And the canvas is empty
    When the user drags "simplify" onto the canvas at position (300, 200)
    Then a new command node appears at approximately (300, 200)
    And the node's label is "simplify"
    And the node's type in Svelte-Flow terms is "command"
    And CommandNode.svelte is the component rendering it

  Scenario: Skill drag payloads are ignored by the canvas drop handler
    Given a drop event carrying a mashed-asset JSON payload with role "skill"
    When the drop handler runs
    Then $nodes is unchanged
    And no console error is raised

  Scenario: Command node survives save + reload
    Given a workflow was saved with one command node
    When the user reopens the workflow
    Then the node is rendered using CommandNode.svelte
    And the node's config keys commandName, commandPath, commandDescription are intact

  Scenario: CommandNode uses only design-system tokens
    Given CommandNode.svelte has been compiled
    When the stylesheet is inspected
    Then zero hex color literals appear
    And zero rgba literals appear
```

## Tasks / Subtasks

- [ ] Task 1 — Build `CommandNode.svelte` (AC-1, AC-3, AC-4)
  - [ ] Copy `ProcessNode.svelte` as a starting point; strip the `in:`/`out:` rows
  - [ ] Wire up `data.label`, `data.config.commandName`, `data.config.commandDescription`, `data.status`
  - [ ] Use `Terminal` (or `Slash`) icon from `lucide-svelte`
  - [ ] Verify every color is a `var(--…)` token
- [ ] Task 2 — Register node type (AC-3)
  - [ ] Import `CommandNode` in `WorkflowBuilder.svelte`
  - [ ] Add `command: CommandNode` to the `nodeTypes` map
  - [ ] Extend the load-side type resolution so `n.nodeType === 'command'` maps to Svelte-Flow `type: 'command'`
- [ ] Task 3 — Extend drop handler (AC-1, AC-2)
  - [ ] Add `application/mashed-asset` branch to `CanvasPane.svelte`'s drop handler
  - [ ] Defensive `JSON.parse` with try/catch
  - [ ] Reject payloads whose `role !== 'command'`
  - [ ] Compute drop position via existing `project()` pattern
  - [ ] Append to `$nodes` with the shape in Developer Notes
- [ ] Task 4 — Tests (AC-1, AC-2, AC-3, AC-4)
  - [ ] Playwright scenario: drag-drop a command onto the canvas via the `/playwright-cli` skill pattern
  - [ ] Unit test: drop handler ignores `role: "skill"` payload
  - [ ] Save + reload assertion via a JSON fixture
  - [ ] Lint/regex check for color literals in `CommandNode.svelte`

## Definition of Done

- [ ] All ACs verified by an automated test (Playwright + unit; no "manually verified")
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean (should be unaffected; still run to confirm)
- [ ] `go test ./... -race -short` clean
- [ ] Frontend build (`wails build` or dev compile) clean; no Svelte warnings
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths, magic numbers, or color literals added
- [ ] Existing tests still pass

## Design Brief

### Layout composition
Card mirrors `ProcessNode.svelte`'s envelope (`min-width: 160px`, `max-width: 220px`, `overflow: hidden`) so mixed canvases feel coherent, but the inner grammar is simpler because commands have no `in:`/`out:` rows. Stack:

1. **Accent stripe** — 4px full-width bar at top (same slot as `.phase-bar`) rendered in `var(--accent-green)` for `command` nodes (the brand "alive" color). This is the primary visual departure from ProcessNode, which colors its stripe by phase.
2. **Body** — `padding: var(--sp-sm) var(--sp-md) var(--sp-xs)` (matches ProcessNode's `8px 10px var(--sp-xs)`). Two rows:
   - Row A: header — `display: flex; align-items: center; gap: var(--sp-sm);` — Terminal icon (13px, tinted `var(--accent-green)`) + command name.
   - Row B: description — single-line `text-overflow: ellipsis` below the header, margin-top `var(--sp-2xs)`.
3. **Status row** — same bordered bottom rail ProcessNode uses: `border-top: 1px solid var(--border-subtle); padding-top: var(--sp-xs); margin-top: var(--sp-xs); gap: var(--sp-xs);`.
4. **Left/right connector handles** at `Position.Left` / `Position.Right` inherit `@xyflow/svelte` defaults — no overrides.

### Typography plan
- **Command name (header)**: `font-family: var(--font-ui)` (Geist), `font-size: var(--text-label)` (11px), `font-weight: 600`, `color: var(--text-primary)`. This is one step down from ProcessNode's mono to hint "prose-ish identifier" vs. ProcessNode's machine-ish role.
- **Description (subtitle)**: `font-family: var(--font-mono)` (token missing a true mono face — currently resolves to `monospace`; use the token anyway so the eventual Geist Mono swap cascades), `font-size: 9px` (matches ProcessNode's `.artifacts`), `color: var(--text-muted)`, `line-height: 1.3`.
- **Status label**: `font-size: 9px`, `text-transform: uppercase`, `letter-spacing: 0.3px`, `color: var(--text-muted)` — identical to ProcessNode's `.status-text` so mixed canvases share a visual rhythm.
- No numeric columns here, so no `tabular-nums` required.

### Color strategy
- **Surface**: `background: var(--bg-elevated)` (same plane as ProcessNode) — canvas nodes all sit on the elevated layer so the deepest `--bg-deepest` canvas reads as ground.
- **Border**: `1px solid var(--border-subtle)` idle → `var(--accent-green)` on `selected` → pulsing `--accent-green` on `running`.
- **Accent stripe**: `var(--accent-green)`. Always. Commands get the brand neon; ProcessNode varies its stripe by `phase`.
- **Icon tint**: `var(--accent-green)` for the Terminal glyph.
- **Selected ring**: `box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-green) 35%, transparent)` (same alpha as ProcessNode selected state).
- **Running pulse**: reuse ProcessNode's `@keyframes node-pulse` — same 2s `var(--ease-move)` cadence keeps the canvas in one motion dialect.
- **Status mapping** (unchanged from ProcessNode): pending → `var(--text-muted)` dot, running → `var(--accent-green)` pulsing dot, complete → `var(--accent-green)` check, failed → `var(--accent-red)` cross, skipped → `var(--text-muted)` dash.

### Interaction model
- **Primary action**: drag source is the sidebar row, not the card itself. Card exposes no click affordance in v1.
- **Hover**: `border-color` transitions to `var(--border-emphasis)` over `var(--duration-short) var(--ease-enter)`. No transform/lift — canvas nodes must not visually drift on hover or drag-to-connect becomes imprecise.
- **Selected**: neon ring (see above) + border shift to `var(--accent-green)`.
- **Focus-visible**: global `button:focus-visible` rule in `style.css` already applies a `2px solid var(--accent-green)` outline to focusable children; the card itself is not natively focusable, matching ProcessNode.
- **Keyboard**: relies on Svelte-Flow's built-in node navigation; no custom key bindings at card level.
- **State transitions**: status changes animate via the existing `dot-pulse` keyframes and `transition: border-color 150ms ease`.

### Component specs
```
.command-node {
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);     /* 4px — same as ProcessNode */
  min-width: 160px;
  max-width: 220px;
  overflow: hidden;
  font-family: var(--font-mono);
  transition: border-color var(--duration-medium) var(--ease-enter),
              box-shadow var(--duration-medium) var(--ease-enter);
}
.accent-stripe { height: 4px; background: var(--accent-green); }
.body          { padding: var(--sp-sm) var(--sp-md) var(--sp-xs); }
.header        { display: flex; align-items: center; gap: var(--sp-sm); }
.name          { font-size: var(--text-label); font-weight: 600; color: var(--text-primary); }
.description   { margin-top: var(--sp-2xs); font-size: 9px; color: var(--text-muted); line-height: 1.3; }
.status-row    { margin-top: var(--sp-xs); padding-top: var(--sp-xs);
                  border-top: 1px solid var(--border-subtle);
                  display: flex; align-items: center; gap: var(--sp-xs); }
.command-node.selected {
  border-color: var(--accent-green);
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--accent-green) 35%, transparent);
}
.command-node.running { animation: node-pulse 2s var(--ease-move) infinite; }
```
All measurements trace to `style.css` tokens. The `9px` description size is not in the `--text-*` scale but matches ProcessNode's existing `.artifacts` value — token missing, use fallback by inheriting ProcessNode's precedent.

### Signature elements
- **Terminal glyph + brand green accent stripe** is the "unmistakably mashed" mark — it reads as "slash-command surface on the canvas" at a glance, against a field of phase-colored ProcessNodes. The green stripe also echoes the sidebar's `.tab.active` and the `.status-pulse` dot, giving the whole app a single neon spine.
- **Density-first body** — two rows of text, no padding bloat, description truncates before wrapping. Matches the "Bloomberg terminal" density target in `DESIGN.md`.
- **Zero idle motion** — the card is still until state changes. When running, the same 2s `node-pulse` as ProcessNode takes over. No hover lift. This restraint IS the aesthetic.

### Distinguishing from existing patterns
- **vs ProcessNode**: no `in:`/`out:` artifact rows, no `artifact-indicators` block, no `story-badge`. Accent stripe is always `var(--accent-green)` not phase-colored. Icon is `Terminal` (lucide) not a role icon. Header name uses `var(--font-ui)` instead of inheriting the mono family ProcessNode uses, so a user scanning the canvas can distinguish the two node types in peripheral vision without reading labels.
- **vs `.process-item` sidebar row**: the node is a full card with its own stripe and status row; the sidebar row is a single-line draggable with just a colored dot. The drag journey is visually: "dot → full card".

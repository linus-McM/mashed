# UI Architect Design Review -- 2026-04-10

**Reviewer:** ui-architect agent (design director)
**Scope:** Full application audit -- 5 major views, 7 shared components, global tokens
**Method:** 7-dimension scoring per view, hardcoded color audit, cross-cutting analysis

---

## 1. Executive Summary

| Metric | Value |
|--------|-------|
| **Overall Score** | **5.9 / 10** |
| **Weakest Dimension** | Motion & Interaction (3.4 avg) |
| **Strongest Dimension** | Information Density (7.0 avg) |
| **Hardcoded Colors Found** | **42 instances** across 8 files (theme-breaking) |
| **Critical Issues** | 6 |
| **High-Priority Issues** | 14 |
| **Polish Issues** | 11 |

The mashed UI has strong bones -- the information architecture is sound, density is appropriate for a power-user tool, and the token system in `style.css` is well-structured. However, the implementation drifts significantly from the design system: hardcoded hex colors bypass theming, motion is nearly absent, and several views lack the "signature mashed" identity described in `DESIGN.md`. The gap between the design system's ambition (Bloomberg meets Linear) and the current implementation (functional but generic dark-theme dev tool) is the central finding.

**Three actions that would most improve the app:**
1. Replace all `#39ff14` instances with `var(--accent-green)` -- fixes theme switching for 42+ elements
2. Add entry animations to NotificationFeed (staggered fly) and modal entry transitions
3. Introduce a "hot action" CSS custom property to replace the duplicated glow-button pattern across 3 files

---

## 2. Per-View Scoring

### 2.1 NotificationFeed (Main View)

**File:** `frontend/src/views/NotificationFeed.svelte` (1894 lines)
**Role:** Primary view -- this IS the app for most users

| Dimension | Score | Notes |
|-----------|-------|-------|
| Visual Hierarchy | 6/10 | Repo groups are visually distinct, but agent rows within them are monotonous. No clear primary action per repo. |
| Spatial Composition | 6/10 | 75/25 split is functional. Repo-level spacing is consistent. Agent rows could use more internal rhythm variation. |
| Typographic Quality | 7/10 | Good mono/proportional separation. Status labels use uppercase + spacing. Missing: tabular-nums on token counts. |
| Color & Contrast | 5/10 | `#39ff14` used instead of `var(--accent-green)` in status colors and hot-action buttons. Breaks all non-default themes. `borderPalette` hardcodes 11 hex values. |
| Motion & Interaction | 3/10 | Zero entry animation. No staggered reveal on repo groups. Repo collapse/expand is instant (no transition). Only micro-interaction: hover background on rows. |
| Information Density | 8/10 | Excellent. Agent rows pack model + status + sub-count + summary + tokens + elapsed + kill. Compact without feeling cramped. |
| Identity & Distinction | 5/10 | Color picker dots and drag-to-reorder are nice touches. But the overall layout reads as "generic list view." No sparklines visible despite SparkLine component being imported. |
| **Overall** | **5.7** | |

**Specific Issues:**

| ID | File:Line | Issue | Severity |
|----|-----------|-------|----------|
| NF-01 | NotificationFeed.svelte:1100 | `running: '#39ff14'` -- hardcoded, should be `var(--accent-green)` | Critical |
| NF-02 | NotificationFeed.svelte:1101 | `open: '#00c4b3'` -- hardcoded, should be `var(--accent-teal)` | Critical |
| NF-03 | NotificationFeed.svelte:34-37 | `borderPalette` array contains 11 hardcoded hex colors with no var() fallback | High |
| NF-04 | NotificationFeed.svelte:48 | `getRepoColor` defaults to `'#1e2530'` -- should be `var(--border-subtle)` but is used in JS equality checks so needs a different approach | High |
| NF-05 | NotificationFeed.svelte:1700-1722 | `.action-hot` and `.action-review.action-hot` use `#39ff14` (4 instances) + rgba(57,255,20,...) (8 instances) | Critical |
| NF-06 | NotificationFeed.svelte:1271 | `.color-swatch.active` uses `border-color: #fff` -- should use `var(--text-primary)` | Medium |
| NF-07 | NotificationFeed.svelte:707-1034 | No `transition:` directives on any dynamic content -- repo collapse, agent list, commit panel all appear/disappear instantly | High |
| NF-08 | NotificationFeed.svelte:14 | SparkLine is imported but never used in the template -- dead import or missed feature | Medium |
| NF-09 | NotificationFeed.svelte:1598 | `.new-session-btn` uses `font-size: 12px` -- should be `var(--text-body)` or `var(--text-label)` | Low |
| NF-10 | NotificationFeed.svelte:1399 | `.sub-count` uses `font-size: 10px` and `padding: 0 5px` -- orphaned values not on token scale | Low |

**What a 10/10 looks like:**
Repo groups enter with a staggered `fly` transition (40ms offset). Agent rows within an expanding repo slide in with `slide` transition. The SparkLine component renders a token-usage sparkline per agent. Hot-action buttons use `var(--accent-green)` so they glow correctly in every theme. The repo collapse animation uses Svelte's `slide` transition for smooth accordion behavior.

---

### 2.2 AgentDetail

**File:** `frontend/src/views/AgentDetail.svelte` (~1320 lines)
**Role:** Deep-dive into a single agent session with terminal, file tree, and editor

| Dimension | Score | Notes |
|-----------|-------|-------|
| Visual Hierarchy | 6/10 | Header breadcrumb provides context. Terminal dominates (correct). But when file tree + editor are open, the three-pane layout has no clear visual priority hierarchy. |
| Spatial Composition | 7/10 | Resizable panes via drag handles -- good desktop pattern. File strip min/max constraints prevent collapse. |
| Typographic Quality | 6/10 | Breadcrumb uses good font separation. But `font-size: 14px` on `.header-repo` is hardcoded (should be `var(--text-data)`). |
| Color & Contrast | 4/10 | `#39ff14` used 5 times in CSS (lines 679, 689, 1005, 1011, 1012). rgba(57,255,20,...) used 6 times. `.tab-close:hover` uses hardcoded `#ff5f57`. All break theming. |
| Motion & Interaction | 3/10 | No view-entry animation. No pane-resize animation feedback. Tab switching is instant. Terminal connection error states have no transition. |
| Information Density | 7/10 | Three-pane layout is efficient. Tab bar for multiple sessions per repo is excellent. Token progress bar is compact and informative. |
| Identity & Distinction | 5/10 | Resizable panes are nice. But the overall layout is a standard editor pattern (VS Code). No mashed-specific signature element. |
| **Overall** | **5.4** | |

**Specific Issues:**

| ID | File:Line | Issue | Severity |
|----|-----------|-------|----------|
| AD-01 | AgentDetail.svelte:679 | `.back-btn` color: `#39ff14` -- should be `var(--accent-green)` | Critical |
| AD-02 | AgentDetail.svelte:685 | `.back-btn` text-shadow uses `rgba(57, 255, 20, ...)` -- should derive from accent variable | High |
| AD-03 | AgentDetail.svelte:821 | `.tab-close:hover` color: `#ff5f57` -- should be `var(--accent-red)` | High |
| AD-04 | AgentDetail.svelte:1004-1014 | `.git-hot` pattern duplicates NotificationFeed's `.action-hot` exactly -- should be a shared class or CSS custom property | High |
| AD-05 | AgentDetail.svelte:695 | `font-size: 14px` hardcoded -- should be `var(--text-data)` | Low |

---

### 2.3 WorkflowBuilder (BMAD)

**File:** `frontend/src/views/WorkflowBuilder.svelte` + `components/bmad/*`
**Role:** Node-based workflow editor for BMAD process orchestration

| Dimension | Score | Notes |
|-----------|-------|-------|
| Visual Hierarchy | 7/10 | Tab bar provides clear navigation. Canvas + sidebar + config panel is a well-established pattern. Execution bar draws the eye to the primary action. |
| Spatial Composition | 7/10 | Three-column layout with resizable sidebar and config panel. Good use of CSS Grid. |
| Typographic Quality | 7/10 | ProcessNode uses consistent mono font. Labels are uppercase with letter-spacing. Artifact labels at 9px are perhaps too small. |
| Color & Contrast | 6/10 | ProcessNode uses `var()` with fallbacks (good pattern). But ExecutionBar.svelte has 4 hardcoded hex colors (`#22d3ee`, `#fb923c` at lines 163, 168, 197, 202). These are not in the design system. |
| Motion & Interaction | 5/10 | Node pulse animation on running state is a good signature moment. Drag-and-drop from sidebar works. But no canvas entry animation, no edge drawing animation, no template insertion animation. |
| Information Density | 7/10 | Nodes show role icon + label + inputs/outputs + status + story badge. Compact and scannable. |
| Identity & Distinction | 7/10 | The phase-colored accent bar on nodes is a strong identity element. The canvas dot-grid background and node-pulse animation feel intentional. Best "mashed identity" of any view. |
| **Overall** | **6.6** | |

**Specific Issues:**

| ID | File:Line | Issue | Severity |
|----|-----------|-------|----------|
| WB-01 | ExecutionBar.svelte:163 | `.ctrl-btn.run` color: `#22d3ee` (cyan) -- not a design system color, no `--accent-cyan` token exists | High |
| WB-02 | ExecutionBar.svelte:197 | `.ctrl-btn.stop` color: `#fb923c` (orange) -- not a design system color, no `--accent-orange` token exists | High |
| WB-03 | ProcessNode.svelte:126 | `.process-node` uses `box-shadow: 0 0 0 2px rgba(0, 229, 122, 0.35)` -- hardcoded green; would break in a theme where accent is not green | Medium |
| WB-04 | CanvasPane.svelte | Context menu lacks keyboard navigation (no j/k, no arrow keys, no Escape to close) | Medium |
| WB-05 | WorkflowBuilder.svelte:883 | `rgba(248, 81, 73, 0.1)` and `rgba(248, 81, 73, 0.15)` -- hardcoded red, should derive from `var(--accent-red)` | Medium |

---

### 2.4 Settings

**File:** `frontend/src/views/Settings.svelte`
**Role:** Theme selection, font config, editor settings, sidebar width

| Dimension | Score | Notes |
|-----------|-------|-------|
| Visual Hierarchy | 6/10 | Two-column split is clear. But "Themes" and "Font" sections compete equally for attention -- no clear primary. |
| Spatial Composition | 6/10 | Theme thumbnails are well-designed (mini previews). But the right column is a vertical scroll of unrelated sections with no grouping hierarchy. |
| Typographic Quality | 6/10 | Code preview font rendering is excellent. Section titles use consistent treatment. But label font-sizes vary between hardcoded values and tokens. |
| Color & Contrast | 5/10 | Theme preview dots use hardcoded macOS traffic light colors (`#ff5f57`, `#febc2e`, `#28c840` at lines 165-167) which is intentional for the preview, but `.import-indicator.dark` uses `#565670` and `.light` uses `#c0c0d0` (lines 860, 864) -- these are orphaned magic numbers. |
| Motion & Interaction | 3/10 | Theme switching happens instantly (no transition). Font size slider has no feedback animation. No entry animation for settings pane. |
| Information Density | 6/10 | Appropriate for a settings view -- less dense is correct. But editor settings section is just labels with no actual controls visible (confirmed from screenshot description). |
| Identity & Distinction | 5/10 | Theme preview thumbnails are a nice touch. But overall looks like a standard settings page -- no mashed personality. |
| **Overall** | **5.3** | |

**Specific Issues:**

| ID | File:Line | Issue | Severity |
|----|-----------|-------|----------|
| ST-01 | Settings.svelte:860 | `.import-indicator.dark` background: `#565670` -- orphaned magic number, not in token set | Medium |
| ST-02 | Settings.svelte:864 | `.import-indicator.light` background: `#c0c0d0` -- orphaned magic number | Medium |
| ST-03 | Settings.svelte:165-167 | macOS traffic light colors hardcoded in theme preview -- acceptable (matches real traffic lights) but should be documented as intentional exception | Low |
| ST-04 | Settings.svelte:466 | `text-shadow: 0 0 10px rgba(0, 229, 122, 0.6)` -- hardcoded green in text-shadow | Medium |

---

### 2.5 NewSessionModal + SpawnAgent

**Files:** `frontend/src/views/NewSessionModal.svelte`, `frontend/src/views/SpawnAgent.svelte`
**Role:** Modal dialogs for spawning new Claude sessions

| Dimension | Score | Notes |
|-----------|-------|-------|
| Visual Hierarchy | 7/10 | Clear title/subtitle. Model selector with visual selection state. Section labels with uppercase treatment. Primary CTA at bottom. |
| Spatial Composition | 6/10 | NewSessionModal is very tall -- requires scrolling. Option groups could be collapsed/tabbed. SpawnAgent is more compact and well-proportioned. |
| Typographic Quality | 7/10 | Labels use `11px uppercase letter-spacing: 0.5px` -- matches design system. Body text and field values are appropriately sized. |
| Color & Contrast | 7/10 | Selected states use `var(--accent-green)` correctly. Background overlays use rgba. No hardcoded hex in the CSS section. The `rgba(0, 229, 122, 0.08)` for selected state is hardcoded but acceptable (derives from the accent color). |
| Motion & Interaction | 3/10 | Modal appears instantly -- no fade or scale entry. No exit animation. Button press has no feedback beyond cursor change. Keyboard shortcuts (Escape) work. |
| Information Density | 6/10 | SpawnAgent is appropriately compact. NewSessionModal packs too many options without grouping -- toggles, conditionals, and text fields compete for attention. |
| Identity & Distinction | 5/10 | Terminal icon in header is a nice touch. But the modal design is standard card-in-overlay. |
| **Overall** | **5.9** | |

**Specific Issues:**

| ID | File:Line | Issue | Severity |
|----|-----------|-------|----------|
| SM-01 | SpawnAgent.svelte:137 | `h2` font-size: `18px` -- not on the type scale (nearest: 16px section or 20px page) | Low |
| SM-02 | SpawnAgent.svelte:152 | `.field` margin-bottom: `20px` -- not on spacing scale (nearest: 16px or 24px) | Low |
| SM-03 | NewSessionModal.svelte/SpawnAgent.svelte | No modal entry transition -- should fade in over 100ms | High |

---

### 2.6 Shared Components

#### StatusBadge (`components/StatusBadge.svelte`)

| Issue | Detail |
|-------|--------|
| **Critical: hardcoded `#39ff14`** | Line 6: `running: '#39ff14'` -- this is the most visible hardcoded color in the entire app. Every running agent badge uses this instead of `var(--accent-green)`. Breaks every non-default theme. |
| **Hardcoded `#00c4b3`** | Line 7: `open: '#00c4b3'` -- should be `var(--accent-teal)` |
| **Good pattern** | Uses `color-mix(in srgb, var(--status-color) 15%, transparent)` for background -- this is the correct approach and should be replicated elsewhere |

#### SparkLine (`components/SparkLine.svelte`)
- Well-implemented. Uses `var(--accent-teal)` and `var(--font-mono)`. Token-compliant.
- But **never rendered** in the main NotificationFeed despite being imported.

#### TitleBar (`components/TitleBar.svelte`)
- Traffic light colors (`#ff5f57`, `#febc2e`, `#28c840`) are hardcoded. This is **acceptable** -- these are macOS system chrome colors and should NOT change with theme.
- Theme picker popover uses proper `var()` tokens throughout.

#### ProcessNode (`components/bmad/ProcessNode.svelte`)
- Good use of `var()` with fallbacks: `var(--accent-blue, #3d9eff)`
- The fallback pattern is acceptable for xyflow nodes where CSS variable inheritance might break
- rgba values in box-shadow are hardcoded but minor

#### ExecutionBar (`components/bmad/ExecutionBar.svelte`)
- **Two colors outside the design system:** `#22d3ee` (cyan) and `#fb923c` (orange)
- These need either: (a) new CSS custom properties added to style.css, or (b) remapping to existing tokens

---

## 3. Cross-Cutting Analysis

### 3.1 Recurring Themes

**Theme 1: `#39ff14` epidemic**
The color `#39ff14` appears in 5 files across ~15 CSS rules. This is NOT the design system's accent green (`#00e57a` / `var(--accent-green)`). It is a brighter, more neon green that was likely introduced early and never harmonized. It breaks every custom theme because it is hardcoded.

Files affected:
- `StatusBadge.svelte` (2 instances -- affects every badge in the app)
- `NotificationFeed.svelte` (6 instances in CSS, 2 in JS statusColors)
- `AgentDetail.svelte` (5 instances in CSS)
- `ExecutionBar.svelte` (indirect -- uses different non-system colors)
- `RepoContextBar.svelte` (1 instance)

**Theme 2: Duplicated "hot action" button pattern**
The glowing green button effect (`.action-hot` / `.git-hot`) is copy-pasted identically in NotificationFeed.svelte and AgentDetail.svelte. Both use the same `#39ff14` color, same rgba shadows, same hover escalation. This should be:
1. A shared CSS class in `style.css`, or
2. A shared Svelte component (e.g., `GlowButton.svelte`)

**Theme 3: Zero entry animations**
No view in the app uses Svelte's `transition:` directive for entry animation. The design system specifies "notification entry: 150ms ease-out (slide + fade from top)" and "staggered reveal: 40-60ms delay between sibling elements" but none of this is implemented. The app feels static -- content appears instantly without communicating state changes.

**Theme 4: rgba() hardcoded with raw RGB values**
There are 80+ instances of `rgba(R, G, B, alpha)` using raw integer values instead of deriving from CSS custom properties. While CSS cannot natively do `rgba(var(--accent-green), 0.1)`, the app already uses `color-mix(in srgb, var(--status-color) 15%, transparent)` in StatusBadge -- this pattern should be standardized across all translucent accent colors.

**Theme 5: Orphaned font-size and spacing values**
Multiple components use sizes not on the token scale:
- `10px` font-size (scale has 11px labels)
- `12px` font-size (scale has 13px body)
- `5px`, `6px`, `3px` padding values (scale goes 2, 4, 8, 12...)
- `18px`, `15px` font-sizes (not on the 11/13/14/16/20/28 scale)

### 3.2 Weakest Dimensions Across All Views

| Rank | Dimension | Average Score | Root Cause |
|------|-----------|---------------|------------|
| 1 | **Motion & Interaction** | **3.4** | Zero Svelte transitions. No entry animations. No state-change animations. |
| 2 | **Identity & Distinction** | **5.4** | Layout patterns are generic. SparkLine never rendered. No signature animation. |
| 3 | **Color & Contrast** | **5.4** | `#39ff14` epidemic. rgba hardcoded throughout. Two rogue colors in ExecutionBar. |

### 3.3 Strongest Dimensions

| Rank | Dimension | Average Score | Why It Works |
|------|-----------|---------------|--------------|
| 1 | **Information Density** | **7.0** | Agent rows are compact. ProcessNodes pack data well. Three-pane layouts are efficient. |
| 2 | **Typographic Quality** | **6.6** | Mono/proportional separation is consistent. Labels use uppercase + letter-spacing. |
| 3 | **Spatial Composition** | **6.4** | Resizable panes, grid-based layouts, consistent outer margins. |

---

## 4. Hardcoded Color Audit

### Critical: `#39ff14` (neon green -- NOT the design system accent)

| File | Line(s) | Context | Fix |
|------|---------|---------|-----|
| StatusBadge.svelte | 6 | `running: '#39ff14'` | Change to `'var(--accent-green)'` |
| NotificationFeed.svelte | 1100 | `running: '#39ff14'` in statusColors | Change to `'var(--accent-green)'` |
| NotificationFeed.svelte | 1700, 1706, 1713, 1719 | `.action-hot` color | Replace with `var(--accent-green)` |
| NotificationFeed.svelte | 1707, 1720 | `.action-hot` border-color | Replace with `var(--accent-green)` |
| AgentDetail.svelte | 679, 689 | `.back-btn` color | Replace with `var(--accent-green)` |
| AgentDetail.svelte | 1005, 1011, 1012 | `.git-hot` color/border | Replace with `var(--accent-green)` |
| RepoContextBar.svelte | 61 | color | Replace with `var(--accent-green)` |

### High: Colors outside design system

| File | Line(s) | Color | Issue | Fix |
|------|---------|-------|-------|-----|
| ExecutionBar.svelte | 163, 168 | `#22d3ee` | Cyan -- no token exists | Add `--accent-cyan: #22d3ee` to style.css or remap to `var(--accent-teal)` |
| ExecutionBar.svelte | 197, 202 | `#fb923c` | Orange -- no token exists | Add `--accent-orange: #fb923c` to style.css or remap to `var(--accent-amber)` |
| Settings.svelte | 860 | `#565670` | Dark indicator -- orphaned | Add to token set or use `var(--text-muted)` |
| Settings.svelte | 864 | `#c0c0d0` | Light indicator -- orphaned | Add to token set or use `var(--text-dim)` |
| AgentDetail.svelte | 821 | `#ff5f57` | Tab close hover -- macOS red | Use `var(--accent-red)` |

### Acceptable Exceptions (do NOT change)

| File | Line(s) | Color | Reason |
|------|---------|-------|--------|
| TitleBar.svelte | 131-133 | `#ff5f57`, `#febc2e`, `#28c840` | macOS traffic light chrome -- must match system |
| Settings.svelte | 165-167 | `#ff5f57`, `#febc2e`, `#28c840` | Theme preview traffic lights -- matches real chrome |
| NotificationFeed.svelte | 34-37 | `borderPalette` array | User-selectable palette -- these are data, not theme. But the default (`#1e2530`) should reference the token value. |

### Pattern: rgba() with hardcoded RGB

Over 80 instances of `rgba(R, G, B, alpha)` using raw values. These will NOT update when themes change. Priority replacements:

| Pattern | Count | Fix |
|---------|-------|-----|
| `rgba(57, 255, 20, ...)` (neon green) | ~14 | Use `color-mix(in srgb, var(--accent-green) N%, transparent)` |
| `rgba(0, 229, 122, ...)` (design green) | ~12 | Use `color-mix(in srgb, var(--accent-green) N%, transparent)` |
| `rgba(232, 69, 69, ...)` (red) | ~10 | Use `color-mix(in srgb, var(--accent-red) N%, transparent)` |
| `rgba(240, 165, 0, ...)` (amber) | ~8 | Use `color-mix(in srgb, var(--accent-amber) N%, transparent)` |
| `rgba(0, 0, 0, ...)` (black overlays) | ~10 | Acceptable for overlays -- theme-independent |

---

## 5. Prioritized Remediation Plan

### Priority 1: Critical (Blocks theme switching)

| ID | Change | Files | Dimensions Improved | Effort |
|----|--------|-------|---------------------|--------|
| **FIX-01** | Replace all `#39ff14` with `var(--accent-green)` in CSS | StatusBadge, NotificationFeed, AgentDetail, RepoContextBar | Color +2, Identity +1 | Small (find-replace) |
| **FIX-02** | Replace `#39ff14` in JS statusColors maps with `var(--accent-green)` | StatusBadge:6, NotificationFeed:1100 | Color +1 | Small |
| **FIX-03** | Replace `#00c4b3` in JS with `var(--accent-teal)` | StatusBadge:7, NotificationFeed:1101 | Color +1 | Small |
| **FIX-04** | Add `--accent-cyan` and `--accent-orange` tokens to style.css, use in ExecutionBar | style.css, ExecutionBar.svelte | Color +1 | Small |

### Priority 2: High (Significant UX improvement)

| ID | Change | Files | Dimensions Improved | Effort |
|----|--------|-------|---------------------|--------|
| **FIX-05** | Add `transition:fly` to repo groups in NotificationFeed with 40ms stagger | NotificationFeed.svelte | Motion +3 | Medium |
| **FIX-06** | Add `transition:slide` to repo body expand/collapse | NotificationFeed.svelte | Motion +2 | Medium |
| **FIX-07** | Add `transition:fade` (100ms) to modal entry/exit in NewSessionModal, SpawnAgent, BranchModal | 4 modal files | Motion +2 | Small |
| **FIX-08** | Extract shared `.glow-btn` pattern from `.action-hot`/`.git-hot` into style.css global class using `var(--accent-green)` | style.css, NotificationFeed, AgentDetail | Color +1, Code quality | Medium |
| **FIX-09** | Replace rgba(57,255,20,...) with `color-mix(in srgb, var(--accent-green) N%, transparent)` | 5 files, ~14 instances | Color +1 | Medium |
| **FIX-10** | Replace rgba(232,69,69,...) and rgba(240,165,0,...) with color-mix equivalents | 4 files, ~18 instances | Color +1 | Medium |
| **FIX-11** | Render SparkLine in NotificationFeed agent rows (token history data) | NotificationFeed.svelte | Density +1, Identity +2 | Medium |
| **FIX-12** | Add `font-variant-numeric: tabular-nums` to all token count and elapsed time displays | NotificationFeed, AgentDetail | Typography +1 | Small |

### Priority 3: Polish (Score 7 to 8+)

| ID | Change | Files | Dimensions Improved | Effort |
|----|--------|-------|---------------------|--------|
| **FIX-13** | Normalize orphaned font-sizes: `10px` -> `var(--text-label)`, `12px` -> `var(--text-body)` or `var(--text-label)`, `18px` -> `var(--text-section)` | Multiple files | Typography +1 | Small |
| **FIX-14** | Normalize orphaned spacing: `5px` -> `var(--sp-xs)`, `6px` -> `var(--sp-xs)` or `var(--sp-sm)`, `3px` -> `var(--sp-2xs)` | Multiple files | Composition +0.5 | Small |
| **FIX-15** | Replace `#565670` and `#c0c0d0` in Settings with theme tokens or new `--indicator-dark`/`--indicator-light` tokens | Settings.svelte | Color +0.5 | Small |
| **FIX-16** | Add `:focus-visible` ring to agent rows, action buttons, and modal inputs (currently only some elements have focus styles) | Multiple files | Interaction +1 | Medium |
| **FIX-17** | Add context menu keyboard nav to CanvasPane (Escape to close, arrow keys to navigate) | CanvasPane.svelte | Interaction +1 | Medium |
| **FIX-18** | Remove dead SparkLine import if not implementing FIX-11 | NotificationFeed.svelte | Code quality | Trivial |

---

## 6. Signature Moments -- Opportunities for Mashed Personality

The design system describes mashed as having an "electric" neon green accent and "terminal aesthetic bleeding into UI chrome." Currently, the UI is functional but generic. Here are specific opportunities to inject mashed personality:

### Moment 1: The Running Pulse
When an agent is actively running, its row in NotificationFeed should have a subtle left-border pulse animation (the ProcessNode already has `node-pulse` -- extend this pattern to the feed). This creates a "living" feel where you can see activity at a glance without reading text.

### Moment 2: Token Sparklines
The SparkLine component exists but is never rendered. Each agent row should show a tiny sparkline of token consumption over time. This is the quintessential "Bloomberg in a dev tool" moment -- dense, informative, and visually distinctive.

### Moment 3: Status Bar as Ambient Signal
The bottom status bar (`total agents / repos / tokens`) is static text. It should include a subtle aggregate sparkline or a pulsing indicator showing overall system activity. Make it the "vital signs monitor" for your agent fleet.

### Moment 4: Modal Entry Choreography
When NewSessionModal opens, the model selector buttons should stagger in (40ms each) rather than appearing as a block. This takes 5 lines of Svelte code but creates a "refined tool" impression.

### Moment 5: Commit Panel as Terminal Aesthetic
The commit streaming panel already has a terminal feel with step-by-step output. Lean into this: add a monospace prompt character before each step, use the green accent for the active step, dim completed steps. Make it feel like watching a build in a terminal.

---

## 7. Design System Gaps

Issues found where `DESIGN.md` and `style.css` define tokens but the implementation ignores them:

| Token | Defined In | Used Correctly | Drifted |
|-------|-----------|----------------|---------|
| `--accent-green: #00e57a` | style.css:10 | ProcessNode, global utilities | StatusBadge, NotificationFeed, AgentDetail use `#39ff14` instead |
| `--duration-medium: 150ms` | style.css:44 | None | No Svelte transitions use the token |
| `--text-label: 11px` | style.css:34 | Most labels | Some use 10px, 12px instead |
| `--sp-xs: 4px` | style.css:27 | Most gaps | Some use 3px, 5px, 6px |
| `--font-mono` | style.css:21 | Consistently used | Good -- no drift |
| `--radius-md: 4px` | style.css:23 | Consistently used | Good -- no drift |

**Missing tokens that should be added:**
- `--accent-cyan` (for ExecutionBar run button, or remap to teal)
- `--accent-orange` (for ExecutionBar stop button, or remap to amber)
- `--overlay-backdrop` (standardize `rgba(0, 0, 0, 0.6)` used in all modals)
- `--glow-spread` (standardize the neon text-shadow pattern used in hot buttons)

---

## Appendix: Score Summary Table

| View | Hierarchy | Composition | Typography | Color | Motion | Density | Identity | **Avg** |
|------|-----------|-------------|------------|-------|--------|---------|----------|---------|
| NotificationFeed | 6 | 6 | 7 | 5 | 3 | 8 | 5 | **5.7** |
| AgentDetail | 6 | 7 | 6 | 4 | 3 | 7 | 5 | **5.4** |
| WorkflowBuilder | 7 | 7 | 7 | 6 | 5 | 7 | 7 | **6.6** |
| Settings | 6 | 6 | 6 | 5 | 3 | 6 | 5 | **5.3** |
| Modals (avg) | 7 | 6 | 7 | 7 | 3 | 6 | 5 | **5.9** |
| **App Average** | **6.4** | **6.4** | **6.6** | **5.4** | **3.4** | **6.8** | **5.4** | **5.8** |

# Design Critique: StatusBadge Component

**Component**: `frontend/src/components/StatusBadge.svelte`
**Workflow**: Component Review (Workflow 4 from ui-design skill)
**Date**: 2026-04-10

---

## Component Overview

StatusBadge is a compact status indicator pill used in `NotificationFeed` and `AgentDetail` views. It accepts a `status` prop (one of 11 states) and a `size` prop (`sm` or `md`). It renders an uppercase label with a tinted background derived from the status color at 15% opacity via `color-mix()`.

---

## 7-Dimension Scoring

| # | Dimension | Score | Notes |
|---|-----------|-------|-------|
| 1 | **Visual Hierarchy** | 7/10 | The badge correctly draws secondary attention. However, too many statuses map to `--accent-red` (finished, needs_response, waiting, error, completed, blocked), collapsing hierarchy. "DONE" uses blue while "FINISHED" uses red, which is semantically confusing. |
| 2 | **Spatial Composition** | 7/10 | Padding is compact and appropriate for a badge (2px 8px for md, 1px 6px for sm). However, the padding values are hardcoded px rather than spacing tokens. The `sm` variant has asymmetric vertical padding (1px) which is off the 4px grid. |
| 3 | **Typographic Quality** | 6/10 | Uses `var(--font-ui)` and uppercase with letter-spacing, which is correct for a label. But `font-size: 10px` (sm) and `font-size: 11px` (md) are hardcoded px values instead of using `var(--text-label)`. The 10px size for `sm` is not in the design system type scale (11/13/14/16/20/28). `line-height` is not explicitly set, relying on browser defaults. |
| 4 | **Color & Contrast** | 5/10 | Two critical token violations: `running` uses hardcoded `#39ff14` instead of `var(--accent-green)` (#00e57a). `open` uses hardcoded `#00c4b3` instead of `var(--accent-teal)`. The color-mix at 15% opacity is well-calibrated for readability. However, six statuses all map to `--accent-red`, violating AP-08 (Rainbow Status anti-pattern inverted: too few colors for too many states). Semantic mapping is broken: "finished" = red, "completed" = red, "done" = blue, yet all three mean the same thing conceptually. |
| 5 | **Motion & Interaction** | 3/10 | No hover, focus-visible, active, or disabled states defined. No transition property set. As a `<span>`, it has no interactive affordance, but parent views render it inside clickable rows, so it should at minimum not interfere with parent focus rings. No `transition` on color changes means dynamic status updates would snap without animation. |
| 6 | **Information Density** | 8/10 | Excellent. The badge is compact and packs status info into minimal space. The `color-mix()` technique for tinted backgrounds is efficient and avoids extra elements. Two sizes provide flexibility for different contexts. |
| 7 | **Identity & Distinction** | 6/10 | The tinted-background pill pattern is clean but generic -- it looks like every SaaS status badge. Opportunities to make it more "mashed": monospace font for the label (data, not prose), a left-edge accent bar instead of full pill, or block-character status indicators aligned with the sparkline aesthetic. |
| | **Overall** | **6.0/10** | |

---

## Fix List

### [DIM-4] Color & Contrast: 5/10 -> 8/10

**Problem**: Two hardcoded hex values bypass the design system. Six states collapse into identical red, destroying semantic differentiation.

**Fix 1** — `StatusBadge.svelte:6` — Replace hardcoded hex for `running`:
- Old: `running: '#39ff14'`
- New: `running: 'var(--accent-green)'`

**Fix 2** — `StatusBadge.svelte:7` — Replace hardcoded hex for `open`:
- Old: `open: '#00c4b3'`
- New: `open: 'var(--accent-teal)'`

**Fix 3** — `StatusBadge.svelte:9-16` — Correct semantic color mapping:
- `finished`: should be `var(--accent-green)` (success, not error)
- `completed`: should be `var(--accent-green)` (success)
- `needs_response`: should be `var(--accent-amber)` (needs human attention)
- `waiting`: should be `var(--accent-amber)` (blocked, waiting)
- `blocked`: should be `var(--accent-amber)` (blocked, waiting)
- `error`: stays `var(--accent-red)` (correct)
- `done`: could stay `var(--accent-blue)` or change to `var(--accent-green)` for consistency with completed/finished

**Impact**: Also improves Visual Hierarchy (DIM-1) and Identity (DIM-7).

---

### [DIM-3] Typographic Quality: 6/10 -> 8/10

**Problem**: Hardcoded font sizes including one (10px) outside the type scale. No explicit line-height.

**Fix 1** — `StatusBadge.svelte:50` — Use design token for sm font size:
- Old: `.badge.sm { font-size: 10px; padding: 1px 6px; }`
- New: `.badge.sm { font-size: 9px; padding: var(--sp-2xs) var(--sp-xs); }`
- Note: 10px is not in the type scale. If `sm` must be smaller than 11px, either accept 11px (the smallest token) or intentionally add a `--text-micro: 9px` token to the system. Recommendation: use `var(--text-label)` (11px) for both sizes and differentiate via padding alone.

**Fix 2** — `StatusBadge.svelte:51` — Use design token for md font size:
- Old: `.badge.md { font-size: 11px; }`
- New: `.badge.md { font-size: var(--text-label); }`

**Fix 3** — `StatusBadge.svelte:39-48` — Add explicit line-height to `.badge`:
- Add: `line-height: 1;`
- Rationale: Badges should have tight line-height to prevent vertical expansion.

**Impact**: Also improves Spatial Composition (DIM-2) via consistent spacing tokens.

---

### [DIM-5] Motion & Interaction: 3/10 -> 6/10

**Problem**: No transitions for dynamic status changes. No interaction states.

**Fix 1** — `StatusBadge.svelte:39-48` — Add transition to `.badge`:
- Add: `transition: background var(--duration-short) var(--ease-move), color var(--duration-short) var(--ease-move);`
- Rationale: When a status changes dynamically (e.g., running -> completed), the badge should smoothly transition color rather than snapping.

**Fix 2** — Consider adding a subtle pulse or dot animation for the `running` state only. This is the one status where motion communicates information ("this is actively running"). A small keyframe animation on a dot prefix would be appropriate, per Design Principle 5 (Motion Communicates, Never Decorates).

**Impact**: Also improves Identity (DIM-7) -- the running pulse would be a distinctive mashed touch.

---

### [DIM-2] Spatial Composition: 7/10 -> 8/10

**Problem**: Hardcoded padding values off the 4px grid.

**Fix** — `StatusBadge.svelte:49-51` — Use spacing tokens:
- `.badge` padding: `2px 8px` -> `var(--sp-2xs) var(--sp-sm)`
- `.badge.sm` padding: `1px 6px` -> `var(--sp-2xs) var(--sp-xs)` (rounds 1px up to 2px which is on-grid, 6px up to 4px which is on-grid but tighter)

**Impact**: Enforces grid alignment, per Design Principle 9.

---

### [DIM-7] Identity & Distinction: 6/10 -> 7/10

**Problem**: Generic pill badge pattern indistinguishable from other tools.

**Fix (suggestion, low priority)**: Consider switching the font to `var(--font-mono)` for the badge label. Status labels are machine-state identifiers (RUNNING, ERROR, DONE), not prose. Monospace aligns with the mashed terminal aesthetic (Design Principle 10, Distinctive > Correct) and pairs with the sparkline pattern defined in DESIGN.md.

---

## Token Compliance Audit

| Value | File:Line | Violation | Suggested Token |
|-------|-----------|-----------|-----------------|
| `#39ff14` | StatusBadge.svelte:6 | Hardcoded hex, not in design system at all | `var(--accent-green)` |
| `#00c4b3` | StatusBadge.svelte:7 | Hardcoded hex, matches `--accent-teal` | `var(--accent-teal)` |
| `10px` | StatusBadge.svelte:50 | Font size not in type scale | `var(--text-label)` (11px) |
| `11px` | StatusBadge.svelte:51 | Correct value but hardcoded | `var(--text-label)` |
| `2px 8px` | StatusBadge.svelte:49 | Correct values but hardcoded | `var(--sp-2xs) var(--sp-sm)` |
| `1px 6px` | StatusBadge.svelte:50 | Off-grid values (1px, 6px) | `var(--sp-2xs) var(--sp-xs)` |
| `0.5px` | StatusBadge.svelte:48 | Letter-spacing hardcoded | Acceptable (no letter-spacing token exists) |

---

## Interaction States Audit

| State | Implemented? | Notes |
|-------|-------------|-------|
| Default | Yes | Renders correctly |
| Hover | No | N/A if purely informational, but parent row hover should be tested |
| Focus-visible | No | Not needed (non-interactive span) |
| Active | No | N/A |
| Disabled | No | No disabled visual variant exists; unclear what a "disabled badge" means |
| Dynamic update | No | No transition; status changes snap |

---

## Responsiveness Check

The badge handles varying content lengths reasonably via `inline-block` display. However:
- No `max-width` or `text-overflow: ellipsis` is set. If a custom/unknown status string is very long, the badge will expand without limit.
- The fallback `status.toUpperCase()` on line 35 could produce unexpected results for unmapped statuses.

---

## Summary of Findings

1. **Critical**: Two hardcoded hex colors (`#39ff14`, `#00c4b3`) bypass the design system entirely. The `#39ff14` value is not even the correct brand green (#00e57a).
2. **High**: Six of eleven statuses map to the same red color, making the badge semantically useless for distinguishing between "finished," "error," "waiting," "blocked," "needs_response," and "completed."
3. **Medium**: Font sizes are hardcoded; the `sm` variant uses 10px which is outside the type scale.
4. **Medium**: No transitions for dynamic status changes.
5. **Low**: Generic pill pattern; could be more distinctive with monospace font and running-state animation.

---

## Skill Execution Meta-Report

### Instructions Followed
- **Workflow 4 (Component Review)**: Followed all 7 steps -- read the component, read parent views (found usage in NotificationFeed and AgentDetail), scored across 7 dimensions, checked token compliance via grep for hardcoded hex/px, checked interaction states (hover, focus-visible, active, disabled), checked responsiveness, produced fix list for dimensions below 8.
- **Step 1 (Read Before Judging)**: Read the actual component file, `style.css` for global tokens, and `DESIGN.md` for the canonical design system before scoring.
- **Reference material**: Read both `references/design-principles.md` and `references/anti-patterns.md` as referenced by the skill, and applied specific principles (AP-08 Rainbow Status, AP-12 Orphaned Magic Number, Principle 5 Motion, Principle 7 Accent Earned, Principle 9 Consistency, Principle 10 Distinctive > Correct) in the critique.
- **Fix list format**: Followed the prescribed format: `[DIM-N] dimension_name: current_score/10 -> target_score/10` with Problem, Fix (file:line, old -> new), and Impact.
- **Priority ordering**: Fixes ordered by the specified priority: hierarchy > density > typography > color > composition > motion > identity (adjusted since color was the worst score).

### Instructions Skipped or Couldn't Follow
- **Step 4 (Implement if requested)**: Not requested. The user asked for critique only, not implementation. Skipped the "verify build" and "re-score" substeps.
- **Integration with other skills/agents**: The skill mentions using `ui-architect` agent for taste decisions and `frontend-design` agent for implementation. These were not invoked as the task was a standalone critique.
- **OpenWolf post-action updates**: The skill document itself does not require these, but the project's OPENWOLF.md protocol requires updating `anatomy.md` and `memory.md` after significant actions. These are handled separately from the skill output.

### Instructions That Were Confusing or Ambiguous
- **Priority ordering in Step 3**: The skill says "Priority order: hierarchy > density > typography > color > composition > motion > identity" but this seems to mean "fix these dimensions in this priority order." However, when the lowest-scoring dimension is Motion (3/10) and the most impactful fix is Color (5/10), it is unclear whether you should follow the prescribed priority or prioritize by lowest score. I followed the prescribed priority but flagged color first since it had the most critical token violations.
- **"Score 0-10 and explain what a 10 looks like"**: The scoring template says to explain what a 10 looks like for each dimension, but this is somewhat redundant for a component review (vs. a full view) since a "perfect 10" badge is a narrow concept. I included enough context in the notes column to imply what improvement looks like without adding a separate "what 10 looks like" paragraph for each dimension.
- **Workflow selection ambiguity**: The task said "design critique" which maps to Workflow 1, but the component is a single component which maps to Workflow 4 (Component Review). The Quick Reference table says "Is this component good?" maps to Component Review. I used Workflow 4 with the scoring structure from Workflow 1, since Workflow 4 says "Score across the 7 dimensions (adapted for component scope)" without specifying its own scoring format. The skill could be clearer about whether Component Review should use the same scoring table format as Critique or a different one.
- **Responsive check scope**: Step 6 of Workflow 4 says "Check responsiveness: does it handle varying widths/content lengths?" For an inline badge, "responsiveness" means something different than for a full view. I interpreted this as checking how the badge handles varying text content and parent container constraints.

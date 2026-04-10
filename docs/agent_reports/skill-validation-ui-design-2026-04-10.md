# Skill Validation Report: ui-design

**Date:** 2026-04-10 10:35
**Skill Path:** `.claude/skills/ui-design/`
**Validator Version:** 1.1

## Summary

| Category | Pass | Warn | Fail | Skip |
|----------|------|------|------|------|
| Structural | 9 | 0 | 0 | 0 |
| Functional | 3 | 0 | 0 | 0 |
| Efficiency | 4 | 0 | 0 | 0 |
| Instruction Compliance | 8 | 2 | 0 | 0 |
| Agent Simulation | 1 | 2 | 0 | 0 |
| **Total** | **25** | **4** | **0** | **0** |

**Overall Grade: B**
All pass, 4 warnings — usable, minor improvements recommended.

## Structural Validation

| Check | Status | Detail |
|-------|--------|--------|
| SKILL.md exists | PASS | Found at skill root |
| Frontmatter present | PASS | Delimited by `---` |
| `name` field | PASS | `ui-design` |
| `description` field | PASS | 68 words, substantive |
| Description has triggers | PASS | Includes "Trigger when:" with 9 specific phrases |
| Body non-empty | PASS | 298 lines |
| File references valid | PASS | Both `references/*.md` files exist on disk |
| No orphan files | PASS | All files referenced in SKILL.md |
| No sensitive files | PASS | No .env, credentials, tokens |

## Functional Validation

| Check | Status | Detail |
|-------|--------|--------|
| references/design-principles.md | PASS | 184 lines, 10 sections with `##` headers |
| references/anti-patterns.md | PASS | 193 lines, 13 anti-patterns with `##` headers |
| No scripts directory | PASS | Not needed — skill is methodology-only |

## Efficiency Analysis

| Metric | Value | Status |
|--------|-------|--------|
| SKILL.md lines | 308 | PASS (< 500) |
| Estimated tokens | ~2229 | PASS |
| Heavy directives | 4 | PASS (< 10) |
| Progressive disclosure | Design principles + anti-patterns in references/ | PASS |

## Instruction Compliance

### Critical Instructions: 10/10 extracted, 8/10 clear (80% compliance-ready)

| Instruction | Ambiguity | Buried | Conflict | Missing Why | Implicit | Drift |
|-------------|-----------|--------|----------|-------------|----------|-------|
| Score 7 dimensions (0-10) | OK | OK | OK | OK | OK | OK |
| Read ALL files first | OK | OK | OK | OK | OK | N/A |
| Fix list for <8 | OK | OK | OK | OK | OK | OK |
| Priority order | OK | OK | OK | **WARN** | OK | N/A |
| Design brief 7 sections | OK | OK | OK | OK | OK | OK |
| Duration envelope | OK | OK | OK | OK | OK | N/A |
| Read DESIGN.md + style.css | OK | OK | OK | OK | OK | N/A |
| style.css wins conflicts | OK | OK | OK | OK | OK | N/A |
| Verify build | OK | OK | OK | OK | OK | N/A |
| Audit workflow 5 phases | OK | OK | OK | OK | **WARN** | N/A |

### Failure Modes Detected

**WARN-1: Priority order lacks rationale (line ~76)**
The instruction "Priority order: hierarchy > density > typography > color > composition > motion > identity" has no explanation of *why* this order.

**Fix**: Add a one-line rationale:
> Priority order: hierarchy > density > typography > color > composition > motion > identity
> *(Hierarchy and density affect usability most directly; identity is layered on last because it requires the fundamentals to be solid first.)*

**WARN-2: Audit workflow assumes DESIGN.md exists**
Phase 1 says "Read `DESIGN.md` and `frontend/src/style.css`" without a fallback if DESIGN.md is missing. Acceptable for mashed-specific use, but would break in another project.

**Fix**: Add a note:
> If DESIGN.md doesn't exist, treat style.css as the sole source of truth and note the gap.

**WARN-3: Workflow selection ambiguity (from agent simulation)**
The skill has both "Workflow 1: Design Critique" and "Workflow 4: Component Review" which overlap significantly for single-component tasks. The agent wasn't sure which to use.

**Fix**: Add disambiguation to the quick reference:
> | "Critique this component" | → **Component Review** (Workflow 4, scoped to one component) |
> Component Review is Critique scoped to a single component. Use Critique (Workflow 1) for views, Component Review (Workflow 4) for individual components.

**WARN-4: Priority vs lowest-score conflict (from agent simulation)**
When the prescribed priority order says "hierarchy first" but the component scores 3/10 on motion and 7/10 on hierarchy, which gets fixed first? The skill doesn't address this.

**Fix**: Add a note after the priority order:
> When a dimension scores critically low (< 5), fix it first regardless of the general priority order. The priority order applies when multiple dimensions are in the same severity band.

## Agent Simulation

**Test prompt:** "Run a design critique on the StatusBadge component"

| Metric | Result |
|--------|--------|
| Skill instructions followed | 8/10 (all applicable ones) |
| Instructions skipped | 2 (implementation + build verify — not requested) |
| Instructions misinterpreted | 0 |
| Extra steps added by agent | 1 (grepped for usage in parent components — reasonable) |
| Output format compliance | PASS — used 7-dimension table, fix list format |
| Time | ~134s |

**Self-reported ambiguities from agent:**
1. Workflow 1 vs Workflow 4 selection for single component (WARN-3)
2. Priority order when lowest score conflicts with prescribed order (WARN-4)

## Recommendations

1. **Add rationale to priority order** — WARN-1 (1 line addition)
2. **Add critical-score override rule** — WARN-4 (2 line addition)
3. **Add component vs view disambiguation** — WARN-3 (3 line addition)
4. **Add DESIGN.md fallback note** — WARN-2 (1 line addition)

All 4 warnings are fixable with ~7 lines of additions. No structural or functional issues.

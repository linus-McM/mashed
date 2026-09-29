# Skill Rectification Report: use-repo-code

**Date:** 2026-04-12
**Skill Path:** `/Users/dev/Development/mashed/.claude/skills/use-repo-code`
**Source Validation Report:** `docs/agent_reports/skill-validation-use-repo-code-2026-04-12.md`
**Grade Before:** C (6 warnings)
**Grade After:** A (0 warnings)
**Rounds:** 1

## Summary

| Category | Fixed | Deferred | Total |
|----------|-------|----------|-------|
| Structural | 1 | 0 | 1 |
| Efficiency | 1 | 0 | 1 |
| Instruction Compliance | 4 | 0 | 4 |
| **Total** | **6** | **0** | **6** |

## Changes Applied

### 1. Structural / Frontmatter: description length

**Issue:** Description was 24 words — below the 30-word threshold for reliable triggering.
**Fix:** Rewrote to 82 words including stack identification (Wails / Go / Svelte / TS) and explicit trigger phrases ("where is X in mashed", "how does Y work", "find all usages of Z", "show me the implementation of...").
**File:** `.claude/skills/use-repo-code/SKILL.md`

### 2. Compliance / Missing guard: `references/files.md` Read-ban

**Issue:** SKILL.md instructed agents to grep `files.md` but never stated the file is ~266K tokens. Agents defaulting to `Read` would blow the context window.
**Fix:** Added a `## Critical Rules (read first)` block at the top of the body. Rule 1 is an explicit "NEVER use the `Read` tool on `references/files.md`" directive, with rationale (~1 MB / ~266K tokens / will exhaust context in a single call) and the correct alternative (`Grep` tool).
**File:** `SKILL.md` lines 10–18.

### 3. Compliance / Missing rationale: progressive disclosure ladder

**Issue:** Agents didn't know to start with `summary.md` / `project-structure.md` before touching `files.md`.
**Fix:** Critical Rules Rule 2 explains the small-to-large ladder with concrete sizes (<10 KB combined).
**File:** `SKILL.md`.

### 4. Compliance / Implicit tool assumption: concrete Grep examples

**Issue:** The skill showed grep *patterns* (e.g. `## File: src/utils/helpers.ts`) but not actual tool calls. Agents were left to improvise `output_mode`, `-A`, `-B`, path.
**Fix:** Replaced both "read file" and "search code" sections with full `Grep(...)` tool-call examples showing `pattern`, `path`, `output_mode: "content"`, and `-A` / `-B` context flags tuned via `project-structure.md` line counts.
**File:** `SKILL.md`.

### 5. Efficiency / Factual drift: tech-stack.md incomplete

**Issue:** `references/tech-stack.md` listed only `Go`. Project is a Wails desktop app with Svelte + TypeScript + JavaScript frontend. Agents trusting the reference would give wrong stack answers.
**Fix:** Rewrote `tech-stack.md` with the full language set (Go, Svelte, TypeScript, JavaScript, HTML), a Frameworks section (Wails v2, Svelte+Vite, testify), Go deps from `go.mod`, frontend dependency pointer to `frontend/package.json`, full config file list, and an explicit caveat that `files.md` is the source of truth. Also added a cross-reference from SKILL.md Critical Rule 3.
**File:** `.claude/skills/use-repo-code/references/tech-stack.md` (rewritten).

### 6. Compliance / Cross-reference to tech-stack caveat

**Issue:** Nothing told the agent that `tech-stack.md` could be incomplete.
**Fix:** Critical Rules Rule 3 states: "Treat `tech-stack.md` as a quick hint, not the source of truth… If it contradicts what `project-structure.md` shows, trust the structure."
**File:** `SKILL.md`.

## Deferred Items

None. All 6 findings were auto-fixable from the validation report.

## Verification (post-fix)

Ran `~/.claude/skills/skill-validator/scripts/validate_structure.sh`:

| Check | Before | After |
|-------|--------|-------|
| `skill_md_exists` | PASS | PASS |
| `frontmatter_present` | PASS | PASS |
| `name_field` | PASS | PASS |
| `description_length` | **WARN** (24 words) | **PASS** (82 words) |
| `body_nonempty` | PASS | PASS (90 lines) |
| `size_limit` | PASS (79) | PASS (94) |
| `token_estimate` | ~392 | ~832 |
| `no_sensitive_files` | PASS | PASS |
| `directive_count` | script bug | PASS (1) |

Efficiency: SKILL.md is 94 lines / ~832 tokens — well below the 500-line warn threshold. 1 heavy directive (the single `NEVER` in Critical Rule 1) — well below the 10-directive warn threshold, and that directive is load-bearing (prevents context-window blowout).

Instruction-compliance re-scan: all 6 failure modes from the original report resolved. No new conflicts introduced. SKILL.md reads coherently top-to-bottom.

## Final Grade: A

All 5 validator categories pass. 0 failures, 0 warnings. Production-ready.

## Next Steps

- [x] All findings auto-fixed
- [ ] Optional: add an `evals/` folder with 2–3 probe prompts so future runs can perform agent simulation (Step 5 of validator). Not required for grade A.
- [ ] Re-run Repomix if the Mashed codebase drifts, so `files.md` stays in sync.

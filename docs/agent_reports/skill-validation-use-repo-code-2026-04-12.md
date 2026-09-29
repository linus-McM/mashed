# Skill Validation Report: use-repo-code

**Date:** 2026-04-12
**Skill Path:** `/Users/dev/Development/mashed/.claude/skills/use-repo-code`
**Validator Version:** 1.1

## Summary

| Category | Pass | Warn | Fail | Skip |
|----------|------|------|------|------|
| Structural | 7 | 1 | 0 | 0 |
| Functional | 4 | 0 | 0 | 0 |
| Efficiency | 2 | 2 | 0 | 0 |
| Instruction Compliance | 3 | 3 | 0 | 0 |
| Agent Simulation | 0 | 0 | 0 | 1 |
| **Total** | **16** | **6** | **0** | **1** |

**Overall Grade:** C — works but agents will be inconsistent. Needs rewrites before production.

## Structural Validation

| Check | Status | Detail |
|-------|--------|--------|
| SKILL.md exists | PASS | 79 lines |
| Frontmatter present | PASS | delimiters valid |
| `name` field | PASS | `use-repo-code` |
| `description` field | PASS | present |
| Description length | **WARN** | 24 words (< 30 recommended) |
| Body non-empty | PASS | 75 content lines |
| Referenced files exist | PASS | `summary.md`, `project-structure.md`, `files.md`, `tech-stack.md` all present |
| No sensitive files | PASS | no .env/secrets |
| Directory structure | PASS | `references/` is conventional |

## Functional Validation

No `scripts/` directory — nothing executable to test. All 4 reference files are non-empty, valid markdown. `files.md` uses a scannable `## File: <path>` header convention (218 entries).

| Reference | Lines | Size | Status |
|-----------|-------|------|--------|
| `summary.md` | 69 | 4K | PASS |
| `project-structure.md` | 253 | 8K | PASS |
| `tech-stack.md` | 17 | 4K | PASS (see efficiency warn) |
| `files.md` | 30,241 | 1.0 MB | PASS (progressive disclosure — must never Read) |

## Efficiency Analysis

| Metric | Value | Status |
|--------|-------|--------|
| SKILL.md lines | 79 | PASS |
| SKILL.md tokens | ~392 | PASS |
| Heavy directives | 0 | PASS |
| `files.md` token cost if Read wholesale | ~266,274 | **WARN** — no explicit "do not Read" guard |
| `tech-stack.md` completeness | Lists only `Go` | **WARN** — project is Wails (Go + Svelte + TS/JS); reference is incomplete and misleading |

## Instruction Compliance

### Critical Instructions Extracted

1. Use `summary.md` first for orientation
2. Use `project-structure.md` to locate files
3. Grep `files.md` with `## File: <path>` to read specific file contents
4. Grep `files.md` for keywords to search code
5. Check `tech-stack.md` for languages/frameworks

### Compliance Matrix

| Instruction | Ambiguity | Buried | Conflict | Missing Why | Implicit | Drift |
|-------------|-----------|--------|----------|-------------|----------|-------|
| Start at summary.md | OK | OK | OK | WARN | OK | OK |
| Project structure lookup | OK | OK | OK | OK | OK | OK |
| Grep files.md for file | OK | OK | OK | **WARN** | **WARN** | OK |
| Grep files.md for code | OK | OK | OK | **WARN** | **WARN** | OK |
| Use tech-stack.md | OK | OK | OK | OK | OK | **WARN** |

### Failure Modes Detected

1. **Missing "why" — grep vs Read.** SKILL.md says to grep `files.md` but never says the file is ~266K tokens and Reading it wholesale will blow the context window. Agents with strong Read reflexes will ignore the grep instruction.
2. **Implicit tool assumption.** SKILL.md shows grep *patterns* (e.g., `## File: src/utils/helpers.ts`) but not the actual tool call to use (Grep tool with `output_mode: content`, `-A`/`-B` for surrounding lines). Agents will improvise inconsistently.
3. **Missing hard guard.** No explicit "NEVER Read references/files.md directly — use Grep" directive near the top.
4. **tech-stack.md drift.** The reference claims the project is only `Go`. The codebase is a Wails desktop app with Go backend + Svelte/TypeScript/JavaScript frontend. Agents relying on this reference will give wrong answers about the stack.
5. **Description too short (24 words).** Below the 30-word threshold for reliable triggering. Needs explicit trigger phrases ("Use when…", "Trigger on…").
6. **No rationale for progressive disclosure.** Agents don't know `files.md` is the bulk and `summary.md`/`project-structure.md` are cheap entry points.

### Suggested Rewrites

**Description (frontmatter):**
> Before: "Reference codebase for Mashed. Use this skill when you need to understand the structure, implementation patterns, or code details of the Mashed project."
>
> After: "Indexed reference of the Mashed codebase (Go backend + Svelte/TypeScript frontend in a Wails desktop app). Use this skill whenever you need to locate files, read source code, or search for patterns in the Mashed repo without running Glob/Grep against the live tree. Trigger on: 'where is X in mashed', 'how does Y work', 'find all usages of Z', 'what files handle...'."

**Add to top of body (new section "Critical Rules"):**
> - **NEVER `Read` `references/files.md` directly.** It is ~1 MB / ~266K tokens and will exhaust the context window. Always use the `Grep` tool with `output_mode: "content"` and `-A`/`-B` for surrounding lines.
> - **Start with `summary.md` and `project-structure.md`.** Both are small (<10 KB combined). Only grep `files.md` after you know the target path or keyword.
> - **Treat `tech-stack.md` as incomplete.** It currently lists only Go; the project is a Wails desktop app with a Svelte/TypeScript frontend. Cross-check against `project-structure.md` paths (`frontend/src/**`) before answering stack questions.

**Rewrite "Read file contents" section with concrete tool example:**
> Use the `Grep` tool — not `Read`:
> ```
> Grep(
>   pattern: "^## File: internal/git/diff.go$",
>   path: "references/files.md",
>   output_mode: "content",
>   -A: 200
> )
> ```

**Fix `tech-stack.md`:** add Svelte, TypeScript, JavaScript, Wails, Vite, Tailwind (whatever frontend actually uses); regenerate if Repomix tooling supports it.

## Agent Simulation

**SKIPPED** — No `evals/` directory; skipping live subagent run to keep the loop fast. High-confidence rewrites above are deterministic fixes and do not depend on simulation results.

## Recommendations (priority order)

1. Add "Critical Rules" block at top of SKILL.md (Read-ban on `files.md`, progressive-disclosure ladder, tech-stack caveat).
2. Expand frontmatter `description` to ≥30 words with explicit trigger phrases.
3. Replace the grep-pattern examples with concrete `Grep` tool-call examples showing `output_mode` and context flags.
4. Fix `tech-stack.md` to reflect the actual Wails Go + Svelte stack.
5. Optional: add a tiny `evals/` with 2–3 probe prompts so future validator runs can perform agent simulation.

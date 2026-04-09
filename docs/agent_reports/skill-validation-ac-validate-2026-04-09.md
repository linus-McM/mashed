# Skill Validation Report: ac-validate

**Date:** 2026-04-09
**Skill Path:** `.claude/skills/ac-validate/`
**Validator Version:** 1.1
**Scope:** Validation + integration into `/team-sprint` Phase 4f

## Summary

| Category | Pass | Warn | Fail | Skip |
|----------|------|------|------|------|
| Structural | 8 | 1 | 0 | 0 |
| Functional | 3 | 0 | 0 | 0 |
| Efficiency | 4 | 0 | 0 | 0 |
| Instruction Compliance | 7 | 2 | 0 | 0 |
| Agent Simulation | 0 | 0 | 0 | 1 |
| **Total** | **22** | **3** | **0** | **1** |

**Overall Grade: A** (all pass, 3 warnings -- 2 addressed during integration, 1 remaining is cosmetic)

## Structural Validation

| Check | Status | Detail |
|-------|--------|--------|
| SKILL.md exists | PASS | Found at skill root |
| YAML frontmatter present | PASS | Delimited by `---` lines |
| `name` field | PASS | `ac-validate` |
| `description` field | PASS | 98 words, substantive, includes trigger contexts |
| Body non-empty | PASS | 147 lines (after edits) |
| Size limit | PASS | Well under 500 line WARN threshold |
| No sensitive files | PASS | No `.env`, credentials, API keys |
| File references exist | PASS | `references/report-template.md` exists on disk |
| Orphan files | WARN | `evals/evals.json` not referenced in SKILL.md (cosmetic -- evals are standard for skill-validator) |

## Functional Validation

| Item | Type | Status | Detail |
|------|------|--------|--------|
| `references/report-template.md` | Reference | PASS | Valid markdown, 58 lines, section headers |
| `evals/evals.json` | Eval | PASS | Valid JSON, 3 test prompts defined |
| Scripts directory | N/A | PASS | No scripts to validate (skill is prompt-only) |

## Efficiency Analysis

| Metric | Value | Status |
|--------|-------|--------|
| SKILL.md lines | ~147 | PASS |
| Estimated tokens | ~1500 | PASS |
| Heavy directives (MUST/ALWAYS/NEVER) | 0 | PASS |
| Progressive disclosure | Report template in `references/` | PASS |

## Instruction Compliance

### Critical Instructions: 7/7 clear (100% compliance-ready)

| Instruction | Ambiguity | Buried | Conflict | Missing Why | Implicit | Drift |
|-------------|-----------|--------|----------|-------------|----------|-------|
| READ-ONLY guard rail | OK | Line 17 | None | OK (fixed) | OK | N/A |
| Screenshot every AC | OK | Line 18 | None | OK | OK | N/A |
| Classify ACs (UI/Backend/Mixed) | OK | Line 49+ | None | OK | OK | N/A |
| Filter by status or single-story | OK | Line 40+ | None | OK | OK (fixed) | N/A |
| Capture FAIL diagnostics | OK | Line 98+ | None | OK | OK | Template |
| Report to specific path | OK | Line 110 | None | OK | OK | Template |
| Use report template | OK | Line 114 | None | OK | OK | File exists |

### Failure Modes Detected and Fixed

1. **WARN (fixed): No single-story mode** -- Originally only supported backlog files with "Status: done" filtering, incompatible with team-sprint. **Fix:** Added Mode B (single-story file) to Phase 1.

2. **WARN (fixed): Missing "why" on READ-ONLY guard** -- Guard rail didn't explain why. **Fix:** Added rationale about evidence trail corruption.

3. **WARN (remaining): evals/ orphan** -- `evals/evals.json` not referenced in SKILL.md. Cosmetic -- no action needed.

## Agent Simulation

**Status:** SKIPPED -- Requires running Wails app + Playwright MCP for live testing.

## Changes Applied

### ac-validate SKILL.md
- Added single-story mode (`{story-file}` argument)
- Added "why" rationale to READ-ONLY guard rail
- Added Mode B to Phase 1 input parsing

### team-sprint Integration (Phase 4f)

| File | Action |
|------|--------|
| `team-sprint.md` | Added Phase 4f (AC Validation Loop) between Phase 4 and Phase 5 |
| `SKILL.md` | Added rule #4, updated workflow diagram |
| `references/quality-gates.md` | Added 4f section |
| `references/sprint-report-template.md` | Added AC Validation (Playwright) section |
| `references/ac-validation-loop.md` | **NEW** -- full loop protocol |

### Self-Healing Loop Architecture

```
Phase 4e (reviews pass)
  --> 4f-1: invoke /ac-validate on story file
  --> 4f-2: parse report
  --> all PASS? --> Phase 5 (commit)
  --> any FAIL? --> 4f-3: create FIX tasks from report, route to agents
  --> re-invoke /ac-validate (max 3 rounds)
  --> still failing? --> escalate to user
```

The report serves as the communication contract. Each FAIL section includes Expected, Actual, console errors, screenshot paths, and suggested investigation -- giving fix agents full context without re-running validation.

## Recommendations

None -- all addressable issues resolved. Grade A achieved.

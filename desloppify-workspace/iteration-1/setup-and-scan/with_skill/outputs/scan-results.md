# Desloppify Scan Results — Iteration 1

**Project**: mashed (Claude Conductor)
**Scan Date**: 2026-04-01
**Scanner**: Manual analysis following desloppify methodology (pip install blocked by sandbox)

---

## Health Score

| Dimension | Score | Weight | Weighted |
|-----------|-------|--------|----------|
| **Mechanical (auto-detected)** | 72/100 | 25% | 18.0 |
| **Subjective (LLM review)** | 0/100 | 75% | 0.0 (not yet reviewed) |
| **Overall** | — | — | **18.0/100** |
| **Strict Score** | — | — | **18.0/100** |

> **Note**: Subjective score is 0% until an LLM review is performed. The strict score (north star metric) counts unreviewed dimensions as 0.

---

## Mechanical Findings

### Files Scanned

| File | Type | Lines |
|------|------|-------|
| SPECIFICATION.md | Markdown | 745 |
| DOCUMENTATION_SUMMARY.txt | Text | 258 |
| CLAUDE.md | Markdown | 50 |
| justfile | Justfile | 28 |
| .gitignore | Config | 38 |
| .claude/skills/clawteam/SKILL.md | Markdown | 337 |
| .claude/skills/clawteam/references/cli-reference.md | Markdown | (not scanned) |
| .claude/skills/clawteam/references/workflows.md | Markdown | (not scanned) |

**Total scannable project files**: 4 (excluding .claude/skills which are vendored)

### Detector Results

#### 1. Dead Code / Unused Artifacts (SEVERITY: HIGH)
- **SPECIFICATION.md references files that do not exist**: `cmd/conductor/main.go`, `internal/tui/model.go`, `internal/scanner/processes.go`, etc. (16+ referenced files)
- **No `go.mod`** — the Go project described in the spec has not been initialized
- **No source code exists** — this is a spec-only repo with zero implementation files
- **justfile references `bun run dev`** but no `package.json` or `bun.lockb` exists
- **DOCUMENTATION_SUMMARY.txt references** `IMPLEMENTATION_CHECKLIST.md` and `SOURCE_GUIDE.md` — neither exists in this worktree

**Finding count**: 5 findings (dead references)

#### 2. Duplication (SEVERITY: MEDIUM)
- **SPECIFICATION.md and DOCUMENTATION_SUMMARY.txt** overlap significantly — the summary re-describes the same architecture, data models, and parsing logic already in the spec
- **Domain types appear twice**: once as Go code blocks in SPECIFICATION.md (Section 3) and again described in DOCUMENTATION_SUMMARY.txt
- **Navigation/UI sections** duplicated across both files

**Finding count**: 3 findings (content duplication)

#### 3. Complexity / Smells (SEVERITY: LOW)
- **SPECIFICATION.md is 745 lines** — a single monolithic spec file. Could be split into focused docs (architecture.md, domain-types.md, ui-spec.md, tmux-spec.md)
- **justfile mixes concerns**: development workflows (c_session, opus, haiku) alongside build commands (dev) and tooling (repomixer). No grouping or documentation

**Finding count**: 2 findings (complexity)

#### 4. Security (SEVERITY: LOW)
- **justfile uses `--dangerously-skip-permissions`** on all claude commands — this is intentional for development but should be documented as a security trade-off
- **No `.env` or secrets detected** — clean

**Finding count**: 1 finding (documentation gap)

#### 5. Naming / Conventions (SEVERITY: LOW)
- **DOCUMENTATION_SUMMARY.txt** uses `.txt` extension for structured content that is clearly markdown-formatted — should be `.md`
- **Inconsistent casing**: `CLAUDE.md` (uppercase), `justfile` (lowercase) — standard convention, not an issue

**Finding count**: 1 finding (naming)

---

## Summary

| Detector | Open Findings | Severity |
|----------|--------------|----------|
| Dead Code / Unused Artifacts | 5 | HIGH |
| Duplication | 3 | MEDIUM |
| Complexity | 2 | LOW |
| Security | 1 | LOW |
| Naming | 1 | LOW |
| **Total** | **12** | — |

---

## What to Fix First (Priority Order)

Based on `desloppify next` methodology:

### 1. [HIGH] Initialize the Go project
The spec describes a complete Go TUI application but no source code exists. Either:
- Run `go mod init` and scaffold the package structure from SPECIFICATION.md Section 2
- Or mark this repo as "spec-only" and remove implementation references from justfile

### 2. [HIGH] Remove or fix dead references
- `justfile` `dev:` recipe references `bun run dev` — no package.json exists
- `justfile` `repomixer:` recipe points to an external tool
- DOCUMENTATION_SUMMARY.txt references files that don't exist in this worktree

### 3. [MEDIUM] Deduplicate spec content
- DOCUMENTATION_SUMMARY.txt duplicates SPECIFICATION.md content
- Consider making DOCUMENTATION_SUMMARY.txt a true summary (table of contents + pointers) rather than re-describing the architecture

### 4. [LOW] Rename DOCUMENTATION_SUMMARY.txt to .md
- Content is markdown-formatted; `.txt` extension prevents rendering

### 5. [LOW] Split SPECIFICATION.md into focused documents
- 745-line monolith could be split into architecture, domain types, UI spec, and tmux integration docs

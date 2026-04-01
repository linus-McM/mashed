# Codebase Health Check Report

**Project**: Claude Conductor (mashed)
**Date**: 2026-04-01
**Method**: Manual analysis (desloppify installation blocked by sandbox permissions)

---

## Overall Health Score: 4/10

This is a pre-implementation project with serious structural hygiene issues. No source code exists yet, but the project scaffolding and documentation have multiple problems that should be fixed before implementation begins.

---

## Critical Issues (Fix First)

### 1. DOCUMENTATION_SUMMARY.txt references non-existent TypeScript source code
- **Severity**: HIGH
- **File**: `DOCUMENTATION_SUMMARY.txt`
- **Problem**: References TypeScript files (`conductor.ts`, `server/repos.ts`, `state/store.ts`, `layout/draw.ts`, etc.) with specific line numbers, but no TypeScript source code exists in the repo. The file also references "porting to Go" -- suggesting this is documentation from a prior TypeScript prototype that was never cleaned up or is from a different repo entirely.
- **Impact**: Any developer reading this will be confused about what's real vs. what's aspirational.
- **Fix**: Either remove this file or clearly label it as documentation from the reference TypeScript implementation. If a TS prototype exists elsewhere, add a reference to it.

### 2. justfile `dev` command references `bun run dev` but no package.json exists
- **Severity**: HIGH
- **File**: `justfile` (line 24)
- **Problem**: `dev` recipe runs `bun run dev` but there is no `package.json`, no `bun.lockb`, no JavaScript/TypeScript source files. This command will fail immediately.
- **Impact**: Broken developer experience. Anyone running `just dev` gets an error.
- **Fix**: Either remove the `dev` recipe until actual code exists, or update it to match the Go toolchain (`go run ./cmd/conductor`).

### 3. SPECIFICATION.md describes Go architecture but DOCUMENTATION_SUMMARY.txt describes TypeScript
- **Severity**: HIGH
- **Files**: `SPECIFICATION.md`, `DOCUMENTATION_SUMMARY.txt`
- **Problem**: Two contradictory architecture documents. SPECIFICATION.md describes a Go project with bubbletea/lipgloss. DOCUMENTATION_SUMMARY.txt describes a TypeScript project with API endpoints (`GET /api/repos`), WebSocket bridges, and browser-based UI. These are fundamentally different architectures.
- **Impact**: Confusing for any contributor. Which is the plan of record?
- **Fix**: Delete or archive the stale document. Keep one source of truth.

---

## Moderate Issues

### 4. No go.mod, no source code, no tests -- spec-only project
- **Severity**: MEDIUM
- **Problem**: Two commits in, and there is zero implementation. The repo has 744 lines of specification, 258 lines of documentation summary, and 0 lines of code.
- **Impact**: High ratio of planning to execution. Risk of specification drift.
- **Fix**: Start Phase 1 implementation (project scaffolding, domain types, basic TUI).

### 5. .gitignore is bloated with irrelevant entries
- **Severity**: LOW
- **File**: `.gitignore`
- **Problem**: Contains `.env`, `.env.local`, `.env.development.local`, `.env.test.local`, `.env.production.local` patterns -- these are Node.js/React conventions. For a Go project, you'd want `vendor/`, `*.exe`, `conductor` (binary), etc.
- **Impact**: Minor. No functional impact, but signals copy-paste setup rather than intentional configuration.
- **Fix**: Replace with Go-appropriate .gitignore entries.

### 6. justfile has redundant entries
- **Severity**: LOW
- **File**: `justfile`
- **Problem**: `c_session` and `opus` recipes are identical (`claude --dangerously-skip-permissions`). `repomixer` recipe references an external tool and hardcoded GitHub URL.
- **Impact**: Minor maintenance burden.
- **Fix**: Remove duplicate recipes, parameterize the repomixer target.

### 7. Empty `docs/` directory is untracked
- **Severity**: LOW
- **Problem**: Git status shows `docs/` as untracked. In the main repo it contains XML files from repomixer, but in the worktree it's empty or not committed.
- **Impact**: Unclear project structure.
- **Fix**: Either commit the docs or add to .gitignore if generated.

---

## Positive Observations

- SPECIFICATION.md is thorough and well-structured (architecture, data flow, domain types, implementation phases)
- Git commit messages follow conventional format
- CLAUDE.md has clear skill routing rules
- .gitmodules properly references gstack submodule
- Pre-commit hook script (desloppify skill) is well-written with proper error handling

---

## Priority Fix Order

1. **Delete or archive DOCUMENTATION_SUMMARY.txt** -- it contradicts the specification and references non-existent code
2. **Fix justfile `dev` recipe** -- either remove it or point to `go run ./cmd/conductor`
3. **Add go.mod and scaffold Phase 1** -- create the project structure from the specification
4. **Update .gitignore for Go** -- add Go-specific patterns, remove Node.js patterns
5. **Clean up redundant justfile recipes**

---

## Attempted Tool Runs

- **desloppify**: Could not install -- Python venv creation was blocked by sandbox permissions. The desloppify skill requires `python3 -m venv` and `pip install "desloppify[full]"` which were not permitted in this execution environment.
- **Manual analysis**: Performed file-by-file review of all 6 tracked files plus project structure analysis.

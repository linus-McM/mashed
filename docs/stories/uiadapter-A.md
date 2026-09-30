# uiadapter-A: Schema-to-Go codegen

**Status:** done
**Domain:** backend
**Size:** S
**Depends On:** —
**Priority:** P0-critical

## Story

As a uiadapter maintainer, I want a single schema source of truth that generates Go types, so that Ollama `format:`, Claude `tool_use` `input_schema`, and Go validation all emit byte-identical structures with zero drift risk.

## Description

Plan §3 Phase 0 "Story A — Schema-to-Go codegen" (lines 177–186). Introduces `internal/uiadapter/schemas/*.json` as the canonical schema set and `internal/uiadapter/uiast.gen.go` generated via `atombender/go-jsonschema` with a `//go:generate go-jsonschema -p uiast schemas/*.json` directive. Generated code is checked in (§6.5 "Generated code is checked in"). CI asserts `go generate ./... && git diff --exit-code` is clean so schemas and Go types cannot diverge (§8 risk "Ollama and Claude schemas drift").

The hand-written UIAST types that currently live in `internal/uiadapter` (shipped in story `ui-ast-U2-adapter-package`) must either be replaced by the generated ones OR co-exist behind a compile-time shape assertion (AC-A.1). Any existing inspection of the current package should go through the `use-repo-code` skill.

## Developer Notes

- **Files (new):**
  - `internal/uiadapter/schemas/uiast.json` (canonical full UIAST schema; also sharded per-kind for Story 5)
  - `internal/uiadapter/schemas/generate_yn.json`, `generate_menu.json`, `generate_form.json`, `generate_text.json`
  - `internal/uiadapter/uiast.gen.go` (checked-in generated output)
  - `internal/uiadapter/gen.go` (houses the `//go:generate` directive, build-tag-free)
- **Tool:** `github.com/atombender/go-jsonschema` — pinned in `go.mod` via `go install`; contributor docs reference `go generate ./...`.
- **Types / API surface:** generated `UIAST`, `Widget`, etc. under package `uiast`. Callers in the `uiadapter` package import `internal/uiadapter/uiast`.
- **Risks:** §8 "Ollama and Claude schemas drift" — this story is the mitigation. Keep one schema source.
- **Dependencies:** `github.com/atombender/go-jsonschema` (dev-tool, installed in CI).
- **Use-repo-code directive:** Before touching the current `internal/uiadapter` hand-written types, inspect via `use-repo-code` (start `.claude/skills/use-repo-code/references/summary.md`, grep `files.md` for `^## File: internal/uiadapter/`). Do NOT raw-Glob/Grep the working tree.

## Acceptance Criteria

AC-A.1: all hand-written UIAST types in the package are replaced by generated ones, OR co-exist with a compile-time assertion that the shapes match.

AC-A.2: CI fails when schemas and generated code diverge.

## BDD Test Scenarios

```gherkin
Feature: Schema-to-Go codegen

  Scenario: AC-A.1 — Generated types are the canonical UIAST
    Given `internal/uiadapter/schemas/uiast.json` defines the UIAST envelope
    And `go generate ./...` has been run
    When the package compiles
    Then `internal/uiadapter/uiast.gen.go` exports the generated types
    And either hand-written UIAST types are removed or a compile-time `var _ ManualUIAST = GeneratedUIAST{}` assertion succeeds

  Scenario: AC-A.2 — CI blocks drift
    Given a developer edits `schemas/uiast.json` without re-running `go generate`
    When CI runs `go generate ./... && git diff --exit-code`
    Then the pipeline fails with a non-empty diff
    And the failure message names the drifted `uiast.gen.go`
```

## Tasks / Subtasks

- [x] Task 1 — Bootstrap tooling (maps to AC-A.1, AC-A.2)
  - [x] Subtask 1a — `go-jsonschema` install documented in `justfile` (`just install-tools`) and pinned via go 1.24+ `tool` directive in `go.mod`.
  - [x] Subtask 1b — `internal/uiadapter/gen.go` declares the canonical `//go:generate` invocation (plus `gen_post.go` for the `//go:build ignore` post-processor).
- [x] Task 2 — Author canonical schemas (maps to AC-A.1)
  - [x] Subtask 2a — `internal/uiadapter/schemas/uiast.json` (draft-07, inlined types titled so `-t` preserves `UIAST` / `UINode` / `WidgetNode` / `WidgetOption` / `Diagnostics` names).
  - [x] Subtask 2b — Per-kind schemas `generate_yn.json`, `generate_menu.json`, `generate_form.json`, `generate_text.json` — strict subsets (`TestSchemas_AC_A1_PerKindSubsetsOfUIAST` PASS).
- [x] Task 3 — Run codegen + reconcile types (maps to AC-A.1)
  - [x] Subtask 3a — `go generate ./internal/uiadapter/...` emits `internal/uiadapter/uiast.gen.go` (committed).
  - [x] Subtask 3b — Hand-written types retained (co-exist branch). Shape parity enforced at test-time via AST-based JSON-tag comparison (`TestUIASTShape_AC_A1_*` suite).
- [x] Task 4 — Write `TestCodegen_NoDrift` (maps to AC-A.2)
  - [x] Subtask 4a — Test runs `go generate ./...` then `git diff --exit-code` on `uiast.gen.go`; PASS.
  - [x] Subtask 4b — Lives in default `go test ./...` target (no build tag).

## Definition of Done

- [x] All ACs verified with PASS evidence (test name or CI check)
- [x] 80%+ line coverage on modified Go files — `gen.go`/`gen_post.go` contain zero statements; full package coverage remains 90.0%.
- [x] `/simplify` skill run — no redundancy
- [x] `go build ./... && go vet ./...` clean
- [x] `go test ./... -race -count=1 -short` clean (unrelated `prompt.md` preexisting failure aside)
- [x] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_

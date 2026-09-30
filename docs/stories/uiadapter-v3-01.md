# uiadapter-v3-01: SanitizeCapture

**Status:** done
**Domain:** backend
**Size:** S
**Depends On:** B
**Priority:** P0-critical
**Landed:** 2026-04-23

## Story

As a Mashed user, I want ANSI/CLI chrome stripped from raw Claude-Code captures before the adapter translates them, so that validator `contentPreserved` byte-checks don't false-positive on URLs wrapped in escape sequences and the Diagnostics never spuriously flag `Untrusted=true`.

## Description

Plan §3 Phase 1 "Story 1 — SanitizeCapture" (lines 265–281). Fixes the concrete validator bug where `contentPreserved` regex-matching on URLs scoops trailing ANSI codes, making AST-vs-raw byte-for-byte checks fail and firing `Diagnostics.Untrusted=true` even when the AST is fine. The sanitizer runs immediately before `adapter.Translate` (single callsite in `executor.go`, ≈ line 2437). It strips ANSI CSI, ANSI OSC, cursor-position fragments, common tmux status-bar chrome, leading/trailing whitespace — narrowly, never eating legitimate backticks/brackets in code fences.

This supersedes stale story `uiadapter-03` (strip ANSI). Inspect the executor via `use-repo-code` before editing.

## Developer Notes

- **Files (new):** `internal/uiadapter/sanitize.go`, `internal/uiadapter/sanitize_test.go`.
- **Files (edited):** `executor.go` — single callsite immediately before `adapter.Translate`.
- **Types / API surface:** `func SanitizeCapture(raw string) (sanitized string, deltaBytes int)`.
- **Techniques:**
  - ANSI CSI: `\x1b\[[0-9;?]*[a-zA-Z]`
  - ANSI OSC: `\x1b\].*?\x07`
  - Cursor-position fragments (e.g. `\x1b[2K`)
  - Common tmux status-bar fragments
  - Leading/trailing whitespace
- **Telemetry:** log `sanitize_delta_bytes` on the `uiadapter.translate` slog line (§6.1).
- **Risks:** aggressive regex eats fenced code — AC-1.3 is the guard.
- **Dependencies:** `regexp` stdlib only.
- **Use-repo-code directive:** inspect `executor.go` callsite and existing validator `contentPreserved` logic via `use-repo-code` (`summary.md` → grep `files.md`).

## Acceptance Criteria

AC-1.1: `TestSanitize_StripsANSI_Golden` passes with four scenarios: bare ANSI color (`\x1b[32m…\x1b[0m`), cursor-position (`\x1b[2K`), OSC-8 hyperlinks, pathological mixed input.

AC-1.2: end-to-end `TestAdapter_Translate_ANSIWrappedURL_NotUntrusted` — raw with ANSI-wrapped URL flows through Translate; `ast.Diagnostics.Untrusted == false`.

AC-1.3: `TestSanitize_PreservesFencedCodeBlocks` — triple-backtick blocks with brackets, braces, and backticks inside round-trip unchanged.

AC-1.4: `sanitize_delta_bytes` attribute present on structured log line.

## BDD Test Scenarios

```gherkin
Feature: SanitizeCapture

  Scenario: AC-1.1 — Strips ANSI variants
    Given raw input containing bare CSI color codes, a cursor-clear, an OSC-8 hyperlink, and pathological mixed escapes
    When SanitizeCapture runs
    Then the output matches the golden string byte-for-byte
    And deltaBytes equals len(raw) - len(sanitized)

  Scenario: AC-1.2 — ANSI-wrapped URL no longer triggers Untrusted
    Given raw with "\x1b[34mhttps://example.com\x1b[0m"
    When adapter.Translate processes the sanitized capture
    Then ast.Diagnostics.Untrusted is false
    And the URL appears in the content node

  Scenario: AC-1.3 — Fenced code blocks round-trip
    Given raw containing triple-backtick blocks with brackets, braces, and backticks inside
    When SanitizeCapture runs
    Then fenced sections are byte-identical to the input

  Scenario: AC-1.4 — Telemetry includes sanitize_delta_bytes
    Given any Translate call
    When the structured log line is captured
    Then the slog attribute `sanitize_delta_bytes` is present
```

## Tasks / Subtasks

- [ ] Task 1 — Implement `SanitizeCapture` (maps to AC-1.1, AC-1.3)
  - [ ] Subtask 1a — Compile regexes at package init.
  - [ ] Subtask 1b — Apply narrow replacements in order; preserve fenced code regions.
- [ ] Task 2 — Wire into executor (maps to AC-1.2)
- [ ] Task 3 — Tests (maps to AC-1.1, AC-1.3)
  - [ ] Subtask 3a — `TestSanitize_StripsANSI_Golden` — four scenarios.
  - [ ] Subtask 3b — `TestSanitize_PreservesFencedCodeBlocks`.
- [ ] Task 4 — End-to-end test `TestAdapter_Translate_ANSIWrappedURL_NotUntrusted` (maps to AC-1.2)
- [ ] Task 5 — Emit `sanitize_delta_bytes` on slog (maps to AC-1.4)

## Definition of Done

- [ ] All ACs verified with PASS evidence (test name)
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_

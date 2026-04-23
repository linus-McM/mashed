# uiadapter-v3-05: Two-stage Classify → Generate

**Status:** ready
**Domain:** backend
**Size:** L
**Depends On:** A, B, C
**Priority:** P0-critical

## Story

As a uiadapter author, I want a per-backend two-stage pipeline (`Classify` → `Generate`) with narrow per-kind schemas and a byte-stable static prefix, so that weaker models (gemma3:4b) produce reliable outputs, stronger models (Sonnet) can skip Stage 1, and prompt-prefix caching (Story 7) remains intact.

## Description

Plan §3 Phase 2 "Story 5 — Two-stage Classify → Generate" (lines 349–386). Replaces the monolithic prompt with a tiny classifier (~20 tokens) plus four per-kind narrow generators. Per-backend policy: Ollama (gemma3:4b) always two-stage; Claude Haiku two-stage by default; Claude Sonnet single-shot by default (router skips Classify and calls `GenerateSingleShot`). Stage-2 schemas are four strict subsets of the full UIAST — bad field names are rejected at decode time, not post-hoc. Slugification rule lives only in `generate_menu.md` (the only kind with `options.value`): lowercase hyphen-separated slug; literal `"1"`..`"5"` when the source uses numbered lists.

Byte budget tests: each `generate_*.md` ≤3 KiB; static prefix ≤4 KiB (§6.3).

Prompt-version discipline: bump `promptVersion` in `prompt.go` and update `TestPromptVersion` (§6.2 — v3 baseline is `"v3"`).

## Developer Notes

- **Files (new):**
  - `internal/uiadapter/stage_classify.go`
  - `internal/uiadapter/stage_generate.go`
  - `internal/uiadapter/prompts/classify.md`
  - `internal/uiadapter/prompts/generate_yn.md`
  - `internal/uiadapter/prompts/generate_menu.md`
  - `internal/uiadapter/prompts/generate_form.md`
  - `internal/uiadapter/prompts/generate_text.md`
- **Files (edited):** `internal/uiadapter/prompt.go` (add `promptVersion = "v3"` and `TestPromptVersion`).
- **Stage 1 schema:** `{ "type":"object", "properties":{"kind":{"enum":["yn","menu","form","text"]}}, "required":["kind"] }` — under `schemas/stage1.json` (codegen'd via Story A).
- **Prompt construction (byte-stable):**
  ```
  [static_prefix]    ← identity, safety, schema reminder, fixed few-shots
  [delimiter]        ← "\n\n---\nRAW CAPTURE:\n"
  [spotlighted_raw]  ← from Story 8
  [delimiter2]       ← "\n\n---\nKIND: "
  [kind_directive]   ← e.g. "menu\n"
  ```
- **Per-backend policy surface:** the router (Story 16) reads `Capabilities().SupportsSingleShot` to decide; Ollama returns false, Sonnet returns true, Haiku returns true with `RouterPolicy=haiku-single-shot` opt-in.
- **Risks:** §8 "Story 5 doubles Ollama call count" — mitigated by prefix cache (Story 7) + short Stage-2 output + narrow schemas.
- **Dependencies:** embedded prompt files via `//go:embed prompts/*.md`.
- **Use-repo-code directive:** the existing `internal/uiadapter/prompt.md` (story U3) is the baseline; use `use-repo-code` to locate it before lifting shared prelude text.

## Acceptance Criteria

AC-5.1: all golden fixtures from the existing `testdata/prompts/*-response.json` pass against the two-stage pipeline (`TestPrompt_Golden_Brainstorming`, `_Elicitation`, `_ProductBrief`, `_PartyMode`, `_PartyMode_CodeBlockPreserved`, `_Freeform`).

AC-5.2: Stage-1 classification accuracy ≥0.99 on the eval corpus.

AC-5.3: Stage-2 terminal-validation rate per kind ≤0.02 on Ollama; ≤0.005 on Claude.

AC-5.4: combined two-call p50 latency ≤1.2× single warm call on Ollama (classifier is small and shares the warm prefix).

AC-5.5: `TestPolicy_BackendDecidesStages` — measured network-call counts: Sonnet path = 1, Haiku path = 2, Ollama path = 2.

## BDD Test Scenarios

```gherkin
Feature: Two-stage Classify → Generate

  Scenario: AC-5.1 — Golden fixtures pass
    Given the six existing golden fixtures in testdata/prompts
    When the two-stage pipeline runs against each on the Ollama stub
    Then every AST matches the expected JSON byte-for-byte (modulo whitelisted fields)

  Scenario: AC-5.2 — Classifier accuracy
    Given the eval corpus with kind labels
    When Stage-1 runs on all 30 entries
    Then classification accuracy is ≥ 0.99

  Scenario: AC-5.3 — Terminal validation rates
    Given the eval corpus
    When Stage-2 runs on Ollama and Claude respectively
    Then terminal validation rate ≤ 0.02 on Ollama and ≤ 0.005 on Claude

  Scenario: AC-5.4 — Two-call latency budget
    Given a warm Ollama backend
    When a single two-stage Translate runs
    Then p50 combined latency is ≤ 1.2× a single warm Generate call

  Scenario: AC-5.5 — Policy decides call count
    Given stub backends capturing call counts
    When the pipeline runs once each against Sonnet, Haiku, Ollama
    Then Sonnet records 1 call, Haiku records 2, Ollama records 2
```

## Tasks / Subtasks

- [ ] Task 1 — Author prompt files (maps to AC-5.1, AC-5.3)
  - [ ] Subtask 1a — `classify.md` (≤1 KiB).
  - [ ] Subtask 1b — `generate_yn.md`, `generate_menu.md`, `generate_form.md`, `generate_text.md`, each ≤3 KiB.
  - [ ] Subtask 1c — Static prefix ≤4 KiB — assert in a test.
- [ ] Task 2 — Implement Stage 1 + Stage 2 dispatch (maps to AC-5.1, AC-5.5)
  - [ ] Subtask 2a — `stage_classify.go` returns `Kind`.
  - [ ] Subtask 2b — `stage_generate.go` picks per-kind prompt + schema.
- [ ] Task 3 — Per-backend policy wiring (maps to AC-5.5)
  - [ ] Subtask 3a — Router consults `Capabilities().SupportsSingleShot`.
  - [ ] Subtask 3b — `Haiku-single-shot` opt-in respected.
- [ ] Task 4 — Tests (maps to AC-5.1, AC-5.2, AC-5.4, AC-5.5)
  - [ ] Subtask 4a — `TestPrompt_Golden_*` variants.
  - [ ] Subtask 4b — `TestStage1_Accuracy`.
  - [ ] Subtask 4c — `TestStage2_Golden_PerKind`.
  - [ ] Subtask 4d — `TestPolicy_BackendDecidesStages`.
  - [ ] Subtask 4e — Byte-budget tests for prompt files.
- [ ] Task 5 — Bump `promptVersion` to "v3" and update `TestPromptVersion` (maps to AC-5.1)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_

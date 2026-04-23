# uiadapter-v3-06: Structured output dispatch

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** M
**Depends On:** A, C
**Priority:** P0-critical

## Story

As a uiadapter author, I want one schema source of truth rendered as three encodings (Ollama `format:<schema>`, Claude API `tool_use` `input_schema`, Claude CLI fenced-JSON prompt), so that bad field names are rejected structurally — not post-hoc in Go — and the three backends stay byte-equivalent on the schema.

## Description

Plan §3 Phase 2 "Story 6 — Structured output dispatch" (lines 388–408). Ollama wire change: `chatRequest.Format` becomes `json.RawMessage` (was `string`). A rollback flag `Config.LooseFormat=true` reverts to the literal string `"json"` for emergencies. Claude API leverages `tool_use` — the provider guarantees a shaped content block matching the `input_schema`; no intermediate JSON parsing. Claude CLI has the softest contract ("fenced JSON in the final assistant message") — that's why Story 13's parse-rate target for CLI is lower (≥0.95).

Supersedes stale story `uiadapter-02` (schema-constrained Ollama format) — narrower scope extended to all three backends.

## Developer Notes

- **Files (edited):**
  - `internal/uiadapter/backend/ollama/client.go` — change `chatRequest.Format` to `json.RawMessage`; respect `LooseFormat`.
  - `internal/uiadapter/backend/claudeapi/encode.go` (new) — builds `tools: [{name:"emit_uiast_<kind>", input_schema:<schema>}]` + `tool_choice: {type:"tool", name:...}`.
  - `internal/uiadapter/backend/claudecli/encode.go` (new) — system prompt asks for fenced JSON; parse + validate on output.
- **Schemas:** codegen'd by Story A. One canonical per-kind schema is re-used in all three paths.
- **Types / API surface:** per-backend `encode(kind Kind, schema json.RawMessage, raw string) ([]byte, error)` helpers.
- **Risks:** §8 "Ollama and Claude schemas drift" — AC-6.4 is the guard.
- **Dependencies:** `encoding/json` stdlib.
- **Use-repo-code directive:** use `use-repo-code` to read the current Ollama `chatRequest` struct before changing `Format`.

## Acceptance Criteria

AC-6.1: (Ollama) wire payload has `format.type == "object"` when `LooseFormat=false`.

AC-6.2: (Ollama) wire payload has `format == "json"` string when `LooseFormat=true`.

AC-6.3: (Claude API) wire payload has `tools[0].input_schema.type == "object"` and `tool_choice.name` matches the kind.

AC-6.4: `TestSchemas_IdenticalAcrossBackends` — the canonical per-kind schema is byte-equal whether served as `format`, as `input_schema`, or described in a CLI prompt.

AC-6.5: (Claude CLI) fenced-JSON parse + schema validation succeeds on ≥0.95 of the eval corpus.

AC-6.6: bad field names rejected structurally on Ollama (`format:`) and Claude API (`tool_use`), not post-hoc in Go.

## BDD Test Scenarios

```gherkin
Feature: Structured output dispatch

  Scenario: AC-6.1 — Ollama format object
    Given LooseFormat=false
    When the Ollama backend sends a Generate request
    Then the JSON body has `format.type == "object"`

  Scenario: AC-6.2 — Ollama loose fallback
    Given LooseFormat=true
    When the Ollama backend sends a Generate request
    Then the JSON body has `format == "json"`

  Scenario: AC-6.3 — Claude API tool_choice
    Given kind = "menu"
    When the Claude API backend sends a Generate request
    Then the JSON body has `tools[0].input_schema.type == "object"`
    And `tool_choice.name == "emit_uiast_menu"`

  Scenario: AC-6.4 — Schema byte-equivalence
    Given the canonical per-kind schema from Story A
    When it is rendered as Ollama format, Claude input_schema, and CLI prompt body
    Then the three renderings are byte-equal after JSON canonicalization

  Scenario: AC-6.5 — CLI parse-rate
    Given the eval corpus
    When the Claude CLI backend processes each entry
    Then fenced-JSON parse + schema validation succeeds on ≥ 0.95

  Scenario: AC-6.6 — Structural rejection
    Given a prompt that would otherwise emit `{ "kindX": ... }` (bad field)
    When Ollama `format:` or Claude `tool_use` decodes
    Then the provider rejects before Go sees the payload
```

## Tasks / Subtasks

- [ ] Task 1 — Ollama format dispatch (maps to AC-6.1, AC-6.2, AC-6.6)
  - [ ] Subtask 1a — `chatRequest.Format json.RawMessage`.
  - [ ] Subtask 1b — Honor `Config.LooseFormat`.
- [ ] Task 2 — Claude API encode (maps to AC-6.3, AC-6.6)
  - [ ] Subtask 2a — `tools[0]` construction per-kind.
  - [ ] Subtask 2b — `tool_choice.name` = `emit_uiast_<kind>`.
- [ ] Task 3 — Claude CLI encode (maps to AC-6.5)
  - [ ] Subtask 3a — System prompt template requests fenced JSON.
  - [ ] Subtask 3b — Parse + schema validate.
- [ ] Task 4 — Tests (maps to AC-6.1–6.5)
  - [ ] Subtask 4a — `TestClient_SendsJSONSchemaFormat` (Ollama).
  - [ ] Subtask 4b — `TestSchemas_IdenticalAcrossBackends`.

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_

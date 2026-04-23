# uiadapter-02: Switch `client.go` to schema-constrained decoding

**Status:** ready
**Domain:** backend
**Size:** M
**Depends On:** none
**Priority:** P0-critical

## Story

As a UIAdapter maintainer, I want `client.go` to send a JSON Schema object in the Ollama `format` field (schema-constrained decoding) instead of the loose `"json"` string, so that parse-rate rises from ~85% to near 100% and widget field-name drift becomes impossible at decode time rather than merely discouraged by the prompt.

## Description

Ollama ≥0.5 accepts a JSON Schema object as the `format` value. The model's decoder is constrained to JSON matching that schema, eliminating most validator churn. This story changes `chatRequest.Format` from `string` to `json.RawMessage`, adds a hand-written UIAST JSON Schema in a new `schema_json.go`, threads a `ClientConfig.LooseFormat bool` rollback flag, and updates affected tests. Runs in parallel with stories 1, 3, 4.

### Scope summary
- `chatRequest.Format` field type: `string` → `json.RawMessage`.
- New file `internal/uiadapter/schema_json.go` exporting `UIASTJSONSchema() json.RawMessage`.
- `Client.chat` marshals `UIASTJSONSchema()` into `Format` when `LooseFormat == false`; sends the literal JSON string `"json"` wrapped as `json.RawMessage` when `LooseFormat == true`.
- New field `ClientConfig.LooseFormat bool` (default `false`).
- Update `TestAdapter_SendsSystemPromptInRequest` to decode `Format` as a schema object.
- New test `TestClient_SendsJSONSchemaFormat` verifying wire payload shape.

### Non-goals
- Do not reflect the schema from Go types. Keep it hand-written (see "Why hand-written" below).
- Do not change `Schema json.RawMessage` inside `WidgetNode` or the widget union.
- Do not alter prompt.md (Story 1 owns that).
- Do not change validator or sanitize logic.

## Developer Notes

### Files to modify / create
- `internal/uiadapter/client.go` — change `chatRequest.Format string` → `chatRequest.Format json.RawMessage`; route `LooseFormat` in `(*Client).chat`.
- `internal/uiadapter/client_test.go` — update any assertion that compared `Format` to the string `"json"`; add `TestClient_SendsJSONSchemaFormat`.
- `internal/uiadapter/schema_json.go` (new) — exports `UIASTJSONSchema() json.RawMessage`.
- `internal/uiadapter/adapter_test.go` — update `TestAdapter_SendsSystemPromptInRequest` to match new `Format` shape.

### Type/symbol inventory (exact names)
- New: `UIASTJSONSchema() json.RawMessage` — package-level function in `schema_json.go`.
- New: `ClientConfig.LooseFormat bool` — default `false`; documented as rollback-only.
- Modified: `chatRequest` struct (package-private) — `Format json.RawMessage` with `json:"format,omitempty"`.

### Why hand-written schema
Plan §1 Story 2: Go types use `omitempty` and `Schema json.RawMessage` in `WidgetNode`. Reflection-based JSON Schema generation produces noisy drafts around those shapes. A hand-written schema is stable, reviewable, and matches UIAST v1 exactly.

The schema must cover:
- Envelope: `version`, `title`, `nodes` (array), `schema` (object|null). Adapter-stamped fields (`generated_by`, `generated_at`, `diagnostics`) must NOT be listed as required; keep the schema permissive for those or omit them entirely.
- Node union: each supported node type with its discriminator.
- Widget union: each supported widget kind (`text`, `choice`, `multi-choice`, `freeform`, etc. — enumerate from existing `schema.go`) with camelCase fields (`maxLength`, `multiline`, `copyable`, `repoRootRelative`).

### Rollback flag semantics
- `LooseFormat == false` (default): `Format` = `UIASTJSONSchema()`.
- `LooseFormat == true`: `Format` = `json.RawMessage(`"json"`)` (a JSON-encoded string). Matches pre-change behaviour byte-for-byte.

### Risks / gotchas
- Ollama builds older than 0.5 reject a schema object with a non-2xx. The `LooseFormat=true` flag is the documented rollback. PR description must state the minimum Ollama version and mention the flag for on-call.
- `chatRequest` is marshalled with `encoding/json`. `json.RawMessage` implements `Marshaler` — it serializes as its raw bytes. Make sure the raw bytes are valid JSON when `LooseFormat=true` (quoted string: `"json"`).
- Any other tests that decode the captured request and do `assert.Equal(req.Format, "json")` will break — grep for string comparisons under `internal/uiadapter/` and migrate them to `json.Unmarshal` of the raw message.
- The schema is embedded as a Go string literal (or `//go:embed schema.json`). Prefer `//go:embed` if a separate `schema.json` file is cleaner; otherwise a `var uiastSchema = json.RawMessage(`…`)` literal is fine. Either way, validate that the bytes parse with `json.Valid` at init time and fail fast if not.
- Keep schema permissive around `Schema json.RawMessage` (allow any JSON) to preserve forward-compat with new widget schemas.

### Telemetry preservation
- Do not remove or rename any existing attributes on the `uiadapter.translate` slog line (plan §3 "Telemetry: additive only").

### Reference files
- `internal/uiadapter/client.go` lines 24–35 (existing error types), 93–106 (current format handling).
- `internal/uiadapter/adapter.go` lines 136–148 (envelope stamping).
- `internal/uiadapter/schema.go` — canonical UIAST types the schema mirrors.
- `internal/uiadapter/adapter_test.go` — existing `TestAdapter_SendsSystemPromptInRequest`.
- `docs/mashed-ui-ast-schema.md` — UIAST v1 reference.
- `docs/plans/IMPLEMENTATION_PLAN.md` §1 Story 2.

## Acceptance Criteria

**AC-2.1: Wire-captured request uses a schema object when `LooseFormat=false`**
- Given a `Client` constructed with `ClientConfig{LooseFormat: false}` (default)
- When `Client.chat` dispatches a request to a stubbed transport
- Then `json.Unmarshal(req.Format, &m)` succeeds with `m["type"] == "object"`
- And the schema contains the UIAST envelope `properties` (at minimum `nodes`)

**AC-2.2: Wire-captured request uses the legacy string when `LooseFormat=true`**
- Given a `Client` constructed with `ClientConfig{LooseFormat: true}`
- When `Client.chat` dispatches a request
- Then `string(req.Format) == "\"json\""`
- And the rollback path is proven by a dedicated test

**AC-2.3: No behaviour regression in existing tests**
- Given all tests under `internal/uiadapter/`
- When `go test ./internal/uiadapter/... -race -short` runs
- Then `client_test.go`, `adapter_test.go`, `prompt_test.go` all pass
- And any format-string assertions have been migrated to schema-shape assertions

**AC-2.4: Eval baseline captured for parse-rate delta**
- Given Story 5's eval scorecard is available (gated on story 5 completion)
- When the scorecard is run once with `LooseFormat=true` and once with `LooseFormat=false`
- Then both runs are recorded in the PR description with the parse-rate delta
- And `LooseFormat=false` shows parse-rate ≥ 0.95
- Note: AC-2.4 is deferred to PR review once Story 5 lands; this story only needs to make the delta measurable.

## BDD Test Scenarios

### Scenario 1: Schema-object format on the wire

```gherkin
Feature: Schema-constrained decoding

  Scenario: Default config sends a schema object
    Given a Client with default ClientConfig
    And a stubbed HTTP transport capturing the request body
    When Client.chat is invoked with any prompt
    Then the captured request.format is a JSON object
    And the object has property "type" equal to "object"
    And the object has a "properties" key containing "nodes"

  Scenario: LooseFormat flag re-enables the legacy string
    Given a Client with ClientConfig.LooseFormat = true
    And a stubbed HTTP transport capturing the request body
    When Client.chat is invoked with any prompt
    Then the captured request.format is the JSON-encoded string "json"
```

### Scenario 2: UIASTJSONSchema shape

```gherkin
Feature: Hand-written JSON Schema for UIAST envelope

  Scenario: Schema parses as valid JSON
    Given the package-level function UIASTJSONSchema
    When the returned bytes are passed to json.Valid
    Then json.Valid returns true

  Scenario: Schema covers the UIAST envelope
    Given the parsed schema object
    Then the schema has type "object"
    And properties includes "version", "title", "nodes"
    And adapter-stamped fields generated_by, generated_at, diagnostics are not required
```

### Scenario 3: No existing-test regressions

```gherkin
Feature: Backward compatibility of the stubbed-Ollama path

  Scenario: Existing adapter prompt test still passes after format migration
    Given the updated TestAdapter_SendsSystemPromptInRequest
    When the test asserts messages[0].content equals the embedded prompt
    Then the assertion passes
    And the test also asserts that request.format unmarshals to a schema object
```

## Tasks / Subtasks

- [ ] Task 1: Create `schema_json.go` with the hand-written UIAST schema (AC: 2.1)
  - [ ] Implement `UIASTJSONSchema() json.RawMessage`
  - [ ] Add an `init()` check using `json.Valid` to fail fast on typos
  - [ ] Cover envelope + node union + widget union with camelCase field names
- [ ] Task 2: Migrate `chatRequest.Format` to `json.RawMessage` (AC: 2.1, 2.2, 2.3)
  - [ ] Change field type and re-run `go build`
  - [ ] Update `(*Client).chat` to select schema-object vs `"json"` string based on `ClientConfig.LooseFormat`
  - [ ] Add `ClientConfig.LooseFormat bool` with godoc explaining rollback-only usage
- [ ] Task 3: Add `TestClient_SendsJSONSchemaFormat` (AC: 2.1, 2.2)
  - [ ] Stub the HTTP transport, capture the request body
  - [ ] Assert `format.type == "object"` when `LooseFormat=false`
  - [ ] Assert `format == "json"` when `LooseFormat=true`
- [ ] Task 4: Update `TestAdapter_SendsSystemPromptInRequest` (AC: 2.3)
  - [ ] Migrate the format assertion from string comparison to schema-shape check
  - [ ] Ensure the `messages[0].content == prompt` assertion is unchanged
- [ ] Task 5: Pre-flight and handoff (AC: all)
  - [ ] `go build ./...`, `go vet ./...`
  - [ ] `go test ./internal/uiadapter/... -race -short`
  - [ ] `/simplify` on `schema_json.go` and `client.go`
  - [ ] Document the rollback flag in the PR description and reference the minimum Ollama version

## Definition of Done

- [ ] All acceptance criteria pass (AC-2.4 captured in PR description once Story 5 lands)
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `client.go` and `schema_json.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./internal/uiadapter/... -race -short` passes
- [ ] `/simplify` run on all modified code
- [ ] Pre-flight: default path uses schema object; `LooseFormat=true` matches legacy byte-for-byte
- [ ] AC Validation Table complete

### AC Validation Table (fill in PR description)

| AC | Test | Status |
|----|------|--------|
| AC-2.1 | `TestClient_SendsJSONSchemaFormat` (default path) | |
| AC-2.2 | `TestClient_SendsJSONSchemaFormat` (`LooseFormat=true`) | |
| AC-2.3 | `go test ./internal/uiadapter/... -race -short` | |
| AC-2.4 | Parse-rate delta recorded in PR description (gated on Story 5) | |

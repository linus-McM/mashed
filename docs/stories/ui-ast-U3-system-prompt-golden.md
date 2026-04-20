# ui-ast-U3: Embedded system prompt + 5 golden-path fixture tests

**Status:** ready
**Domain:** backend
**Size:** M
**Depends on:** ui-ast-U2
**Priority:** P1-high

## Story

As the UI AST adapter owner, I want a production-quality embedded system prompt (`prompt.md`) and five golden-path fixtures driving the validator end-to-end via a mock HTTP server, so that the adapter's hot path is exercised as a whole — not piecewise — and future prompt edits cannot silently regress the five canonical interaction shapes (brainstorming, elicitation, product-brief, party-mode, freeform chat).

## Description

Implements spec §4.6 (system prompt outline) and §4.2's `prompt_test.go` placeholder. Ships `prompt.md` embedded via `//go:embed` plus a `promptVersion = "v1"` constant used by the telemetry line (§4.7.7). Drives U2's `defaultAdapter.Translate` end-to-end with a stubbed HTTP server that returns hand-authored "Gemma output" JSON for five representative Claude turns. Each fixture asserts the adapter produces a non-fallback AST that passes the validator and has the expected widget-shape mix. This is the first story where the adapter runs as an integrated whole — every earlier test mocks pieces.

### Scope summary

- `internal/uiadapter/prompt.go` — embed `prompt.md` via `//go:embed`; export `SystemPrompt() string` + `promptVersion`.
- `internal/uiadapter/prompt.md` (new) — rendered per spec §4.6 outline: role, schema declaration, 5 golden examples, output contract, safety rules.
- `internal/uiadapter/prompt_test.go` — 5 golden-path fixtures driving `defaultAdapter.Translate` via `httptest.Server`.
- `internal/uiadapter/testdata/prompts/` — the five raw Claude captures + five "Gemma response" JSON files (bidirectional fixtures).
- Extend `defaultAdapter.Translate` to wire `SystemPrompt()` into the U1 `Client.Chat` call (U2 shipped an empty system message).

### Non-goals

- No live Ollama calls (U9 owns the eval harness).
- No per-process system prompt variants (spec §11 Q5 — not a gating requirement for v1).
- No prompt-injection adversarial corpus (that's a future security review outside this story).
- No telemetry slog wiring beyond passing `promptVersion` to the log line (U4 wires the actual log sink).

## Developer Notes

### Files to create/modify

- `internal/uiadapter/prompt.go` (new):
  ```go
  package uiadapter

  import _ "embed"

  //go:embed prompt.md
  var embeddedSystemPrompt string

  // promptVersion pins the telemetry "prompt_version" field (§4.7.7).
  // Bump on any prompt.md edit that changes behaviour.
  const promptVersion = "v1"

  // SystemPrompt returns the embedded adapter system prompt.
  func SystemPrompt() string { return embeddedSystemPrompt }
  ```
- `internal/uiadapter/prompt.md` (new) — per §4.6 outline:
  1. Role paragraph.
  2. UIAST v1 schema — compact but complete.
  3. Five golden examples (brainstorming, elicitation, product-brief, party-mode, freeform).
  4. Output contract: "Return only a JSON object matching the schema. No prose, no code fences."
  5. Safety rules (never fabricate options, never rewrite user-facing prose, never emit file paths/URLs/commands the source didn't contain, fallback to a single markdown node on any rule violation).
- `internal/uiadapter/adapter.go` — change the `a.client.Chat(ctx, a.model, "", raw)` call to `a.client.Chat(ctx, a.model, SystemPrompt(), raw)`.
- `internal/uiadapter/prompt_test.go` — five end-to-end tests with `httptest.Server` per fixture.
- `internal/uiadapter/testdata/prompts/`:
  - `brainstorming-raw.txt` + `brainstorming-response.json`
  - `elicitation-raw.txt` + `elicitation-response.json`
  - `product-brief-raw.txt` + `product-brief-response.json`
  - `party-mode-raw.txt` + `party-mode-response.json`
  - `freeform-raw.txt` + `freeform-response.json`

### Golden fixture contract

Each fixture pair exercises a distinct widget mix:

| Fixture | Raw-capture shape | Expected AST feature |
|---|---|---|
| `brainstorming` | Method picker (5 methods listed) | One `decision_group` with `choice` widget, 5 options |
| `elicitation` | 5 elicitation methods + "or r/a/x" | One `decision_group` with `multi` widget (1–5 numeric options) |
| `product-brief` | 3 stage questions + stage heading | Three `decision_group` nodes, mixed `choice`/`free` widgets |
| `party-mode` | Freeform "continue?" prompt with code block | `markdown` node preserving the code block, one `approval` widget |
| `freeform` | Single open-ended question | No `decision_group`; `fallback_answer_shape: "free"` |

The stubbed `httptest.Server` returns exactly `{"message":{"role":"assistant","content":<fixture.response>}}` so the fixture is what would come back from a perfect Gemma call. The validator runs as-is.

### Prompt-test skeleton

```go
func TestPrompt_Golden_Brainstorming(t *testing.T) {
    raw := readFixture(t, "brainstorming-raw.txt")
    response := readFixture(t, "brainstorming-response.json")
    srv := newStubbedOllama(t, response)
    defer srv.Close()

    adapter := NewDefault(Config{Enabled: true, Model: "gemma3:4b", TimeoutMs: 3000}, slog.Default())
    // test-only host override (see U1's ollamaHost var-rather-than-const note).
    withHost(t, srv.URL)

    ast := adapter.Translate(context.Background(), raw, "bmad-brainstorming")

    require.NotNil(t, ast)
    require.Equal(t, "1", ast.Version)
    require.False(t, strings.HasPrefix(ast.GeneratedBy, "fallback:"),
        "expected non-fallback, got %s", ast.GeneratedBy)
    require.Len(t, decisionGroups(ast), 1)
    require.Equal(t, "choice", decisionGroups(ast)[0].Widget.Type)
    require.Len(t, decisionGroups(ast)[0].Widget.Options, 5)
}
```

Same pattern for the other four, asserting the "Expected AST feature" column above.

### `promptVersion` telemetry hook

U4 wires the full slog line; U3 establishes the constant and passes it in the client call metadata (future-proof). Add a `TestPromptVersion_IsV1` that simply asserts `promptVersion == "v1"` so any prompt edit forces a conscious version bump via code review.

### Risks / gotchas

- **Prompt bloat.** `prompt.md` must stay under ~1500 tokens or it eats the input budget. Keep schema examples tight. Target: 2500 chars.
- **Embed path resolution.** `//go:embed prompt.md` requires `prompt.md` to live next to `prompt.go`. It does NOT support `testdata/` glob from the same directive.
- **Fixture staleness.** Any schema change in U2 breaks fixtures. Part of U3's Definition of Done is re-running the fixtures and documenting the failing/regenerated set in the commit message.
- **Mock server URL injection.** The U1 `ollamaHost` var-not-const is the only way to point `Client` at a test server. Codify via a test-only `withHost(t, url)` helper guarded by `t.Cleanup` to restore.
- **Prompt-injection example in fixtures.** DO NOT include adversarial payloads in U3 fixtures — those belong to U9's eval corpus. U3 fixtures are happy-path only; error paths are covered by U2.

### Reference files

- `internal/uiadapter/adapter.go` (U2) — `Translate` call site to mutate.
- `internal/uiadapter/client.go` (U1) — `Chat` signature.
- `docs/mashed-ui-ast-schema.md` §4.6 (system-prompt outline), §4.5 (model drift detection via `prompt_version`).

## Acceptance Criteria

**AC-1: `SystemPrompt()` returns the embedded `prompt.md` content**
- Given the compiled `internal/uiadapter` package
- When `SystemPrompt()` is invoked
- Then the returned string contains all five section headings from §4.6 (role, schema, examples, output contract, safety rules)
- And the string length is between 1500 and 6000 bytes (readability + token-budget guard)
- Verified by `TestSystemPrompt_ContainsAllSections`

**AC-2: `promptVersion == "v1"`**
- Given the `promptVersion` constant
- When tested
- Then it equals the literal `"v1"`
- Verified by `TestPromptVersion_IsV1`

**AC-3: Brainstorming golden — choice widget with 5 options**
- Given the `brainstorming-raw.txt` + `brainstorming-response.json` fixtures
- When `Translate` runs against a stubbed Ollama returning the fixture response
- Then the returned AST has exactly one `decision_group` with `widget.type == "choice"` and 5 options
- And the AST is NOT a fallback (`GeneratedBy` does not start with `fallback:`)
- Verified by `TestPrompt_Golden_Brainstorming`

**AC-4: Elicitation golden — multi widget**
- Given the `elicitation` fixtures
- When `Translate` runs
- Then the AST contains one `decision_group` with `widget.type == "multi"` and 5 numeric options
- Verified by `TestPrompt_Golden_Elicitation`

**AC-5: Product-brief golden — three decision_groups**
- Given the `product-brief` fixtures
- When `Translate` runs
- Then the AST contains three `decision_group` nodes
- And their widget types include both `choice` and `free`
- And the collapse rule (spec §3.4) would flag `Diagnostics.Collapsed` only at the frontend — validator leaves `Collapsed = false`
- Verified by `TestPrompt_Golden_ProductBrief`

**AC-6: Party-mode golden — code block preserved, approval widget**
- Given the `party-mode` fixtures (raw contains a fenced code block)
- When `Translate` runs
- Then the AST contains a `markdown` node whose content includes the verbatim fenced code block
- And the `contentPreserved` check passes → `Diagnostics.Untrusted == false`
- And the AST contains one `decision_group` with `widget.type == "approval"`
- Verified by `TestPrompt_Golden_PartyMode` + `TestPrompt_Golden_PartyMode_CodeBlockPreserved`

**AC-7: Freeform golden — no decision_group, fallback_answer_shape "free"**
- Given the `freeform` fixtures
- When `Translate` runs
- Then the AST has zero `decision_group` nodes
- And `FallbackAnswerShape == "free"`
- Verified by `TestPrompt_Golden_Freeform`

**AC-8: Adapter sends the embedded prompt in the HTTP body**
- Given a recording `httptest.Server` and the adapter configured with a non-empty prompt
- When `Translate` is called
- Then the decoded request body's `messages[0].role == "system"` and `messages[0].content == SystemPrompt()`
- Verified by `TestAdapter_SendsSystemPromptInRequest`

## BDD Test Scenarios

```gherkin
Feature: Embedded adapter system prompt + golden fixtures

  Scenario: SystemPrompt() returns all required sections
    Given the compiled uiadapter package
    When SystemPrompt() is invoked
    Then the returned string contains "Role", "Schema", "Examples", "Output contract", and "Safety"

  Scenario: promptVersion is v1
    Given the compiled uiadapter package
    When the promptVersion constant is read
    Then it equals "v1"

  Scenario: Brainstorming fixture produces a 5-option choice widget
    Given the brainstorming raw capture and stubbed Gemma response
    When Translate is called
    Then the AST contains one decision_group with choice widget and 5 options
    And GeneratedBy does not start with "fallback:"

  Scenario: Product-brief fixture produces three decision groups
    Given the product-brief raw capture and stubbed Gemma response
    When Translate is called
    Then the AST contains three decision_group nodes
    And the widget types include both choice and free

  Scenario: Party-mode fixture preserves fenced code block
    Given a raw capture containing a fenced code block
    And a Gemma response that copies the code block into a markdown node
    When Translate is called
    Then the AST markdown node contains the code block verbatim
    And Diagnostics.Untrusted is false

  Scenario: Freeform fixture has no decision_group
    Given a raw single-question capture
    And a Gemma response with no decision_group nodes
    When Translate is called
    Then the AST has zero decision_group nodes
    And FallbackAnswerShape equals "free"

  Scenario: Adapter sends the embedded system prompt in the HTTP body
    Given a recording test server and adapter pointed at it
    When Translate is called
    Then the request messages[0] has role system and content equal to SystemPrompt()
```

## Tasks / Subtasks

- [ ] Task 1: Draft `prompt.md` per §4.6 (AC-1, AC-2)
  - [ ] Write role + schema + 5 examples + output contract + safety rules
  - [ ] RED: `TestSystemPrompt_ContainsAllSections`, `TestPromptVersion_IsV1`
  - [ ] GREEN: wire `//go:embed` in `prompt.go`, declare `promptVersion`
- [ ] Task 2: Wire `SystemPrompt()` into `Translate` (AC-8)
  - [ ] RED: `TestAdapter_SendsSystemPromptInRequest` against a recording server
  - [ ] GREEN: swap the empty-string system in `Translate` for `SystemPrompt()`
  - [ ] REFACTOR: `/simplify`
- [ ] Task 3: Draft the five fixture pairs (AC-3…AC-7)
  - [ ] Write `brainstorming-raw.txt` + `brainstorming-response.json`
  - [ ] Write the other four pairs
- [ ] Task 4: End-to-end golden tests (AC-3…AC-7)
  - [ ] RED: one test per fixture hitting `Translate` via stubbed server
  - [ ] GREEN: iterate fixture JSON until each test passes
  - [ ] Include the party-mode content-preservation sub-test
- [ ] Task 5: Test-only host override helper
  - [ ] Add `withHost(t, url)` in an `internal/uiadapter/export_test.go` or `_test.go` helper
  - [ ] Use `t.Cleanup` to restore `ollamaHost`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `prompt.go` and the adapter code paths the tests reach
- [ ] `go build ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `go vet ./...` passes
- [ ] `/simplify` run on all modified Go files; no CRITICAL/HIGH findings
- [ ] `prompt.md` < 6000 bytes (token-budget guard)
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch

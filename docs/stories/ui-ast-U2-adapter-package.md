# ui-ast-U2: `internal/uiadapter` package — Adapter, schema, validator, semaphore, mock

**Status:** done
**Domain:** backend
**Size:** L
**Depends on:** ui-ast-U1
**Priority:** P0-critical
**Landed:** 2026-04-21 (commit bd96df3)

## Story

As the UI AST translation owner, I want the full `internal/uiadapter` Go package — `Adapter` interface, `defaultAdapter` struct, `UIAST` schema types, deterministic validator per §3.3, bounded-concurrency semaphore, and a `//go:build testing` `MockAdapter` — so that U4 can inject an adapter into the executor and every §3.3 validator rule is enforced at the trust boundary before any downstream consumer sees a node.

## Description

Implements spec §4.2 (package layout), §4.3 (Adapter contract — never returns nil, no error return), §4.7 (reliability layers 1/2/3/4/5/6/7), §3.3 (deterministic validator rules 1–8), and the §7.1/§7.2/§7.3 trust boundaries. Wires the U1 `Client` for HTTP and ships a `MockAdapter` gated behind `//go:build testing` so tests can inject a fixed AST. This is the single biggest story in the backlog — do not combine with the U3 system prompt or U4 executor wiring. The adapter has zero production callers after U2 lands; U4 is the call-site.

### Scope summary

- `internal/uiadapter/adapter.go` — `Adapter` interface + `defaultAdapter` struct + `NewDefault(cfg, logger) Adapter`.
- `internal/uiadapter/schema.go` — `UIAST`, `UINode`, `WidgetNode`, `Diagnostics` types with `encoding/json` tags per §3.1 / §3.2.
- `internal/uiadapter/validator.go` — post-unmarshal pass enforcing §3.3 rules 1–8 deterministically in order.
- `internal/uiadapter/semaphore.go` — bounded in-flight gate (§4.7.5), default capacity 1.
- `internal/uiadapter/mock.go` — `//go:build testing` `MockAdapter` returning a fixed AST.
- `internal/uiadapter/adapter_test.go` — table-driven coverage of every §3.3 rule.
- `internal/uiadapter/fallback.go` — `FallbackAST(raw, reason string) *UIAST` helper (spec §4.8 literal construction).
- No system prompt yet (U3 owns `prompt.go`).

### Non-goals

- No `prompt.go` + embedded system prompt (U3).
- No golden-set eval harness (U9).
- No `PendingPrompt.Structured` wiring (U4 — §5.2).
- No settings UI (U5).
- No frontend (U6/U7/U8).

## Developer Notes

### Files to create

Per spec §4.2, verbatim:

```
internal/uiadapter/
    adapter.go       // Adapter interface + defaultAdapter implementing it
    schema.go        // UIAST types
    validator.go     // §3.3 rules 1–8
    semaphore.go     // bounded in-flight gate
    fallback.go      // FallbackAST helper (§4.8 literal)
    mock.go          // //go:build testing
    adapter_test.go  // table-driven
    validator_test.go // one test case per §3.3 rule
```

`client.go` already exists from U1. Do not duplicate.

### `Adapter` contract (spec §4.3, lines 268–284)

```go
// Adapter translates a raw Claude-turn capture into a Mashed UI AST tree.
type Adapter interface {
    // Translate never returns nil. On ANY failure the returned AST has
    // GeneratedBy="fallback:<reason>". ctx.Err() surfaces via
    // returned.Diagnostics.CancelReason when non-nil.
    Translate(ctx context.Context, raw, procID string) *UIAST
}
```

No error return (§4.3 line 283 — "Dropping the error simplifies every call site"). Every failure mode maps to a fallback AST:

| Failure | `GeneratedBy` |
|---|---|
| Ollama unreachable | `fallback:unreachable` |
| HTTP timeout (adapter context) | `fallback:timeout` |
| HTTP status != 200 | `fallback:server:<code>` |
| JSON unmarshal failure | `fallback:validation:malformed` |
| §3.3 rule (6) drops a `required=true` group | `fallback:validation:required_dropped` |
| §3.3 rule (8) serialised size overflow | `fallback:validation:oversize` |
| Adapter disabled (`UIAdapterEnabled=false` OR `OllamaEnabled=false`) | `fallback:disabled` |
| Semaphore blocked + ctx cancel | `fallback:saturated` |
| Context canceled mid-call | `fallback:canceled` (also sets `Diagnostics.CancelReason`) |

### `defaultAdapter` skeleton (§4.3 lines 287–302)

```go
type defaultAdapter struct {
    client  *Client        // from U1
    model   string
    timeout time.Duration
    sem     chan struct{}  // bounded in-flight, cap 1 by default (§4.7.5)
    logger  *slog.Logger
}

type Config struct {
    Enabled      bool
    Model        string
    TimeoutMs    int
    MaxInflight  int  // default 1
}

func NewDefault(cfg Config, logger *slog.Logger) Adapter {
    if !cfg.Enabled { return disabledAdapter{} } // always-fallback
    if cfg.MaxInflight <= 0 { cfg.MaxInflight = 1 }
    return &defaultAdapter{
        client:  NewClient(ClientConfig{TimeoutMs: cfg.TimeoutMs}),
        model:   cfg.Model,
        timeout: time.Duration(cfg.TimeoutMs) * time.Millisecond,
        sem:     make(chan struct{}, cfg.MaxInflight),
        logger:  logger,
    }
}
```

### `Translate` flow (spec §4.1 + §4.7)

```go
func (a *defaultAdapter) Translate(ctx context.Context, raw, procID string) *UIAST {
    select {
    case a.sem <- struct{}{}:
        defer func() { <-a.sem }()
    case <-ctx.Done():
        return FallbackAST(raw, "saturated")
    }

    ctx, cancel := context.WithTimeout(ctx, a.timeout)
    defer cancel()

    // empty system prompt for U2 — U3 will embed via //go:embed.
    body, err := a.client.Chat(ctx, a.model, "", raw)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) { return FallbackAST(raw, "timeout") }
        if errors.Is(err, context.Canceled)         { ast := FallbackAST(raw, "canceled"); ast.Diagnostics.CancelReason = ctx.Err().Error(); return ast }
        if errors.Is(err, ErrOllamaUnreachable)     { return FallbackAST(raw, "unreachable") }
        return FallbackAST(raw, "server:"+httpStatusOr(err))
    }

    ast := &UIAST{}
    if err := json.Unmarshal([]byte(body), ast); err != nil {
        return FallbackAST(raw, "validation:malformed")
    }
    if reasons := Validate(ast, raw); len(reasons) > 0 {
        // rule-6/7 required-drop is inside Validate — it returns a SENTINEL
        // reason "required_dropped" that Validate's caller converts to full
        // fallback. Other reasons are appended to Diagnostics and the AST
        // is kept.
        if containsRequiredDrop(reasons) || containsOversize(reasons) {
            return FallbackAST(raw, "validation:"+firstTerminal(reasons))
        }
        ast.Diagnostics.FallbackReasons = append(ast.Diagnostics.FallbackReasons, reasons...)
    }

    // §4.7.4 content preservation — only URLs + code blocks.
    if !contentPreserved(raw, ast) {
        ast.Diagnostics.Untrusted = true
    }

    return ast
}
```

### Schema types (spec §3.1 + §3.2)

```go
// schema.go
type UIAST struct {
    Version              string     `json:"version"`
    GeneratedBy          string     `json:"generated_by"`
    GeneratedAt          int64      `json:"generated_at"`
    TurnSummary          string     `json:"turn_summary"`
    Nodes                []UINode   `json:"nodes"`
    FallbackAnswerShape  string     `json:"fallback_answer_shape"`
    Diagnostics          Diagnostics `json:"diagnostics"`
}

type Diagnostics struct {
    InputBytes       int      `json:"input_bytes,omitempty"`
    OutputBytes      int      `json:"output_bytes,omitempty"`
    LatencyMs        int      `json:"latency_ms,omitempty"`
    FallbackReasons  []string `json:"fallback_reasons,omitempty"`
    Untrusted        bool     `json:"untrusted,omitempty"`
    Collapsed        bool     `json:"collapsed,omitempty"`
    CancelReason     string   `json:"-"`  // internal only — NOT serialised
}

type UINode struct {
    Type        string        `json:"type"`
    // shared/aliased fields — only the ones relevant to the node's Type are populated
    Content     string        `json:"content,omitempty"`
    Tone        string        `json:"tone,omitempty"`       // hint
    Heading     string        `json:"heading,omitempty"`    // summary, table, decision_group
    Bullets     []string      `json:"bullets,omitempty"`    // summary
    Lang        string        `json:"lang,omitempty"`       // code
    Copyable    bool          `json:"copyable,omitempty"`   // code
    Columns     []string      `json:"columns,omitempty"`    // table
    Rows        [][]string    `json:"rows,omitempty"`       // table
    Prompt      string        `json:"prompt,omitempty"`     // decision_group
    Help        string        `json:"help,omitempty"`       // decision_group
    Required    bool          `json:"required,omitempty"`   // decision_group
    Widget      *WidgetNode   `json:"widget,omitempty"`     // decision_group
    ResponseKey string        `json:"response_key,omitempty"` // decision_group
}

type WidgetNode struct {
    Type        string         `json:"type"`
    Options     []WidgetOption `json:"options,omitempty"`
    Default     string         `json:"default,omitempty"`
    Min         int            `json:"min,omitempty"`
    Max         int            `json:"max,omitempty"`
    YesLabel    string         `json:"yes_label,omitempty"`
    NoLabel     string         `json:"no_label,omitempty"`
    Placeholder string         `json:"placeholder,omitempty"`
    MaxLength   int            `json:"maxLength,omitempty"`
    Multiline   bool           `json:"multiline,omitempty"`
    Accept      []string       `json:"accept,omitempty"`
    RepoRootRel bool           `json:"repoRootRelative,omitempty"`
    Schema      json.RawMessage `json:"schema,omitempty"` // JSON Schema subset
}

type WidgetOption struct {
    Value string `json:"value"`
    Label string `json:"label,omitempty"`
}
```

### Validator (§3.3 rules 1–8, declaration order)

Dedicated test per rule. Exact names:

| Rule | Test |
|---|---|
| (1) Unknown node type rewrites to markdown | `TestValidate_Rule1_UnknownTypeBecomesMarkdown` |
| (2) Empty options on choice/multi drops enclosing decision_group | `TestValidate_Rule2_EmptyOptionsDropsGroup` |
| (3) decision_group with no widget is discarded | `TestValidate_Rule3_NoWidgetDropsGroup` |
| (4) Duplicate response_key suffixed `-2`, `-3`, … | `TestValidate_Rule4_DuplicateKeySuffix` |
| (5) response_key > 64 chars truncated | `TestValidate_Rule5_LongKeyTruncated` |
| (6) decision_group > 8: keep first 8; required-drop → full fallback | `TestValidate_Rule6_CountCapRequiredDropFallback` + `TestValidate_Rule6_CountCapOptionalDrop` |
| (7) Total nodes > 32: same required-drop guard | `TestValidate_Rule7_NodeCapRequiredDropFallback` |
| (8) Serialized size > 6 KiB → full fallback | `TestValidate_Rule8_OversizeFullFallback` |

`Validate(ast, raw) []string` returns a list of reasons. Terminal reasons (`required_dropped`, `oversize`) cause the caller (`Translate`) to substitute the full fallback AST. Non-terminal reasons accumulate in `ast.Diagnostics.FallbackReasons`.

### Security ACs (§7.1, §7.2, §7.3) — first-class in this story

| Rule | Test |
|---|---|
| §7.1 `ShapeFile` still routes through `resolveFileInput` at submission time (adapter cannot bypass) | `TestValidate_Security71_FileWidgetDoesNotBypassShapeFile` (asserts that the validator does not strip `repoRootRelative`, does not transform path strings, and the emitted widget type is `"file"` — the submission-time path-traversal rejection test lives in `internal/bmad/validate_test.go` already) |
| §7.2 URL preservation: missing URL sets `untrusted` | `TestContentPreservation_URLDropped_SetsUntrusted` |
| §7.2 Code-block preservation: missing fenced block sets `untrusted` | `TestContentPreservation_CodeBlockDropped_SetsUntrusted` |
| §7.2 Numbered list conversion does NOT flag untrusted | `TestContentPreservation_NumberedListToChoice_NoUntrusted` |
| §7.3 Network pinning (inherits from U1 but re-assert here) | `TestAdapter_NetworkPinned_NoNonLocalhostReach` (grep uiadapter/*.go for `http://` — only match is localhost:11434) |

### `//go:build testing` mock (§4.2)

```go
//go:build testing

package uiadapter

type MockAdapter struct { Fixed *UIAST }
func NewMock(fixed *UIAST) Adapter { return &MockAdapter{Fixed: fixed} }
func (m *MockAdapter) Translate(_ context.Context, raw, _ string) *UIAST {
    if m.Fixed == nil { return FallbackAST(raw, "mock:nil") }
    return m.Fixed
}
```

The `testing` build tag isolates the mock from production binaries. Executor tests in U4 must build with `-tags testing` to link it.

### Risks / gotchas

- **Validator order matters.** Spec §3.3 line 181: "Validator walks `nodes` in declaration order. Rules (applied in this exact sequence):" — the test harness MUST drive one fixture per rule and assert the exact `FallbackReasons` string ordering.
- **`Diagnostics.CancelReason` is not serialised.** JSON tag `"-"` per spec §4.3 (in-memory signalling only for U4's round-loop).
- **Serialized size check at rule (8).** Must happen AFTER rules (1)–(7) mutate. Use `json.Marshal(ast); len(b) > 6*1024`.
- **`contentPreserved` is URL + code block only.** Per §4.7.4 / §7.2: "Numbered-list entries are not checked — turning '1. stdout / 2. file / 3. both' into a choice widget is the feature's main job and never flags untrusted." RFC-3986 URL regex + fenced-code-block regex (```` ``` ... ``` ````).
- **`disabledAdapter`** is a separate tiny struct implementing `Adapter` — returns `FallbackAST(raw, "disabled")` always. Cleaner than a nil-check on every call-site.
- **Semaphore capacity 0 would deadlock.** `NewDefault` bumps `MaxInflight <= 0` to 1.
- **`ResponseKey` uniqueness** applies after rule-(1) rewrites could change node types; the dedupe walk iterates ONLY `decision_group` nodes (markdown-ified nodes have no `response_key`).
- **Fallback AST fields.** `FallbackAST` must exactly match spec §4.8 lines 378–389 — version "1", `turn_summary = firstLine(raw, 120)`, a single markdown node with the raw content, `fallback_answer_shape = "free"`. Put this as its own file (`fallback.go`) so it can be imported by U4's snapshot-migration regression test without pulling the full adapter.

### Reference files

- `internal/uiadapter/client.go` (U1, existing).
- `docs/mashed-ui-ast-schema.md` §3 (schema), §4 (adapter), §7 (security).
- `internal/bmad/validate.go:13` — existing `validateInput` path that `ShapeFile` routes through.

## Acceptance Criteria

**AC-1: `Adapter.Translate` never returns nil and always has `Version == "1"`**
- Given any invocation path (happy, unreachable, timeout, malformed JSON, canceled, saturated, disabled)
- When `Translate` is called
- Then the returned `*UIAST` is non-nil and `ast.Version == "1"`
- Verified by `TestAdapter_Translate_NeverNil` (table of every failure path)

**AC-2: `GeneratedBy` fallback reasons map 1:1 to the spec §4.8 table**
- Given each row in the §4.8 failure table
- When the corresponding failure is triggered (unreachable / timeout / HTTP 500 / malformed / canceled / saturated / disabled / required-dropped / oversize)
- Then the returned `ast.GeneratedBy` equals `"fallback:<reason>"` with the exact reason string
- Verified by `TestAdapter_FallbackReasonTable`

**AC-3: Validator §3.3 rules (1)–(8) each have a passing dedicated test**
- Given fixtures crafted per spec §3.3 rules (1)–(8)
- When `Validate(ast, raw)` is called
- Then the listed test names (see Developer Notes table) each pass
- And rule-ordering determinism is asserted: an input that triggers rules 1, 4, and 5 emits reasons in exactly `["unknown_type", "dup_key", "key_truncated"]` order
- Verified by `TestValidate_Rule{1..8}_*` and `TestValidate_RuleOrdering_Deterministic`

**AC-4: Content-preservation check flags untrusted only for URL/code-block drops (§4.7.4 / §7.2)**
- Given a raw capture containing `https://example.com/foo` and the adapter drops it from the AST
- When `contentPreserved(raw, ast)` runs
- Then it returns `false` and the caller sets `ast.Diagnostics.Untrusted = true`
- And given a raw capture with numbered options `1. stdout 2. file 3. both` converted into a `choice` widget with all three values
- When the check runs
- Then it returns `true` and `Untrusted` stays `false`
- Verified by `TestContentPreservation_URLDropped_SetsUntrusted`, `TestContentPreservation_CodeBlockDropped_SetsUntrusted`, `TestContentPreservation_NumberedListToChoice_NoUntrusted`

**AC-5: Bounded concurrency — semaphore prevents parallel fan-out (§4.7.5)**
- Given `defaultAdapter` with `MaxInflight: 1` and a slow test server (500ms response)
- When two `Translate` calls are issued concurrently with a 100ms context on the second
- Then the second call returns `fallback:saturated` via the ctx.Done select branch
- Verified by `TestAdapter_Semaphore_BlockedCallSaturates`

**AC-6: Security §7.1 — file widget validator preserves `repoRootRelative` and `accept`**
- Given an AST with a `decision_group{widget.type:"file", widget.repoRootRelative:true, widget.accept:[".md"]}`
- When `Validate` runs
- Then the widget fields survive the validator pass unchanged
- And no validator rule silently strips or rewrites file-widget metadata
- Verified by `TestValidate_Security71_FileWidgetPreserved`

**AC-7: Security §7.3 — network pinning enforced at the package boundary**
- Given a grep of `internal/uiadapter/*.go` for `http://`
- When results are collected
- Then the only match is the literal `"http://localhost:11434"` in `client.go`
- Verified by `TestAdapter_NetworkPinned_NoNonLocalhostReach`

**AC-8: `MockAdapter` compiles only under `-tags testing` and returns the fixed AST**
- Given `//go:build testing` mock compiled via `go test -tags testing ./internal/uiadapter/...`
- When a caller invokes `NewMock(ast)` followed by `Translate(ctx, "", "")`
- Then the returned AST is the same object passed to `NewMock`
- And `go build ./...` (without `-tags testing`) does not link `MockAdapter`
- Verified by `TestMockAdapter_FixedReturn` + `TestMockAdapter_NotInProductionBuild`

**AC-9: §11 Q6 resolved — explicit opt-out produces `fallback:disabled`; unreachable Ollama does NOT auto-disable (user decision 2026-04-21)**
- Given `UIAdapterEnabled: false` or `OllamaEnabled: false` passed to `NewDefault`
- When `Translate` is invoked
- Then the returned AST has `GeneratedBy == "fallback:disabled"` and a single markdown node with the raw input
- And no HTTP call is made (verified via a stub client that panics on Call)
- And when `UIAdapterEnabled: true` + `OllamaEnabled: true` but the Ollama socket is closed, `Translate` returns `GeneratedBy == "fallback:unreachable"` (NOT `fallback:disabled`) — the user-decision default TRUE means an unreachable Ollama degrades to raw markdown per §4.8, never silently auto-disables
- And `adapter_test.go` carries a header comment: `// NOTE: §11 Q6 resolved 2026-04-21 — adapter enabled by default; unreachable Ollama degrades to fallback:unreachable, not fallback:disabled`
- Verified by `TestAdapter_DisabledProducesFallback`, `TestAdapter_DisabledNoHTTPCall`, `TestAdapter_UnreachableNotAutoDisabled`

## BDD Test Scenarios

```gherkin
Feature: UI AST adapter package

  Scenario: Translate returns a non-nil AST for every failure path
    Given an adapter configured with a dead socket
    When Translate is called
    Then the returned AST is non-nil with Version "1" and GeneratedBy "fallback:unreachable"

  Scenario: Validator rewrites unknown node types to markdown (rule 1)
    Given a UIAST with a node whose type is "snarkfish"
    When Validate runs
    Then the node is rewritten to type "markdown" with content preserved
    And the AST is retained, not fallback-replaced

  Scenario: Validator drops decision_group with empty choice options (rule 2)
    Given a decision_group whose widget is choice with options []
    When Validate runs
    Then the decision_group is discarded
    And "empty_options" is in Diagnostics.FallbackReasons

  Scenario: Duplicate response_key is suffixed -2 (rule 4)
    Given two decision_groups both with response_key "confirm"
    When Validate runs
    Then the first retains "confirm" and the second becomes "confirm-2"
    And "dup_key" is in Diagnostics.FallbackReasons

  Scenario: Decision_group cap with required drop triggers full fallback (rule 6)
    Given 10 decision_groups where group 9 has required true
    When Validate runs
    Then the caller returns FallbackAST with reason "validation:required_dropped"

  Scenario: Serialised size over 6 KiB forces full fallback (rule 8)
    Given a valid UIAST whose JSON serialises to 7 KiB
    When Validate runs
    Then the caller returns FallbackAST with reason "validation:oversize"

  Scenario: Content preservation flags URL drop as untrusted
    Given raw text containing https://example.com/foo
    And an AST whose nodes do not mention that URL
    When the content check runs
    Then Diagnostics.Untrusted is true

  Scenario: Numbered-list-to-choice conversion does not flag untrusted
    Given raw text "1. stdout 2. file 3. both"
    And an AST with a choice widget containing all three values
    When the content check runs
    Then Diagnostics.Untrusted stays false

  Scenario: Semaphore saturation returns fallback:saturated
    Given an adapter with MaxInflight 1 and a blocking in-flight call
    When a second Translate is issued with a short context
    Then the second returns GeneratedBy "fallback:saturated"

  Scenario: File widget metadata survives the validator pass
    Given a decision_group whose widget is file with repoRootRelative true
    When Validate runs
    Then the widget retains repoRootRelative true and accept list intact

  Scenario: Grep confirms localhost-only network
    Given the internal/uiadapter package source files
    When grepped for http://
    Then the only match is "http://localhost:11434" in client.go

  Scenario: Mock adapter returns the fixture under -tags testing
    Given MockAdapter constructed with a fixed AST
    When Translate is invoked under go test -tags testing
    Then the returned pointer equals the fixed AST
```

## Tasks / Subtasks

- [ ] Task 1: Schema types + `FallbackAST` helper (AC-1)
  - [ ] RED: `TestFallbackAST_MatchesSpec48` asserting the exact field values from spec §4.8
  - [ ] GREEN: write `schema.go`, `fallback.go`
  - [ ] REFACTOR: `/simplify`
- [ ] Task 2: Validator with §3.3 rules 1–8 (AC-3, AC-4, AC-6)
  - [ ] RED: 10 failing tests — one per rule + ordering determinism + file-widget preservation
  - [ ] GREEN: implement `Validate(ast, raw)` driving rules 1–8 in order
  - [ ] Implement `contentPreserved(raw, ast)` with URL (RFC-3986) + fenced code block regex
  - [ ] REFACTOR: `/simplify` on `validator.go`
- [ ] Task 3: `defaultAdapter` + `Translate` + semaphore (AC-1, AC-2, AC-5)
  - [ ] RED: `TestAdapter_Translate_NeverNil`, `TestAdapter_FallbackReasonTable`, `TestAdapter_Semaphore_BlockedCallSaturates`
  - [ ] GREEN: implement `defaultAdapter`, `Translate`, `disabledAdapter`, `semaphore.go`
  - [ ] REFACTOR: `/simplify`
- [ ] Task 4: `MockAdapter` under `//go:build testing` (AC-8)
  - [ ] RED: `TestMockAdapter_FixedReturn` behind `-tags testing`
  - [ ] RED: `TestMockAdapter_NotInProductionBuild` — `go build ./...` without tag passes; grep for `MockAdapter` in the built binary returns nothing
  - [ ] GREEN: write `mock.go`
- [ ] Task 5: Security network-pin assertion (AC-7)
  - [ ] RED: `TestAdapter_NetworkPinned_NoNonLocalhostReach` greps package sources
  - [ ] GREEN: already satisfied by U1 — test codifies
- [ ] Task 6: §11 Q6 resolved — disabled-vs-unreachable semantics (AC-9)
  - [ ] RED: `TestAdapter_DisabledProducesFallback`, `TestAdapter_DisabledNoHTTPCall`, `TestAdapter_UnreachableNotAutoDisabled`
  - [ ] GREEN: `disabledAdapter` returns `FallbackAST(raw, "disabled")`; `defaultAdapter` returns `FallbackAST(raw, "unreachable")` when the socket is closed — does NOT auto-flip config
  - [ ] Document the 2026-04-21 Q6 decision in `CLAUDE.md` under "UI AST Adapter": default `UIAdapterEnabled=true`, `OllamaEnabled=true`; unreachable Ollama → `fallback:unreachable`; explicit opt-out → `fallback:disabled`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on every new file in `internal/uiadapter/`
- [ ] `go build ./...` passes (no `-tags testing`)
- [ ] `go test -tags testing ./internal/uiadapter/... -race` passes
- [ ] `go vet ./...` passes
- [ ] `/simplify` run on every new Go file; no CRITICAL/HIGH findings
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch

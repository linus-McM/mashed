# Story breadcrumbs-10: Ollama HTTP client (DEFERRED)

**Priority:** P3-low
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** breadcrumbs-09
**Status:** deferred

> DEFERRED — Phase 3 per plan lines 202, 212-217. Requires explicit human decision to move to `ready`.

## Description

Phase 3 Task 3.1 (plan lines 163-167). Create `internal/ollama/client.go` — a small HTTP client against `localhost:11434`. Expose `IsOllamaAvailable()` (installed + running + model-selected) and `AskOllama(prompt, schema)` where the JSON schema constrains response shape. Config-driven: model name from `mashedConfig.OllamaModel`, timeout default 2s for routing calls, respects `OllamaEnabled`.

## Developer Notes

### Architecture

New package `internal/ollama/client.go` (plan line 165). Exports:

```go
type Client struct {
    httpClient *http.Client
    baseURL    string
    model      string
    timeout    time.Duration
    enabled    bool
}

func NewClient(cfg mashedConfig) *Client
func (c *Client) IsAvailable(ctx context.Context) bool
func (c *Client) Ask(ctx context.Context, prompt string, schema json.RawMessage) (json.RawMessage, error)
```

`Ask` POSTs to `/api/chat` (or `/api/generate` with `format` field for JSON mode). Use Ollama's structured outputs to constrain the response to the schema.

### Technical Considerations

- **Short timeout** (2s default). Routing decisions must be fast — long LLM latency blocks the UI.
- **Zero-retry**: one shot. If it fails, caller logs + falls through to deterministic behavior.
- **No streaming** (plan line 208 — "no streaming"). Buffered full JSON response only.
- **Error handling**: project pattern — sentinels (`ErrOllamaDisabled`, `ErrOllamaUnavailable`, `ErrOllamaInvalidJSON`), wrapping with `%w`, `errors.Is`/`As` friendly.

### Wails binding requirements

Expose `IsOllamaAvailable()` in `app.go` so the frontend can check before asking (plan line 167). The `AskOllama` method stays Go-internal for this story — Task 3.2 (breadcrumbs-11) is the caller.

### Risks & Edge Cases

- Model not yet loaded into Ollama memory: first call may take longer than the 2s timeout. Document as acceptable: deterministic fallback engages.
- Schema validation failure: return `ErrOllamaInvalidJSON` wrapping the parse error; caller logs raw response.
- Ollama returns non-JSON (plain text): same → `ErrOllamaInvalidJSON`.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 163-167.
- Skills: `/golang-testing`, `/golang-error-handling`.

## Acceptance Criteria

AC-1: IsAvailable composes installed + running + enabled
- Given Ollama enabled + running + a model selected
- When IsAvailable is called
- Then it returns true

AC-2: Ask returns schema-conformant JSON
- Given a valid prompt and schema
- When Ask is called with the local Ollama mock
- Then the returned bytes parse as the schema
- And no error is returned

AC-3: Timeout respected
- Given a slow Ollama server (mocked with 5s latency)
- When Ask runs with a 2s timeout
- Then the call returns `ErrOllamaUnavailable` within ~2s
- And the underlying error wraps `context.DeadlineExceeded`

AC-4: Disabled short-circuits
- Given `OllamaEnabled = false` in config
- When Ask is called
- Then it returns `ErrOllamaDisabled` without making an HTTP request

## BDD Test Scenarios

```gherkin
Feature: Ollama client

  Scenario: Happy path Ask
    Given a mock Ollama returning valid JSON matching schema S
    When Ask(prompt, S) runs
    Then it returns the parsed JSON with no error

  Scenario: Timeout
    Given a mock with 5s latency and a 2s client timeout
    When Ask runs
    Then the error wraps context.DeadlineExceeded
    And errors.Is(err, ErrOllamaUnavailable) is true

  Scenario: Disabled
    Given OllamaEnabled = false
    When Ask is called
    Then errors.Is(err, ErrOllamaDisabled) is true
    And no HTTP request is made

  Scenario: Invalid JSON
    Given a mock returning plain text "I cannot comply"
    When Ask runs
    Then errors.Is(err, ErrOllamaInvalidJSON) is true
```

## Tasks / Subtasks

- [ ] Task 1: Package skeleton (AC-1)
  - [ ] `internal/ollama/client.go`, `errors.go` (sentinels), `types.go` (OllamaModel, ChatRequest, ChatResponse)
- [ ] Task 2: Ask + schema (AC-2, AC-3)
  - [ ] POST `/api/chat` with `format` field
  - [ ] Response validation against schema
- [ ] Task 3: IsAvailable (AC-1, AC-4)
  - [ ] Check enabled flag + HTTP `/api/version` probe
- [ ] Task 4: Wails binding IsOllamaAvailable (AC-1)
- [ ] Task 5: Table-driven tests with httptest.Server (AC-2, AC-3, AC-4)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues

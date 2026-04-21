# ui-ast-U1: Inline Ollama HTTP client + mashedConfig Ollama fields

**Status:** done
**Domain:** backend
**Size:** M
**Depends on:** none
**Priority:** P0-critical
**Landed:** 2026-04-21 (commit 5123856)

## Story

As the UI AST foundation owner, I want a thin, localhost-pinned Ollama HTTP client (with a `Chat` call + a `ListModels` discovery call) and four new `mashedConfig` fields (`OllamaEnabled`, `OllamaModel`, `UIAdapterEnabled`, `UIAdapterTimeoutMs`), so that U2's `defaultAdapter` has a well-defined HTTP call-site and U5's settings UI has both config targets and a dynamic model picker backed by whatever Ollama has pulled locally — without waiting for the deferred breadcrumbs-09/-10 stories.

## Description

Implements §4.4 config fields and the "thin Ollama HTTP wrapper" of §4.2. Ships a package-level singleton `http.Client` (per §4.7.6), a `POST /api/chat` wrapper that sets `format: "json"` and hardcodes `http://localhost:11434` (per §7.3), and the four new config fields with defaults captured in §4.4. No adapter, no schema types, no prompt — those land in U2. This story is pure infrastructure: the client compiles and is reachable from `internal/uiadapter` but has zero callers in production code until U2 wires it.

### Scope summary

- Create `internal/uiadapter/client.go` with a `Client` struct wrapping `*http.Client`, a `Chat(ctx, model, system, user string) (string, error)` method, and a `ListModels(ctx context.Context) ([]string, error)` method that calls `GET /api/tags` and returns the names of pulled models.
- Singleton package-level `http.Client` with `Timeout = configured + 500ms` (buffer so the in-adapter `context.WithTimeout` fires first — §4.7.3, §4.7.6).
- Hardcoded host `http://localhost:11434` — no config override (§7.3).
- Extend `mashedConfig` (in `app.go`) with `OllamaEnabled bool`, `OllamaModel string`, `UIAdapterEnabled bool`, `UIAdapterTimeoutMs int`. **Defaults changed by 2026-04-21 user decision:** `UIAdapterEnabled` default is **TRUE** (overrides spec §4.4's `false`); `OllamaEnabled` default is **TRUE**; `OllamaModel` default `"gemma3:4b"`; `UIAdapterTimeoutMs` default `3000`.
- Unit tests with a `httptest.Server` that asserts both `Chat` and `ListModels` request shapes, asserts `ListModels` parses Ollama's `{"models":[{"name":"gemma3:4b"},...]}` response, and verifies the client rejects any non-localhost override via a compile-time constant.

### Non-goals

- No `Adapter` interface, no `Translate` (U2).
- No schema / validator (U2).
- No system prompt (U3).
- No settings UI (U5).
- No per-call retry or circuit breaker (spec §10 non-goal "no auto-retry").

## Developer Notes

### Files to create/modify

- `internal/uiadapter/` (new directory) — `client.go` only in U1.
- `internal/uiadapter/client.go` (new) — `Client`, `NewClient(cfg)`, `Chat`, the singleton `http.Client`.
- `internal/uiadapter/client_test.go` (new) — `httptest.Server` test driving the happy path + error paths + localhost pinning.
- `app.go` — extend `mashedConfig` struct with the four new fields; update `loadConfig` / `saveConfig` / default-construction.
- `app.go` — no new Wails bindings in U1 (bindings land in U5).
- `app_test.go` or a new `app_config_test.go` — round-trip test for the four new fields.

### `Client` shape (§4.2 / §4.7.6)

```go
// internal/uiadapter/client.go
package uiadapter

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"
)

// ollamaHost is hardcoded per §7.3 — the client never talks to any other host.
const ollamaHost = "http://localhost:11434"

// sharedTransport + sharedClient are package-level per §4.7.6.
var (
    sharedTransport = &http.Transport{ /* defaults; MaxIdleConns=2 */ }
    sharedClient    = &http.Client{Transport: sharedTransport}
)

type Client struct {
    timeout time.Duration
}

type Config struct {
    TimeoutMs int // adapter-side; client.Timeout = TimeoutMs + 500ms
}

func NewClient(cfg Config) *Client {
    return &Client{timeout: time.Duration(cfg.TimeoutMs+500) * time.Millisecond}
}

// Chat POSTs /api/chat with format:"json" per §4.1. Returns the raw JSON
// body the model produced (the "message.content" field). Does NOT parse
// into a UIAST — U2's adapter does that.
func (c *Client) Chat(ctx context.Context, model, system, user string) (string, error) {
    // body := {"model":..., "format":"json", "stream":false,
    //         "messages":[{"role":"system","content":system},{"role":"user","content":user}]}
    // Apply c.timeout via context; return assistant message content or error.
}

// ListModels GETs /api/tags and returns the names of locally-pulled models.
// Used by U5's Settings dropdown (dynamic model picker) and by app startup
// to verify the configured OllamaModel is actually pulled.
// The /api/tags response shape is {"models":[{"name":"gemma3:4b","modified_at":"...",...}, ...]}
// — ListModels returns just the name strings, sorted lexicographically.
// Returns ErrOllamaUnreachable on transport failure (same sentinel as Chat).
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
    // GET http://localhost:11434/api/tags
    // Decode { Models []struct{ Name string } }
    // Return sorted names; on transport error wrap ErrOllamaUnreachable.
}

var ErrOllamaUnreachable = errors.New("uiadapter: ollama unreachable")
```

### Localhost-pinning guard

Add a compile-time constant plus a runtime assertion:

```go
// The request URL is always ollamaHost + "/api/chat" — concatenated in Chat,
// never accepting an override parameter. Any future attempt to parameterise
// the host must pass a code-review gate citing §7.3.
```

The unit test asserts: passing a fake host via environment variables / ENV is ignored; the request always goes to `localhost:11434`.

### `mashedConfig` diff

In `app.go` near line where `mashedConfig` is declared:

```go
type mashedConfig struct {
    // existing fields…
    OllamaEnabled      bool   `json:"ollamaEnabled"`                // default TRUE (user decision 2026-04-21)
    OllamaModel        string `json:"ollamaModel,omitempty"`        // default "gemma3:4b"
    UIAdapterEnabled   bool   `json:"uiAdapterEnabled"`             // default TRUE (user decision 2026-04-21, overrides spec §4.4 default-false)
    UIAdapterTimeoutMs int    `json:"uiAdapterTimeoutMs,omitempty"` // default 3000 (§4.4)
}
```

Defaults are applied in `loadConfig` when the key is absent. Because the 2026-04-21 user decision flipped `UIAdapterEnabled` and `OllamaEnabled` to default **TRUE**, the zero-value + omitempty trick no longer works for them — a missing key must be treated as `true`, not `false`. Implementation: drop `omitempty` on those two JSON tags (keep the keys always present) AND explicitly fill them in `loadConfig` when the raw bytes had no such key (detect via a `*bool` staging struct or a dedicated `json.RawMessage` presence check). `OllamaModel` fills via `if cfg.OllamaModel == "" { cfg.OllamaModel = "gemma3:4b" }`; `UIAdapterTimeoutMs` fills via `if cfg.UIAdapterTimeoutMs == 0 { cfg.UIAdapterTimeoutMs = 3000 }`. Write a pre-U1 legacy-config fixture test (`TestMashedConfig_LegacyLoad_DefaultsApplied`) that asserts a config written before U1 loads with `UIAdapterEnabled == true`, `OllamaEnabled == true`, `OllamaModel == "gemma3:4b"`, `UIAdapterTimeoutMs == 3000`.

### Risks / gotchas

- **Singleton `http.Client` leaks across tests.** Tests that use `httptest.Server` must point the test through an injected URL. Compromise: keep `ollamaHost` as a `var` (not `const`) but document in a comment that it is configurable only via tests through a `_test.go`-only setter. Production code never writes it.
- **Timeout buffer.** `http.Client.Timeout = adapterTimeout + 500ms`. The adapter's `context.WithTimeout` must fire first so the fallback path (§4.8 `fallback:timeout`) attributes cause correctly. Otherwise we'd see `net/http: Client.Timeout exceeded` errors bubble up as `fallback:server`.
- **JSON-format request body.** `format: "json"` lives at the top level of the `/api/chat` body (alongside `model`, `messages`, `stream`). Easy to misplace — test the request-body shape with an `httptest.Server` that decodes and asserts.
- **`OllamaModel` default.** Spec §4.5 locks default at `gemma3:4b`. Do not read a different default from environment variables — it must be constant.
- **`loadConfig` migration.** Existing configs on disk have none of these keys; `json.Unmarshal` into the extended struct leaves them at zero values. The explicit default-fill at load-time is required for `OllamaModel` + `UIAdapterTimeoutMs`. Write a regression test that loads a pre-U1 config fixture and asserts the defaults take effect.

### Reference files

- `app.go` — `mashedConfig` struct, `loadConfig`, `saveConfig`.
- `docs/mashed-ui-ast-schema.md` §4.4 (config), §4.7.6 (singleton client), §7.3 (network pinning), §8 (latency budget).

## Acceptance Criteria

**AC-1: `Client.Chat` POSTs the correct body shape to localhost**
- Given an `httptest.Server` stood up with a request recorder and the `Client` pointed at its URL (via test-only setter)
- When `Chat(ctx, "gemma3:4b", "system text", "user text")` is invoked
- Then the request body is JSON-decodable with `format == "json"`, `model == "gemma3:4b"`, `stream == false`, and `messages` containing both system and user roles in order
- And the `Content-Type` header is `application/json`
- Verified by `TestClient_Chat_RequestBodyShape`

**AC-2: Production host is pinned to localhost:11434 (§7.3)**
- Given the `ollamaHost` package variable
- When the caller attempts to change it via any exported API
- Then no such exported API exists (compile-time guarantee)
- And a grep of `internal/uiadapter/*.go` for `http://` returns only the literal `"http://localhost:11434"`
- Verified by `TestClient_LocalhostPinned` (greps the package source file) + package structure

**AC-3: Timeout behaviour — context deadline beats `http.Client.Timeout` by 500ms**
- Given a `Client` with `TimeoutMs: 100` and a test server that delays 300ms before responding
- When `Chat` is called with a 100ms context deadline
- Then the call returns a context-deadline error, not a client-timeout error
- And the error wraps `context.DeadlineExceeded`
- Verified by `TestClient_Chat_ContextDeadlineBeatsClientTimeout`

**AC-4: `mashedConfig` round-trips all four new fields**
- Given a `mashedConfig` with `OllamaEnabled: true`, `OllamaModel: "gemma3:4b"`, `UIAdapterEnabled: true`, `UIAdapterTimeoutMs: 3000`
- When written to disk via `saveConfig` and re-read via `loadConfig`
- Then all four fields round-trip with the expected values
- And a pre-U1 fixture config with none of the keys loads with the **user-decision defaults**: `UIAdapterEnabled == true`, `OllamaEnabled == true`, `OllamaModel == "gemma3:4b"`, `UIAdapterTimeoutMs == 3000`
- And a config explicitly setting `uiAdapterEnabled: false` round-trips that value (opt-out honoured)
- Verified by `TestMashedConfig_UIAdapterFields_RoundTrip`, `TestMashedConfig_LegacyLoad_DefaultsApplied`, and `TestMashedConfig_ExplicitOptOut_Honored`

**AC-7: `Client.ListModels` parses `/api/tags` and returns sorted names**
- Given an `httptest.Server` that returns `{"models":[{"name":"qwen2.5:3b"},{"name":"gemma3:4b"},{"name":"llama3.2:3b"}]}`
- When `ListModels(ctx)` is called
- Then the returned slice equals `["gemma3:4b","llama3.2:3b","qwen2.5:3b"]` (sorted lexicographically)
- And the HTTP method is `GET`, path is `/api/tags`, and no body is sent
- Verified by `TestClient_ListModels_ParsesAndSorts`

**AC-8: `Client.ListModels` surfaces `ErrOllamaUnreachable` on transport failure**
- Given a `Client` pointed at a closed socket
- When `ListModels(ctx)` is called
- Then the returned error wraps `ErrOllamaUnreachable`
- And the returned slice is empty (not nil-vs-empty contract-sensitive — callers must tolerate both)
- Verified by `TestClient_ListModels_UnreachableError`

**AC-9: `Client.ListModels` returns an empty slice (not error) when Ollama has no models pulled**
- Given an `httptest.Server` that returns `{"models":[]}` with HTTP 200
- When `ListModels(ctx)` is called
- Then err is nil and the slice has zero length
- Verified by `TestClient_ListModels_EmptyOK`

**AC-5: `Chat` surfaces transport errors as `ErrOllamaUnreachable`**
- Given a `Client` pointed at a closed socket (server stopped immediately)
- When `Chat` is called
- Then the error wraps `ErrOllamaUnreachable`
- And `errors.Is(err, ErrOllamaUnreachable)` returns true
- Verified by `TestClient_Chat_UnreachableSurfaceError`

**AC-6: `Chat` parses the `message.content` field from a valid Ollama response**
- Given a test server that returns `{"message":{"role":"assistant","content":"{\"version\":\"1\"}"}}`
- When `Chat` is called
- Then the returned string equals `{"version":"1"}` (the raw content, unparsed)
- Verified by `TestClient_Chat_ExtractsMessageContent`

## BDD Test Scenarios

```gherkin
Feature: Ollama HTTP client for UI AST adapter

  Scenario: Chat posts the JSON-format body to localhost
    Given a recording httptest.Server and a Client pointed at it
    When Chat is called with model "gemma3:4b" and system/user messages
    Then the request body has format "json", stream false, and both messages in order
    And the response content field is returned verbatim

  Scenario: Context deadline is hit before HTTP client timeout
    Given a Client with TimeoutMs 100 and a server that delays 300ms
    When Chat is called with a 100ms context
    Then the error wraps context.DeadlineExceeded
    And the error is not "Client.Timeout exceeded"

  Scenario: Unreachable server surfaces ErrOllamaUnreachable
    Given a Client pointed at a closed socket
    When Chat is called
    Then errors.Is(err, ErrOllamaUnreachable) returns true

  Scenario: Host is pinned to localhost:11434
    Given the internal/uiadapter package sources
    When grepped for "http://"
    Then the only match is "http://localhost:11434"

  Scenario: New mashedConfig fields round-trip on disk
    Given a mashedConfig with OllamaEnabled, OllamaModel, UIAdapterEnabled, UIAdapterTimeoutMs populated
    When saveConfig then loadConfig runs
    Then all four values survive byte-for-byte

  Scenario: Legacy config loads with new-field defaults (user decision 2026-04-21)
    Given a config JSON fixture missing all four new keys
    When loadConfig runs
    Then OllamaModel equals "gemma3:4b"
    And UIAdapterTimeoutMs equals 3000
    And UIAdapterEnabled is true
    And OllamaEnabled is true

  Scenario: Explicit opt-out is preserved
    Given a config JSON fixture with "uiAdapterEnabled": false
    When loadConfig runs
    Then UIAdapterEnabled is false

  Scenario: ListModels parses /api/tags and returns sorted names
    Given a test server returning three models in reverse order
    When ListModels is called
    Then the returned slice is sorted lexicographically

  Scenario: ListModels returns empty slice when nothing pulled
    Given a test server returning {"models":[]}
    When ListModels is called
    Then err is nil and len(result) is 0

  Scenario: ListModels surfaces ErrOllamaUnreachable on transport failure
    Given a Client pointed at a closed socket
    When ListModels is called
    Then errors.Is(err, ErrOllamaUnreachable) is true
```

## Tasks / Subtasks

- [ ] Task 1: `Client` skeleton + localhost pin (AC-1, AC-2)
  - [ ] RED: `TestClient_Chat_RequestBodyShape` against a stubbed recording server
  - [ ] RED: `TestClient_LocalhostPinned` grep test
  - [ ] GREEN: implement `Client`, `NewClient`, `Chat`, `ollamaHost` constant
  - [ ] REFACTOR: `/simplify` run on `internal/uiadapter/client.go`
- [ ] Task 2: Context deadline + error semantics (AC-3, AC-5, AC-6)
  - [ ] RED: three failing tests (deadline, unreachable, extract content)
  - [ ] GREEN: wire `context.WithTimeout` + error wrap (`ErrOllamaUnreachable`) + response parsing
  - [ ] REFACTOR: `/simplify`
- [ ] Task 3: `mashedConfig` extension with TRUE defaults (AC-4)
  - [ ] RED: `TestMashedConfig_UIAdapterFields_RoundTrip`, `TestMashedConfig_LegacyLoad_DefaultsApplied`, `TestMashedConfig_ExplicitOptOut_Honored`
  - [ ] GREEN: add four fields, drop `omitempty` on `uiAdapterEnabled`/`ollamaEnabled`, presence-aware default-fill in `loadConfig` (e.g., two-pass unmarshal via `json.RawMessage`), no-op in `saveConfig`
  - [ ] REFACTOR: `/simplify` on `app.go`
- [ ] Task 4: `Client.ListModels` for dynamic Settings picker (AC-7, AC-8, AC-9)
  - [ ] RED: `TestClient_ListModels_ParsesAndSorts`, `TestClient_ListModels_UnreachableError`, `TestClient_ListModels_EmptyOK`
  - [ ] GREEN: implement `ListModels`; reuse singleton `http.Client`, same localhost-pin guarantee
  - [ ] REFACTOR: `/simplify`
- [ ] Task 5: Documentation cross-ref
  - [ ] Note the client location in `CLAUDE.md` under a new "UI AST Adapter" section (one line, one link)
  - [ ] Point `docs/mashed-ui-ast-schema.md` §4.2 / §4.4 at the concrete files (link substitution)
  - [ ] Record the 2026-04-21 user decision (`UIAdapterEnabled`/`OllamaEnabled` default TRUE, dynamic model list) in the spec's §11 (resolved open questions)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `internal/uiadapter/client.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on modified files; no CRITICAL/HIGH findings
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch

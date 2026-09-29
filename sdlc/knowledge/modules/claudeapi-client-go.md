---
type: Module
title: claudeapi/client.go
description: "Graphify community 85: internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/claudeapi/client_test.go, internal/uiadapter/backend/claudeapi/"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e3e3d9c28e77effe }
  - { id: stub, resource: internal/uiadapter/backend/claudeapi/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 0c95368a8fcd18d3 }
  - { id: stub, resource: internal/uiadapter/backend/claudecli/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 9d9685929a699a14 }
  - { id: stub, resource: internal/uiadapter/backend/ollama/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 82a69a451dac2ce8 }
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 34796db160eeb018 }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-04-23T11:09:52+10:00", digest: aa5a14fdab76f8f6 }
---

# Files
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`
- `internal/uiadapter/backend/claudeapi/stub.go`
- `internal/uiadapter/backend/claudecli/stub.go`
- `internal/uiadapter/backend/ollama/stub.go`
- `internal/uiadapter/backend/registry.go`
- `internal/uiadapter/backend/registry_test.go`

# Symbols
- Kind (internal/uiadapter/backend/backend.go:L17)
- claudeapi/client.go (internal/uiadapter/backend/claudeapi/client.go:L1)
- .Generate() (internal/uiadapter/backend/claudeapi/client.go:L114)
- .GenerateSingleShot() (internal/uiadapter/backend/claudeapi/client.go:L139)
- requestBody (internal/uiadapter/backend/claudeapi/client.go:L147)
- classifyRequest() (internal/uiadapter/backend/claudeapi/client.go:L158)
- generateRequest() (internal/uiadapter/backend/claudeapi/client.go:L174)
- .call() (internal/uiadapter/backend/claudeapi/client.go:L211)
- Client (internal/uiadapter/backend/claudeapi/client.go:L23)
- waitRetryAfter() (internal/uiadapter/backend/claudeapi/client.go:L270)
- sanitiseErrorBody() (internal/uiadapter/backend/claudeapi/client.go:L284)
- anthropicVersion() (internal/uiadapter/backend/claudeapi/client.go:L292)
- init() (internal/uiadapter/backend/claudeapi/client.go:L302)
- apiKey (internal/uiadapter/backend/claudeapi/client.go:L34)
- .String() (internal/uiadapter/backend/claudeapi/client.go:L36)
- NewClient() (internal/uiadapter/backend/claudeapi/client.go:L41)
- .Name() (internal/uiadapter/backend/claudeapi/client.go:L62)
- .WarmUp() (internal/uiadapter/backend/claudeapi/client.go:L80)
- .Health() (internal/uiadapter/backend/claudeapi/client.go:L89)
- .Classify() (internal/uiadapter/backend/claudeapi/client.go:L93)
- claudeapi/client_test.go (internal/uiadapter/backend/claudeapi/client_test.go:L1)
- TestClaudeAPI_PromptCacheMarkers() (internal/uiadapter/backend/claudeapi/client_test.go:L111)
- TestClaudeAPI_AnthropicVersionHeader() (internal/uiadapter/backend/claudeapi/client_test.go:L122)
- TestClaudeAPI_StopReasonNotToolUseErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L142)
- TestClaudeAPI_ToolUseRoundtrip() (internal/uiadapter/backend/claudeapi/client_test.go:L22)
- TestClaudeAPI_RetryAfter429() (internal/uiadapter/backend/claudeapi/client_test.go:L53)
- TestClaudeAPI_KeyNotLogged() (internal/uiadapter/backend/claudeapi/client_test.go:L89)
- TestClaudeAPI_WarmUpMissingKeyErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L99)
- claudeapi/stub.go (internal/uiadapter/backend/claudeapi/stub.go:L1)
- init() (internal/uiadapter/backend/claudeapi/stub.go:L13)
- claudecli/stub.go (internal/uiadapter/backend/claudecli/stub.go:L1)
- init() (internal/uiadapter/backend/claudecli/stub.go:L12)
- ollama/stub.go (internal/uiadapter/backend/ollama/stub.go:L1)
- init() (internal/uiadapter/backend/ollama/stub.go:L13)
- backend/registry.go (internal/uiadapter/backend/registry.go:L1)
- Constructor (internal/uiadapter/backend/registry.go:L13)
- Register() (internal/uiadapter/backend/registry.go:L24)
- From() (internal/uiadapter/backend/registry.go:L41)
- Available() (internal/uiadapter/backend/registry.go:L53)
- reset() (internal/uiadapter/backend/registry.go:L67)
- backend/registry_test.go (internal/uiadapter/backend/registry_test.go:L1)
- TestBackend_SingleShotUnsupportedIsSentinel() (internal/uiadapter/backend/registry_test.go:L119)
- TestBackend_Available() (internal/uiadapter/backend/registry_test.go:L132)
- TestBackend_FromUnknownName() (internal/uiadapter/backend/registry_test.go:L22)
- TestBackend_FromKnownName() (internal/uiadapter/backend/registry_test.go:L37)
- TestBackend_InterfaceStressConcurrent() (internal/uiadapter/backend/registry_test.go:L53)

# Depends on
- [Accountant](/modules/accountant.md)
- [Config](/modules/config.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [go_pkg_testing](/modules/go-pkg-testing.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [sessions.go](/modules/sessions-go.md)
- [StubBackend](/modules/stubbackend.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

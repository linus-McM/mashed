---
type: Module
title: Client
description: "Graphify community 85: internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/claudeapi/client_test.go"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-09-29T07:07:25Z", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-09-29T07:07:25Z", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-09-29T07:07:25Z", digest: e3e3d9c28e77effe }
---

# Files
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`

# Symbols
- Kind (internal/uiadapter/backend/backend.go:L17)
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
- apiKey (internal/uiadapter/backend/claudeapi/client.go:L34)
- .String() (internal/uiadapter/backend/claudeapi/client.go:L36)
- NewClient() (internal/uiadapter/backend/claudeapi/client.go:L41)
- .Name() (internal/uiadapter/backend/claudeapi/client.go:L62)
- .Classify() (internal/uiadapter/backend/claudeapi/client.go:L93)
- TestClaudeAPI_PromptCacheMarkers() (internal/uiadapter/backend/claudeapi/client_test.go:L111)
- TestClaudeAPI_AnthropicVersionHeader() (internal/uiadapter/backend/claudeapi/client_test.go:L122)
- TestClaudeAPI_StopReasonNotToolUseErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L142)
- TestClaudeAPI_ToolUseRoundtrip() (internal/uiadapter/backend/claudeapi/client_test.go:L22)
- TestClaudeAPI_RetryAfter429() (internal/uiadapter/backend/claudeapi/client_test.go:L53)
- TestClaudeAPI_KeyNotLogged() (internal/uiadapter/backend/claudeapi/client_test.go:L89)
- TestClaudeAPI_WarmUpMissingKeyErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L99)

# Depends on
- [Accountant](/modules/accountant.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [stages_test.go](/modules/stages-test-go.md)
- [StubBackend](/modules/stubbackend.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

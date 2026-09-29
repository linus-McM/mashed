---
type: Module
title: claudeapi/client.go
description: "Graphify community 85: internal/uiadapter/accountant.go, internal/uiadapter/accountant_test.go, internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapte"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: accountant, resource: internal/uiadapter/accountant.go, last_modified: "2026-04-23T11:29:34+10:00", digest: ae4389c89ff37e93 }
  - { id: accountant_test, resource: internal/uiadapter/accountant_test.go, last_modified: "2026-04-23T11:29:34+10:00", digest: 60c344b0cdff2ce8 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e3e3d9c28e77effe }
---

# Files
- `internal/uiadapter/accountant.go`
- `internal/uiadapter/accountant_test.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`

# Symbols
- accountant.go (internal/uiadapter/accountant.go:L1)
- .Record() (internal/uiadapter/accountant.go:L120)
- .Snapshot() (internal/uiadapter/accountant.go:L130)
- .expireLocked() (internal/uiadapter/accountant.go:L143)
- Pricing (internal/uiadapter/accountant.go:L19)
- PricingFor() (internal/uiadapter/accountant.go:L32)
- Usage (internal/uiadapter/accountant.go:L45)
- CostUSD() (internal/uiadapter/accountant.go:L54)
- Accountant (internal/uiadapter/accountant.go:L68)
- tokenTick (internal/uiadapter/accountant.go:L77)
- NewAccountant() (internal/uiadapter/accountant.go:L84)
- .CheckPrecall() (internal/uiadapter/accountant.go:L91)
- accountant_test.go (internal/uiadapter/accountant_test.go:L1)
- TestAccountant_CostMatchesBilling() (internal/uiadapter/accountant_test.go:L13)
- TestAccountant_TripsBeforeHard429() (internal/uiadapter/accountant_test.go:L27)
- TestAccountant_USDBudgetSoftLimit() (internal/uiadapter/accountant_test.go:L45)
- TestAccountant_Snapshot() (internal/uiadapter/accountant_test.go:L59)
- TestAccountant_SlidingWindow() (internal/uiadapter/accountant_test.go:L72)
- TestAccountant_UnknownModelIsFree() (internal/uiadapter/accountant_test.go:L90)
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
- .Classify() (internal/uiadapter/backend/claudeapi/client.go:L93)
- claudeapi/client_test.go (internal/uiadapter/backend/claudeapi/client_test.go:L1)
- TestClaudeAPI_PromptCacheMarkers() (internal/uiadapter/backend/claudeapi/client_test.go:L111)
- TestClaudeAPI_AnthropicVersionHeader() (internal/uiadapter/backend/claudeapi/client_test.go:L122)
- TestClaudeAPI_StopReasonNotToolUseErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L142)
- TestClaudeAPI_ToolUseRoundtrip() (internal/uiadapter/backend/claudeapi/client_test.go:L22)
- TestClaudeAPI_RetryAfter429() (internal/uiadapter/backend/claudeapi/client_test.go:L53)
- TestClaudeAPI_KeyNotLogged() (internal/uiadapter/backend/claudeapi/client_test.go:L89)
- TestClaudeAPI_WarmUpMissingKeyErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L99)

# Depends on
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [sessions.go](/modules/sessions-go.md)
- [StubBackend](/modules/stubbackend.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)

# Features
- no feature plan names these files

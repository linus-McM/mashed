---
type: Module
title: mashed/internal/uiadapter.UIAST
description: "Graphify community 85: app_uiadapter_claudecli.go, internal/bmad/executor_adapter_test.go, internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/bac"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: executor_adapter_test, resource: internal/bmad/executor_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f79102c9f400d28f }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e3e3d9c28e77effe }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: scorecard_test, resource: internal/uiadapter/eval/scorecard_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: c10507f5ef6c8f7b }
  - { id: fallback, resource: internal/uiadapter/fallback.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 187709266229767c }
  - { id: fallback_test, resource: internal/uiadapter/fallback_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 2fa70a931cc96a6b }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/bmad/executor_adapter_test.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/eval/scorecard_test.go`
- `internal/uiadapter/fallback.go`
- `internal/uiadapter/fallback_test.go`

# Symbols
- truncReason() (app_uiadapter_claudecli.go:L105)
- splitHeadTail() (app_uiadapter_claudecli.go:L118)
- dumpRawIfRequested() (app_uiadapter_claudecli.go:L129)
- capModel() (app_uiadapter_claudecli.go:L150)
- claudeCLIAdapter (app_uiadapter_claudecli.go:L21)
- .Translate() (app_uiadapter_claudecli.go:L39)
- delayAdapter (internal/bmad/executor_adapter_test.go:L34)
- .Translate() (internal/bmad/executor_adapter_test.go:L39)
- Kind (internal/uiadapter/backend/backend.go:L17)
- Capabilities (internal/uiadapter/backend/backend.go:L45)
- .Generate() (internal/uiadapter/backend/claudeapi/client.go:L114)
- .GenerateSingleShot() (internal/uiadapter/backend/claudeapi/client.go:L139)
- .call() (internal/uiadapter/backend/claudeapi/client.go:L211)
- Client (internal/uiadapter/backend/claudeapi/client.go:L23)
- waitRetryAfter() (internal/uiadapter/backend/claudeapi/client.go:L270)
- sanitiseErrorBody() (internal/uiadapter/backend/claudeapi/client.go:L284)
- anthropicVersion() (internal/uiadapter/backend/claudeapi/client.go:L292)
- apiKey (internal/uiadapter/backend/claudeapi/client.go:L34)
- .String() (internal/uiadapter/backend/claudeapi/client.go:L36)
- NewClient() (internal/uiadapter/backend/claudeapi/client.go:L41)
- .Name() (internal/uiadapter/backend/claudeapi/client.go:L62)
- .Capabilities() (internal/uiadapter/backend/claudeapi/client.go:L65)
- .WarmUp() (internal/uiadapter/backend/claudeapi/client.go:L80)
- .Health() (internal/uiadapter/backend/claudeapi/client.go:L89)
- TestClaudeAPI_AnthropicVersionHeader() (internal/uiadapter/backend/claudeapi/client_test.go:L122)
- TestClaudeAPI_StopReasonNotToolUseErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L142)
- TestClaudeAPI_ToolUseRoundtrip() (internal/uiadapter/backend/claudeapi/client_test.go:L22)
- TestClaudeAPI_RetryAfter429() (internal/uiadapter/backend/claudeapi/client_test.go:L53)
- TestClaudeAPI_KeyNotLogged() (internal/uiadapter/backend/claudeapi/client_test.go:L89)
- TestClaudeAPI_WarmUpMissingKeyErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L99)
- .GenerateSingleShot() (internal/uiadapter/backend/claudecli/client.go:L116)
- .TranslateWithFullPrompt() (internal/uiadapter/backend/claudecli/client.go:L132)
- .runRaw() (internal/uiadapter/backend/claudecli/client.go:L152)
- Client (internal/uiadapter/backend/claudecli/client.go:L21)
- extractFencedJSON() (internal/uiadapter/backend/claudecli/client.go:L326)
- .Name() (internal/uiadapter/backend/claudecli/client.go:L41)
- .Capabilities() (internal/uiadapter/backend/claudecli/client.go:L44)
- TestClaudeCLI_ParsingHandlesShapes() (internal/uiadapter/backend/claudecli/client_test.go:L27)
- slowBackend (internal/uiadapter/backend/lifecycle_test.go:L16)
- .WarmUp() (internal/uiadapter/backend/lifecycle_test.go:L22)
- StubBackend (internal/uiadapter/backend/stubs.go:L14)
- .Name() (internal/uiadapter/backend/stubs.go:L32)
- .Classify() (internal/uiadapter/backend/stubs.go:L34)
- .Generate() (internal/uiadapter/backend/stubs.go:L39)
- .GenerateSingleShot() (internal/uiadapter/backend/stubs.go:L44)
- .WarmUp() (internal/uiadapter/backend/stubs.go:L52)
- .Health() (internal/uiadapter/backend/stubs.go:L57)
- .Capabilities() (internal/uiadapter/backend/stubs.go:L62)
- .Calls() (internal/uiadapter/backend/stubs.go:L66)
- synthUIAST() (internal/uiadapter/backend/stubs.go:L77)
- sequentialMockAdapter (internal/uiadapter/eval/scorecard_test.go:L19)
- .Translate() (internal/uiadapter/eval/scorecard_test.go:L24)
- FallbackAST() (internal/uiadapter/fallback.go:L23)
- firstLine() (internal/uiadapter/fallback.go:L47)
- TestFallbackAST_LiteralShape() (internal/uiadapter/fallback_test.go:L14)
- TestFallbackAST_TurnSummaryTruncates() (internal/uiadapter/fallback_test.go:L38)
- TestFallbackAST_EmptyRawIsSafe() (internal/uiadapter/fallback_test.go:L51)

# Depends on
- [Accountant](/modules/accountant.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [prompt_test.go](/modules/prompt-test-go.md)
- [SanitizeCapture](/modules/sanitizecapture.md)

# Inferred
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

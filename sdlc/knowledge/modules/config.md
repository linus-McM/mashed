---
type: Module
title: Config
description: "Graphify community 85: app_uiadapter_claudecli.go, internal/uiadapter/adapter.go, internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/clau"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e3e3d9c28e77effe }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: config, resource: internal/uiadapter/config.go, last_modified: "2026-04-26T09:22:14+10:00", digest: d9832db7180bc48b }
  - { id: config_test, resource: internal/uiadapter/config_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: 748e88156e161a65 }
  - { id: contextguard, resource: internal/uiadapter/contextguard.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 417a5e0974dc194b }
  - { id: logging_test, resource: internal/uiadapter/logging_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 744e5721d44f353b }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/config.go`
- `internal/uiadapter/config_test.go`
- `internal/uiadapter/contextguard.go`
- `internal/uiadapter/logging_test.go`

# Symbols
- newClaudeCLIAdapter() (app_uiadapter_claudecli.go:L31)
- Config (internal/uiadapter/adapter.go:L32)
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
- NewClient() (internal/uiadapter/backend/claudecli/client.go:L32)
- TestClaudeCLI_VersionCheck() (internal/uiadapter/backend/claudecli/client_test.go:L15)
- TestClaudeCLI_Capabilities() (internal/uiadapter/backend/claudecli/client_test.go:L56)
- config.go (internal/uiadapter/config.go:L1)
- mergeWithDefaults() (internal/uiadapter/config.go:L61)
- configFieldSet() (internal/uiadapter/config.go:L98)
- config_test.go (internal/uiadapter/config_test.go:L1)
- TestConfig_PrivacyPatternsUsable() (internal/uiadapter/config_test.go:L113)
- TestDefaultConfig_Bootable_AllBackends() (internal/uiadapter/config_test.go:L20)
- TestDefaultConfig_ZeroValueBootsAdapter() (internal/uiadapter/config_test.go:L59)
- TestDefaultConfig_MergeOntoZeroValued() (internal/uiadapter/config_test.go:L75)
- TestConfig_HasEveryPlanField() (internal/uiadapter/config_test.go:L87)
- contextguard.go (internal/uiadapter/contextguard.go:L1)
- truncateToBudget() (internal/uiadapter/contextguard.go:L105)
- .ApplyClaude() (internal/uiadapter/contextguard.go:L135)
- .OllamaOptions() (internal/uiadapter/contextguard.go:L176)
- ContextGuard (internal/uiadapter/contextguard.go:L20)
- .ApplyOllama() (internal/uiadapter/contextguard.go:L55)
- TestStory1_AC5_ConfigDefaults() (internal/uiadapter/logging_test.go:L260)

# Depends on
- [Accountant](/modules/accountant.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [go_pkg_log_slog](/modules/go-pkg-log-slog.md)
- [sessions.go](/modules/sessions-go.md)
- [stages_test.go](/modules/stages-test-go.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Features
- no feature plan names these files

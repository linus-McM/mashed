---
type: Module
title: go_pkg_log_slog
description: "Graphify community 18: app_uiadapter_claudecli.go, internal/uiadapter/fallback.go, internal/uiadapter/fallback_test.go, internal/uiadapter/fallback_tiers.go, internal/uiadapter/fallback_tiers_test.go,"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: fallback, resource: internal/uiadapter/fallback.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 187709266229767c }
  - { id: fallback_test, resource: internal/uiadapter/fallback_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 2fa70a931cc96a6b }
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 88e0c867731816cd }
  - { id: fallback_tiers_test, resource: internal/uiadapter/fallback_tiers_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 73bb2c32262274bd }
  - { id: logging_plumbing_mock_test, resource: internal/uiadapter/logging_plumbing_mock_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: bce0a1e9685e2603 }
  - { id: mock, resource: internal/uiadapter/mock.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ee77ed81915c20b0 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e6b7941070a1bd3a }
  - { id: sanitize, resource: internal/uiadapter/sanitize.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 6cdb312fe6e18be6 }
  - { id: sanitize_test, resource: internal/uiadapter/sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: c90f1e17fbe3e70f }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/uiadapter/fallback.go`
- `internal/uiadapter/fallback_test.go`
- `internal/uiadapter/fallback_tiers.go`
- `internal/uiadapter/fallback_tiers_test.go`
- `internal/uiadapter/logging_plumbing_mock_test.go`
- `internal/uiadapter/mock.go`
- `internal/uiadapter/mock_test.go`
- `internal/uiadapter/sanitize.go`
- `internal/uiadapter/sanitize_test.go`

# Symbols
- app_uiadapter_claudecli.go (app_uiadapter_claudecli.go:L1)
- truncReason() (app_uiadapter_claudecli.go:L105)
- splitHeadTail() (app_uiadapter_claudecli.go:L118)
- dumpRawIfRequested() (app_uiadapter_claudecli.go:L129)
- capModel() (app_uiadapter_claudecli.go:L150)
- claudeCLIAdapter (app_uiadapter_claudecli.go:L21)
- .Translate() (app_uiadapter_claudecli.go:L39)
- fallback.go (internal/uiadapter/fallback.go:L1)
- FallbackAST() (internal/uiadapter/fallback.go:L23)
- firstLine() (internal/uiadapter/fallback.go:L47)
- fallback_test.go (internal/uiadapter/fallback_test.go:L1)
- TestFallbackAST_LiteralShape() (internal/uiadapter/fallback_test.go:L14)
- TestFallbackAST_TurnSummaryTruncates() (internal/uiadapter/fallback_test.go:L38)
- TestFallbackAST_EmptyRawIsSafe() (internal/uiadapter/fallback_test.go:L51)
- fallback_tiers.go (internal/uiadapter/fallback_tiers.go:L1)
- tierFailureReason() (internal/uiadapter/fallback_tiers.go:L19)
- FallbackTier (internal/uiadapter/fallback_tiers.go:L43)
- RunWithFallback() (internal/uiadapter/fallback_tiers.go:L57)
- fallback_tiers_test.go (internal/uiadapter/fallback_tiers_test.go:L1)
- TestFallback_TieredRecovery() (internal/uiadapter/fallback_tiers_test.go:L15)
- TestFallback_MinimalKindWhenAllBackendsFail() (internal/uiadapter/fallback_tiers_test.go:L37)
- TestFallback_PlaintextLastResort() (internal/uiadapter/fallback_tiers_test.go:L54)
- TestFallback_PrimarySuccessEscalatedFromEmpty() (internal/uiadapter/fallback_tiers_test.go:L69)
- TestFallback_ContextCancellation() (internal/uiadapter/fallback_tiers_test.go:L82)
- TestStory2_AC2_AC1_NewMockAcceptsNilLogger() (internal/uiadapter/logging_plumbing_mock_test.go:L19)
- mock.go (internal/uiadapter/mock.go:L1)
- MockAdapter (internal/uiadapter/mock.go:L19)
- NewMock() (internal/uiadapter/mock.go:L31)
- fixtureName() (internal/uiadapter/mock.go:L45)
- .Translate() (internal/uiadapter/mock.go:L58)
- mock_test.go (internal/uiadapter/mock_test.go:L1)
- runGoList() (internal/uiadapter/mock_test.go:L121)
- TestU2_AC8_MockAdapter_FixedReturn() (internal/uiadapter/mock_test.go:L19)
- TestU2_AC8_MockAdapter_NilReturnsFallback() (internal/uiadapter/mock_test.go:L33)
- TestU2_AC8_MockAdapter_NotInProductionBuild() (internal/uiadapter/mock_test.go:L46)
- sanitize.go (internal/uiadapter/sanitize.go:L1)
- sanitizeChrome() (internal/uiadapter/sanitize.go:L130)
- SanitizeCapture() (internal/uiadapter/sanitize.go:L29)
- stripOutsideFences() (internal/uiadapter/sanitize.go:L62)
- sanitize_test.go (internal/uiadapter/sanitize_test.go:L1)
- TestSanitize_StripsANSI_Golden() (internal/uiadapter/sanitize_test.go:L13)
- TestSanitize_PreservesFencedCodeBlocks() (internal/uiadapter/sanitize_test.go:L52)
- TestSanitize_Idempotent() (internal/uiadapter/sanitize_test.go:L70)
- TestSanitize_EmptyInput() (internal/uiadapter/sanitize_test.go:L80)
- TestSanitize_WhitespaceTrimmed() (internal/uiadapter/sanitize_test.go:L89)
- TestSanitize_LargeInput() (internal/uiadapter/sanitize_test.go:L98)

# Depends on
- [context.Context](/modules/context-context.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files

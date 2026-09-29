---
type: Module
title: FallbackAST
description: "Graphify community 91: app_uiadapter_claudecli.go, internal/uiadapter/fallback.go, internal/uiadapter/fallback_test.go, internal/uiadapter/sanitize.go, internal/uiadapter/sanitize_adapter_test.go, int"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-09-29T07:07:25Z", digest: 9c423a0a128af5b2 }
  - { id: fallback, resource: internal/uiadapter/fallback.go, last_modified: "2026-09-29T07:07:25Z", digest: 187709266229767c }
  - { id: fallback_test, resource: internal/uiadapter/fallback_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 2fa70a931cc96a6b }
  - { id: sanitize, resource: internal/uiadapter/sanitize.go, last_modified: "2026-09-29T07:07:25Z", digest: 6cdb312fe6e18be6 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 2951e338741da597 }
  - { id: sanitize_test, resource: internal/uiadapter/sanitize_test.go, last_modified: "2026-09-29T07:07:25Z", digest: c90f1e17fbe3e70f }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/uiadapter/fallback.go`
- `internal/uiadapter/fallback_test.go`
- `internal/uiadapter/sanitize.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/sanitize_test.go`

# Symbols
- app_uiadapter_claudecli.go (app_uiadapter_claudecli.go:L1)
- truncReason() (app_uiadapter_claudecli.go:L105)
- splitHeadTail() (app_uiadapter_claudecli.go:L118)
- dumpRawIfRequested() (app_uiadapter_claudecli.go:L129)
- capModel() (app_uiadapter_claudecli.go:L150)
- claudeCLIAdapter (app_uiadapter_claudecli.go:L21)
- .Translate() (app_uiadapter_claudecli.go:L39)
- FallbackAST() (internal/uiadapter/fallback.go:L23)
- firstLine() (internal/uiadapter/fallback.go:L47)
- TestFallbackAST_LiteralShape() (internal/uiadapter/fallback_test.go:L14)
- TestFallbackAST_TurnSummaryTruncates() (internal/uiadapter/fallback_test.go:L38)
- TestFallbackAST_EmptyRawIsSafe() (internal/uiadapter/fallback_test.go:L51)
- SanitizeCapture() (internal/uiadapter/sanitize.go:L29)
- TestAdapter_Translate_SanitizeRunsBeforeChat() (internal/uiadapter/sanitize_adapter_test.go:L88)
- sanitize_test.go (internal/uiadapter/sanitize_test.go:L1)
- TestSanitize_StripsANSI_Golden() (internal/uiadapter/sanitize_test.go:L13)
- TestSanitize_PreservesFencedCodeBlocks() (internal/uiadapter/sanitize_test.go:L52)
- TestSanitize_Idempotent() (internal/uiadapter/sanitize_test.go:L70)
- TestSanitize_EmptyInput() (internal/uiadapter/sanitize_test.go:L80)
- TestSanitize_WhitespaceTrimmed() (internal/uiadapter/sanitize_test.go:L89)
- TestSanitize_LargeInput() (internal/uiadapter/sanitize_test.go:L98)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [stages_test.go](/modules/stages-test-go.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

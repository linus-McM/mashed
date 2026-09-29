---
type: Module
title: SanitizeCapture
description: "Graphify community 140: internal/uiadapter/encode.go, internal/uiadapter/encode_test.go, internal/uiadapter/logging_story4_sanitize_test.go, internal/uiadapter/logging_story5_sanitize_test.go, interna"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-04-26T11:30:52+10:00", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: a0cd96e6b9f32bfb }
  - { id: logging_story4_sanitize_test, resource: internal/uiadapter/logging_story4_sanitize_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: b4229ff1bc6c01ad }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
  - { id: sanitize, resource: internal/uiadapter/sanitize.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 6cdb312fe6e18be6 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: sanitize_test, resource: internal/uiadapter/sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: c90f1e17fbe3e70f }
  - { id: spotlight, resource: internal/uiadapter/spotlight.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 82f02eccfdc52095 }
  - { id: spotlight_test, resource: internal/uiadapter/spotlight_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: f9f81a13aa4a0344 }
---

# Files
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/logging_story4_sanitize_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/sanitize.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/sanitize_test.go`
- `internal/uiadapter/spotlight.go`
- `internal/uiadapter/spotlight_test.go`

# Symbols
- ClaudeToolName() (internal/uiadapter/encode.go:L93)
- TestClaudeToolName_StableShape() (internal/uiadapter/encode_test.go:L63)
- story4RandomPayload() (internal/uiadapter/logging_story4_sanitize_test.go:L67)
- TestStory5_AC8_SanitizeDisciplineHolds() (internal/uiadapter/logging_story5_sanitize_test.go:L138)
- TestStory5_AC9_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story5_sanitize_test.go:L173)
- runStory5PayloadShapingSurface() (internal/uiadapter/logging_story5_sanitize_test.go:L56)
- SanitizeCapture() (internal/uiadapter/sanitize.go:L29)
- TestAdapter_Translate_SanitizeRunsBeforeChat() (internal/uiadapter/sanitize_adapter_test.go:L88)
- sanitize_test.go (internal/uiadapter/sanitize_test.go:L1)
- TestSanitize_StripsANSI_Golden() (internal/uiadapter/sanitize_test.go:L13)
- TestSanitize_PreservesFencedCodeBlocks() (internal/uiadapter/sanitize_test.go:L52)
- TestSanitize_Idempotent() (internal/uiadapter/sanitize_test.go:L70)
- TestSanitize_EmptyInput() (internal/uiadapter/sanitize_test.go:L80)
- TestSanitize_WhitespaceTrimmed() (internal/uiadapter/sanitize_test.go:L89)
- TestSanitize_LargeInput() (internal/uiadapter/sanitize_test.go:L98)
- Unspotlight() (internal/uiadapter/spotlight.go:L100)
- Spotlight() (internal/uiadapter/spotlight.go:L42)
- spotlight_test.go (internal/uiadapter/spotlight_test.go:L1)
- TestSpotlight_ReplacesWhitespace() (internal/uiadapter/spotlight_test.go:L13)
- TestStory5_AC2_UnspotlightRemoved() (internal/uiadapter/spotlight_test.go:L141)
- TestSpotlight_DisabledBypass() (internal/uiadapter/spotlight_test.go:L21)
- TestSpotlight_RoundTripLossless() (internal/uiadapter/spotlight_test.go:L29)
- TestSpotlight_InjectionCorpus() (internal/uiadapter/spotlight_test.go:L39)
- TestSpotlight_EmptyAndUnicode() (internal/uiadapter/spotlight_test.go:L65)

# Depends on
- [log/slog.Logger](/modules/log-slog-logger.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [NewContextGuard](/modules/newcontextguard.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files

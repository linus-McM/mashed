---
type: Module
title: logging_comprehensive_test.go
description: "Graphify community 105: internal/uiadapter/logging.go, internal/uiadapter/logging_comprehensive_test.go, internal/uiadapter/logging_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
  - { id: logging_comprehensive_test, resource: internal/uiadapter/logging_comprehensive_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 446816f0c9fa2be3 }
  - { id: logging_test, resource: internal/uiadapter/logging_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 744e5721d44f353b }
---

# Files
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_comprehensive_test.go`
- `internal/uiadapter/logging_test.go`

# Symbols
- NewProductionLogger() (internal/uiadapter/logging.go:L51)
- parseSlogLevel() (internal/uiadapter/logging.go:L78)
- logging_comprehensive_test.go (internal/uiadapter/logging_comprehensive_test.go:L1)
- TestStory6_AC1_WithAttrs_BothSinks() (internal/uiadapter/logging_comprehensive_test.go:L164)
- TestStory6_AC1_WithGroup_NestsJSON() (internal/uiadapter/logging_comprehensive_test.go:L192)
- TestStory6_AC1_FileNameUTC() (internal/uiadapter/logging_comprehensive_test.go:L221)
- TestStory6_AC1_RaceSafe() (internal/uiadapter/logging_comprehensive_test.go:L256)
- TestStory6_AC1_NilSafeLoggerComprehensive() (internal/uiadapter/logging_comprehensive_test.go:L305)
- readJSONFileLines() (internal/uiadapter/logging_comprehensive_test.go:L31)
- TestStory6_AC1_FanoutToBothSinks_Comprehensive() (internal/uiadapter/logging_comprehensive_test.go:L57)
- TestStory6_AC1_LevelFilter_BothSinks() (internal/uiadapter/logging_comprehensive_test.go:L95)
- logging_test.go (internal/uiadapter/logging_test.go:L1)
- TestStory1_AC1_LevelFilter() (internal/uiadapter/logging_test.go:L110)
- TestStory1_AC2_CloserIdempotent() (internal/uiadapter/logging_test.go:L145)
- TestStory1_AC3_StdoutOnlyFallback() (internal/uiadapter/logging_test.go:L173)
- TestStory1_AC4_ParseSlogLevel() (internal/uiadapter/logging_test.go:L208)
- TestStory1_AC4_EnvVarBootLevel() (internal/uiadapter/logging_test.go:L241)
- captureStdout() (internal/uiadapter/logging_test.go:L35)
- expectedLogFile() (internal/uiadapter/logging_test.go:L57)
- TestStory1_AC1_NewProductionLogger_FanoutToBothSinks() (internal/uiadapter/logging_test.go:L66)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files

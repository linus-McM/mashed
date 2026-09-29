---
type: Module
title: allowlist_test.go
description: "Graphify community 181: internal/uiadapter/allowlist.go, internal/uiadapter/allowlist_test.go, internal/uiadapter/logging_story3_sanitize_test.go, internal/uiadapter/logging_story4_sanitize_test.go, i"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: allowlist, resource: internal/uiadapter/allowlist.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ec96ab6a3976170d }
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 865e6c967ac886df }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
  - { id: logging_story4_sanitize_test, resource: internal/uiadapter/logging_story4_sanitize_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: b4229ff1bc6c01ad }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
---

# Files
- `internal/uiadapter/allowlist.go`
- `internal/uiadapter/allowlist_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/logging_story4_sanitize_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`

# Symbols
- allowlist.go (internal/uiadapter/allowlist.go:L1)
- CheckModelAllowlist() (internal/uiadapter/allowlist.go:L40)
- joinAllowlist() (internal/uiadapter/allowlist.go:L78)
- allowlist_test.go (internal/uiadapter/allowlist_test.go:L1)
- TestStory5_AC6_AllowlistNilLoggerStillSafe() (internal/uiadapter/allowlist_test.go:L110)
- TestAllowlist_WarnsOnUnvetted() (internal/uiadapter/allowlist_test.go:L18)
- TestAllowlist_AllowsKnownModels() (internal/uiadapter/allowlist_test.go:L32)
- TestAllowlist_OverrideFlag() (internal/uiadapter/allowlist_test.go:L44)
- TestAllowlist_WarnsOnUnvettedClaude() (internal/uiadapter/allowlist_test.go:L57)
- TestStory5_AC6_AllowlistDefaultRemoved() (internal/uiadapter/allowlist_test.go:L93)
- containsSlice() (internal/uiadapter/logging_story3_sanitize_test.go:L55)
- story4RandomPayload() (internal/uiadapter/logging_story4_sanitize_test.go:L67)
- TestStory4_AC7_SanitizeDisciplineAcrossPipeline() (internal/uiadapter/logging_story4_sanitize_test.go:L90)
- logging_story5_sanitize_test.go (internal/uiadapter/logging_story5_sanitize_test.go:L1)
- TestStory5_AC8_SanitizeDisciplineHolds() (internal/uiadapter/logging_story5_sanitize_test.go:L138)

# Depends on
- [Config](/modules/config.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [NewFastPathClassifier](/modules/newfastpathclassifier.md)
- [NewRepairer](/modules/newrepairer.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files

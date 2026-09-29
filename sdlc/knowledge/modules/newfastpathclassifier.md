---
type: Module
title: NewFastPathClassifier
description: "Graphify community 252: internal/uiadapter/fastpath.go, internal/uiadapter/fastpath_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: fastpath, resource: internal/uiadapter/fastpath.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 4ce7a34b961d63bc }
  - { id: fastpath_test, resource: internal/uiadapter/fastpath_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: a7f912e39ab3f4b0 }
---

# Files
- `internal/uiadapter/fastpath.go`
- `internal/uiadapter/fastpath_test.go`

# Symbols
- NewFastPathClassifier() (internal/uiadapter/fastpath.go:L45)
- fastpath_test.go (internal/uiadapter/fastpath_test.go:L1)
- TestFastPath_FirstMatchWins() (internal/uiadapter/fastpath_test.go:L102)
- BenchmarkFastPath() (internal/uiadapter/fastpath_test.go:L113)
- TestFastPath_NoRuleMatches() (internal/uiadapter/fastpath_test.go:L130)
- TestFastPath_CoversCommonCases() (internal/uiadapter/fastpath_test.go:L14)
- TestFastPath_DisabledReturnsMiss() (internal/uiadapter/fastpath_test.go:L45)
- TestFastPath_HitRateAndCounters() (internal/uiadapter/fastpath_test.go:L55)
- TestFastPath_NumberedMenuRequiresTwoItems() (internal/uiadapter/fastpath_test.go:L73)
- TestFastPath_NumberedMenuExtractsOptions() (internal/uiadapter/fastpath_test.go:L84)

# Depends on
- [fastpath.go](/modules/fastpath-go.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

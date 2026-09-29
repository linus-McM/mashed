---
type: Module
title: NewFastPathClassifier
description: "Graphify community 252: internal/uiadapter/fastpath.go, internal/uiadapter/fastpath_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: fastpath, resource: internal/uiadapter/fastpath.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 4ce7a34b961d63bc }
  - { id: fastpath_test, resource: internal/uiadapter/fastpath_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: a7f912e39ab3f4b0 }
---

# Files
- `internal/uiadapter/fastpath.go`
- `internal/uiadapter/fastpath_test.go`

# Symbols
- fastpath.go (internal/uiadapter/fastpath.go:L1)
- .HitRate() (internal/uiadapter/fastpath.go:L108)
- .HitsPerRule() (internal/uiadapter/fastpath.go:L117)
- defaultFastRules() (internal/uiadapter/fastpath.go:L127)
- buildYN() (internal/uiadapter/fastpath.go:L157)
- buildPressEnter() (internal/uiadapter/fastpath.go:L172)
- buildFileConfirm() (internal/uiadapter/fastpath.go:L186)
- buildNumberedMenu() (internal/uiadapter/fastpath.go:L202)
- buildFreeText() (internal/uiadapter/fastpath.go:L241)
- firstNonEmptyLine() (internal/uiadapter/fastpath.go:L256)
- FastPathClassifier (internal/uiadapter/fastpath.go:L26)
- truncate() (internal/uiadapter/fastpath.go:L266)
- fastRule (internal/uiadapter/fastpath.go:L35)
- NewFastPathClassifier() (internal/uiadapter/fastpath.go:L45)
- .Classify() (internal/uiadapter/fastpath.go:L56)
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
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

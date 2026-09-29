---
type: Module
title: fastpath.go
description: "Graphify community 260: internal/uiadapter/cache.go, internal/uiadapter/eval/scorecard_v3.go, internal/uiadapter/fastpath.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
  - { id: fastpath, resource: internal/uiadapter/fastpath.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 4ce7a34b961d63bc }
---

# Files
- `internal/uiadapter/cache.go`
- `internal/uiadapter/eval/scorecard_v3.go`
- `internal/uiadapter/fastpath.go`

# Symbols
- .Store() (internal/uiadapter/cache.go:L111)
- .logCacheGet() (internal/uiadapter/cache.go:L127)
- .HitRate() (internal/uiadapter/cache.go:L144)
- .DoShared() (internal/uiadapter/cache.go:L162)
- hashKey() (internal/uiadapter/cache.go:L17)
- .Metrics() (internal/uiadapter/cache.go:L174)
- ResponseCache (internal/uiadapter/cache.go:L35)
- .onEvicted() (internal/uiadapter/cache.go:L66)
- .Lookup() (internal/uiadapter/cache.go:L94)
- .AggregateByBackend() (internal/uiadapter/eval/scorecard_v3.go:L113)
- percentileIdx() (internal/uiadapter/eval/scorecard_v3.go:L170)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard_v3.go:L183)
- .PrettyPrint() (internal/uiadapter/eval/scorecard_v3.go:L215)
- .CrossBackendDelta() (internal/uiadapter/eval/scorecard_v3.go:L230)
- joinBackends() (internal/uiadapter/eval/scorecard_v3.go:L251)
- ShadowSampler (internal/uiadapter/eval/scorecard_v3.go:L262)
- .ShouldSample() (internal/uiadapter/eval/scorecard_v3.go:L274)
- .Counts() (internal/uiadapter/eval/scorecard_v3.go:L296)
- Row (internal/uiadapter/eval/scorecard_v3.go:L48)
- ScorecardV3 (internal/uiadapter/eval/scorecard_v3.go:L73)
- .Add() (internal/uiadapter/eval/scorecard_v3.go:L82)
- .Rows() (internal/uiadapter/eval/scorecard_v3.go:L89)
- Aggregate (internal/uiadapter/eval/scorecard_v3.go:L99)
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
- .Classify() (internal/uiadapter/fastpath.go:L56)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

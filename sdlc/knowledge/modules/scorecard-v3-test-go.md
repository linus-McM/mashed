---
type: Module
title: scorecard_v3_test.go
description: "Graphify community 247: internal/uiadapter/eval/scorecard_v3.go, internal/uiadapter/eval/scorecard_v3_test.go"
resource: internal/uiadapter/eval
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
  - { id: scorecard_v3_test, resource: internal/uiadapter/eval/scorecard_v3_test.go, last_modified: "2026-04-23T11:36:37+10:00", digest: cd219473982ce8b5 }
---

# Files
- `internal/uiadapter/eval/scorecard_v3.go`
- `internal/uiadapter/eval/scorecard_v3_test.go`

# Symbols
- .AggregateByBackend() (internal/uiadapter/eval/scorecard_v3.go:L113)
- percentileIdx() (internal/uiadapter/eval/scorecard_v3.go:L170)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard_v3.go:L183)
- .PrettyPrint() (internal/uiadapter/eval/scorecard_v3.go:L215)
- .CrossBackendDelta() (internal/uiadapter/eval/scorecard_v3.go:L230)
- joinBackends() (internal/uiadapter/eval/scorecard_v3.go:L251)
- NewShadowSampler() (internal/uiadapter/eval/scorecard_v3.go:L270)
- .ShouldSample() (internal/uiadapter/eval/scorecard_v3.go:L274)
- Row (internal/uiadapter/eval/scorecard_v3.go:L48)
- ScorecardV3 (internal/uiadapter/eval/scorecard_v3.go:L73)
- NewScorecardV3() (internal/uiadapter/eval/scorecard_v3.go:L79)
- .Add() (internal/uiadapter/eval/scorecard_v3.go:L82)
- .Rows() (internal/uiadapter/eval/scorecard_v3.go:L89)
- Aggregate (internal/uiadapter/eval/scorecard_v3.go:L99)
- scorecard_v3_test.go (internal/uiadapter/eval/scorecard_v3_test.go:L1)
- TestScorecardV3_ClaudeAPIModelSpecificThreshold() (internal/uiadapter/eval/scorecard_v3_test.go:L113)
- TestScorecardV3_AggregateByBackend() (internal/uiadapter/eval/scorecard_v3_test.go:L14)
- TestScorecardV3_MeetsThresholds() (internal/uiadapter/eval/scorecard_v3_test.go:L29)
- TestScorecardV3_MeetsThresholds_Pass() (internal/uiadapter/eval/scorecard_v3_test.go:L46)
- TestScorecardV3_PrettyPrint_GroupsByBackend() (internal/uiadapter/eval/scorecard_v3_test.go:L57)
- TestScorecardV3_CrossBackendDelta() (internal/uiadapter/eval/scorecard_v3_test.go:L71)
- TestShadowSampler_Rate() (internal/uiadapter/eval/scorecard_v3_test.go:L82)
- TestShadowSampler_Extremes() (internal/uiadapter/eval/scorecard_v3_test.go:L99)

# Depends on
- [context.Context](/modules/context-context.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

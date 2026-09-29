---
type: Module
title: time.Duration
description: "Graphify community 323: internal/uiadapter/eval/scorecard_v3.go, internal/uiadapter/eval/scorecard_v3_test.go"
resource: internal/uiadapter/eval
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
  - { id: scorecard_v3_test, resource: internal/uiadapter/eval/scorecard_v3_test.go, last_modified: "2026-04-23T11:36:37+10:00", digest: cd219473982ce8b5 }
---

# Files
- `internal/uiadapter/eval/scorecard_v3.go`
- `internal/uiadapter/eval/scorecard_v3_test.go`

# Symbols
- scorecard_v3.go (internal/uiadapter/eval/scorecard_v3.go:L1)
- .AggregateByBackend() (internal/uiadapter/eval/scorecard_v3.go:L113)
- percentileIdx() (internal/uiadapter/eval/scorecard_v3.go:L170)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard_v3.go:L183)
- BackendThresholds (internal/uiadapter/eval/scorecard_v3.go:L20)
- .PrettyPrint() (internal/uiadapter/eval/scorecard_v3.go:L215)
- .CrossBackendDelta() (internal/uiadapter/eval/scorecard_v3.go:L230)
- joinBackends() (internal/uiadapter/eval/scorecard_v3.go:L251)
- ShadowSampler (internal/uiadapter/eval/scorecard_v3.go:L262)
- NewShadowSampler() (internal/uiadapter/eval/scorecard_v3.go:L270)
- .ShouldSample() (internal/uiadapter/eval/scorecard_v3.go:L274)
- .Counts() (internal/uiadapter/eval/scorecard_v3.go:L296)
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
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

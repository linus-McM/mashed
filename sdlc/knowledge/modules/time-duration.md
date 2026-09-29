---
type: Module
title: time.Duration
description: "Graphify community 323: app_review.go, internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/lifecycle_test.go, internal/uiadapter/backend/s"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: app_review, resource: app_review.go, last_modified: "2026-05-07T18:18:02+10:00", digest: f186f322bd66e914 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
---

# Files
- `app_review.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/eval/scorecard_v3.go`

# Symbols
- .StreamCodeReviewSummary() (app_review.go:L112)
- runClaudePrompt() (app_review.go:L459)
- isReviewableFile() (app_review.go:L84)
- Capabilities (internal/uiadapter/backend/backend.go:L45)
- .Capabilities() (internal/uiadapter/backend/claudeapi/client.go:L65)
- slowBackend (internal/uiadapter/backend/lifecycle_test.go:L16)
- .WarmUp() (internal/uiadapter/backend/lifecycle_test.go:L22)
- StubBackend (internal/uiadapter/backend/stubs.go:L14)
- .Name() (internal/uiadapter/backend/stubs.go:L32)
- .Classify() (internal/uiadapter/backend/stubs.go:L34)
- .WarmUp() (internal/uiadapter/backend/stubs.go:L52)
- .Health() (internal/uiadapter/backend/stubs.go:L57)
- .Capabilities() (internal/uiadapter/backend/stubs.go:L62)
- .Calls() (internal/uiadapter/backend/stubs.go:L66)
- scorecard_v3.go (internal/uiadapter/eval/scorecard_v3.go:L1)
- .AggregateByBackend() (internal/uiadapter/eval/scorecard_v3.go:L113)
- percentileIdx() (internal/uiadapter/eval/scorecard_v3.go:L170)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard_v3.go:L183)
- BackendThresholds (internal/uiadapter/eval/scorecard_v3.go:L20)
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

# Depends on
- [app_review_test.go](/modules/app-review-test-go.md)
- [Config](/modules/config.md)
- [diff.go](/modules/diff-go.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [scorecard_v3_test.go](/modules/scorecard-v3-test-go.md)

# Inferred
- [App](/modules/app.md)

# Features
- no feature plan names these files

---
type: Module
title: time.Duration
description: "Graphify community 260: app_uiadapter_claudecli.go, internal/uiadapter/eval/scorecard_v3.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/uiadapter/eval/scorecard_v3.go`

# Symbols
- app_uiadapter_claudecli.go (app_uiadapter_claudecli.go:L1)
- truncReason() (app_uiadapter_claudecli.go:L105)
- splitHeadTail() (app_uiadapter_claudecli.go:L118)
- dumpRawIfRequested() (app_uiadapter_claudecli.go:L129)
- capModel() (app_uiadapter_claudecli.go:L150)
- claudeCLIAdapter (app_uiadapter_claudecli.go:L21)
- .Translate() (app_uiadapter_claudecli.go:L39)
- scorecard_v3.go (internal/uiadapter/eval/scorecard_v3.go:L1)
- .AggregateByBackend() (internal/uiadapter/eval/scorecard_v3.go:L113)
- percentileIdx() (internal/uiadapter/eval/scorecard_v3.go:L170)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard_v3.go:L183)
- BackendThresholds (internal/uiadapter/eval/scorecard_v3.go:L20)
- .PrettyPrint() (internal/uiadapter/eval/scorecard_v3.go:L215)
- .CrossBackendDelta() (internal/uiadapter/eval/scorecard_v3.go:L230)
- joinBackends() (internal/uiadapter/eval/scorecard_v3.go:L251)
- .ShouldSample() (internal/uiadapter/eval/scorecard_v3.go:L274)
- Row (internal/uiadapter/eval/scorecard_v3.go:L48)
- ScorecardV3 (internal/uiadapter/eval/scorecard_v3.go:L73)
- .Add() (internal/uiadapter/eval/scorecard_v3.go:L82)
- .Rows() (internal/uiadapter/eval/scorecard_v3.go:L89)
- Aggregate (internal/uiadapter/eval/scorecard_v3.go:L99)

# Depends on
- [context.Context](/modules/context-context.md)
- [go_pkg_testing](/modules/go-pkg-testing.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [Repairer](/modules/repairer.md)
- [scorecard_v3_test.go](/modules/scorecard-v3-test-go.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

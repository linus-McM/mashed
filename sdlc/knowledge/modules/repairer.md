---
type: Module
title: Repairer
description: "Graphify community 185: internal/uiadapter/backend/lifecycle_test.go, internal/uiadapter/eval/scorecard_v3.go, internal/uiadapter/repair.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
  - { id: repair, resource: internal/uiadapter/repair.go, last_modified: "2026-04-26T10:43:55+10:00", digest: cc21726252779f85 }
---

# Files
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/eval/scorecard_v3.go`
- `internal/uiadapter/repair.go`

# Symbols
- slowBackend (internal/uiadapter/backend/lifecycle_test.go:L16)
- .WarmUp() (internal/uiadapter/backend/lifecycle_test.go:L22)
- ShadowSampler (internal/uiadapter/eval/scorecard_v3.go:L262)
- .Counts() (internal/uiadapter/eval/scorecard_v3.go:L296)
- .MaxRetries() (internal/uiadapter/repair.go:L118)
- .Run() (internal/uiadapter/repair.go:L124)
- .Metrics() (internal/uiadapter/repair.go:L202)
- firstReason() (internal/uiadapter/repair.go:L23)
- Repairer (internal/uiadapter/repair.go:L99)

# Depends on
- [context.Context](/modules/context-context.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [time.Duration](/modules/time-duration.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

---
type: Module
title: StubBackend
description: "Graphify community 323: internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/claudecli/client.go, internal/uiadapter/backend/lifecycle_test"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
---

# Files
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/eval/scorecard_v3.go`

# Symbols
- Capabilities (internal/uiadapter/backend/backend.go:L45)
- .Capabilities() (internal/uiadapter/backend/claudeapi/client.go:L65)
- .Capabilities() (internal/uiadapter/backend/claudecli/client.go:L44)
- slowBackend (internal/uiadapter/backend/lifecycle_test.go:L16)
- .WarmUp() (internal/uiadapter/backend/lifecycle_test.go:L22)
- StubBackend (internal/uiadapter/backend/stubs.go:L14)
- .Name() (internal/uiadapter/backend/stubs.go:L32)
- .Classify() (internal/uiadapter/backend/stubs.go:L34)
- .WarmUp() (internal/uiadapter/backend/stubs.go:L52)
- .Health() (internal/uiadapter/backend/stubs.go:L57)
- .Capabilities() (internal/uiadapter/backend/stubs.go:L62)
- .Calls() (internal/uiadapter/backend/stubs.go:L66)
- ShadowSampler (internal/uiadapter/eval/scorecard_v3.go:L262)
- .Counts() (internal/uiadapter/eval/scorecard_v3.go:L296)

# Depends on
- [claudeapi/client.go](/modules/claudeapi-client-go.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [ScorecardV3](/modules/scorecardv3.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

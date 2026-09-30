---
type: Module
title: engine_test.go
description: "Graphify community 343: internal/agent/engine.go, internal/agent/engine_test.go, internal/agent/tokensamples_test.go"
resource: internal/agent
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-09-29T07:07:25Z", digest: 1042190db571e4e7 }
  - { id: engine_test, resource: internal/agent/engine_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 745feda4505cf50c }
  - { id: tokensamples_test, resource: internal/agent/tokensamples_test.go, last_modified: "2026-09-29T07:07:25Z", digest: e73f3aa2bb785df4 }
---

# Files
- `internal/agent/engine.go`
- `internal/agent/engine_test.go`
- `internal/agent/tokensamples_test.go`

# Symbols
- SortByPriority() (internal/agent/engine.go:L156)
- NewNotificationEngine() (internal/agent/engine.go:L60)
- engine_test.go (internal/agent/engine_test.go:L1)
- drainEvents() (internal/agent/engine_test.go:L11)
- TestSortByPriority() (internal/agent/engine_test.go:L180)
- TestProcessAgentUpdate_EmptyID() (internal/agent/engine_test.go:L207)
- TestActiveAgentCount() (internal/agent/engine_test.go:L222)
- TestStateTransition_RunningToError() (internal/agent/engine_test.go:L28)
- TestDeduplication_SameStateTwice() (internal/agent/engine_test.go:L62)
- TestAgentRestart_CompletedStartedRunning() (internal/agent/engine_test.go:L94)
- TestProcessAgentUpdate_PropagatesTokenSamples() (internal/agent/tokensamples_test.go:L142)

# Depends on
- [domain/types.go](/modules/domain-types-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

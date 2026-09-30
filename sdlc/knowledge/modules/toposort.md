---
type: Module
title: topoSort
description: "Graphify community 400: internal/bmad/executor.go, internal/bmad/executor_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_test.go`

# Symbols
- extractRegex() (internal/bmad/executor.go:L2118)
- topoSort() (internal/bmad/executor.go:L2312)
- TestTopoSort_Sequential() (internal/bmad/executor_test.go:L133)
- TestExtractRegex_WithCaptureGroup() (internal/bmad/executor_test.go:L1462)
- TestTopoSort_Parallel() (internal/bmad/executor_test.go:L149)
- TestExtractRegex_WithoutCaptureGroup() (internal/bmad/executor_test.go:L1490)
- TestExtractRegex_NoMatch() (internal/bmad/executor_test.go:L1518)
- TestExtractRegex_InvalidRegex() (internal/bmad/executor_test.go:L1523)
- TestTopoSort_Cycle() (internal/bmad/executor_test.go:L164)
- TestTopoSort_Diamond() (internal/bmad/executor_test.go:L174)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

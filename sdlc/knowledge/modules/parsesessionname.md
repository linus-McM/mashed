---
type: Module
title: ParseSessionName
description: "Graphify community 474: internal/bmad/executor_test.go, internal/bmad/session_naming.go, internal/bmad/session_naming_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: session_naming, resource: internal/bmad/session_naming.go, last_modified: "2026-04-10T15:50:15+10:00", digest: acbdad6853f5eaed }
  - { id: session_naming_test, resource: internal/bmad/session_naming_test.go, last_modified: "2026-04-10T15:50:15+10:00", digest: 42de95158f5b442b }
---

# Files
- `internal/bmad/executor_test.go`
- `internal/bmad/session_naming.go`
- `internal/bmad/session_naming_test.go`

# Symbols
- runExecuteNodeSessionCase() (internal/bmad/executor_test.go:L852)
- TestExecuteNode_AC7_UsesDescriptiveName() (internal/bmad/executor_test.go:L904)
- TestExecuteNode_AC8_BranchLookupFailureFallsBackToDetached() (internal/bmad/executor_test.go:L922)
- ParseSessionName() (internal/bmad/session_naming.go:L99)
- TestParseSessionName_AC6_InvalidInputs() (internal/bmad/session_naming_test.go:L404)

# Depends on
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [newHarness](/modules/newharness.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

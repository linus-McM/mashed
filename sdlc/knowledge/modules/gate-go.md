---
type: Module
title: gate.go
description: "Graphify community 278: internal/bmad/executor_anyuseranswer_test.go, internal/bmad/gate.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: executor_anyuseranswer_test, resource: internal/bmad/executor_anyuseranswer_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: c5b7c0a594991958 }
  - { id: gate, resource: internal/bmad/gate.go, last_modified: "2026-04-21T20:21:32+10:00", digest: faeb23bd0495c1c9 }
---

# Files
- `internal/bmad/executor_anyuseranswer_test.go`
- `internal/bmad/gate.go`

# Symbols
- TestGate_AnyUserAnswerMatches_LastRoundWindow() (internal/bmad/executor_anyuseranswer_test.go:L44)
- gate.go (internal/bmad/gate.go:L1)
- collectSubAnswersForSpec() (internal/bmad/gate.go:L115)
- findNodeProcessID() (internal/bmad/gate.go:L132)
- astStructuredInUse() (internal/bmad/gate.go:L152)
- flattenSubAnswers() (internal/bmad/gate.go:L169)
- .checkGate() (internal/bmad/gate.go:L19)
- Executor (internal/bmad/gate.go:L19)
- containsToken() (internal/bmad/gate.go:L76)
- anyUserAnswerMatches() (internal/bmad/gate.go:L95)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [NewExecutor](/modules/newexecutor.md)
- [ProcessByID](/modules/processbyid.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)

# Features
- no feature plan names these files

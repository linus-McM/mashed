---
type: Module
title: wait_idle_test.go
description: "Graphify community 329: internal/bmad/mock_helpers_test.go, internal/bmad/wait_idle_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: mock_helpers_test, resource: internal/bmad/mock_helpers_test.go, last_modified: "2026-04-12T16:42:48+10:00", digest: dd7b5edf17e36713 }
  - { id: wait_idle_test, resource: internal/bmad/wait_idle_test.go, last_modified: "2026-04-23T13:05:14+10:00", digest: 1dda9adef91d1698 }
---

# Files
- `internal/bmad/mock_helpers_test.go`
- `internal/bmad/wait_idle_test.go`

# Symbols
- mock_helpers_test.go (internal/bmad/mock_helpers_test.go:L1)
- makeIdleOutput() (internal/bmad/mock_helpers_test.go:L117)
- idleMockRunner() (internal/bmad/mock_helpers_test.go:L86)
- wait_idle_test.go (internal/bmad/wait_idle_test.go:L1)
- TestWaitForIdleCompletion_CtxCancel() (internal/bmad/wait_idle_test.go:L119)
- newWaitIdleState() (internal/bmad/wait_idle_test.go:L17)
- TestWaitForIdleCompletion_PaneDeathSessionGone() (internal/bmad/wait_idle_test.go:L184)
- TestWaitForIdleCompletion_IdlePromptRequiresStableHash() (internal/bmad/wait_idle_test.go:L212)
- TestWaitForIdleCompletion_HappyPath() (internal/bmad/wait_idle_test.go:L37)
- TestWaitForIdleCompletion_NoWork_Timeout() (internal/bmad/wait_idle_test.go:L65)
- TestWaitForIdleCompletion_PaneDeath() (internal/bmad/wait_idle_test.go:L81)

# Depends on
- [newHarness](/modules/newharness.md)

# Inferred
- [Executor](/modules/executor.md)
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files

---
type: Module
title: seedResponseState
description: "Graphify community 185: internal/bmad/executor_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
---

# Files
- `internal/bmad/executor_test.go`

# Symbols
- responseRunner() (internal/bmad/executor_test.go:L2737)
- findCall() (internal/bmad/executor_test.go:L2757)
- seedResponseState() (internal/bmad/executor_test.go:L2780)
- TestStory2_AC1_RespondToQuestion_Success() (internal/bmad/executor_test.go:L2805)
- TestStory2_AC1_RespondToQuestion_MenuOption() (internal/bmad/executor_test.go:L2832)
- TestStory2_AC2_RespondToQuestion_DeadPane() (internal/bmad/executor_test.go:L2848)
- TestStory2_AC2_RespondToQuestion_PaneCheckFails() (internal/bmad/executor_test.go:L2866)
- TestStory2_AC4_RespondToQuestion_ClearsHash() (internal/bmad/executor_test.go:L2883)
- TestStory2_AC4_RespondToQuestion_HashPreservedOnFailure() (internal/bmad/executor_test.go:L2898)
- TestStory2_RespondToQuestion_UnknownNode() (internal/bmad/executor_test.go:L2921)
- TestStory2_RespondToQuestion_LongAnswer() (internal/bmad/executor_test.go:L2933)
- TestStory2_RespondToQuestion_NoTmuxTarget() (internal/bmad/executor_test.go:L2953)
- TestStory2_RespondToQuestion_SendKeysLiteralFails() (internal/bmad/executor_test.go:L2970)
- TestStory2_RespondToQuestion_SendKeysEnterFails() (internal/bmad/executor_test.go:L2989)
- TestStory2_RespondToQuestion_EmptyAnswer() (internal/bmad/executor_test.go:L3014)

# Depends on
- [cleanup_test.go](/modules/cleanup-test-go.md)
- [newHarness](/modules/newharness.md)
- [sync.Mutex](/modules/sync-mutex.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

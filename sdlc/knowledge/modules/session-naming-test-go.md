---
type: Module
title: session_naming_test.go
description: "Graphify community 101: internal/bmad/executor_test.go, internal/bmad/session_naming.go, internal/bmad/session_naming_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
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
- slugifyComponent() (internal/bmad/session_naming.go:L130)
- shortHashOf() (internal/bmad/session_naming.go:L147)
- BuildSessionName() (internal/bmad/session_naming.go:L69)
- ParseSessionName() (internal/bmad/session_naming.go:L99)
- session_naming_test.go (internal/bmad/session_naming_test.go:L1)
- TestBuildSessionName_AC3_AllEmptyFallback() (internal/bmad/session_naming_test.go:L184)
- TestBuildSessionName_AC3_EmptyComponentFallbacks() (internal/bmad/session_naming_test.go:L191)
- buildAndSplit() (internal/bmad/session_naming_test.go:L21)
- TestBuildSessionName_AC4_LengthCapWith500ByteInputs() (internal/bmad/session_naming_test.go:L262)
- TestBuildSessionName_AC4_HashPreservedUnderTruncation() (internal/bmad/session_naming_test.go:L288)
- TestBuildSessionName_HashDeterminismAndUniqueness() (internal/bmad/session_naming_test.go:L304)
- TestParseSessionName_AC5_RoundTrip() (internal/bmad/session_naming_test.go:L364)
- TestParseSessionName_AC5_FeatureBranchAppearsInName() (internal/bmad/session_naming_test.go:L387)
- assertValidHash() (internal/bmad/session_naming_test.go:L40)
- TestParseSessionName_AC6_InvalidInputs() (internal/bmad/session_naming_test.go:L404)
- TestSlugifyComponent() (internal/bmad/session_naming_test.go:L442)
- TestSessionNaming_ConstantsSanity() (internal/bmad/session_naming_test.go:L483)
- TestBuildSessionName_AC1_HappyPath() (internal/bmad/session_naming_test.go:L49)
- TestBuildSessionName_AC2_SlugificationRules() (internal/bmad/session_naming_test.go:L74)

# Depends on
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [newHarness](/modules/newharness.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

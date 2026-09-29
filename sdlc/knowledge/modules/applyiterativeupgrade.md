---
type: Module
title: applyIterativeUpgrade
description: "Graphify community 26: internal/bmad/interactive_defaults.go, internal/bmad/interactive_defaults_test.go, internal/bmad/registry_interactive_phase2.go, internal/bmad/registry_interactive_phase2_helper"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: interactive_defaults, resource: internal/bmad/interactive_defaults.go, last_modified: "2026-09-29T07:07:25Z", digest: d72d23f8f4651178 }
  - { id: interactive_defaults_test, resource: internal/bmad/interactive_defaults_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 3238ce12f829fa21 }
  - { id: registry_interactive_phase2, resource: internal/bmad/registry_interactive_phase2.go, last_modified: "2026-09-29T07:07:25Z", digest: 4e4d06a37c78b67b }
  - { id: registry_interactive_phase2_helper_test, resource: internal/bmad/registry_interactive_phase2_helper_test.go, last_modified: "2026-09-29T07:07:25Z", digest: b442030b66b1e92a }
  - { id: registry_interactive_phase3a, resource: internal/bmad/registry_interactive_phase3a.go, last_modified: "2026-09-29T07:07:25Z", digest: c79c3f4b9c008816 }
  - { id: registry_interactive_phase3b, resource: internal/bmad/registry_interactive_phase3b.go, last_modified: "2026-09-29T07:07:25Z", digest: 57eb0e74dbdb175b }
  - { id: registry_interactive_phase3c, resource: internal/bmad/registry_interactive_phase3c.go, last_modified: "2026-09-29T07:07:25Z", digest: 2ca78d7d99de3329 }
  - { id: registry_interactive_phase3d, resource: internal/bmad/registry_interactive_phase3d.go, last_modified: "2026-09-29T07:07:25Z", digest: 1192a8ff19586453 }
  - { id: registry_interactive_phase4, resource: internal/bmad/registry_interactive_phase4.go, last_modified: "2026-09-29T07:07:25Z", digest: 581736515ce5163f }
  - { id: registry_interactive_phase5, resource: internal/bmad/registry_interactive_phase5.go, last_modified: "2026-09-29T07:07:25Z", digest: 81eb5eb463eb4c28 }
---

# Files
- `internal/bmad/interactive_defaults.go`
- `internal/bmad/interactive_defaults_test.go`
- `internal/bmad/registry_interactive_phase2.go`
- `internal/bmad/registry_interactive_phase2_helper_test.go`
- `internal/bmad/registry_interactive_phase3a.go`
- `internal/bmad/registry_interactive_phase3b.go`
- `internal/bmad/registry_interactive_phase3c.go`
- `internal/bmad/registry_interactive_phase3d.go`
- `internal/bmad/registry_interactive_phase4.go`
- `internal/bmad/registry_interactive_phase5.go`

# Symbols
- interactive_defaults.go (internal/bmad/interactive_defaults.go:L1)
- applyIterativeUpgrade() (internal/bmad/interactive_defaults.go:L109)
- applyGuidedUpgrade() (internal/bmad/interactive_defaults.go:L144)
- applyPartyUpgrade() (internal/bmad/interactive_defaults.go:L177)
- mergeAcceptTokens() (internal/bmad/interactive_defaults.go:L214)
- buildGate() (internal/bmad/interactive_defaults.go:L233)
- defaultMaxRounds() (internal/bmad/interactive_defaults.go:L242)
- processIndex() (internal/bmad/interactive_defaults.go:L253)
- memoryOutput() (internal/bmad/interactive_defaults.go:L264)
- fileOutput() (internal/bmad/interactive_defaults.go:L271)
- IterativeUpgradeSpec (internal/bmad/interactive_defaults.go:L50)
- GuidedUpgradeSpec (internal/bmad/interactive_defaults.go:L74)
- PartyUpgradeSpec (internal/bmad/interactive_defaults.go:L90)
- interactive_defaults_test.go (internal/bmad/interactive_defaults_test.go:L1)
- TestStoryRollout01_AC2_PreservesArtifactInputsBeforeRoundResponse() (internal/bmad/interactive_defaults_test.go:L126)
- TestStoryRollout01_AC3_PreSetModePreservedForBrainstorming() (internal/bmad/interactive_defaults_test.go:L155)
- freshAutonomousDef() (internal/bmad/interactive_defaults_test.go:L19)
- TestStoryRollout01_AC4_GateAcceptTokenMergeAndDefaults() (internal/bmad/interactive_defaults_test.go:L207)
- TestStoryRollout01_AC5_IterationInputAfterIterativeUpgrade() (internal/bmad/interactive_defaults_test.go:L296)
- TestStoryRollout01_AC6_GuidedUpgrade() (internal/bmad/interactive_defaults_test.go:L324)
- defaultIterativeSpec() (internal/bmad/interactive_defaults_test.go:L38)
- TestStoryRollout01_AC7_PartyUpgrade() (internal/bmad/interactive_defaults_test.go:L412)
- TestStoryRollout01_AC2_IterativeIdempotent() (internal/bmad/interactive_defaults_test.go:L53)
- registry_interactive_phase2.go (internal/bmad/registry_interactive_phase2.go:L1)
- init() (internal/bmad/registry_interactive_phase2.go:L7)
- TestProcessIndex_Known() (internal/bmad/registry_interactive_phase2_helper_test.go:L14)
- TestProcessIndex_UnknownPanics() (internal/bmad/registry_interactive_phase2_helper_test.go:L28)
- registry_interactive_phase3a.go (internal/bmad/registry_interactive_phase3a.go:L1)
- init() (internal/bmad/registry_interactive_phase3a.go:L5)
- registry_interactive_phase3b.go (internal/bmad/registry_interactive_phase3b.go:L1)
- init() (internal/bmad/registry_interactive_phase3b.go:L5)
- registry_interactive_phase3c.go (internal/bmad/registry_interactive_phase3c.go:L1)
- init() (internal/bmad/registry_interactive_phase3c.go:L5)
- registry_interactive_phase3d.go (internal/bmad/registry_interactive_phase3d.go:L1)
- init() (internal/bmad/registry_interactive_phase3d.go:L5)
- registry_interactive_phase4.go (internal/bmad/registry_interactive_phase4.go:L1)
- init() (internal/bmad/registry_interactive_phase4.go:L5)
- registry_interactive_phase5.go (internal/bmad/registry_interactive_phase5.go:L1)
- init() (internal/bmad/registry_interactive_phase5.go:L5)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files

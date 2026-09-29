---
type: Module
title: ProcessByID
description: "Graphify community 2: internal/bmad/artifacts.go, internal/bmad/executor.go, internal/bmad/executor_outputpaths_test.go, internal/bmad/executor_test.go, internal/bmad/registry.go, internal/bmad/regist"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: registry, resource: internal/bmad/registry.go, last_modified: "2026-04-21T09:23:33+10:00", digest: df9f16ce4aa6d2e3 }
  - { id: registry_interactive_phase2_test, resource: internal/bmad/registry_interactive_phase2_test.go, last_modified: "2026-04-28T11:16:27+10:00", digest: 145cf97aa6278c93 }
  - { id: registry_interactive_phase3a_test, resource: internal/bmad/registry_interactive_phase3a_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 3ee6634ddc1cc172 }
  - { id: registry_interactive_phase3b_test, resource: internal/bmad/registry_interactive_phase3b_test.go, last_modified: "2026-04-28T12:11:01+10:00", digest: 17de966ed5dcdd05 }
  - { id: registry_interactive_phase3c_test, resource: internal/bmad/registry_interactive_phase3c_test.go, last_modified: "2026-04-28T12:11:01+10:00", digest: 0d7a12799878eff1 }
  - { id: registry_interactive_phase3d_test, resource: internal/bmad/registry_interactive_phase3d_test.go, last_modified: "2026-04-28T12:11:01+10:00", digest: df8022944c956694 }
  - { id: registry_interactive_phase4_test, resource: internal/bmad/registry_interactive_phase4_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 85bea3b0aca5ce36 }
  - { id: registry_interactive_phase5_test, resource: internal/bmad/registry_interactive_phase5_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: b7a284fda88c6382 }
  - { id: registry_interactive_test, resource: internal/bmad/registry_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 2394d3c33e43a481 }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
  - { id: allowlist, resource: internal/uiadapter/allowlist.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ec96ab6a3976170d }
---

# Files
- `internal/bmad/artifacts.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_test.go`
- `internal/bmad/registry.go`
- `internal/bmad/registry_interactive_phase2_test.go`
- `internal/bmad/registry_interactive_phase3a_test.go`
- `internal/bmad/registry_interactive_phase3b_test.go`
- `internal/bmad/registry_interactive_phase3c_test.go`
- `internal/bmad/registry_interactive_phase3d_test.go`
- `internal/bmad/registry_interactive_phase4_test.go`
- `internal/bmad/registry_interactive_phase5_test.go`
- `internal/bmad/registry_interactive_test.go`
- `internal/bmad/testutil_interactive_test.go`
- `internal/uiadapter/allowlist.go`

# Symbols
- ResolveArtifactPath() (internal/bmad/artifacts.go:L44)
- GetArtifactStatus() (internal/bmad/artifacts.go:L59)
- resolveOutputPaths() (internal/bmad/executor.go:L1660)
- buildContextStringV3() (internal/bmad/executor.go:L2191)
- buildNodeIndex() (internal/bmad/executor.go:L2302)
- topoSort() (internal/bmad/executor.go:L2312)
- TestAC2_UnmappedArtifacts_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L83)
- TestTopoSort_Sequential() (internal/bmad/executor_test.go:L133)
- TestTopoSort_Parallel() (internal/bmad/executor_test.go:L149)
- TestTopoSort_Cycle() (internal/bmad/executor_test.go:L164)
- TestTopoSort_Diamond() (internal/bmad/executor_test.go:L174)
- TestBuildContextStringV3_IncludesTransformData() (internal/bmad/executor_test.go:L1748)
- TestBuildContextStringV3_TruncatesLongData() (internal/bmad/executor_test.go:L1765)
- TestBuildContextStringV3_IncludesArtifactMatching() (internal/bmad/executor_test.go:L1784)
- TestBuildContextStringV3_EmptyTransformData() (internal/bmad/executor_test.go:L1799)
- TestBuildContextStringV3_SkipsNonCompleteTransforms() (internal/bmad/executor_test.go:L1815)
- TestBuildContextStringV3_FilePathResolution() (internal/bmad/executor_test.go:L1830)
- TestBuildContextStringV3_TransformDataPreservedWithRepoPath() (internal/bmad/executor_test.go:L1938)
- TestBuildContextStringV3_EdgeBasedContext_EmptyInputs() (internal/bmad/executor_test.go:L1967)
- TestBuildContextStringV3_EdgeBasedContext_WithFileResolution() (internal/bmad/executor_test.go:L1987)
- TestBuildContextStringV3_EdgeBasedContext_NoDuplicates() (internal/bmad/executor_test.go:L2006)
- TestBuildContextStringV3_EdgeBasedContext_NonConnectedNodeIgnored() (internal/bmad/executor_test.go:L2027)
- TestBuildContextStringV3_EdgeBasedContext_SkipsNonCompleteUpstream() (internal/bmad/executor_test.go:L2044)
- TestGetArtifactStatus_Exists() (internal/bmad/executor_test.go:L2506)
- TestGetArtifactStatus_Missing() (internal/bmad/executor_test.go:L2517)
- TestGetArtifactStatus_UnmappedArtifact() (internal/bmad/executor_test.go:L2525)
- ProcessByID() (internal/bmad/registry.go:L510)
- registry_interactive_phase2_test.go (internal/bmad/registry_interactive_phase2_test.go:L1)
- expectedAcceptTokens() (internal/bmad/registry_interactive_phase2_test.go:L101)
- outputContains() (internal/bmad/registry_interactive_phase2_test.go:L118)
- outputTuples() (internal/bmad/registry_interactive_phase2_test.go:L128)
- TestStoryRollout02_AC1_ProcessesAreIterativeWithAstAdapter() (internal/bmad/registry_interactive_phase2_test.go:L142)
- TestStoryRollout02_AC2_RoundResponseIterationSlot() (internal/bmad/registry_interactive_phase2_test.go:L163)
- TestStoryRollout02_AC3_GatePopulated() (internal/bmad/registry_interactive_phase2_test.go:L194)
- TestStoryRollout02_AC4_OutputSpecsReflectArtifactMapping() (internal/bmad/registry_interactive_phase2_test.go:L222)
- TestStoryRollout02_AC5_RolloutIDsDifferFromPreRolloutGoldens() (internal/bmad/registry_interactive_phase2_test.go:L260)
- outputTuple (internal/bmad/registry_interactive_phase2_test.go:L29)
- TestStoryRollout02_AC5_Phase2RolloutIDsDeclaredInRegistryTest() (internal/bmad/registry_interactive_phase2_test.go:L297)
- extractPhase2RolloutIDsBlock() (internal/bmad/registry_interactive_phase2_test.go:L337)
- rollout02Expect (internal/bmad/registry_interactive_phase2_test.go:L35)
- TestStoryRollout02_BDD_DevStoryIterativeWith10RoundCapAndShipToken() (internal/bmad/registry_interactive_phase2_test.go:L364)
- TestStoryRollout02_BDD_CodeReviewCarriesReviewReportFileOutput() (internal/bmad/registry_interactive_phase2_test.go:L382)
- TestStoryRollout02_BDD_ValidatePrdAcceptsPassAndApproved() (internal/bmad/registry_interactive_phase2_test.go:L392)
- TestStoryRollout02_BDD_EditPrdWritesPrdMdAsFileOutput() (internal/bmad/registry_interactive_phase2_test.go:L401)
- TestStoryRollout02_BDD_QuickDevCapsAt10Rounds() (internal/bmad/registry_interactive_phase2_test.go:L411)
- TestStoryRollout02_BDD_CreateStoryHasOnlyBaselineAcceptTokens() (internal/bmad/registry_interactive_phase2_test.go:L419)
- TestStoryRollout02_BDD_PersistedExecutionStillResumes() (internal/bmad/registry_interactive_phase2_test.go:L470)
- TestStoryRollout02_BDD_RoundResponseIsShapeJSON() (internal/bmad/registry_interactive_phase2_test.go:L503)
- registry_interactive_phase3a_test.go (internal/bmad/registry_interactive_phase3a_test.go:L1)
- phase3aRolloutExpect (internal/bmad/registry_interactive_phase3a_test.go:L18)
- TestStoryRollout03_AC1_AnalysisProcessesAreIterative() (internal/bmad/registry_interactive_phase3a_test.go:L29)
- TestStoryRollout03_AC2_RoundResponseSlot() (internal/bmad/registry_interactive_phase3a_test.go:L43)
- TestStoryRollout03_AC3_GateBaselineOnly() (internal/bmad/registry_interactive_phase3a_test.go:L61)
- TestStoryRollout03_AC4_OutputsAreFileMapped() (internal/bmad/registry_interactive_phase3a_test.go:L82)
- TestStoryRollout03_AC5_SkipUnionIncludesPhase3a() (internal/bmad/registry_interactive_phase3a_test.go:L99)
- registry_interactive_phase3b_test.go (internal/bmad/registry_interactive_phase3b_test.go:L1)
- TestStoryRollout04_AC4_OutputSpecsMapping() (internal/bmad/registry_interactive_phase3b_test.go:L113)
- TestStoryRollout04_AC5_SkipUnionIncludesPhase3b() (internal/bmad/registry_interactive_phase3b_test.go:L131)
- phase3bRolloutExpect (internal/bmad/registry_interactive_phase3b_test.go:L17)
- TestStoryRollout04_AC1_PlanningProcessesAreIterative() (internal/bmad/registry_interactive_phase3b_test.go:L62)
- TestStoryRollout04_AC2_RoundResponseSlot() (internal/bmad/registry_interactive_phase3b_test.go:L76)
- TestStoryRollout04_AC3_GatePerProcess() (internal/bmad/registry_interactive_phase3b_test.go:L93)
- registry_interactive_phase3c_test.go (internal/bmad/registry_interactive_phase3c_test.go:L1)
- TestStoryRollout05_AC1_QaProcessIsIterative() (internal/bmad/registry_interactive_phase3c_test.go:L19)
- TestStoryRollout05_AC2_RoundResponseSlot() (internal/bmad/registry_interactive_phase3c_test.go:L26)
- TestStoryRollout05_AC3_GateBaselineOnly() (internal/bmad/registry_interactive_phase3c_test.go:L35)
- TestStoryRollout05_AC4_MemoryOnlyOutputs() (internal/bmad/registry_interactive_phase3c_test.go:L47)
- TestStoryRollout05_AC5_SkipUnionIncludesPhase3c() (internal/bmad/registry_interactive_phase3c_test.go:L59)
- registry_interactive_phase3d_test.go (internal/bmad/registry_interactive_phase3d_test.go:L1)
- TestStoryRollout06_AC4_OutputSpecsMapping() (internal/bmad/registry_interactive_phase3d_test.go:L102)
- TestStoryRollout06_AC5_SkipUnionIncludesPhase3d() (internal/bmad/registry_interactive_phase3d_test.go:L119)
- phase3dRolloutExpect (internal/bmad/registry_interactive_phase3d_test.go:L17)
- TestStoryRollout06_AC1_SupportProcessesAreIterative() (internal/bmad/registry_interactive_phase3d_test.go:L60)
- TestStoryRollout06_AC2_RoundResponseSlot() (internal/bmad/registry_interactive_phase3d_test.go:L72)
- TestStoryRollout06_AC3_GatePerProcess() (internal/bmad/registry_interactive_phase3d_test.go:L86)
- registry_interactive_phase4_test.go (internal/bmad/registry_interactive_phase4_test.go:L1)
- TestStoryRollout07_AC5_OutputSpecsFileMapped() (internal/bmad/registry_interactive_phase4_test.go:L106)
- TestStoryRollout07_AC6_GateDefaults() (internal/bmad/registry_interactive_phase4_test.go:L122)
- TestStoryRollout07_AC7_SkipUnionIncludesPhase4() (internal/bmad/registry_interactive_phase4_test.go:L137)
- phase4RolloutExpect (internal/bmad/registry_interactive_phase4_test.go:L19)
- TestStoryRollout07_AC1_GuidedProcessesAreGuided() (internal/bmad/registry_interactive_phase4_test.go:L47)
- TestStoryRollout07_AC2_NoIterationSlot() (internal/bmad/registry_interactive_phase4_test.go:L59)
- TestStoryRollout07_AC3_StagedInputsInDeclaredOrder() (internal/bmad/registry_interactive_phase4_test.go:L72)
- TestStoryRollout07_AC4_FinalApprovalAppended() (internal/bmad/registry_interactive_phase4_test.go:L88)
- registry_interactive_phase5_test.go (internal/bmad/registry_interactive_phase5_test.go:L1)
- TestStoryRollout08_AC5_OutputSpecsMapping() (internal/bmad/registry_interactive_phase5_test.go:L110)
- TestStoryRollout08_AC6_OptionalRetroNotes() (internal/bmad/registry_interactive_phase5_test.go:L124)
- TestStoryRollout08_AC7_SkipUnionIncludesPhase5() (internal/bmad/registry_interactive_phase5_test.go:L132)
- phase5RolloutExpect (internal/bmad/registry_interactive_phase5_test.go:L19)
- TestStoryRollout08_AC1_PartyProcessesAreParty() (internal/bmad/registry_interactive_phase5_test.go:L49)
- TestStoryRollout08_AC2_TopicAndMessageInputs() (internal/bmad/registry_interactive_phase5_test.go:L61)
- TestStoryRollout08_AC3_IterationInputIsMessage() (internal/bmad/registry_interactive_phase5_test.go:L81)
- TestStoryRollout08_AC4_GateDefaults() (internal/bmad/registry_interactive_phase5_test.go:L95)
- registry_interactive_test.go (internal/bmad/registry_interactive_test.go:L1)
- TestInteractiveRegistryIterationInput() (internal/bmad/registry_interactive_test.go:L110)
- TestOptionsRefResolution() (internal/bmad/registry_interactive_test.go:L138)
- TestInteractiveRegistryShape() (internal/bmad/registry_interactive_test.go:L18)
- .writeArtifact() (internal/bmad/testutil_interactive_test.go:L262)
- allowlist.go (internal/uiadapter/allowlist.go:L1)
- joinAllowlist() (internal/uiadapter/allowlist.go:L78)

# Depends on
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)
- [bmad/types.go](/modules/bmad-types-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [Executor](/modules/executor.md)

# Inferred
- [go_pkg_strings](/modules/go-pkg-strings.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files

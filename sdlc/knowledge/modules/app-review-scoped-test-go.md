---
type: Module
title: app_review_scoped_test.go
description: "Graphify community 98: app_git_file_test.go, app_review_scoped.go, app_review_scoped_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_git_file_test, resource: app_git_file_test.go, last_modified: "2026-09-30T06:49:47+10:00", digest: e5f67c0708646e56 }
  - { id: app_review_scoped, resource: app_review_scoped.go, last_modified: "2026-09-30T01:02:27+10:00", digest: ff0004675f50a581 }
  - { id: app_review_scoped_test, resource: app_review_scoped_test.go, last_modified: "2026-09-30T01:02:27+10:00", digest: 97225be1ebe84af0 }
---

# Files
- `app_git_file_test.go`
- `app_review_scoped.go`
- `app_review_scoped_test.go`

# Symbols
- TestReadFileAtHead_DeletedTrackedFileStillReadable() (app_git_file_test.go:L44)
- app_review_scoped.go (app_review_scoped.go:L1)
- assembleScopedPayload() (app_review_scoped.go:L106)
- scopedAdviceEvent() (app_review_scoped.go:L115)
- App (app_review_scoped.go:L128)
- .StreamScopedAdvice() (app_review_scoped.go:L128)
- buildScopedDiff() (app_review_scoped.go:L25)
- containedPath() (app_review_scoped.go:L75)
- listTrackedFiles() (app_review_scoped.go:L89)
- app_review_scoped_test.go (app_review_scoped_test.go:L1)
- TestBuildScopedDiff_AC2_UntrackedFallback() (app_review_scoped_test.go:L100)
- TestBuildScopedDiff_SymlinkEscapeSkipped() (app_review_scoped_test.go:L120)
- TestBuildScopedDiff_MixedTrackedUntracked() (app_review_scoped_test.go:L135)
- TestBuildScopedDiff_AC4_EmptyInput() (app_review_scoped_test.go:L161)
- TestBuildScopedDiff_AllFilesUnchanged() (app_review_scoped_test.go:L192)
- initTestGitRepo() (app_review_scoped_test.go:L20)
- TestBuildScopedDiff_RejectsPathTraversal() (app_review_scoped_test.go:L211)
- TestBuildScopedDiff_SeparatorBetweenFiles() (app_review_scoped_test.go:L249)
- TestAssembleScopedPayload_AC3_AdditionalContextPrepended() (app_review_scoped_test.go:L267)
- TestAssembleScopedPayload_EmptyContext() (app_review_scoped_test.go:L324)
- gitRun() (app_review_scoped_test.go:L34)
- TestScopedAdviceEvent_AC5_MatchesStreamAdviceShape() (app_review_scoped_test.go:L345)
- commitFile() (app_review_scoped_test.go:L44)
- modifyTrackedFile() (app_review_scoped_test.go:L53)
- createUntrackedFile() (app_review_scoped_test.go:L60)
- TestBuildScopedDiff_AC1_OnlySelectedFiles() (app_review_scoped_test.go:L70)

# Depends on
- [loader_test.go](/modules/loader-test-go.md)
- [pathguard.go](/modules/pathguard-go.md)

# Inferred
- [App](/modules/app-73.md)
- [app_git_guard_test.go](/modules/app-git-guard-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

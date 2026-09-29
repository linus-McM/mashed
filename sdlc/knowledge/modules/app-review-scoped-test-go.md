---
type: Module
title: app_review_scoped_test.go
description: "Graphify community 46: app_review_scoped.go, app_review_scoped_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app_review_scoped, resource: app_review_scoped.go, last_modified: "2026-04-10T13:49:09+10:00", digest: 83c692114f39a0c9 }
  - { id: app_review_scoped_test, resource: app_review_scoped_test.go, last_modified: "2026-04-10T13:49:09+10:00", digest: 495a403ad8cace57 }
---

# Files
- `app_review_scoped.go`
- `app_review_scoped_test.go`

# Symbols
- scopedAdviceEvent() (app_review_scoped.go:L108)
- App (app_review_scoped.go:L121)
- .StreamScopedAdvice() (app_review_scoped.go:L121)
- buildScopedDiff() (app_review_scoped.go:L24)
- containedPath() (app_review_scoped.go:L68)
- listTrackedFiles() (app_review_scoped.go:L82)
- assembleScopedPayload() (app_review_scoped.go:L99)
- app_review_scoped_test.go (app_review_scoped_test.go:L1)
- TestBuildScopedDiff_MixedTrackedUntracked() (app_review_scoped_test.go:L121)
- TestBuildScopedDiff_AC4_EmptyInput() (app_review_scoped_test.go:L147)
- TestBuildScopedDiff_AllFilesUnchanged() (app_review_scoped_test.go:L178)
- TestBuildScopedDiff_RejectsPathTraversal() (app_review_scoped_test.go:L197)
- initTestGitRepo() (app_review_scoped_test.go:L20)
- TestBuildScopedDiff_SeparatorBetweenFiles() (app_review_scoped_test.go:L235)
- TestAssembleScopedPayload_AC3_AdditionalContextPrepended() (app_review_scoped_test.go:L253)
- TestAssembleScopedPayload_EmptyContext() (app_review_scoped_test.go:L310)
- TestScopedAdviceEvent_AC5_MatchesStreamAdviceShape() (app_review_scoped_test.go:L331)
- gitRun() (app_review_scoped_test.go:L34)
- commitFile() (app_review_scoped_test.go:L43)
- modifyTrackedFile() (app_review_scoped_test.go:L52)
- createUntrackedFile() (app_review_scoped_test.go:L59)
- TestBuildScopedDiff_AC1_OnlySelectedFiles() (app_review_scoped_test.go:L69)
- TestBuildScopedDiff_AC2_UntrackedFallback() (app_review_scoped_test.go:L99)

# Depends on
- [loader_test.go](/modules/loader-test-go.md)

# Inferred
- [App](/modules/app-63.md)

# Features
- no feature plan names these files

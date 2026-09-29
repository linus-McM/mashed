---
type: Module
title: app_review_scoped_test.go
description: "Graphify community 98: app_review_scoped.go, app_review_scoped_test.go, testutil_git_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_review_scoped, resource: app_review_scoped.go, last_modified: "2026-04-10T13:49:09+10:00", digest: 83c692114f39a0c9 }
  - { id: app_review_scoped_test, resource: app_review_scoped_test.go, last_modified: "2026-09-30T00:42:54+10:00", digest: ee0b06c52b831fad }
  - { id: testutil_git_test, resource: testutil_git_test.go, last_modified: "2026-09-30T00:42:54+10:00", digest: 936f0dfbf4954e35 }
---

# Files
- `app_review_scoped.go`
- `app_review_scoped_test.go`
- `testutil_git_test.go`

# Symbols
- scopedAdviceEvent() (app_review_scoped.go:L108)
- App (app_review_scoped.go:L121)
- .StreamScopedAdvice() (app_review_scoped.go:L121)
- buildScopedDiff() (app_review_scoped.go:L24)
- containedPath() (app_review_scoped.go:L68)
- listTrackedFiles() (app_review_scoped.go:L82)
- assembleScopedPayload() (app_review_scoped.go:L99)
- app_review_scoped_test.go (app_review_scoped_test.go:L1)
- TestBuildScopedDiff_AC2_UntrackedFallback() (app_review_scoped_test.go:L100)
- TestBuildScopedDiff_MixedTrackedUntracked() (app_review_scoped_test.go:L122)
- TestBuildScopedDiff_AC4_EmptyInput() (app_review_scoped_test.go:L148)
- TestBuildScopedDiff_AllFilesUnchanged() (app_review_scoped_test.go:L179)
- TestBuildScopedDiff_RejectsPathTraversal() (app_review_scoped_test.go:L198)
- initTestGitRepo() (app_review_scoped_test.go:L20)
- TestBuildScopedDiff_SeparatorBetweenFiles() (app_review_scoped_test.go:L236)
- TestAssembleScopedPayload_AC3_AdditionalContextPrepended() (app_review_scoped_test.go:L254)
- TestAssembleScopedPayload_EmptyContext() (app_review_scoped_test.go:L311)
- TestScopedAdviceEvent_AC5_MatchesStreamAdviceShape() (app_review_scoped_test.go:L332)
- gitRun() (app_review_scoped_test.go:L34)
- commitFile() (app_review_scoped_test.go:L44)
- modifyTrackedFile() (app_review_scoped_test.go:L53)
- createUntrackedFile() (app_review_scoped_test.go:L60)
- TestBuildScopedDiff_AC1_OnlySelectedFiles() (app_review_scoped_test.go:L70)
- cleanGitEnv() (testutil_git_test.go:L24)

# Depends on
- [loader_test.go](/modules/loader-test-go.md)

# Inferred
- [App](/modules/app-73.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

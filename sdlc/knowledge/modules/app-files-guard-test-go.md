---
type: Module
title: app_files_guard_test.go
description: "Graphify community 65: app_files_guard_test.go, app_git_file_test.go, app_git_guard_test.go, testutil_git_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_files_guard_test, resource: app_files_guard_test.go, last_modified: "2026-09-30T00:48:40+10:00", digest: 80bcd44de2130820 }
  - { id: app_git_file_test, resource: app_git_file_test.go, last_modified: "2026-09-30T01:02:27+10:00", digest: cbd68b146dc0b467 }
  - { id: app_git_guard_test, resource: app_git_guard_test.go, last_modified: "2026-09-30T01:02:27+10:00", digest: e9f9b1cdc53663ff }
  - { id: testutil_git_test, resource: testutil_git_test.go, last_modified: "2026-09-30T00:42:54+10:00", digest: 936f0dfbf4954e35 }
---

# Files
- `app_files_guard_test.go`
- `app_git_file_test.go`
- `app_git_guard_test.go`
- `testutil_git_test.go`

# Symbols
- app_files_guard_test.go (app_files_guard_test.go:L1)
- TestFileBindings_AllowInsideRoots() (app_files_guard_test.go:L100)
- TestWriteFile_SymlinkInsideRootPreservesLink() (app_files_guard_test.go:L120)
- guardFixture() (app_files_guard_test.go:L14)
- mustWrite() (app_files_guard_test.go:L39)
- TestFileBindings_RejectOutsideRoots() (app_files_guard_test.go:L46)
- TestWriteFile_RejectsDenylist() (app_files_guard_test.go:L88)
- TestReadFileDiff_RejectsTraversal() (app_git_file_test.go:L16)
- initTestGitRepoAt() (app_git_file_test.go:L71)
- gitOut() (app_git_guard_test.go:L19)
- dirtyRepoInHome() (app_git_guard_test.go:L33)
- TestGitBindings_RejectInvalidRefs() (app_git_guard_test.go:L47)
- TestGitBindings_ValidRefStillWorks() (app_git_guard_test.go:L78)
- TestGitBindings_RejectRepoOutsideRoots() (app_git_guard_test.go:L89)
- testutil_git_test.go (testutil_git_test.go:L1)
- TestMain() (testutil_git_test.go:L13)
- cleanGitEnv() (testutil_git_test.go:L24)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [app_review_scoped_test.go](/modules/app-review-scoped-test-go.md)
- [Config](/modules/config.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

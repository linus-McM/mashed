---
type: Module
title: app_git_guard_test.go
description: "Graphify community 65: app_files_guard_test.go, app_git.go, app_git_file_test.go, app_git_guard_test.go, app_scan_lifecycle_test.go, internal/uiadapter/testmain_git_test.go, readfilebase64_test.go, te"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_files_guard_test, resource: app_files_guard_test.go, last_modified: "2026-09-30T06:49:47+10:00", digest: e98008d261325259 }
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 9681e22dfb0b07f4 }
  - { id: app_git_file_test, resource: app_git_file_test.go, last_modified: "2026-09-30T06:49:47+10:00", digest: e5f67c0708646e56 }
  - { id: app_git_guard_test, resource: app_git_guard_test.go, last_modified: "2026-09-30T01:02:27+10:00", digest: e9f9b1cdc53663ff }
  - { id: app_scan_lifecycle_test, resource: app_scan_lifecycle_test.go, last_modified: "2026-09-30T06:49:47+10:00", digest: e96b860c5391806f }
  - { id: testmain_git_test, resource: internal/uiadapter/testmain_git_test.go, last_modified: "2026-09-30T06:37:42+10:00", digest: df2916e85d9ab5b9 }
  - { id: readfilebase64_test, resource: readfilebase64_test.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 560bdb04bd6aaba6 }
  - { id: testutil_git_test, resource: testutil_git_test.go, last_modified: "2026-09-30T00:42:54+10:00", digest: 936f0dfbf4954e35 }
---

# Files
- `app_files_guard_test.go`
- `app_git.go`
- `app_git_file_test.go`
- `app_git_guard_test.go`
- `app_scan_lifecycle_test.go`
- `internal/uiadapter/testmain_git_test.go`
- `readfilebase64_test.go`
- `testutil_git_test.go`

# Symbols
- app_files_guard_test.go (app_files_guard_test.go:L1)
- TestWriteFile_RejectsDenylistCaseVariants() (app_files_guard_test.go:L103)
- TestFileBindings_AllowInsideRoots() (app_files_guard_test.go:L126)
- guardFixture() (app_files_guard_test.go:L14)
- TestWriteFile_SymlinkInsideRootPreservesLink() (app_files_guard_test.go:L146)
- mustWrite() (app_files_guard_test.go:L39)
- TestFileBindings_RejectOutsideRoots() (app_files_guard_test.go:L46)
- TestWriteFile_RejectsDenylist() (app_files_guard_test.go:L88)
- mimeForExt() (app_git.go:L891)
- app_git_file_test.go (app_git_file_test.go:L1)
- TestReadFileDiff_RejectsTraversal() (app_git_file_test.go:L15)
- TestReadFileAtHead_RejectsAbsolute() (app_git_file_test.go:L34)
- TestReadFileDiff_DashPathCreatesNoFile() (app_git_file_test.go:L59)
- initTestGitRepoAt() (app_git_file_test.go:L70)
- app_git_guard_test.go (app_git_guard_test.go:L1)
- gitOut() (app_git_guard_test.go:L19)
- dirtyRepoInHome() (app_git_guard_test.go:L33)
- TestGitBindings_RejectInvalidRefs() (app_git_guard_test.go:L47)
- TestGitBindings_ValidRefStillWorks() (app_git_guard_test.go:L78)
- TestGitBindings_RejectRepoOutsideRoots() (app_git_guard_test.go:L89)
- appWithDevDir() (app_scan_lifecycle_test.go:L21)
- appWithDevDirCtx() (app_scan_lifecycle_test.go:L28)
- testmain_git_test.go (internal/uiadapter/testmain_git_test.go:L1)
- TestMain() (internal/uiadapter/testmain_git_test.go:L13)
- TestReadFileBase64_AC1_Base64Encoding() (readfilebase64_test.go:L15)
- TestReadFileBase64_AC2_MimeTypes() (readfilebase64_test.go:L40)
- TestReadFileBase64_AC3_FileSizeLimit() (readfilebase64_test.go:L86)
- testutil_git_test.go (testutil_git_test.go:L1)
- TestMain() (testutil_git_test.go:L13)
- cleanGitEnv() (testutil_git_test.go:L24)

# Depends on
- [app_review_scoped_test.go](/modules/app-review-scoped-test-go.md)

# Inferred
- [app_review_scoped_test.go](/modules/app-review-scoped-test-go.md)
- [claudeapi/client.go](/modules/claudeapi-client-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

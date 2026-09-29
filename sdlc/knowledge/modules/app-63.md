---
type: Module
title: App
description: "Graphify community 63: app_git.go, readfilebase64_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_git, resource: app_git.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 6abf9d08d507b1db }
  - { id: readfilebase64_test, resource: readfilebase64_test.go, last_modified: "2026-04-09T11:58:53+10:00", digest: 82165929a5efaa1e }
---

# Files
- `app_git.go`
- `readfilebase64_test.go`

# Symbols
- .GitListBranches() (app_git.go:L172)
- BranchInfo (app_git.go:L19)
- .GitSwitchBranch() (app_git.go:L199)
- .GitCreateBranch() (app_git.go:L221)
- RepoStatusInfo (app_git.go:L25)
- .RepoStatus() (app_git.go:L251)
- RepoChoice (app_git.go:L34)
- App (app_git.go:L42)
- .RepoMtimes() (app_git.go:L42)
- .GitCommit() (app_git.go:L468)
- .GitCommitAndPush() (app_git.go:L523)
- .GitPush() (app_git.go:L541)
- .GitForcePush() (app_git.go:L563)
- .GitPull() (app_git.go:L577)
- .GitMergeInto() (app_git.go:L592)
- .ListRepoChoices() (app_git.go:L61)
- .GitCommitPushAndPR() (app_git.go:L639)
- .ListRepoFiles() (app_git.go:L734)
- .WriteFile() (app_git.go:L757)
- .ReadFile() (app_git.go:L765)
- mimeForExt() (app_git.go:L783)
- .ReadFileBase64() (app_git.go:L805)
- .CreateRepo() (app_git.go:L82)
- .ReadFileDiff() (app_git.go:L828)
- .ReadFileAtHead() (app_git.go:L847)
- .MarkRead() (app_git.go:L863)
- TestReadFileBase64_AC2_MimeTypes() (readfilebase64_test.go:L40)

# Depends on
- [claudeCommand](/modules/claudecommand.md)
- [diff.go](/modules/diff-go.md)
- [ModelInfo](/modules/modelinfo.md)
- [worktree.go](/modules/worktree-go.md)

# Inferred
- [claudeCommand](/modules/claudecommand.md)
- [Executor](/modules/executor.md)

# Features
- no feature plan names these files

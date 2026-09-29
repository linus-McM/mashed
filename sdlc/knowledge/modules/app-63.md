---
type: Module
title: App
description: "Graphify community 63: app_git.go"
resource: .
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_git, resource: app_git.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 6abf9d08d507b1db }
---

# Files
- `app_git.go`

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
- .CreateRepo() (app_git.go:L82)
- .ReadFileDiff() (app_git.go:L828)
- .ReadFileAtHead() (app_git.go:L847)
- .MarkRead() (app_git.go:L863)
- .SpawnPRReview() (app_git.go:L876)

# Depends on
- [claudeCommand](/modules/claudecommand.md)
- [mimeForExt](/modules/mimeforext.md)
- [ModelInfo](/modules/modelinfo.md)
- [ScopedDiff](/modules/scopeddiff.md)
- [worktree.go](/modules/worktree-go.md)

# Inferred
- [claudeCommand](/modules/claudecommand.md)
- [Executor](/modules/executor.md)

# Features
- no feature plan names these files

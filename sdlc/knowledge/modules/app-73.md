---
type: Module
title: App
description: "Graphify community 73: app_claude.go, app_git.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_claude, resource: app_claude.go, last_modified: "2026-04-10T11:36:27+10:00", digest: aff648c4f900ca53 }
  - { id: app_git, resource: app_git.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 6abf9d08d507b1db }
---

# Files
- `app_claude.go`
- `app_git.go`

# Symbols
- claudeCommand() (app_claude.go:L13)
- envWithoutKey() (app_claude.go:L20)
- .GitListBranches() (app_git.go:L172)
- BranchInfo (app_git.go:L19)
- .GitSwitchBranch() (app_git.go:L199)
- .GitCreateBranch() (app_git.go:L221)
- RepoStatusInfo (app_git.go:L25)
- .RepoStatus() (app_git.go:L251)
- .gitCommitCore() (app_git.go:L319)
- RepoChoice (app_git.go:L34)
- App (app_git.go:L42)
- .RepoMtimes() (app_git.go:L42)
- .generateCommitMessage() (app_git.go:L425)
- .GitCommit() (app_git.go:L468)
- .GitCommitStreaming() (app_git.go:L476)
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

# Depends on
- [diff.go](/modules/diff-go.md)
- [go_pkg_testing](/modules/go-pkg-testing.md)
- [ModelInfo](/modules/modelinfo.md)
- [worktree.go](/modules/worktree-go.md)

# Inferred
- [Executor](/modules/executor.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

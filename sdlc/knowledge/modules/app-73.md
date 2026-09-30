---
type: Module
title: App
description: "Graphify community 73: app_claude.go, app_git.go, internal/git/refs.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_claude, resource: app_claude.go, last_modified: "2026-04-10T11:36:27+10:00", digest: aff648c4f900ca53 }
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 9681e22dfb0b07f4 }
  - { id: refs, resource: internal/git/refs.go, last_modified: "2026-09-30T00:58:21+10:00", digest: 33d9ff4117bf01b8 }
---

# Files
- `app_claude.go`
- `app_git.go`
- `internal/git/refs.go`

# Symbols
- claudeCommand() (app_claude.go:L13)
- envWithoutKey() (app_claude.go:L20)
- .MarkRead() (app_git.go:L1006)
- .SpawnPRReview() (app_git.go:L1019)
- isPRNumber() (app_git.go:L1067)
- .GitListBranches() (app_git.go:L177)
- .GitSwitchBranch() (app_git.go:L207)
- BranchInfo (app_git.go:L22)
- .GitCreateBranch() (app_git.go:L236)
- .RepoStatus() (app_git.go:L278)
- RepoStatusInfo (app_git.go:L28)
- .gitCommitCore() (app_git.go:L349)
- RepoChoice (app_git.go:L37)
- App (app_git.go:L45)
- .RepoMtimes() (app_git.go:L45)
- .generateCommitMessage() (app_git.go:L455)
- .GitCommit() (app_git.go:L498)
- .GitCommitStreaming() (app_git.go:L509)
- .GitCommitAndPush() (app_git.go:L562)
- .GitPush() (app_git.go:L583)
- .GitForcePush() (app_git.go:L608)
- .GitPull() (app_git.go:L625)
- .ListRepoChoices() (app_git.go:L64)
- .GitMergeInto() (app_git.go:L643)
- .GitCommitPushAndPR() (app_git.go:L697)
- .GetScopedDiff() (app_git.go:L779)
- .ListRepoFiles() (app_git.go:L801)
- .repoDir() (app_git.go:L834)
- .CreateRepo() (app_git.go:L86)
- repoRelPath() (app_git.go:L942)
- .ReadFileDiff() (app_git.go:L952)
- .ReadFileAtHead() (app_git.go:L983)
- ValidateBranchName() (internal/git/refs.go:L16)

# Depends on
- [diff.go](/modules/diff-go.md)
- [ModelInfo](/modules/modelinfo.md)
- [pathguard.go](/modules/pathguard-go.md)
- [worktree.go](/modules/worktree-go.md)

# Inferred
- [Executor](/modules/executor.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

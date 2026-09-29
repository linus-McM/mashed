---
type: Module
title: App
description: "Graphify community 73: app_claude.go, app_git.go, internal/domain/types.go, internal/git/diff.go, internal/git/diff_test.go, internal/git/refs.go, internal/git/refs_test.go, internal/git/worktree.go,"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_claude, resource: app_claude.go, last_modified: "2026-04-10T11:36:27+10:00", digest: aff648c4f900ca53 }
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-30T01:11:52+10:00", digest: f61e71fb9e1ebdc6 }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: diff, resource: internal/git/diff.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 6abe6e537093b3dc }
  - { id: diff_test, resource: internal/git/diff_test.go, last_modified: "2026-04-02T07:38:07+11:00", digest: a9b7a8c942479bf7 }
  - { id: refs, resource: internal/git/refs.go, last_modified: "2026-09-30T00:58:21+10:00", digest: 33d9ff4117bf01b8 }
  - { id: refs_test, resource: internal/git/refs_test.go, last_modified: "2026-09-30T00:58:21+10:00", digest: 557cff720e95f689 }
  - { id: worktree, resource: internal/git/worktree.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 930f6a3a3c657dad }
  - { id: worktree_test, resource: internal/git/worktree_test.go, last_modified: "2026-04-02T07:38:07+11:00", digest: e527b275a87593df }
  - { id: pathguard, resource: internal/pathguard/pathguard.go, last_modified: "2026-09-30T00:45:50+10:00", digest: 2d1cc617647f15db }
  - { id: pathguard_test, resource: internal/pathguard/pathguard_test.go, last_modified: "2026-09-30T00:45:50+10:00", digest: 28f3a6502ddf38df }
---

# Files
- `app_claude.go`
- `app_git.go`
- `internal/domain/types.go`
- `internal/git/diff.go`
- `internal/git/diff_test.go`
- `internal/git/refs.go`
- `internal/git/refs_test.go`
- `internal/git/worktree.go`
- `internal/git/worktree_test.go`
- `internal/pathguard/pathguard.go`
- `internal/pathguard/pathguard_test.go`

# Symbols
- claudeCommand() (app_claude.go:L13)
- envWithoutKey() (app_claude.go:L20)
- .MarkRead() (app_git.go:L1003)
- .SpawnPRReview() (app_git.go:L1016)
- .GitListBranches() (app_git.go:L174)
- .GitSwitchBranch() (app_git.go:L204)
- BranchInfo (app_git.go:L21)
- .GitCreateBranch() (app_git.go:L233)
- RepoStatusInfo (app_git.go:L27)
- .RepoStatus() (app_git.go:L275)
- .gitCommitCore() (app_git.go:L346)
- RepoChoice (app_git.go:L36)
- App (app_git.go:L44)
- .RepoMtimes() (app_git.go:L44)
- .generateCommitMessage() (app_git.go:L452)
- .GitCommit() (app_git.go:L495)
- .GitCommitStreaming() (app_git.go:L506)
- .GitCommitAndPush() (app_git.go:L559)
- .GitPush() (app_git.go:L580)
- .GitForcePush() (app_git.go:L605)
- .GitPull() (app_git.go:L622)
- .ListRepoChoices() (app_git.go:L63)
- .GitMergeInto() (app_git.go:L640)
- .GitCommitPushAndPR() (app_git.go:L694)
- .GetScopedDiff() (app_git.go:L776)
- .GetWorktrees() (app_git.go:L787)
- .ListRepoFiles() (app_git.go:L798)
- .fileRoots() (app_git.go:L825)
- .repoDir() (app_git.go:L831)
- .CreateRepo() (app_git.go:L84)
- .WriteFile() (app_git.go:L848)
- .ReadFile() (app_git.go:L866)
- .ReadFileBase64() (app_git.go:L910)
- repoRelPath() (app_git.go:L939)
- .ReadFileDiff() (app_git.go:L949)
- .ReadFileAtHead() (app_git.go:L980)
- WorktreeInfo (internal/domain/types.go:L198)
- ScopedDiff (internal/domain/types.go:L205)
- DiffFileStat (internal/domain/types.go:L210)
- parseDiffStatLine() (internal/git/diff.go:L114)
- ScopedDiff() (internal/git/diff.go:L13)
- parseUntrackedFiles() (internal/git/diff.go:L162)
- filterIgnored() (internal/git/diff.go:L38)
- parseDiffStat() (internal/git/diff.go:L82)
- diff_test.go (internal/git/diff_test.go:L1)
- TestParseDiffStatLine() (internal/git/diff_test.go:L7)
- TestParseDiffStatLine_EmptyPath() (internal/git/diff_test.go:L88)
- ValidateBranchName() (internal/git/refs.go:L16)
- TestValidateBranchName() (internal/git/refs_test.go:L8)
- worktree.go (internal/git/worktree.go:L1)
- DetectWorktrees() (internal/git/worktree.go:L12)
- splitWorktreeBlocks() (internal/git/worktree.go:L51)
- parseWorktreeBlock() (internal/git/worktree.go:L80)
- isBranchOrphaned() (internal/git/worktree.go:L99)
- worktree_test.go (internal/git/worktree_test.go:L1)
- TestSplitWorktreeBlocks_WhitespaceOnly() (internal/git/worktree_test.go:L14)
- TestParseWorktreeBlock_Porcelain() (internal/git/worktree_test.go:L21)
- TestSplitWorktreeBlocks_Empty() (internal/git/worktree_test.go:L7)
- TestParseWorktreeBlock_NoWorktreeLine() (internal/git/worktree_test.go:L88)
- pathguard.go (internal/pathguard/pathguard.go:L1)
- within() (internal/pathguard/pathguard.go:L110)
- resolveRoot() (internal/pathguard/pathguard.go:L125)
- HomeRoot() (internal/pathguard/pathguard.go:L23)
- AllowedRoots() (internal/pathguard/pathguard.go:L28)
- ResolveExisting() (internal/pathguard/pathguard.go:L44)
- ResolveForWrite() (internal/pathguard/pathguard.go:L63)
- CheckWriteDenylist() (internal/pathguard/pathguard.go:L90)
- pathguard_test.go (internal/pathguard/pathguard_test.go:L1)
- TestResolveForWrite() (internal/pathguard/pathguard_test.go:L105)
- layout() (internal/pathguard/pathguard_test.go:L11)
- TestResolveForWrite_SymlinkResolvesToTarget() (internal/pathguard/pathguard_test.go:L136)
- TestCheckWriteDenylist() (internal/pathguard/pathguard_test.go:L148)
- TestAllowedRoots() (internal/pathguard/pathguard_test.go:L47)
- TestResolveExisting() (internal/pathguard/pathguard_test.go:L62)
- TestResolveExisting_MissingFile() (internal/pathguard/pathguard_test.go:L98)

# Depends on
- [go_pkg_testing](/modules/go-pkg-testing.md)
- [ModelInfo](/modules/modelinfo.md)
- [WriteFileAtomic](/modules/writefileatomic.md)

# Inferred
- [Executor](/modules/executor.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

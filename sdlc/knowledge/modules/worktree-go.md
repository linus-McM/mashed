---
type: Module
title: worktree.go
description: "Graphify community 350: app_git.go, internal/domain/types.go, internal/git/worktree.go, internal/git/worktree_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 9681e22dfb0b07f4 }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: worktree, resource: internal/git/worktree.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 930f6a3a3c657dad }
  - { id: worktree_test, resource: internal/git/worktree_test.go, last_modified: "2026-04-02T07:38:07+11:00", digest: e527b275a87593df }
---

# Files
- `app_git.go`
- `internal/domain/types.go`
- `internal/git/worktree.go`
- `internal/git/worktree_test.go`

# Symbols
- .GetWorktrees() (app_git.go:L790)
- WorktreeInfo (internal/domain/types.go:L198)
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

# Depends on
- [App](/modules/app-73.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

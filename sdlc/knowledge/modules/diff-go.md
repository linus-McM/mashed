---
type: Module
title: diff.go
description: "Graphify community 293: app_git.go, internal/domain/types.go, internal/git/diff.go, internal/git/diff_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_git, resource: app_git.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 6abf9d08d507b1db }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: diff, resource: internal/git/diff.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 6abe6e537093b3dc }
  - { id: diff_test, resource: internal/git/diff_test.go, last_modified: "2026-04-02T07:38:07+11:00", digest: a9b7a8c942479bf7 }
---

# Files
- `app_git.go`
- `internal/domain/types.go`
- `internal/git/diff.go`
- `internal/git/diff_test.go`

# Symbols
- .GetScopedDiff() (app_git.go:L718)
- ScopedDiff (internal/domain/types.go:L205)
- DiffFileStat (internal/domain/types.go:L210)
- diff.go (internal/git/diff.go:L1)
- parseDiffStatLine() (internal/git/diff.go:L114)
- ScopedDiff() (internal/git/diff.go:L13)
- parseUntrackedFiles() (internal/git/diff.go:L162)
- filterIgnored() (internal/git/diff.go:L38)
- parseDiffStat() (internal/git/diff.go:L82)
- diff_test.go (internal/git/diff_test.go:L1)
- TestParseDiffStatLine() (internal/git/diff_test.go:L7)
- TestParseDiffStatLine_EmptyPath() (internal/git/diff_test.go:L88)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

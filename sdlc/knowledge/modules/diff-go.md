---
type: Module
title: diff.go
description: "Graphify community 310: app_review.go, internal/domain/types.go, internal/git/diff.go, internal/git/diff_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_review, resource: app_review.go, last_modified: "2026-09-30T06:45:21+10:00", digest: f17cdf1f6df7fb1a }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: diff, resource: internal/git/diff.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 6abe6e537093b3dc }
  - { id: diff_test, resource: internal/git/diff_test.go, last_modified: "2026-04-02T07:38:07+11:00", digest: a9b7a8c942479bf7 }
---

# Files
- `app_review.go`
- `internal/domain/types.go`
- `internal/git/diff.go`
- `internal/git/diff_test.go`

# Symbols
- .StreamCodeReviewSummary() (app_review.go:L112)
- runClaudePrompt() (app_review.go:L480)
- isReviewableFile() (app_review.go:L84)
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
- [testing.T](/modules/testing-t.md)

# Inferred
- [App](/modules/app-73.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

---
type: Module
title: claudeCommand
description: "Graphify community 348: app_claude.go, app_git.go, app_review.go, internal/advice/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_claude, resource: app_claude.go, last_modified: "2026-04-10T11:36:27+10:00", digest: aff648c4f900ca53 }
  - { id: app_git, resource: app_git.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 6abf9d08d507b1db }
  - { id: app_review, resource: app_review.go, last_modified: "2026-05-07T18:18:02+10:00", digest: f186f322bd66e914 }
  - { id: types, resource: internal/advice/types.go, last_modified: "2026-04-10T10:10:32+10:00", digest: 91b5fd3d9c0081b2 }
---

# Files
- `app_claude.go`
- `app_git.go`
- `app_review.go`
- `internal/advice/types.go`

# Symbols
- claudeCommand() (app_claude.go:L13)
- envWithoutKey() (app_claude.go:L20)
- .gitCommitCore() (app_git.go:L319)
- .generateCommitMessage() (app_git.go:L425)
- .GitCommitStreaming() (app_git.go:L476)
- App (app_review.go:L105)
- .ListAdviceModes() (app_review.go:L105)
- .StreamCodeReviewSummary() (app_review.go:L112)
- .StreamAdvice() (app_review.go:L237)
- runClaudePrompt() (app_review.go:L459)
- isReviewableFile() (app_review.go:L84)
- advice/types.go (internal/advice/types.go:L1)
- AdviceMode (internal/advice/types.go:L4)

# Depends on
- [app_review_test.go](/modules/app-review-test-go.md)
- [diff.go](/modules/diff-go.md)
- [loader_test.go](/modules/loader-test-go.md)

# Inferred
- [Executor](/modules/executor.md)

# Features
- no feature plan names these files

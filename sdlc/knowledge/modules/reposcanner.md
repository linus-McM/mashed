---
type: Module
title: RepoScanner
description: "Graphify community 453: internal/domain/types.go, internal/scanner/repos.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-05-07T10:33:04+10:00", digest: c8bd28e2f7bd59b8 }
---

# Files
- `internal/domain/types.go`
- `internal/scanner/repos.go`

# Symbols
- AgentSession (internal/domain/types.go:L137)
- RepoInfo (internal/domain/types.go:L146)
- .fetchGitInfo() (internal/scanner/repos.go:L122)
- RepoScanner (internal/scanner/repos.go:L17)
- .InvalidateCache() (internal/scanner/repos.go:L173)
- repoCacheEntry (internal/scanner/repos.go:L25)
- NewRepoScanner() (internal/scanner/repos.go:L31)
- .ScanRepos() (internal/scanner/repos.go:L48)
- .getRepoInfo() (internal/scanner/repos.go:L96)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

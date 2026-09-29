---
type: Module
title: RepoScanner
description: "Graphify community 453: internal/domain/types.go, internal/scanner/processes.go, internal/scanner/repos.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: processes, resource: internal/scanner/processes.go, last_modified: "2026-04-07T10:03:32+10:00", digest: f095332194a18619 }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-05-07T10:33:04+10:00", digest: c8bd28e2f7bd59b8 }
---

# Files
- `internal/domain/types.go`
- `internal/scanner/processes.go`
- `internal/scanner/repos.go`

# Symbols
- AgentSession (internal/domain/types.go:L137)
- RepoInfo (internal/domain/types.go:L146)
- .GetWorkingDir() (internal/scanner/processes.go:L120)
- getWorkingDirDarwin() (internal/scanner/processes.go:L151)
- getWorkingDirLinux() (internal/scanner/processes.go:L176)
- ClaudeCodeProvider (internal/scanner/processes.go:L47)
- .ScanProcesses() (internal/scanner/processes.go:L47)
- .fetchGitInfo() (internal/scanner/repos.go:L122)
- RepoScanner (internal/scanner/repos.go:L17)
- .InvalidateCache() (internal/scanner/repos.go:L173)
- repoCacheEntry (internal/scanner/repos.go:L25)
- .ScanRepos() (internal/scanner/repos.go:L48)
- .getRepoInfo() (internal/scanner/repos.go:L96)

# Depends on
- [parsePSLine](/modules/parsepsline.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

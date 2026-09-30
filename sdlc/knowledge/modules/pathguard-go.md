---
type: Module
title: pathguard.go
description: "Graphify community 80: app_git.go, internal/pathguard/pathguard.go, internal/pathguard/pathguard_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 9681e22dfb0b07f4 }
  - { id: pathguard, resource: internal/pathguard/pathguard.go, last_modified: "2026-09-30T06:30:15+10:00", digest: b3c153ca67f097c2 }
  - { id: pathguard_test, resource: internal/pathguard/pathguard_test.go, last_modified: "2026-09-30T06:30:15+10:00", digest: 31783f95a1fa7632 }
---

# Files
- `app_git.go`
- `internal/pathguard/pathguard.go`
- `internal/pathguard/pathguard_test.go`

# Symbols
- .fileRoots() (app_git.go:L828)
- .WriteFile() (app_git.go:L851)
- .ReadFile() (app_git.go:L869)
- .ReadFileBase64() (app_git.go:L913)
- pathguard.go (internal/pathguard/pathguard.go:L1)
- hasFoldPrefix() (internal/pathguard/pathguard.go:L145)
- within() (internal/pathguard/pathguard.go:L150)
- resolveRoot() (internal/pathguard/pathguard.go:L165)
- HomeRoot() (internal/pathguard/pathguard.go:L23)
- AllowedRoots() (internal/pathguard/pathguard.go:L28)
- ResolveExisting() (internal/pathguard/pathguard.go:L44)
- ResolveForWrite() (internal/pathguard/pathguard.go:L63)
- CheckWriteDenylist() (internal/pathguard/pathguard.go:L98)
- pathguard_test.go (internal/pathguard/pathguard_test.go:L1)
- TestResolveForWrite() (internal/pathguard/pathguard_test.go:L105)
- layout() (internal/pathguard/pathguard_test.go:L11)
- TestResolveForWrite_SymlinkResolvesToTarget() (internal/pathguard/pathguard_test.go:L136)
- TestCheckWriteDenylist() (internal/pathguard/pathguard_test.go:L148)
- TestCheckWriteDenylist_CaseAndAliases() (internal/pathguard/pathguard_test.go:L172)
- TestAllowedRoots() (internal/pathguard/pathguard_test.go:L47)
- TestResolveExisting() (internal/pathguard/pathguard_test.go:L62)
- TestResolveExisting_MissingFile() (internal/pathguard/pathguard_test.go:L98)

# Depends on
- [app_git_guard_test.go](/modules/app-git-guard-test-go.md)
- [WriteFileAtomic](/modules/writefileatomic.md)

# Inferred
- [app_git_guard_test.go](/modules/app-git-guard-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

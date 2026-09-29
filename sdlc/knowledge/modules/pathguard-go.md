---
type: Module
title: pathguard.go
description: "Graphify community 236: internal/pathguard/pathguard.go, internal/pathguard/pathguard_test.go"
resource: internal/pathguard
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: pathguard, resource: internal/pathguard/pathguard.go, last_modified: "2026-09-30T00:45:50+10:00", digest: 2d1cc617647f15db }
  - { id: pathguard_test, resource: internal/pathguard/pathguard_test.go, last_modified: "2026-09-30T00:45:50+10:00", digest: 28f3a6502ddf38df }
---

# Files
- `internal/pathguard/pathguard.go`
- `internal/pathguard/pathguard_test.go`

# Symbols
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
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

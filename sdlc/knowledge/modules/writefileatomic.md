---
type: Module
title: WriteFileAtomic
description: "Graphify community 416: internal/fsutil/atomic.go, internal/fsutil/atomic_test.go"
resource: internal/fsutil
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: atomic, resource: internal/fsutil/atomic.go, last_modified: "2026-09-30T01:08:11+10:00", digest: 7b0fe58799b5232f }
  - { id: atomic_test, resource: internal/fsutil/atomic_test.go, last_modified: "2026-09-30T01:08:11+10:00", digest: 5511de8351e87dc7 }
---

# Files
- `internal/fsutil/atomic.go`
- `internal/fsutil/atomic_test.go`

# Symbols
- WriteFileAtomic() (internal/fsutil/atomic.go:L17)
- atomic_test.go (internal/fsutil/atomic_test.go:L1)
- TestWriteFileAtomic_WritesAndSetsMode() (internal/fsutil/atomic_test.go:L10)
- TestWriteFileAtomic_ExistingTargetKeepsMode() (internal/fsutil/atomic_test.go:L25)
- TestWriteFileAtomic_FailureLeavesOriginalAndNoTemp() (internal/fsutil/atomic_test.go:L42)
- TestWriteFileAtomic_NoTempLeftOnSuccess() (internal/fsutil/atomic_test.go:L70)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

---
type: Module
title: repair.go
description: "Graphify community 202: internal/uiadapter/repair.go, internal/uiadapter/repair_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: repair, resource: internal/uiadapter/repair.go, last_modified: "2026-04-26T10:43:55+10:00", digest: cc21726252779f85 }
  - { id: repair_test, resource: internal/uiadapter/repair_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 7d95d0c21f4ee4e3 }
---

# Files
- `internal/uiadapter/repair.go`
- `internal/uiadapter/repair_test.go`

# Symbols
- repair.go (internal/uiadapter/repair.go:L1)
- .MaxRetries() (internal/uiadapter/repair.go:L118)
- .Run() (internal/uiadapter/repair.go:L124)
- .Metrics() (internal/uiadapter/repair.go:L202)
- firstReason() (internal/uiadapter/repair.go:L23)
- RepairAttempt (internal/uiadapter/repair.go:L39)
- BuildRepairPrompt() (internal/uiadapter/repair.go:L61)
- buildRepairPromptString() (internal/uiadapter/repair.go:L77)
- Repairer (internal/uiadapter/repair.go:L99)
- TestBuildRepairPrompt_Shape() (internal/uiadapter/repair_test.go:L111)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [stages_test.go](/modules/stages-test-go.md)

# Inferred
- [go_pkg_log_slog](/modules/go-pkg-log-slog.md)

# Features
- no feature plan names these files

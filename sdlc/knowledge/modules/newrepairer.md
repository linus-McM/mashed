---
type: Module
title: NewRepairer
description: "Graphify community 202: internal/uiadapter/repair.go, internal/uiadapter/repair_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: repair, resource: internal/uiadapter/repair.go, last_modified: "2026-04-26T10:43:55+10:00", digest: cc21726252779f85 }
  - { id: repair_test, resource: internal/uiadapter/repair_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 7d95d0c21f4ee4e3 }
---

# Files
- `internal/uiadapter/repair.go`
- `internal/uiadapter/repair_test.go`

# Symbols
- NewRepairer() (internal/uiadapter/repair.go:L109)
- .MaxRetries() (internal/uiadapter/repair.go:L118)
- .Run() (internal/uiadapter/repair.go:L124)
- .Metrics() (internal/uiadapter/repair.go:L202)
- firstReason() (internal/uiadapter/repair.go:L23)
- RepairAttempt (internal/uiadapter/repair.go:L39)
- BuildRepairPrompt() (internal/uiadapter/repair.go:L61)
- buildRepairPromptString() (internal/uiadapter/repair.go:L77)
- Repairer (internal/uiadapter/repair.go:L99)
- repair_test.go (internal/uiadapter/repair_test.go:L1)
- TestBuildRepairPrompt_Shape() (internal/uiadapter/repair_test.go:L111)
- TestRepair_PropagatesGenerateError() (internal/uiadapter/repair_test.go:L129)
- TestRepair_RecoveryRate() (internal/uiadapter/repair_test.go:L17)
- TestRepair_NeverExceedsBudget() (internal/uiadapter/repair_test.go:L50)
- TestRepair_FastPathNoWastedCalls() (internal/uiadapter/repair_test.go:L71)
- TestRepair_Gated_PerBackend() (internal/uiadapter/repair_test.go:L92)

# Depends on
- [Config](/modules/config.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

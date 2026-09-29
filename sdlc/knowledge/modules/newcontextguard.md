---
type: Module
title: NewContextGuard
description: "Graphify community 294: internal/uiadapter/contextguard.go, internal/uiadapter/contextguard_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: contextguard, resource: internal/uiadapter/contextguard.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 417a5e0974dc194b }
  - { id: contextguard_test, resource: internal/uiadapter/contextguard_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 0cf07afee9915c31 }
---

# Files
- `internal/uiadapter/contextguard.go`
- `internal/uiadapter/contextguard_test.go`

# Symbols
- truncateToBudget() (internal/uiadapter/contextguard.go:L105)
- .ApplyClaude() (internal/uiadapter/contextguard.go:L135)
- .OllamaOptions() (internal/uiadapter/contextguard.go:L176)
- ContextGuard (internal/uiadapter/contextguard.go:L20)
- NewContextGuard() (internal/uiadapter/contextguard.go:L41)
- .ApplyOllama() (internal/uiadapter/contextguard.go:L55)
- contextguard_test.go (internal/uiadapter/contextguard_test.go:L1)
- TestContextGuard_TruncatesLongCapture_Ollama() (internal/uiadapter/contextguard_test.go:L15)
- TestContextGuard_NoTruncationUnderBudget() (internal/uiadapter/contextguard_test.go:L30)
- TestContextGuard_OllamaOptions() (internal/uiadapter/contextguard_test.go:L42)
- TestContextGuard_RefusesLongCapture_Claude() (internal/uiadapter/contextguard_test.go:L53)
- TestContextGuard_AcceptsShortClaude() (internal/uiadapter/contextguard_test.go:L68)
- TestContextGuard_ZeroConfigDefaults() (internal/uiadapter/contextguard_test.go:L80)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

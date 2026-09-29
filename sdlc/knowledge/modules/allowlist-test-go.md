---
type: Module
title: allowlist_test.go
description: "Graphify community 181: internal/uiadapter/allowlist.go, internal/uiadapter/allowlist_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: allowlist, resource: internal/uiadapter/allowlist.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ec96ab6a3976170d }
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 865e6c967ac886df }
---

# Files
- `internal/uiadapter/allowlist.go`
- `internal/uiadapter/allowlist_test.go`

# Symbols
- allowlist.go (internal/uiadapter/allowlist.go:L1)
- CheckModelAllowlist() (internal/uiadapter/allowlist.go:L40)
- joinAllowlist() (internal/uiadapter/allowlist.go:L78)
- allowlist_test.go (internal/uiadapter/allowlist_test.go:L1)
- TestStory5_AC6_AllowlistNilLoggerStillSafe() (internal/uiadapter/allowlist_test.go:L110)
- TestAllowlist_WarnsOnUnvetted() (internal/uiadapter/allowlist_test.go:L18)
- TestAllowlist_AllowsKnownModels() (internal/uiadapter/allowlist_test.go:L32)
- TestAllowlist_OverrideFlag() (internal/uiadapter/allowlist_test.go:L44)
- TestAllowlist_WarnsOnUnvettedClaude() (internal/uiadapter/allowlist_test.go:L57)

# Depends on
- [Config](/modules/config.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [testing.T](/modules/testing-t.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files

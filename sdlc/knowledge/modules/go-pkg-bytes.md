---
type: Module
title: go_pkg_bytes
description: "Graphify community 305: internal/uiadapter/allowlist.go, internal/uiadapter/allowlist_test.go, internal/uiadapter/mock.go, internal/uiadapter/mock_test.go, internal/uiadapter/schema.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: allowlist, resource: internal/uiadapter/allowlist.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ec96ab6a3976170d }
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 865e6c967ac886df }
  - { id: mock, resource: internal/uiadapter/mock.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ee77ed81915c20b0 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e6b7941070a1bd3a }
  - { id: schema, resource: internal/uiadapter/schema.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 31cb03cd04da1f49 }
---

# Files
- `internal/uiadapter/allowlist.go`
- `internal/uiadapter/allowlist_test.go`
- `internal/uiadapter/mock.go`
- `internal/uiadapter/mock_test.go`
- `internal/uiadapter/schema.go`

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
- TestStory5_AC6_AllowlistDefaultRemoved() (internal/uiadapter/allowlist_test.go:L93)
- mock.go (internal/uiadapter/mock.go:L1)
- MockAdapter (internal/uiadapter/mock.go:L19)
- NewMock() (internal/uiadapter/mock.go:L31)
- fixtureName() (internal/uiadapter/mock.go:L45)
- .Translate() (internal/uiadapter/mock.go:L58)
- mock_test.go (internal/uiadapter/mock_test.go:L1)
- runGoList() (internal/uiadapter/mock_test.go:L121)
- TestU2_AC8_MockAdapter_FixedReturn() (internal/uiadapter/mock_test.go:L19)
- TestU2_AC8_MockAdapter_NilReturnsFallback() (internal/uiadapter/mock_test.go:L33)
- TestU2_AC8_MockAdapter_NotInProductionBuild() (internal/uiadapter/mock_test.go:L46)
- schema.go (internal/uiadapter/schema.go:L1)
- .UnmarshalJSON() (internal/uiadapter/schema.go:L112)
- UIAST (internal/uiadapter/schema.go:L12)
- Diagnostics (internal/uiadapter/schema.go:L25)
- UINode (internal/uiadapter/schema.go:L38)
- WidgetNode (internal/uiadapter/schema.go:L58)
- WidgetOption (internal/uiadapter/schema.go:L76)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files

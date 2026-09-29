---
type: Module
title: NewMock
description: "Graphify community 391: internal/uiadapter/logging_plumbing_mock_test.go, internal/uiadapter/mock.go, internal/uiadapter/mock_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: logging_plumbing_mock_test, resource: internal/uiadapter/logging_plumbing_mock_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: bce0a1e9685e2603 }
  - { id: mock, resource: internal/uiadapter/mock.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ee77ed81915c20b0 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e6b7941070a1bd3a }
---

# Files
- `internal/uiadapter/logging_plumbing_mock_test.go`
- `internal/uiadapter/mock.go`
- `internal/uiadapter/mock_test.go`

# Symbols
- TestStory2_AC2_AC1_NewMockAcceptsNilLogger() (internal/uiadapter/logging_plumbing_mock_test.go:L19)
- mock.go (internal/uiadapter/mock.go:L1)
- MockAdapter (internal/uiadapter/mock.go:L19)
- NewMock() (internal/uiadapter/mock.go:L31)
- fixtureName() (internal/uiadapter/mock.go:L45)
- .Translate() (internal/uiadapter/mock.go:L58)
- TestU2_AC8_MockAdapter_FixedReturn() (internal/uiadapter/mock_test.go:L19)
- TestU2_AC8_MockAdapter_NilReturnsFallback() (internal/uiadapter/mock_test.go:L33)

# Depends on
- [NewDefault](/modules/newdefault.md)

# Inferred
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

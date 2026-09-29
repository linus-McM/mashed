---
type: Module
title: NewMock
description: "Graphify community 414: internal/uiadapter/logging_plumbing_mock_test.go, internal/uiadapter/mock.go, internal/uiadapter/mock_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: logging_plumbing_mock_test, resource: internal/uiadapter/logging_plumbing_mock_test.go, last_modified: "2026-09-29T07:07:25Z", digest: bce0a1e9685e2603 }
  - { id: mock, resource: internal/uiadapter/mock.go, last_modified: "2026-09-29T07:07:25Z", digest: ee77ed81915c20b0 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-09-29T07:07:25Z", digest: e6b7941070a1bd3a }
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
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [FallbackAST](/modules/fallbackast.md)
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files

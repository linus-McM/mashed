---
type: Module
title: fastpath.go
description: "Graphify community 6: internal/uiadapter/backend/backend.go, internal/uiadapter/backend/lifecycle.go, internal/uiadapter/backend/router.go, internal/uiadapter/fastpath.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e57a307fa43599dd }
  - { id: router, resource: internal/uiadapter/backend/router.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 456b1b3b77e3a38e }
  - { id: fastpath, resource: internal/uiadapter/fastpath.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 4ce7a34b961d63bc }
---

# Files
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/router.go`
- `internal/uiadapter/fastpath.go`

# Symbols
- LLMBackend (internal/uiadapter/backend/backend.go:L31)
- .healthProbeAll() (internal/uiadapter/backend/lifecycle.go:L115)
- Lifecycle (internal/uiadapter/backend/lifecycle.go:L15)
- .WarmUpAll() (internal/uiadapter/backend/lifecycle.go:L42)
- .WarmUpState() (internal/uiadapter/backend/lifecycle.go:L69)
- .StartHealthTicker() (internal/uiadapter/backend/lifecycle.go:L86)
- .decideCostAware() (internal/uiadapter/backend/router.go:L116)
- .matchesPrivacy() (internal/uiadapter/backend/router.go:L126)
- defaultPrivacyPatterns() (internal/uiadapter/backend/router.go:L136)
- .Resolve() (internal/uiadapter/backend/router.go:L152)
- .Dispatch() (internal/uiadapter/backend/router.go:L162)
- Router (internal/uiadapter/backend/router.go:L30)
- .SetHealth() (internal/uiadapter/backend/router.go:L58)
- Decision (internal/uiadapter/backend/router.go:L70)
- .Decide() (internal/uiadapter/backend/router.go:L79)
- fastpath.go (internal/uiadapter/fastpath.go:L1)
- .HitRate() (internal/uiadapter/fastpath.go:L108)
- .HitsPerRule() (internal/uiadapter/fastpath.go:L117)
- defaultFastRules() (internal/uiadapter/fastpath.go:L127)
- buildYN() (internal/uiadapter/fastpath.go:L157)
- buildPressEnter() (internal/uiadapter/fastpath.go:L172)
- buildFileConfirm() (internal/uiadapter/fastpath.go:L186)
- buildNumberedMenu() (internal/uiadapter/fastpath.go:L202)
- buildFreeText() (internal/uiadapter/fastpath.go:L241)
- firstNonEmptyLine() (internal/uiadapter/fastpath.go:L256)
- FastPathClassifier (internal/uiadapter/fastpath.go:L26)
- truncate() (internal/uiadapter/fastpath.go:L266)
- fastRule (internal/uiadapter/fastpath.go:L35)
- .Classify() (internal/uiadapter/fastpath.go:L56)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [go_pkg_testing](/modules/go-pkg-testing.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

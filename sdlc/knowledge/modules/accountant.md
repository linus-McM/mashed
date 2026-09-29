---
type: Module
title: Accountant
description: "Graphify community 314: internal/uiadapter/accountant.go, internal/uiadapter/accountant_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: accountant, resource: internal/uiadapter/accountant.go, last_modified: "2026-04-23T11:29:34+10:00", digest: ae4389c89ff37e93 }
  - { id: accountant_test, resource: internal/uiadapter/accountant_test.go, last_modified: "2026-04-23T11:29:34+10:00", digest: 60c344b0cdff2ce8 }
---

# Files
- `internal/uiadapter/accountant.go`
- `internal/uiadapter/accountant_test.go`

# Symbols
- accountant.go (internal/uiadapter/accountant.go:L1)
- .Record() (internal/uiadapter/accountant.go:L120)
- .Snapshot() (internal/uiadapter/accountant.go:L130)
- .expireLocked() (internal/uiadapter/accountant.go:L143)
- Pricing (internal/uiadapter/accountant.go:L19)
- PricingFor() (internal/uiadapter/accountant.go:L32)
- Usage (internal/uiadapter/accountant.go:L45)
- CostUSD() (internal/uiadapter/accountant.go:L54)
- Accountant (internal/uiadapter/accountant.go:L68)
- tokenTick (internal/uiadapter/accountant.go:L77)
- .CheckPrecall() (internal/uiadapter/accountant.go:L91)
- TestAccountant_CostMatchesBilling() (internal/uiadapter/accountant_test.go:L13)
- TestAccountant_UnknownModelIsFree() (internal/uiadapter/accountant_test.go:L90)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

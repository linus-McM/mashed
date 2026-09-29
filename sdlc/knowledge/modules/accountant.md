---
type: Module
title: Accountant
description: "Graphify community 255: internal/uiadapter/accountant.go, internal/uiadapter/accountant_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
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
- [Config](/modules/config.md)
- [go_pkg_testing](/modules/go-pkg-testing.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

---
type: Module
title: bridge_auth_test.go
description: "Graphify community 150: internal/terminal/bridge_auth_dev_test.go, internal/terminal/bridge_auth_test.go, internal/terminal/bridge_origin_prod_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: bridge_auth_dev_test, resource: internal/terminal/bridge_auth_dev_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 55bf715a2b15fa81 }
  - { id: bridge_auth_test, resource: internal/terminal/bridge_auth_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: e232c80fb7794ab7 }
  - { id: bridge_origin_prod_test, resource: internal/terminal/bridge_origin_prod_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 8eb122b662553df1 }
---

# Files
- `internal/terminal/bridge_auth_dev_test.go`
- `internal/terminal/bridge_auth_test.go`
- `internal/terminal/bridge_origin_prod_test.go`

# Symbols
- TestBridge_DevAcceptsLocalhost34115() (internal/terminal/bridge_auth_dev_test.go:L13)
- bridge_auth_test.go (internal/terminal/bridge_auth_test.go:L1)
- TestBridge_TokenNotLogged() (internal/terminal/bridge_auth_test.go:L117)
- TestBridge_ValidTokenAndOriginUpgrades() (internal/terminal/bridge_auth_test.go:L133)
- failingReader (internal/terminal/bridge_auth_test.go:L145)
- .Read() (internal/terminal/bridge_auth_test.go:L147)
- authHeader() (internal/terminal/bridge_auth_test.go:L25)
- getWS() (internal/terminal/bridge_auth_test.go:L34)
- authBridge() (internal/terminal/bridge_auth_test.go:L51)
- TestBridge_RejectsMissingToken() (internal/terminal/bridge_auth_test.go:L60)
- TestBridge_RejectsWrongToken() (internal/terminal/bridge_auth_test.go:L67)
- TestBridge_UnknownSessionWithoutToken403() (internal/terminal/bridge_auth_test.go:L77)
- TestBridge_RejectsForeignOrigin() (internal/terminal/bridge_auth_test.go:L85)
- TestBridge_RejectsEmptyOrigin() (internal/terminal/bridge_auth_test.go:L92)
- TestBridge_RejectsForeignHost() (internal/terminal/bridge_auth_test.go:L99)
- TestBridge_ProdRejectsDevOrigin() (internal/terminal/bridge_origin_prod_test.go:L13)

# Depends on
- [Bridge](/modules/bridge.md)
- [bridge_test.go](/modules/bridge-test-go.md)

# Inferred
- [bridge_test.go](/modules/bridge-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

---
type: Module
title: GetModules
description: "Graphify community 313: app_bmad.go, internal/bmad/modules.go, internal/bmad/modules_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: modules, resource: internal/bmad/modules.go, last_modified: "2026-04-08T22:05:53+10:00", digest: 1ff35ac25389335c }
  - { id: modules_test, resource: internal/bmad/modules_test.go, last_modified: "2026-04-08T22:05:53+10:00", digest: 8e080b822e24b103 }
---

# Files
- `app_bmad.go`
- `internal/bmad/modules.go`
- `internal/bmad/modules_test.go`

# Symbols
- .GetBmadModules() (app_bmad.go:L443)
- modules.go (internal/bmad/modules.go:L1)
- GetModules() (internal/bmad/modules.go:L24)
- ModuleDef (internal/bmad/modules.go:L4)
- modules_test.go (internal/bmad/modules_test.go:L1)
- TestGetModules_Count() (internal/bmad/modules_test.go:L10)
- TestGetModules_UnpopulatedModulesEmpty() (internal/bmad/modules_test.go:L110)
- TestGetModules_CoreHas25Processes() (internal/bmad/modules_test.go:L15)
- TestGetModules_CoreProcessIDsExistInRegistry() (internal/bmad/modules_test.go:L28)
- TestGetModules_UniqueIDs() (internal/bmad/modules_test.go:L42)
- TestAC2_ModulesPopulated_CIS() (internal/bmad/modules_test.go:L52)
- TestAC2_ModulesPopulated_BMGD() (internal/bmad/modules_test.go:L68)
- TestAC2_ModulesPopulated_CoreIncludesNew() (internal/bmad/modules_test.go:L84)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files

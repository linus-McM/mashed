---
type: Module
title: main.go
description: "Graphify community 64: app_uiadapter.go, internal/bmad/registry_fs.go, internal/uiadapter/encode.go, main.go, main_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: registry_fs, resource: internal/bmad/registry_fs.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 144ec4b753400fe8 }
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-04-26T11:30:52+10:00", digest: bf82e888f12311d6 }
  - { id: main, resource: main.go, last_modified: "2026-04-26T09:22:14+10:00", digest: c9962291a4cdea89 }
  - { id: main_test, resource: main_test.go, last_modified: "2026-04-08T23:16:35+10:00", digest: e65839083faadcd9 }
---

# Files
- `app_uiadapter.go`
- `internal/bmad/registry_fs.go`
- `internal/uiadapter/encode.go`
- `main.go`
- `main_test.go`

# Symbols
- app_uiadapter.go (app_uiadapter.go:L1)
- tagsListResponse (app_uiadapter.go:L141)
- init() (app_uiadapter.go:L34)
- registry_fs.go (internal/bmad/registry_fs.go:L1)
- encode.go (internal/uiadapter/encode.go:L1)
- main.go (main.go:L1)
- resolveHelperPath() (main.go:L103)
- waitForSocket() (main.go:L121)
- main() (main.go:L132)
- buildMenu() (main.go:L30)
- main_test.go (main_test.go:L1)
- TestBuildMenu_AC4_ViewSubmenu() (main_test.go:L116)
- menuItem (main_test.go:L13)
- TestBuildMenu_AC5_HelpSubmenu() (main_test.go:L137)
- assertMenuItems() (main_test.go:L20)
- TestBuildMenu_AC1_FiveMenus() (main_test.go:L43)
- TestBuildMenu_AC1_MashedSubmenu() (main_test.go:L73)
- TestBuildMenu_AC2_FileSubmenu() (main_test.go:L94)

# Depends on
- [loadConfig](/modules/loadconfig.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [server_test.go](/modules/server-test-go.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Inferred
- [NewApp](/modules/newapp.md)

# Features
- no feature plan names these files

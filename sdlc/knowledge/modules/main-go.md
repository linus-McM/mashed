---
type: Module
title: main.go
description: "Graphify community 64: app.go, app_terminal_registry_test.go, internal/terminal/panes.go, internal/terminal/tmux_adapter_coverage_test.go, internal/terminal/tmux_escape.go, main.go, main_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-29T07:07:25Z", digest: 295875db4f0bdc1e }
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 2f85a0b04c1b8ba7 }
  - { id: panes, resource: internal/terminal/panes.go, last_modified: "2026-09-29T07:07:25Z", digest: 45986d36cbef9af3 }
  - { id: tmux_adapter_coverage_test, resource: internal/terminal/tmux_adapter_coverage_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 24b353b6c1480a46 }
  - { id: tmux_escape, resource: internal/terminal/tmux_escape.go, last_modified: "2026-09-29T07:07:25Z", digest: d61b1108f3251770 }
  - { id: main, resource: main.go, last_modified: "2026-09-29T07:07:25Z", digest: c9962291a4cdea89 }
  - { id: main_test, resource: main_test.go, last_modified: "2026-09-29T07:07:25Z", digest: e65839083faadcd9 }
---

# Files
- `app.go`
- `app_terminal_registry_test.go`
- `internal/terminal/panes.go`
- `internal/terminal/tmux_adapter_coverage_test.go`
- `internal/terminal/tmux_escape.go`
- `main.go`
- `main_test.go`

# Symbols
- NewApp() (app.go:L232)
- TestStory1_NewAppInitializesMap() (app_terminal_registry_test.go:L327)
- NewPaneDiscovery() (internal/terminal/panes.go:L67)
- TestIsTmuxAvailable_ReturnsBoolWithoutPanic() (internal/terminal/tmux_adapter_coverage_test.go:L55)
- tmux_escape.go (internal/terminal/tmux_escape.go:L1)
- IsTmuxAvailable() (internal/terminal/tmux_escape.go:L67)
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
- [App](/modules/app.md)
- [Bridge](/modules/bridge.md)
- [manager_test.go](/modules/manager-test-go.md)
- [server_test.go](/modules/server-test-go.md)
- [sync.Mutex](/modules/sync-mutex.md)
- [tmux_adapter_coverage_test.go](/modules/tmux-adapter-coverage-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

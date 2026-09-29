---
type: Module
title: NewApp
description: "Graphify community 465: app.go, app_terminal_registry_test.go, internal/terminal/tmux_adapter_coverage_test.go, internal/terminal/tmux_escape.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
  - { id: tmux_adapter_coverage_test, resource: internal/terminal/tmux_adapter_coverage_test.go, last_modified: "2026-04-11T19:58:57+10:00", digest: 24b353b6c1480a46 }
  - { id: tmux_escape, resource: internal/terminal/tmux_escape.go, last_modified: "2026-04-10T16:37:58+10:00", digest: d61b1108f3251770 }
---

# Files
- `app.go`
- `app_terminal_registry_test.go`
- `internal/terminal/tmux_adapter_coverage_test.go`
- `internal/terminal/tmux_escape.go`

# Symbols
- NewApp() (app.go:L232)
- TestStory1_NewAppInitializesMap() (app_terminal_registry_test.go:L327)
- TestIsTmuxAvailable_ReturnsBoolWithoutPanic() (internal/terminal/tmux_adapter_coverage_test.go:L55)
- IsTmuxAvailable() (internal/terminal/tmux_escape.go:L67)

# Depends on
- [App](/modules/app.md)
- [Bridge](/modules/bridge.md)
- [manager_test.go](/modules/manager-test-go.md)
- [server_test.go](/modules/server-test-go.md)
- [tmux_adapter_coverage_test.go](/modules/tmux-adapter-coverage-test-go.md)
- [TmuxPane](/modules/tmuxpane.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

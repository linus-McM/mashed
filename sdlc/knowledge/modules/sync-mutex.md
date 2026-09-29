---
type: Module
title: sync.Mutex
description: "Graphify community 166: app.go, app_terminal_registry_test.go, internal/bmad/executor_session_test.go, internal/terminal/helper/client.go, internal/terminal/panes.go, internal/terminal/tmux_adapter_co"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 774e87e5f070ece0 }
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
  - { id: executor_session_test, resource: internal/bmad/executor_session_test.go, last_modified: "2026-04-12T15:24:29+10:00", digest: e8e25bf75a4a8c21 }
  - { id: client, resource: internal/terminal/helper/client.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 46fc572e21238bfb }
  - { id: panes, resource: internal/terminal/panes.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 45986d36cbef9af3 }
  - { id: tmux_adapter_coverage_test, resource: internal/terminal/tmux_adapter_coverage_test.go, last_modified: "2026-04-11T19:58:57+10:00", digest: 24b353b6c1480a46 }
  - { id: tmux_escape, resource: internal/terminal/tmux_escape.go, last_modified: "2026-04-10T16:37:58+10:00", digest: d61b1108f3251770 }
---

# Files
- `app.go`
- `app_terminal_registry_test.go`
- `internal/bmad/executor_session_test.go`
- `internal/terminal/helper/client.go`
- `internal/terminal/panes.go`
- `internal/terminal/tmux_adapter_coverage_test.go`
- `internal/terminal/tmux_escape.go`

# Symbols
- NewApp() (app.go:L289)
- fakePaneDiscovery (app_terminal_registry_test.go:L20)
- .ListPanes() (app_terminal_registry_test.go:L26)
- .InvalidateCache() (app_terminal_registry_test.go:L30)
- .FindPaneForPID() (app_terminal_registry_test.go:L32)
- TestStory1_NewAppInitializesMap() (app_terminal_registry_test.go:L327)
- capturedArgv (internal/bmad/executor_session_test.go:L18)
- .record() (internal/bmad/executor_session_test.go:L23)
- .last() (internal/bmad/executor_session_test.go:L31)
- TestSpawnCommandSession_RegressionFromRefactor() (internal/bmad/executor_session_test.go:L81)
- .Close() (internal/terminal/helper/client.go:L106)
- Client (internal/terminal/helper/client.go:L16)
- .Kill() (internal/terminal/helper/client.go:L98)
- .InvalidateCache() (internal/terminal/panes.go:L127)
- discoverPanes() (internal/terminal/panes.go:L133)
- getParentPID() (internal/terminal/panes.go:L185)
- TmuxPane (internal/terminal/panes.go:L40)
- .Target() (internal/terminal/panes.go:L50)
- PaneDiscovery (internal/terminal/panes.go:L60)
- NewPaneDiscovery() (internal/terminal/panes.go:L67)
- .ListPanes() (internal/terminal/panes.go:L72)
- .FindPaneForPID() (internal/terminal/panes.go:L93)
- TestIsTmuxAvailable_ReturnsBoolWithoutPanic() (internal/terminal/tmux_adapter_coverage_test.go:L55)
- IsTmuxAvailable() (internal/terminal/tmux_escape.go:L67)

# Depends on
- [App](/modules/app-109.md)
- [Bridge](/modules/bridge.md)
- [manager_test.go](/modules/manager-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [server_test.go](/modules/server-test-go.md)
- [tmux_adapter_coverage_test.go](/modules/tmux-adapter-coverage-test-go.md)

# Inferred
- [newHarness](/modules/newharness.md)
- [server_test.go](/modules/server-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

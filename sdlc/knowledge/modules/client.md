---
type: Module
title: Client
description: "Graphify community 166: app.go, app_terminal_registry_test.go, internal/terminal/helper/client.go, internal/terminal/panes.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
  - { id: client, resource: internal/terminal/helper/client.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 46fc572e21238bfb }
  - { id: panes, resource: internal/terminal/panes.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 45986d36cbef9af3 }
---

# Files
- `app.go`
- `app_terminal_registry_test.go`
- `internal/terminal/helper/client.go`
- `internal/terminal/panes.go`

# Symbols
- NewApp() (app.go:L232)
- fakePaneDiscovery (app_terminal_registry_test.go:L20)
- .ListPanes() (app_terminal_registry_test.go:L26)
- .InvalidateCache() (app_terminal_registry_test.go:L30)
- .FindPaneForPID() (app_terminal_registry_test.go:L32)
- TestStory1_NewAppInitializesMap() (app_terminal_registry_test.go:L327)
- .Close() (internal/terminal/helper/client.go:L106)
- Client (internal/terminal/helper/client.go:L16)
- .InvalidateCache() (internal/terminal/panes.go:L127)
- discoverPanes() (internal/terminal/panes.go:L133)
- getParentPID() (internal/terminal/panes.go:L185)
- TmuxPane (internal/terminal/panes.go:L40)
- .Target() (internal/terminal/panes.go:L50)
- PaneDiscovery (internal/terminal/panes.go:L60)
- NewPaneDiscovery() (internal/terminal/panes.go:L67)
- .ListPanes() (internal/terminal/panes.go:L72)
- .FindPaneForPID() (internal/terminal/panes.go:L93)

# Depends on
- [App](/modules/app-109.md)
- [Bridge](/modules/bridge.md)
- [manager_test.go](/modules/manager-test-go.md)
- [server_test.go](/modules/server-test-go.md)
- [tmux_adapter_coverage_test.go](/modules/tmux-adapter-coverage-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

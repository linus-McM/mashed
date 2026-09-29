---
type: Module
title: main.go
description: "Graphify community 143: app.go, app_terminal_registry_test.go, internal/bmad/registry_fs.go, internal/terminal/tmux_adapter_coverage_test.go, internal/terminal/tmux_escape.go, main.go, main_socket_tes"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T06:52:18+10:00", digest: 19038d72636ae53b }
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-09-30T06:45:21+10:00", digest: 2844ed2cca730a69 }
  - { id: registry_fs, resource: internal/bmad/registry_fs.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 144ec4b753400fe8 }
  - { id: tmux_adapter_coverage_test, resource: internal/terminal/tmux_adapter_coverage_test.go, last_modified: "2026-04-11T19:58:57+10:00", digest: 24b353b6c1480a46 }
  - { id: tmux_escape, resource: internal/terminal/tmux_escape.go, last_modified: "2026-04-10T16:37:58+10:00", digest: d61b1108f3251770 }
  - { id: main, resource: main.go, last_modified: "2026-09-30T01:04:54+10:00", digest: 0f960b8e19b80a82 }
  - { id: main_socket_test, resource: main_socket_test.go, last_modified: "2026-09-30T01:04:54+10:00", digest: 4f5ff8b677dd3330 }
---

# Files
- `app.go`
- `app_terminal_registry_test.go`
- `internal/bmad/registry_fs.go`
- `internal/terminal/tmux_adapter_coverage_test.go`
- `internal/terminal/tmux_escape.go`
- `main.go`
- `main_socket_test.go`

# Symbols
- NewApp() (app.go:L305)
- TestStory1_NewAppInitializesMap() (app_terminal_registry_test.go:L343)
- registry_fs.go (internal/bmad/registry_fs.go:L1)
- TestIsTmuxAvailable_ReturnsBoolWithoutPanic() (internal/terminal/tmux_adapter_coverage_test.go:L55)
- IsTmuxAvailable() (internal/terminal/tmux_escape.go:L67)
- main.go (main.go:L1)
- resolveHelperPath() (main.go:L103)
- setupHelperSocketDir() (main.go:L122)
- dialHelperSecure() (main.go:L135)
- waitForSocket() (main.go:L150)
- main() (main.go:L161)
- main_socket_test.go (main_socket_test.go:L1)
- TestSetupHelperSocketDir_Mode0700() (main_socket_test.go:L11)
- TestHelperShutdown_RemovesSocketDir() (main_socket_test.go:L27)
- TestDialRefusesWorldAccessibleSocket() (main_socket_test.go:L39)

# Depends on
- [App](/modules/app-109.md)
- [Bridge](/modules/bridge.md)
- [main_test.go](/modules/main-test-go.md)
- [NewSessionManager](/modules/newsessionmanager.md)
- [panes.go](/modules/panes-go.md)
- [server_test.go](/modules/server-test-go.md)
- [tmux_adapter_coverage_test.go](/modules/tmux-adapter-coverage-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

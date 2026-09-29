---
type: Module
title: panes.go
description: "Graphify community 166: app_terminal_registry_test.go, internal/terminal/panes.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-09-30T06:45:21+10:00", digest: 2844ed2cca730a69 }
  - { id: panes, resource: internal/terminal/panes.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 45986d36cbef9af3 }
---

# Files
- `app_terminal_registry_test.go`
- `internal/terminal/panes.go`

# Symbols
- fakePaneDiscovery (app_terminal_registry_test.go:L20)
- .ListPanes() (app_terminal_registry_test.go:L26)
- .InvalidateCache() (app_terminal_registry_test.go:L30)
- .FindPaneForPID() (app_terminal_registry_test.go:L32)
- panes.go (internal/terminal/panes.go:L1)
- .InvalidateCache() (internal/terminal/panes.go:L127)
- discoverPanes() (internal/terminal/panes.go:L133)
- getParentPID() (internal/terminal/panes.go:L185)
- TerminalError (internal/terminal/panes.go:L22)
- .Error() (internal/terminal/panes.go:L28)
- .Unwrap() (internal/terminal/panes.go:L35)
- TmuxPane (internal/terminal/panes.go:L40)
- .Target() (internal/terminal/panes.go:L50)
- PaneDiscovery (internal/terminal/panes.go:L60)
- NewPaneDiscovery() (internal/terminal/panes.go:L67)
- .ListPanes() (internal/terminal/panes.go:L72)
- .FindPaneForPID() (internal/terminal/panes.go:L93)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

---
type: Module
title: TmuxPane
description: "Graphify community 303: app_terminal_registry_test.go, internal/terminal/panes.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
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
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

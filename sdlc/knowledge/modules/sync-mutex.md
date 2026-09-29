---
type: Module
title: sync.Mutex
description: "Graphify community 303: app_terminal_registry_test.go, internal/terminal/panes.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 2f85a0b04c1b8ba7 }
  - { id: panes, resource: internal/terminal/panes.go, last_modified: "2026-09-29T07:07:25Z", digest: 45986d36cbef9af3 }
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
- .ListPanes() (internal/terminal/panes.go:L72)
- .FindPaneForPID() (internal/terminal/panes.go:L93)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

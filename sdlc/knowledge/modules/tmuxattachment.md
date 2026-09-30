---
type: Module
title: TmuxAttachment
description: "Graphify community 138: internal/terminal/tmux_adapter.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: tmux_adapter, resource: internal/terminal/tmux_adapter.go, last_modified: "2026-09-29T07:07:25Z", digest: 1a5bf00ab35bc698 }
---

# Files
- `internal/terminal/tmux_adapter.go`

# Symbols
- TmuxAdapter (internal/terminal/tmux_adapter.go:L110)
- TmuxAttachment (internal/terminal/tmux_adapter.go:L140)
- .Attach() (internal/terminal/tmux_adapter.go:L189)
- .setupFIFO() (internal/terminal/tmux_adapter.go:L257)
- .startPipePane() (internal/terminal/tmux_adapter.go:L281)
- .startFIFOReader() (internal/terminal/tmux_adapter.go:L318)
- .startPaneDeathWatcher() (internal/terminal/tmux_adapter.go:L396)
- .startPollingLoop() (internal/terminal/tmux_adapter.go:L425)
- .Read() (internal/terminal/tmux_adapter.go:L466)
- .loadReadErr() (internal/terminal/tmux_adapter.go:L502)
- .runTmux() (internal/terminal/tmux_adapter.go:L515)
- .SendInput() (internal/terminal/tmux_adapter.go:L555)
- formatSendKeysHex() (internal/terminal/tmux_adapter.go:L569)
- .SendInputToTarget() (internal/terminal/tmux_adapter.go:L580)
- .SendKey() (internal/terminal/tmux_adapter.go:L593)
- .Resize() (internal/terminal/tmux_adapter.go:L598)
- .Close() (internal/terminal/tmux_adapter.go:L610)
- .cleanupFIFO() (internal/terminal/tmux_adapter.go:L637)
- .Close() (internal/terminal/tmux_adapter.go:L650)

# Depends on
- [go_pkg_fmt](/modules/go-pkg-fmt.md)
- [mockTmuxSession](/modules/mocktmuxsession.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

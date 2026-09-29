---
type: Module
title: App
description: "Graphify community 45: app.go, internal/explain/explain.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: explain, resource: internal/explain/explain.go, last_modified: "2026-04-10T11:36:27+10:00", digest: c8a0f1b0e6aad414 }
---

# Files
- `app.go`
- `internal/explain/explain.go`

# Symbols
- MarkdownMenuSettings (app.go:L108)
- mashedConfig (app.go:L122)
- themesPath() (app.go:L161)
- .startup() (app.go:L252)
- paneDiscoverer (app.go:L29)
- sessionManager (app.go:L36)
- .PickDirectory() (app.go:L400)
- .PickFile() (app.go:L412)
- .SetActiveContext() (app.go:L423)
- App (app.go:L45)
- .GetDevDir() (app.go:L520)
- .GetConfig() (app.go:L525)
- .DefaultEditorSettings() (app.go:L575)
- .GetEditorSettings() (app.go:L594)
- validateEditorSettings() (app.go:L603)
- .SetEditorSettings() (app.go:L642)
- .DefaultMarkdownMenuSettings() (app.go:L655)
- .GetMarkdownMenuSettings() (app.go:L668)
- .SetMarkdownMenuSettings() (app.go:L678)
- .GetSavedThemes() (app.go:L691)
- .SaveTheme() (app.go:L700)
- .RemoveTheme() (app.go:L724)
- .GetTerminalPort() (app.go:L752)
- .initSessionLog() (app.go:L757)
- .WriteConsoleLog() (app.go:L771)
- EditorSettings (app.go:L90)
- Explainer (internal/explain/explain.go:L15)
- New() (internal/explain/explain.go:L21)

# Depends on
- [App](/modules/app-349.md)
- [asset_watcher_test.go](/modules/asset-watcher-test-go.md)
- [AssetWatcher](/modules/assetwatcher.md)
- [Bridge](/modules/bridge.md)
- [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- [context.Context](/modules/context-context.md)
- [Executor](/modules/executor.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [go_pkg_strings](/modules/go-pkg-strings.md)
- [loadConfig](/modules/loadconfig.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [NotificationEngine](/modules/notificationengine.md)
- [registerTestProcess](/modules/registertestprocess.md)
- [RepoScanner](/modules/reposcanner.md)
- [screenshot_fullstack_test.go](/modules/screenshot-fullstack-test-go.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files

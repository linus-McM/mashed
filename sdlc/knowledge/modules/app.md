---
type: Module
title: App
description: "Graphify community 45: app.go, internal/explain/explain.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: explain, resource: internal/explain/explain.go, last_modified: "2026-04-10T11:36:27+10:00", digest: c8a0f1b0e6aad414 }
---

# Files
- `app.go`
- `internal/explain/explain.go`

# Symbols
- MarkdownMenuSettings (app.go:L108)
- themesPath() (app.go:L161)
- .startup() (app.go:L252)
- paneDiscoverer (app.go:L29)
- sessionManager (app.go:L36)
- .shutdown() (app.go:L374)
- .PickDirectory() (app.go:L400)
- .PickFile() (app.go:L412)
- .SetActiveContext() (app.go:L423)
- App (app.go:L45)
- .GetDevDir() (app.go:L520)
- .DefaultEditorSettings() (app.go:L575)
- .GetEditorSettings() (app.go:L594)
- validateEditorSettings() (app.go:L603)
- .SetEditorSettings() (app.go:L642)
- .DefaultMarkdownMenuSettings() (app.go:L655)
- .GetMarkdownMenuSettings() (app.go:L668)
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
- [Executor](/modules/executor.md)
- [executor_adapter_test.go](/modules/executor-adapter-test-go.md)
- [loadConfig](/modules/loadconfig.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [NewDefault](/modules/newdefault.md)
- [NewExecutor](/modules/newexecutor.md)
- [NotificationEngine](/modules/notificationengine.md)
- [question.go](/modules/question-go.md)
- [screenshot_fullstack_test.go](/modules/screenshot-fullstack-test-go.md)
- [Storage](/modules/storage.md)
- [time.Time](/modules/time-time.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)

# Features
- no feature plan names these files

---
type: Module
title: App
description: "Graphify community 50: app.go"
resource: .
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
---

# Files
- `app.go`

# Symbols
- MarkdownMenuSettings (app.go:L108)
- mashedConfig (app.go:L122)
- themesPath() (app.go:L161)
- paneDiscoverer (app.go:L29)
- sessionManager (app.go:L36)
- .PickDirectory() (app.go:L400)
- .PickFile() (app.go:L412)
- .SetActiveContext() (app.go:L423)
- App (app.go:L45)
- .GetDevDir() (app.go:L520)
- .GetConfig() (app.go:L525)
- .SetVSCodiumExtPath() (app.go:L539)
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
- .WriteConsoleLog() (app.go:L771)
- EditorSettings (app.go:L90)

# Depends on
- [App](/modules/app-349.md)
- [AssetWatcher](/modules/assetwatcher.md)
- [Bridge](/modules/bridge.md)
- [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- [context.Context](/modules/context-context.md)
- [ensureGitignoreEntry](/modules/ensuregitignoreentry.md)
- [Executor](/modules/executor.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [loadConfig](/modules/loadconfig.md)
- [NewExecutor](/modules/newexecutor.md)
- [NotificationEngine](/modules/notificationengine.md)
- [question.go](/modules/question-go.md)
- [time.Time](/modules/time-time.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

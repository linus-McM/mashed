---
type: Module
title: App
description: "Graphify community 50: app.go, screenshot_fullstack_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: screenshot_fullstack_test, resource: screenshot_fullstack_test.go, last_modified: "2026-04-21T10:01:10+10:00", digest: da72543a3716b145 }
---

# Files
- `app.go`
- `screenshot_fullstack_test.go`

# Symbols
- MarkdownMenuSettings (app.go:L108)
- mashedConfig (app.go:L122)
- themesPath() (app.go:L161)
- .startup() (app.go:L252)
- paneDiscoverer (app.go:L29)
- sessionManager (app.go:L36)
- .shutdown() (app.go:L374)
- .PickDirectory() (app.go:L400)
- .PickFile() (app.go:L412)
- .SetActiveContext() (app.go:L423)
- .TakeScreenshot() (app.go:L431)
- App (app.go:L45)
- ensureGitignoreEntry() (app.go:L470)
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
- TestEnsureGitignoreEntry_AC5() (screenshot_fullstack_test.go:L60)

# Depends on
- [asset_watcher_test.go](/modules/asset-watcher-test-go.md)
- [AssetWatcher](/modules/assetwatcher.md)
- [Bridge](/modules/bridge.md)
- [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- [domain/types.go](/modules/domain-types-go.md)
- [engine_test.go](/modules/engine-test-go.md)
- [Executor](/modules/executor.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [loadConfig](/modules/loadconfig.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [registerTestProcess](/modules/registertestprocess.md)
- [RepoScanner](/modules/reposcanner.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files

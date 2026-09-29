---
type: Module
title: app_terminal_registry_test.go
description: "Graphify community 90: app.go, app_terminal_registry_test.go, internal/terminal/panes.go, internal/terminal/tmux_adapter_coverage_test.go, internal/terminal/tmux_escape.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
  - { id: panes, resource: internal/terminal/panes.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 45986d36cbef9af3 }
  - { id: tmux_adapter_coverage_test, resource: internal/terminal/tmux_adapter_coverage_test.go, last_modified: "2026-04-11T19:58:57+10:00", digest: 24b353b6c1480a46 }
  - { id: tmux_escape, resource: internal/terminal/tmux_escape.go, last_modified: "2026-04-10T16:37:58+10:00", digest: d61b1108f3251770 }
---

# Files
- `app.go`
- `app_terminal_registry_test.go`
- `internal/terminal/panes.go`
- `internal/terminal/tmux_adapter_coverage_test.go`
- `internal/terminal/tmux_escape.go`

# Symbols
- NewApp() (app.go:L232)
- app_terminal_registry_test.go (app_terminal_registry_test.go:L1)
- testApp() (app_terminal_registry_test.go:L101)
- TestStory1_AC1_TerminalSessionJSONTags() (app_terminal_registry_test.go:L111)
- TestStory1_AC1_TerminalSessionRoundTrip() (app_terminal_registry_test.go:L137)
- TestStory1_AC2_FilterByRepo() (app_terminal_registry_test.go:L167)
- TestStory1_AC2_EmptyRepoPath() (app_terminal_registry_test.go:L187)
- TestStory1_AC2_SortBySpawnedAt() (app_terminal_registry_test.go:L199)
- TestStory1_AC3_PruneDeadSessions() (app_terminal_registry_test.go:L221)
- TestStory1_AC3_AllSessionsDead() (app_terminal_registry_test.go:L243)
- TestStory1_AC4_KillRemovesFromRegistry() (app_terminal_registry_test.go:L260)
- TestStory1_AC5_KillAlreadyDead() (app_terminal_registry_test.go:L282)
- TestStory1_AC5_KillNotFound() (app_terminal_registry_test.go:L297)
- TestStory1_RegisterOverwrite() (app_terminal_registry_test.go:L308)
- TestStory1_NewAppInitializesMap() (app_terminal_registry_test.go:L327)
- TestStory2_AC1_RecoverTerminalAndAgent() (app_terminal_registry_test.go:L335)
- TestStory2_AC2_FilterNonMatchingPrefixes() (app_terminal_registry_test.go:L358)
- TestStory2_AC3_TmuxNotRunning() (app_terminal_registry_test.go:L374)
- TestStory2_AC4_IdempotentRecovery() (app_terminal_registry_test.go:L385)
- TestStory2_ParsesRepoNameFromSessionName() (app_terminal_registry_test.go:L400)
- TestStory2_NoTimestampSuffix() (app_terminal_registry_test.go:L429)
- TestStory2_EmptyAndWhitespaceLines() (app_terminal_registry_test.go:L441)
- TestStory3_AC1_SpawnRegistersAgentSession() (app_terminal_registry_test.go:L454)
- TestStory3_AC2_SpawnRegistersTerminalSession() (app_terminal_registry_test.go:L481)
- TestStory3_AC3_KillAgentDeregisters() (app_terminal_registry_test.go:L505)
- newFakeManager() (app_terminal_registry_test.go:L51)
- TestStory3_AC4_FailedSpawnNoRegistration() (app_terminal_registry_test.go:L535)
- TestStory1_RegisterSessionConcurrent() (app_terminal_registry_test.go:L549)
- TestStory4_AC7_RecoverSessionsIsNoOp() (app_terminal_registry_test.go:L575)
- TestStory5_ResolveTmuxTarget() (app_terminal_registry_test.go:L589)
- NewPaneDiscovery() (internal/terminal/panes.go:L67)
- TestIsTmuxAvailable_ReturnsBoolWithoutPanic() (internal/terminal/tmux_adapter_coverage_test.go:L55)
- IsTmuxAvailable() (internal/terminal/tmux_escape.go:L67)

# Depends on
- [App](/modules/app.md)
- [Bridge](/modules/bridge.md)
- [ManagedSession](/modules/managedsession.md)
- [manager_test.go](/modules/manager-test-go.md)
- [server_test.go](/modules/server-test-go.md)
- [tmux_adapter_coverage_test.go](/modules/tmux-adapter-coverage-test-go.md)
- [TmuxPane](/modules/tmuxpane.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

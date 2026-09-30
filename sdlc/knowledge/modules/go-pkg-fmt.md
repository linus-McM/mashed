---
type: Module
title: go_pkg_fmt
description: "Graphify community 8: app_uiadapter.go, app_uiadapter_v3.go, cmd/pty-helper/main.go, internal/agent/engine.go, internal/bmad/cleanup.go, internal/bmad/prompts_test.go, internal/git/errors.go, internal"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-09-29T07:07:25Z", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_v3, resource: app_uiadapter_v3.go, last_modified: "2026-09-29T07:07:25Z", digest: c6ede9f0c3335d2e }
  - { id: main, resource: cmd/pty-helper/main.go, last_modified: "2026-09-29T07:07:25Z", digest: 5b417efc875ee944 }
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-09-29T07:07:25Z", digest: 1042190db571e4e7 }
  - { id: cleanup, resource: internal/bmad/cleanup.go, last_modified: "2026-09-29T07:07:25Z", digest: 78b71ea7a045636f }
  - { id: prompts_test, resource: internal/bmad/prompts_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 6a4efc33b34610a9 }
  - { id: errors, resource: internal/git/errors.go, last_modified: "2026-09-29T07:07:25Z", digest: 1795cc59c6846604 }
  - { id: errors, resource: internal/scanner/errors.go, last_modified: "2026-09-29T07:07:25Z", digest: 2b7bc0225851334e }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-09-29T07:07:25Z", digest: c8bd28e2f7bd59b8 }
  - { id: bridge, resource: internal/terminal/bridge.go, last_modified: "2026-09-29T07:07:25Z", digest: a42cec584372611b }
  - { id: client, resource: internal/terminal/helper/client.go, last_modified: "2026-09-29T07:07:25Z", digest: 46fc572e21238bfb }
  - { id: protocol, resource: internal/terminal/helper/protocol.go, last_modified: "2026-09-29T07:07:25Z", digest: 07f801b52b8e41da }
  - { id: server, resource: internal/terminal/helper/server.go, last_modified: "2026-09-29T07:07:25Z", digest: e54a6a9ea36ca59d }
  - { id: manager, resource: internal/terminal/manager.go, last_modified: "2026-09-29T07:07:25Z", digest: 4c7107fb7905c1b1 }
  - { id: panes, resource: internal/terminal/panes.go, last_modified: "2026-09-29T07:07:25Z", digest: 45986d36cbef9af3 }
  - { id: session, resource: internal/terminal/session.go, last_modified: "2026-09-29T07:07:25Z", digest: 6213e8a0ece661b9 }
  - { id: tmux_adapter, resource: internal/terminal/tmux_adapter.go, last_modified: "2026-09-29T07:07:25Z", digest: 1a5bf00ab35bc698 }
  - { id: tmux_adapter_coverage_test, resource: internal/terminal/tmux_adapter_coverage_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 24b353b6c1480a46 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-09-29T07:07:25Z", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-09-29T07:07:25Z", digest: 3e20643bf928e2f2 }
  - { id: stub, resource: internal/uiadapter/backend/claudeapi/stub.go, last_modified: "2026-09-29T07:07:25Z", digest: 0c95368a8fcd18d3 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-09-29T07:07:25Z", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 4342fc0ff1e40237 }
  - { id: stub, resource: internal/uiadapter/backend/claudecli/stub.go, last_modified: "2026-09-29T07:07:25Z", digest: 9d9685929a699a14 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-09-29T07:07:25Z", digest: e57a307fa43599dd }
  - { id: stub, resource: internal/uiadapter/backend/ollama/stub.go, last_modified: "2026-09-29T07:07:25Z", digest: 82a69a451dac2ce8 }
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-09-29T07:07:25Z", digest: 34796db160eeb018 }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-09-29T07:07:25Z", digest: aa5a14fdab76f8f6 }
  - { id: router, resource: internal/uiadapter/backend/router.go, last_modified: "2026-09-29T07:07:25Z", digest: 456b1b3b77e3a38e }
  - { id: contextguard, resource: internal/uiadapter/contextguard.go, last_modified: "2026-09-29T07:07:25Z", digest: 417a5e0974dc194b }
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-09-29T07:07:25Z", digest: 88e0c867731816cd }
---

# Files
- `app_uiadapter.go`
- `app_uiadapter_v3.go`
- `cmd/pty-helper/main.go`
- `internal/agent/engine.go`
- `internal/bmad/cleanup.go`
- `internal/bmad/prompts_test.go`
- `internal/git/errors.go`
- `internal/scanner/errors.go`
- `internal/scanner/repos.go`
- `internal/terminal/bridge.go`
- `internal/terminal/helper/client.go`
- `internal/terminal/helper/protocol.go`
- `internal/terminal/helper/server.go`
- `internal/terminal/manager.go`
- `internal/terminal/panes.go`
- `internal/terminal/session.go`
- `internal/terminal/tmux_adapter.go`
- `internal/terminal/tmux_adapter_coverage_test.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/stub.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/backend/claudecli/stub.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/ollama/stub.go`
- `internal/uiadapter/backend/registry.go`
- `internal/uiadapter/backend/registry_test.go`
- `internal/uiadapter/backend/router.go`
- `internal/uiadapter/contextguard.go`
- `internal/uiadapter/fallback_tiers.go`

# Symbols
- app_uiadapter.go (app_uiadapter.go:L1)
- tagsListResponse (app_uiadapter.go:L141)
- init() (app_uiadapter.go:L34)
- app_uiadapter_v3.go (app_uiadapter_v3.go:L1)
- pty-helper/main.go (cmd/pty-helper/main.go:L1)
- engine.go (internal/agent/engine.go:L1)
- EngineError (internal/agent/engine.go:L25)
- .Error() (internal/agent/engine.go:L31)
- .Unwrap() (internal/agent/engine.go:L38)
- cleanup.go (internal/bmad/cleanup.go:L1)
- prompts_test.go (internal/bmad/prompts_test.go:L1)
- readFixture() (internal/bmad/prompts_test.go:L13)
- emitAwaitingInputForAll() (internal/bmad/prompts_test.go:L25)
- TestSnapshot_PreU4LoadRoundTrip() (internal/bmad/prompts_test.go:L41)
- git/errors.go (internal/git/errors.go:L1)
- .Error() (internal/git/errors.go:L12)
- .Unwrap() (internal/git/errors.go:L16)
- GitError (internal/git/errors.go:L6)
- scanner/errors.go (internal/scanner/errors.go:L1)
- ScanError (internal/scanner/errors.go:L21)
- .Error() (internal/scanner/errors.go:L27)
- .Unwrap() (internal/scanner/errors.go:L34)
- ParseError (internal/scanner/errors.go:L37)
- .Error() (internal/scanner/errors.go:L44)
- .Unwrap() (internal/scanner/errors.go:L51)
- WatchError (internal/scanner/errors.go:L54)
- .Error() (internal/scanner/errors.go:L60)
- .Unwrap() (internal/scanner/errors.go:L64)
- repos.go (internal/scanner/repos.go:L1)
- bridge.go (internal/terminal/bridge.go:L1)
- resizeMsg (internal/terminal/bridge.go:L54)
- helper/client.go (internal/terminal/helper/client.go:L1)
- protocol.go (internal/terminal/helper/protocol.go:L1)
- server.go (internal/terminal/helper/server.go:L1)
- manager.go (internal/terminal/manager.go:L1)
- panes.go (internal/terminal/panes.go:L1)
- TerminalError (internal/terminal/panes.go:L22)
- .Error() (internal/terminal/panes.go:L28)
- .Unwrap() (internal/terminal/panes.go:L35)
- session.go (internal/terminal/session.go:L1)
- tmux_adapter.go (internal/terminal/tmux_adapter.go:L1)
- CommandRunner (internal/terminal/tmux_adapter.go:L67)
- DefaultCommandRunner() (internal/terminal/tmux_adapter.go:L71)
- TestDefaultCommandRunner_EchoesStdout() (internal/terminal/tmux_adapter_coverage_test.go:L42)
- backend.go (internal/uiadapter/backend/backend.go:L1)
- claudeapi/client.go (internal/uiadapter/backend/claudeapi/client.go:L1)
- init() (internal/uiadapter/backend/claudeapi/client.go:L302)
- claudeapi/stub.go (internal/uiadapter/backend/claudeapi/stub.go:L1)
- init() (internal/uiadapter/backend/claudeapi/stub.go:L13)
- claudecli/client.go (internal/uiadapter/backend/claudecli/client.go:L1)
- extractFencedJSON() (internal/uiadapter/backend/claudecli/client.go:L326)
- init() (internal/uiadapter/backend/claudecli/client.go:L351)
- TestClaudeCLI_ParsingHandlesShapes() (internal/uiadapter/backend/claudecli/client_test.go:L27)
- claudecli/stub.go (internal/uiadapter/backend/claudecli/stub.go:L1)
- init() (internal/uiadapter/backend/claudecli/stub.go:L12)
- lifecycle.go (internal/uiadapter/backend/lifecycle.go:L1)
- ollama/stub.go (internal/uiadapter/backend/ollama/stub.go:L1)
- init() (internal/uiadapter/backend/ollama/stub.go:L13)
- backend/registry.go (internal/uiadapter/backend/registry.go:L1)
- Constructor (internal/uiadapter/backend/registry.go:L13)
- Register() (internal/uiadapter/backend/registry.go:L24)
- Available() (internal/uiadapter/backend/registry.go:L53)
- reset() (internal/uiadapter/backend/registry.go:L67)
- backend/registry_test.go (internal/uiadapter/backend/registry_test.go:L1)
- TestBackend_Available() (internal/uiadapter/backend/registry_test.go:L132)
- router.go (internal/uiadapter/backend/router.go:L1)
- contextguard.go (internal/uiadapter/contextguard.go:L1)
- fallback_tiers.go (internal/uiadapter/fallback_tiers.go:L1)

# Depends on
- [app_uiadapter_bindings_test.go](/modules/app-uiadapter-bindings-test-go.md)
- [Bridge](/modules/bridge.md)
- [Client](/modules/client.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [domain/types.go](/modules/domain-types-go.md)
- [.doScan](/modules/doscan.md)
- [engine_test.go](/modules/engine-test-go.md)
- [loadConfig](/modules/loadconfig.md)
- [loadOllamaBaseURL](/modules/loadollamabaseurl.md)
- [main.go](/modules/main-go.md)
- [ManagedSession](/modules/managedsession.md)
- [manager_test.go](/modules/manager-test-go.md)
- [mockTmuxSession](/modules/mocktmuxsession.md)
- [NewStub](/modules/newstub.md)
- [resume.go](/modules/resume-go.md)
- [Router](/modules/router.md)
- [server_test.go](/modules/server-test-go.md)
- [session_test.go](/modules/session-test-go.md)
- [sessions.go](/modules/sessions-go.md)
- [stages_test.go](/modules/stages-test-go.md)
- [StubBackend](/modules/stubbackend.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)
- [sync.Mutex](/modules/sync-mutex.md)
- [time.Time](/modules/time-time.md)
- [tmux_adapter_coverage_test.go](/modules/tmux-adapter-coverage-test-go.md)
- [TmuxAttachment](/modules/tmuxattachment.md)

# Inferred
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files

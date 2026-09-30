---
type: Module
title: claudeapi/client.go
description: "Graphify community 190: cmd/pty-helper/main.go, cmd/pty-helper/main_test.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/claudeapi/stub.go, internal/uiadapter/backend/cl"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: main, resource: cmd/pty-helper/main.go, last_modified: "2026-09-30T01:04:54+10:00", digest: 37f082bc5d1dbd62 }
  - { id: main_test, resource: cmd/pty-helper/main_test.go, last_modified: "2026-09-30T01:04:54+10:00", digest: 4016e2b9942476ec }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: stub, resource: internal/uiadapter/backend/claudeapi/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 0c95368a8fcd18d3 }
  - { id: stub, resource: internal/uiadapter/backend/claudecli/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 9d9685929a699a14 }
  - { id: stub, resource: internal/uiadapter/backend/ollama/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 82a69a451dac2ce8 }
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 34796db160eeb018 }
---

# Files
- `cmd/pty-helper/main.go`
- `cmd/pty-helper/main_test.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/stub.go`
- `internal/uiadapter/backend/claudecli/stub.go`
- `internal/uiadapter/backend/ollama/stub.go`
- `internal/uiadapter/backend/registry.go`

# Symbols
- pty-helper/main.go (cmd/pty-helper/main.go:L1)
- main() (cmd/pty-helper/main.go:L15)
- listenSocket() (cmd/pty-helper/main.go:L55)
- pty-helper/main_test.go (cmd/pty-helper/main_test.go:L1)
- TestListenSocket_Mode0600() (cmd/pty-helper/main_test.go:L10)
- claudeapi/client.go (internal/uiadapter/backend/claudeapi/client.go:L1)
- responseBody (internal/uiadapter/backend/claudeapi/client.go:L195)
- .call() (internal/uiadapter/backend/claudeapi/client.go:L211)
- waitRetryAfter() (internal/uiadapter/backend/claudeapi/client.go:L270)
- sanitiseErrorBody() (internal/uiadapter/backend/claudeapi/client.go:L284)
- anthropicVersion() (internal/uiadapter/backend/claudeapi/client.go:L292)
- init() (internal/uiadapter/backend/claudeapi/client.go:L302)
- claudeapi/stub.go (internal/uiadapter/backend/claudeapi/stub.go:L1)
- init() (internal/uiadapter/backend/claudeapi/stub.go:L13)
- claudecli/stub.go (internal/uiadapter/backend/claudecli/stub.go:L1)
- init() (internal/uiadapter/backend/claudecli/stub.go:L12)
- ollama/stub.go (internal/uiadapter/backend/ollama/stub.go:L1)
- init() (internal/uiadapter/backend/ollama/stub.go:L13)
- backend/registry.go (internal/uiadapter/backend/registry.go:L1)
- Constructor (internal/uiadapter/backend/registry.go:L13)
- Register() (internal/uiadapter/backend/registry.go:L24)
- reset() (internal/uiadapter/backend/registry.go:L67)

# Depends on
- [Accountant](/modules/accountant.md)
- [backend/registry_test.go](/modules/backend-registry-test-go.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [server_test.go](/modules/server-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

---
type: Module
title: time.Time
description: "Graphify community 123: internal/domain/types.go, internal/scanner/repos.go, internal/uiadapter/adapter.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-05-07T10:33:04+10:00", digest: c8bd28e2f7bd59b8 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
---

# Files
- `internal/domain/types.go`
- `internal/scanner/repos.go`
- `internal/uiadapter/adapter.go`

# Symbols
- AgentSession (internal/domain/types.go:L137)
- RepoInfo (internal/domain/types.go:L146)
- .fetchGitInfo() (internal/scanner/repos.go:L122)
- RepoScanner (internal/scanner/repos.go:L17)
- .InvalidateCache() (internal/scanner/repos.go:L173)
- repoCacheEntry (internal/scanner/repos.go:L25)
- .ScanRepos() (internal/scanner/repos.go:L48)
- .getRepoInfo() (internal/scanner/repos.go:L96)
- defaultAdapter (internal/uiadapter/adapter.go:L106)
- disabledAdapter (internal/uiadapter/adapter.go:L118)
- .Translate() (internal/uiadapter/adapter.go:L156)
- .Translate() (internal/uiadapter/adapter.go:L164)
- .chat() (internal/uiadapter/adapter.go:L210)
- .stampSuccessMetadata() (internal/uiadapter/adapter.go:L220)
- .emitFallback() (internal/uiadapter/adapter.go:L239)
- .logTelemetry() (internal/uiadapter/adapter.go:L255)
- .logTelemetryWithSanitize() (internal/uiadapter/adapter.go:L263)
- extractJSONObject() (internal/uiadapter/adapter.go:L293)

# Depends on
- [context.Context](/modules/context-context.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Features
- no feature plan names these files

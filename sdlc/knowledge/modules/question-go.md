---
type: Module
title: question.go
description: "Graphify community 32: app.go, app_scan.go, internal/bmad/question.go, internal/explain/explain.go, internal/scanner/claude.go, internal/scanner/repos.go, internal/scanner/sessions.go, internal/uiadap"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_scan, resource: app_scan.go, last_modified: "2026-05-07T10:33:04+10:00", digest: e5b2c4798c9f1cad }
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: explain, resource: internal/explain/explain.go, last_modified: "2026-04-10T11:36:27+10:00", digest: c8a0f1b0e6aad414 }
  - { id: claude, resource: internal/scanner/claude.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 2c44ec40e17d728e }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-05-07T10:33:04+10:00", digest: c8bd28e2f7bd59b8 }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
  - { id: breaker, resource: internal/uiadapter/breaker.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ce7fb713c254dda5 }
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
---

# Files
- `app.go`
- `app_scan.go`
- `internal/bmad/question.go`
- `internal/explain/explain.go`
- `internal/scanner/claude.go`
- `internal/scanner/repos.go`
- `internal/scanner/sessions.go`
- `internal/uiadapter/breaker.go`
- `internal/uiadapter/cache.go`

# Symbols
- sanitizeID() (app.go:L779)
- repoNameFromDir() (app.go:L794)
- App (app_scan.go:L21)
- .initScanning() (app_scan.go:L21)
- .scanLoop() (app_scan.go:L47)
- .resolveTmuxTarget() (app_scan.go:L66)
- .doScan() (app_scan.go:L77)
- question.go (internal/bmad/question.go:L1)
- QuestionEvent (internal/bmad/question.go:L55)
- IdleEvent (internal/bmad/question.go:L77)
- explain.go (internal/explain/explain.go:L1)
- Explainer (internal/explain/explain.go:L15)
- New() (internal/explain/explain.go:L21)
- .Explain() (internal/explain/explain.go:L29)
- cacheKey() (internal/explain/explain.go:L71)
- truncateHunk() (internal/explain/explain.go:L77)
- envWithoutAPIKey() (internal/explain/explain.go:L86)
- ClaudeCodeProvider (internal/scanner/claude.go:L18)
- pidDirEntry (internal/scanner/claude.go:L32)
- NewClaudeCodeProvider() (internal/scanner/claude.go:L40)
- .SessionDir() (internal/scanner/claude.go:L70)
- .DevDir() (internal/scanner/claude.go:L76)
- .cachePidDir() (internal/scanner/claude.go:L80)
- .getCachedPidDir() (internal/scanner/claude.go:L86)
- NewRepoScanner() (internal/scanner/repos.go:L31)
- sessionParserState (internal/scanner/sessions.go:L17)
- .StateOf() (internal/uiadapter/breaker.go:L118)
- .Do() (internal/uiadapter/breaker.go:L134)
- breakerStateName() (internal/uiadapter/breaker.go:L16)
- BreakerSet (internal/uiadapter/breaker.go:L38)
- .For() (internal/uiadapter/breaker.go:L62)
- cache.go (internal/uiadapter/cache.go:L1)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [Executor](/modules/executor.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [ModelInfo](/modules/modelinfo.md)
- [question_test.go](/modules/question-test-go.md)
- [ResponseCache](/modules/responsecache.md)
- [time.Time](/modules/time-time.md)
- [tokensamples_test.go](/modules/tokensamples-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

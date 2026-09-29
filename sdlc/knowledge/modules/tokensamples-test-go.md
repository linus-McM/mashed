---
type: Module
title: tokensamples_test.go
description: "Graphify community 363: internal/agent/tokensamples.go, internal/agent/tokensamples_test.go"
resource: internal/agent
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: tokensamples, resource: internal/agent/tokensamples.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 281db4a1d2ab32a0 }
  - { id: tokensamples_test, resource: internal/agent/tokensamples_test.go, last_modified: "2026-04-11T17:08:28+10:00", digest: e73f3aa2bb785df4 }
---

# Files
- `internal/agent/tokensamples.go`
- `internal/agent/tokensamples_test.go`

# Symbols
- tokensamples.go (internal/agent/tokensamples.go:L1)
- MaybeAppendTokenSample() (internal/agent/tokensamples.go:L24)
- tokensamples_test.go (internal/agent/tokensamples_test.go:L1)
- TestNotificationEventJSON_IncludesTokenSamples() (internal/agent/tokensamples_test.go:L112)
- TestNotificationEventJSON_OmitEmptyTokenSamples() (internal/agent/tokensamples_test.go:L129)
- TestMaybeAppendTokenSample_CapEnforced() (internal/agent/tokensamples_test.go:L15)
- TestMaybeAppendTokenSample_FirstSampleAlwaysAppends() (internal/agent/tokensamples_test.go:L34)
- TestMaybeAppendTokenSample_DeltaThrottle() (internal/agent/tokensamples_test.go:L41)
- TestMaybeAppendTokenSample_DeltaBypassedForDecreases() (internal/agent/tokensamples_test.go:L61)
- TestMaybeAppendTokenSample_ThreadSafeRace() (internal/agent/tokensamples_test.go:L70)
- TestAgentJSON_IncludesTokenSamples() (internal/agent/tokensamples_test.go:L97)

# Depends on
- [NotificationEngine](/modules/notificationengine.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

---
type: Module
title: session_naming_test.go
description: "Graphify community 122: internal/bmad/session_naming.go, internal/bmad/session_naming_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: session_naming, resource: internal/bmad/session_naming.go, last_modified: "2026-04-10T15:50:15+10:00", digest: acbdad6853f5eaed }
  - { id: session_naming_test, resource: internal/bmad/session_naming_test.go, last_modified: "2026-04-10T15:50:15+10:00", digest: 42de95158f5b442b }
---

# Files
- `internal/bmad/session_naming.go`
- `internal/bmad/session_naming_test.go`

# Symbols
- slugifyComponent() (internal/bmad/session_naming.go:L130)
- shortHashOf() (internal/bmad/session_naming.go:L147)
- BuildSessionName() (internal/bmad/session_naming.go:L69)
- session_naming_test.go (internal/bmad/session_naming_test.go:L1)
- TestBuildSessionName_AC3_AllEmptyFallback() (internal/bmad/session_naming_test.go:L184)
- TestBuildSessionName_AC3_EmptyComponentFallbacks() (internal/bmad/session_naming_test.go:L191)
- buildAndSplit() (internal/bmad/session_naming_test.go:L21)
- TestBuildSessionName_AC4_LengthCapWith500ByteInputs() (internal/bmad/session_naming_test.go:L262)
- TestBuildSessionName_AC4_HashPreservedUnderTruncation() (internal/bmad/session_naming_test.go:L288)
- TestBuildSessionName_HashDeterminismAndUniqueness() (internal/bmad/session_naming_test.go:L304)
- TestParseSessionName_AC5_RoundTrip() (internal/bmad/session_naming_test.go:L364)
- TestParseSessionName_AC5_FeatureBranchAppearsInName() (internal/bmad/session_naming_test.go:L387)
- assertValidHash() (internal/bmad/session_naming_test.go:L40)
- TestSlugifyComponent() (internal/bmad/session_naming_test.go:L442)
- TestSessionNaming_ConstantsSanity() (internal/bmad/session_naming_test.go:L483)
- TestBuildSessionName_AC1_HappyPath() (internal/bmad/session_naming_test.go:L49)
- TestBuildSessionName_AC2_SlugificationRules() (internal/bmad/session_naming_test.go:L74)

# Depends on
- [ParseSessionName](/modules/parsesessionname.md)

# Inferred
- [ParseSessionName](/modules/parsesessionname.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

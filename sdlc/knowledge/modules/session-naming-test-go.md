---
type: Module
title: session_naming_test.go
description: "Graphify community 101: internal/bmad/session_naming.go, internal/bmad/session_naming_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
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
- ParseSessionName() (internal/bmad/session_naming.go:L99)
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
- TestParseSessionName_AC6_InvalidInputs() (internal/bmad/session_naming_test.go:L404)
- TestSlugifyComponent() (internal/bmad/session_naming_test.go:L442)
- TestSessionNaming_ConstantsSanity() (internal/bmad/session_naming_test.go:L483)
- TestBuildSessionName_AC1_HappyPath() (internal/bmad/session_naming_test.go:L49)
- TestBuildSessionName_AC2_SlugificationRules() (internal/bmad/session_naming_test.go:L74)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

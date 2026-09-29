---
type: Module
title: loader.go
description: "Graphify community 66: internal/advice/loader.go, internal/advice/loader_test.go"
resource: internal/advice
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: loader, resource: internal/advice/loader.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 8dbc36587dafa6a5 }
  - { id: loader_test, resource: internal/advice/loader_test.go, last_modified: "2026-04-10T10:10:32+10:00", digest: 99b099c62c602a3e }
---

# Files
- `internal/advice/loader.go`
- `internal/advice/loader_test.go`

# Symbols
- loader.go (internal/advice/loader.go:L1)
- loadModesFromDir() (internal/advice/loader.go:L103)
- LoadAdviceBody() (internal/advice/loader.go:L127)
- loadAdviceBodyFromDirs() (internal/advice/loader.go:L142)
- adviceFrontmatter (internal/advice/loader.go:L18)
- findModeInDir() (internal/advice/loader.go:L184)
- parseAdviceFile() (internal/advice/loader.go:L207)
- parseAdviceFileFromFS() (internal/advice/loader.go:L222)
- parseAdviceContent() (internal/advice/loader.go:L237)
- LoadAdviceModes() (internal/advice/loader.go:L28)
- globalAdviceDir() (internal/advice/loader.go:L304)
- localAdviceDir() (internal/advice/loader.go:L313)
- TestParseAdviceFileFromFS() (internal/advice/loader_test.go:L380)
- TestParseAdviceFile_NonexistentFile() (internal/advice/loader_test.go:L396)
- TestFindModeInDir_NonexistentDir() (internal/advice/loader_test.go:L603)
- TestLoadModesFromDir_NonexistentDir() (internal/advice/loader_test.go:L616)
- TestLoadAdviceBody_PublicAPI() (internal/advice/loader_test.go:L622)
- TestLocalAdviceDir() (internal/advice/loader_test.go:L652)

# Depends on
- [loader_test.go](/modules/loader-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

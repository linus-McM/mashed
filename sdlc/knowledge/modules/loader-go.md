---
type: Module
title: loader.go
description: "Graphify community 191: internal/advice/loader.go, internal/advice/loader_test.go"
resource: internal/advice
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
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
- adviceFrontmatter (internal/advice/loader.go:L18)
- findModeInDir() (internal/advice/loader.go:L184)
- parseAdviceFile() (internal/advice/loader.go:L207)
- parseAdviceContent() (internal/advice/loader.go:L237)
- splitFrontmatter() (internal/advice/loader.go:L269)
- TestParseAdviceFile_NonexistentFile() (internal/advice/loader_test.go:L396)
- TestFindModeInDir_NonexistentDir() (internal/advice/loader_test.go:L603)
- TestSplitFrontmatter_EmptyBody() (internal/advice/loader_test.go:L608)
- TestLoadModesFromDir_NonexistentDir() (internal/advice/loader_test.go:L616)
- TestSplitFrontmatter_CRLFLineEndings() (internal/advice/loader_test.go:L629)
- TestSplitFrontmatter_OpeningCRLF() (internal/advice/loader_test.go:L637)

# Depends on
- [LoadAdviceBody](/modules/loadadvicebody.md)
- [loader_test.go](/modules/loader-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

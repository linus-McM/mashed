---
type: Module
title: loader_test.go
description: "Graphify community 20: internal/advice/loader.go, internal/advice/loader_test.go"
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
- loadAdviceBodyFromDirs() (internal/advice/loader.go:L142)
- parseAdviceFileFromFS() (internal/advice/loader.go:L222)
- loadAdviceModesFromDirs() (internal/advice/loader.go:L44)
- loader_test.go (internal/advice/loader_test.go:L1)
- writeAdviceFile() (internal/advice/loader_test.go:L14)
- TestLoadAdviceModes_MergePriority() (internal/advice/loader_test.go:L167)
- TestLoadAdviceModes_SortOrderThenDisplayName() (internal/advice/loader_test.go:L234)
- TestParseAdviceFile() (internal/advice/loader_test.go:L24)
- TestLoadAdviceModes_MissingDirectories() (internal/advice/loader_test.go:L269)
- TestLoadAdviceModes_NonMdFilesIgnored() (internal/advice/loader_test.go:L276)
- TestLoadAdviceModes_MalformedFilesSkipped() (internal/advice/loader_test.go:L298)
- TestLoadAdviceBody() (internal/advice/loader_test.go:L335)
- TestParseAdviceFileFromFS() (internal/advice/loader_test.go:L380)
- TestLoadAdviceBody_GlobalOnly() (internal/advice/loader_test.go:L402)
- TestBundledDefaults_AllFilesValid() (internal/advice/loader_test.go:L419)
- TestBundledDefaults_BodyLoadable() (internal/advice/loader_test.go:L471)
- TestBundledDefaults_LocalOverridesBundled() (internal/advice/loader_test.go:L490)
- TestLoadAdviceBody_BundledFallback() (internal/advice/loader_test.go:L524)
- TestLoadAdviceModes_BundledWithEmbeddedFS() (internal/advice/loader_test.go:L531)
- TestLoadAdviceModes_AllThreeSources() (internal/advice/loader_test.go:L549)
- TestFindModeInDir_NotFound() (internal/advice/loader_test.go:L588)

# Depends on
- [LoadAdviceBody](/modules/loadadvicebody.md)
- [loader.go](/modules/loader-go.md)

# Inferred
- [loader.go](/modules/loader-go.md)

# Features
- no feature plan names these files

---
type: Module
title: loader_test.go
description: "Graphify community 15: internal/advice/loader.go, internal/advice/loader_test.go"
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
- splitFrontmatter() (internal/advice/loader.go:L269)
- loadAdviceModesFromDirs() (internal/advice/loader.go:L44)
- loader_test.go (internal/advice/loader_test.go:L1)
- writeAdviceFile() (internal/advice/loader_test.go:L14)
- TestLoadAdviceModes_MergePriority() (internal/advice/loader_test.go:L167)
- TestLoadAdviceModes_SortOrderThenDisplayName() (internal/advice/loader_test.go:L234)
- TestParseAdviceFile() (internal/advice/loader_test.go:L24)
- TestLoadAdviceModes_MissingDirectories() (internal/advice/loader_test.go:L269)
- TestLoadAdviceModes_NonMdFilesIgnored() (internal/advice/loader_test.go:L276)
- TestLoadAdviceModes_MalformedFilesSkipped() (internal/advice/loader_test.go:L298)
- TestLoadAdviceModes_EmptyRepoPathSkipsLocal() (internal/advice/loader_test.go:L327)
- TestLoadAdviceBody() (internal/advice/loader_test.go:L335)
- TestLoadAdviceBody_GlobalOnly() (internal/advice/loader_test.go:L402)
- TestBundledDefaults_AllFilesValid() (internal/advice/loader_test.go:L419)
- TestBundledDefaults_BodyLoadable() (internal/advice/loader_test.go:L471)
- TestBundledDefaults_LocalOverridesBundled() (internal/advice/loader_test.go:L490)
- TestLoadAdviceBody_BundledFallback() (internal/advice/loader_test.go:L524)
- TestLoadAdviceModes_BundledWithEmbeddedFS() (internal/advice/loader_test.go:L531)
- TestLoadAdviceModes_AllThreeSources() (internal/advice/loader_test.go:L549)
- TestFindModeInDir_NotFound() (internal/advice/loader_test.go:L588)
- TestSplitFrontmatter_EmptyBody() (internal/advice/loader_test.go:L608)
- TestSplitFrontmatter_CRLFLineEndings() (internal/advice/loader_test.go:L629)
- TestSplitFrontmatter_OpeningCRLF() (internal/advice/loader_test.go:L637)
- TestGlobalAdviceDir() (internal/advice/loader_test.go:L645)

# Depends on
- [loader.go](/modules/loader-go.md)

# Inferred
- [loader.go](/modules/loader-go.md)

# Features
- no feature plan names these files

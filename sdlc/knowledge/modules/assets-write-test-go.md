---
type: Module
title: assets_write_test.go
description: "Graphify community 257: internal/bmad/assets.go, internal/bmad/assets_test.go, internal/bmad/assets_write.go, internal/bmad/assets_write_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: assets, resource: internal/bmad/assets.go, last_modified: "2026-04-12T16:58:02+10:00", digest: ae5984d95ad01e8c }
  - { id: assets_test, resource: internal/bmad/assets_test.go, last_modified: "2026-04-11T22:06:50+10:00", digest: 906af472f47c7866 }
  - { id: assets_write, resource: internal/bmad/assets_write.go, last_modified: "2026-04-12T17:09:26+10:00", digest: 4f3137dec4993601 }
  - { id: assets_write_test, resource: internal/bmad/assets_write_test.go, last_modified: "2026-04-12T17:09:26+10:00", digest: ea9c116e2df2aa11 }
---

# Files
- `internal/bmad/assets.go`
- `internal/bmad/assets_test.go`
- `internal/bmad/assets_write.go`
- `internal/bmad/assets_write_test.go`

# Symbols
- extractFrontmatter() (internal/bmad/assets.go:L167)
- TestExtractFrontmatter() (internal/bmad/assets_test.go:L30)
- validateAssetPath() (internal/bmad/assets_write.go:L146)
- WriteMashedAssetFrontmatter() (internal/bmad/assets_write.go:L48)
- assets_write_test.go (internal/bmad/assets_write_test.go:L1)
- TestWriteMashedAssetFrontmatter_PathValidation() (internal/bmad/assets_write_test.go:L169)
- TestWriteMashedAssetFrontmatter_WriteFailure() (internal/bmad/assets_write_test.go:L209)
- TestWriteMashedAssetFrontmatter_MissingFile() (internal/bmad/assets_write_test.go:L236)
- TestWriteMashedAssetFrontmatter_RoundTrip() (internal/bmad/assets_write_test.go:L245)
- TestWriteMashedAssetFrontmatter_AddNewField() (internal/bmad/assets_write_test.go:L281)
- TestWriteMashedAssetFrontmatter_EmptyFrontmatter() (internal/bmad/assets_write_test.go:L313)
- TestWriteMashedAssetFrontmatter() (internal/bmad/assets_write_test.go:L41)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [assets_test.go](/modules/assets-test-go.md)

# Features
- no feature plan names these files

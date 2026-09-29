---
type: Module
title: font_scanner.go
description: "Graphify community 175: font_scanner.go"
resource: .
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: font_scanner, resource: font_scanner.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 766f77afbc920665 }
---

# Files
- `font_scanner.go`

# Symbols
- font_scanner.go (font_scanner.go:L1)
- matchNerdFont() (font_scanner.go:L100)
- LocalFontFile (font_scanner.go:L126)
- LocalFontFamily (font_scanner.go:L135)
- fontsDir() (font_scanner.go:L143)
- NerdFontEntry (font_scanner.go:L15)
- App (font_scanner.go:L159)
- .GetFontsDir() (font_scanner.go:L159)
- .OpenFontsDir() (font_scanner.go:L164)
- fontFormat() (font_scanner.go:L170)
- detectWeightStyle() (font_scanner.go:L206)
- familyFromFilename() (font_scanner.go:L217)
- addFontToFamilies() (font_scanner.go:L230)
- scanZipForFonts() (font_scanner.go:L243)
- .ListLocalFonts() (font_scanner.go:L270)
- .ListNerdFonts() (font_scanner.go:L312)
- fontDirs() (font_scanner.go:L84)
- isFontFile() (font_scanner.go:L93)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

---
type: Module
title: parsePSLine
description: "Graphify community 157: internal/scanner/processes.go, internal/scanner/processes_test.go"
resource: internal/scanner
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: processes, resource: internal/scanner/processes.go, last_modified: "2026-04-07T10:03:32+10:00", digest: f095332194a18619 }
  - { id: processes_test, resource: internal/scanner/processes_test.go, last_modified: "2026-04-02T07:38:07+11:00", digest: 8b2996785426d5fd }
---

# Files
- `internal/scanner/processes.go`
- `internal/scanner/processes_test.go`

# Symbols
- .GetWorkingDir() (internal/scanner/processes.go:L120)
- getWorkingDirDarwin() (internal/scanner/processes.go:L151)
- getWorkingDirLinux() (internal/scanner/processes.go:L176)
- parsePSLine() (internal/scanner/processes.go:L186)
- parseEtime() (internal/scanner/processes.go:L219)
- extractModel() (internal/scanner/processes.go:L278)
- extractSessionID() (internal/scanner/processes.go:L286)
- isExcludedProcess() (internal/scanner/processes.go:L294)
- rawProcess (internal/scanner/processes.go:L38)
- ClaudeCodeProvider (internal/scanner/processes.go:L47)
- .ScanProcesses() (internal/scanner/processes.go:L47)
- processes_test.go (internal/scanner/processes_test.go:L1)
- TestParseEtime() (internal/scanner/processes_test.go:L114)
- TestExtractModel() (internal/scanner/processes_test.go:L172)
- TestExtractSessionID() (internal/scanner/processes_test.go:L205)
- TestIsExcludedProcess() (internal/scanner/processes_test.go:L66)
- TestParsePSLine_HappyPath() (internal/scanner/processes_test.go:L8)

# Depends on
- [time.Time](/modules/time-time.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

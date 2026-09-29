---
type: Module
title: processes.go
description: "Graphify community 210: internal/scanner/processes.go, internal/scanner/processes_test.go, repo_hygiene_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: processes, resource: internal/scanner/processes.go, last_modified: "2026-04-07T10:03:32+10:00", digest: f095332194a18619 }
  - { id: processes_test, resource: internal/scanner/processes_test.go, last_modified: "2026-04-02T07:38:07+11:00", digest: 8b2996785426d5fd }
  - { id: repo_hygiene_test, resource: repo_hygiene_test.go, last_modified: "2026-09-30T00:42:54+10:00", digest: cb0db1482ba89e70 }
---

# Files
- `internal/scanner/processes.go`
- `internal/scanner/processes_test.go`
- `repo_hygiene_test.go`

# Symbols
- processes.go (internal/scanner/processes.go:L1)
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
- repo_hygiene_test.go (repo_hygiene_test.go:L1)
- repoRoot() (repo_hygiene_test.go:L15)
- trackedFiles() (repo_hygiene_test.go:L25)
- TestRepo_GraphifyOutIgnored() (repo_hygiene_test.go:L37)

# Depends on
- [RepoScanner](/modules/reposcanner.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

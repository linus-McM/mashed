---
type: Module
title: WriteFile
description: "Graphify community 73: docs/plans/repo-health-remediation.md, docs/stories/old_stories/meditor-02-readfilebase64-binding.md, docs/stories/old_stories/meditor-04-markdown-editor.md, frontend/wailsjs/go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: repo-health-remediation, resource: docs/plans/repo-health-remediation.md, last_modified: "2026-09-29T06:03:48Z", digest: f5a96b797d990d0e }
  - { id: meditor-02-readfilebase64-binding, resource: docs/stories/old_stories/meditor-02-readfilebase64-binding.md, last_modified: "2026-04-12T10:43:48+10:00", digest: c3e539f11ac58af8 }
  - { id: meditor-04-markdown-editor, resource: docs/stories/old_stories/meditor-04-markdown-editor.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 2c11f30e526240e8 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/plans/repo-health-remediation.md`
- `docs/stories/old_stories/meditor-02-readfilebase64-binding.md`
- `docs/stories/old_stories/meditor-04-markdown-editor.md`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Phase 1 — Security *(P0)* (docs/plans/repo-health-remediation.md:L30)
- Task 1.1 — Lock down the terminal WebSocket (docs/plans/repo-health-remediation.md:L32)
- Task 1.2 — Stop git option injection (docs/plans/repo-health-remediation.md:L39)
- Task 1.3 — Restrict file-access bindings (docs/plans/repo-health-remediation.md:L45)
- Task 1.4 — Smaller hardening (docs/plans/repo-health-remediation.md:L51)
- Developer Notes (docs/stories/old_stories/meditor-02-readfilebase64-binding.md:L13)
- Architecture (docs/stories/old_stories/meditor-02-readfilebase64-binding.md:L15)
- Technical Considerations (docs/stories/old_stories/meditor-02-readfilebase64-binding.md:L21)
- Risks & Edge Cases (docs/stories/old_stories/meditor-02-readfilebase64-binding.md:L28)
- Reference Files (docs/stories/old_stories/meditor-02-readfilebase64-binding.md:L34)
- meditor-04-markdown-editor.md (docs/stories/old_stories/meditor-04-markdown-editor.md:L1)
- Story 4: MarkdownEditor with Milkdown Crepe WYSIWYG (docs/stories/old_stories/meditor-04-markdown-editor.md:L1)
- Scenario 2: Auto-save behavior (docs/stories/old_stories/meditor-04-markdown-editor.md:L117)
- Developer Notes (docs/stories/old_stories/meditor-04-markdown-editor.md:L13)
- Scenario 3: Read-only and error handling (docs/stories/old_stories/meditor-04-markdown-editor.md:L147)
- Architecture (docs/stories/old_stories/meditor-04-markdown-editor.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/meditor-04-markdown-editor.md:L163)
- Definition of Done (docs/stories/old_stories/meditor-04-markdown-editor.md:L194)
- Technical Considerations (docs/stories/old_stories/meditor-04-markdown-editor.md:L24)
- Risks & Edge Cases (docs/stories/old_stories/meditor-04-markdown-editor.md:L39)
- Reference Files (docs/stories/old_stories/meditor-04-markdown-editor.md:L46)
- Acceptance Criteria (docs/stories/old_stories/meditor-04-markdown-editor.md:L52)
- Description (docs/stories/old_stories/meditor-04-markdown-editor.md:L9)
- BDD Test Scenarios (docs/stories/old_stories/meditor-04-markdown-editor.md:L97)
- Scenario 1: WYSIWYG rendering and editing (docs/stories/old_stories/meditor-04-markdown-editor.md:L99)
- ReadFile() (frontend/wailsjs/go/main/App.js:L281)
- ReadFileAtHead() (frontend/wailsjs/go/main/App.js:L285)
- WriteFile() (frontend/wailsjs/go/main/App.js:L469)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [GetConfig](/modules/getconfig.md)
- [ReadFileBase64](/modules/readfilebase64.md)
- [themeInit.js](/modules/themeinit-js.md)

# Features
- no feature plan names these files

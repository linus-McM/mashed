---
type: Module
title: SetEditorSettings
description: "Graphify community 117: docs/SPECIFICATION.md, docs/stories/markdown-toolbar-01-backend-config.md, docs/stories/old_stories/edset-01-backend-editor-settings.md, docs/stories/old_stories/edset-02-edito"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: markdown-toolbar-01-backend-config, resource: docs/stories/markdown-toolbar-01-backend-config.md, last_modified: "2026-04-23T10:53:25+10:00", digest: 14619a436e67dd63 }
  - { id: edset-01-backend-editor-settings, resource: docs/stories/old_stories/edset-01-backend-editor-settings.md, last_modified: "2026-04-12T10:43:48+10:00", digest: dadc9968678e6077 }
  - { id: edset-02-editor-settings-ui, resource: docs/stories/old_stories/edset-02-editor-settings-ui.md, last_modified: "2026-04-12T10:43:48+10:00", digest: e8a511023c499c44 }
  - { id: editorSettings, resource: frontend/src/lib/stores/editorSettings.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8f55b394c63663b4 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/markdown-toolbar-01-backend-config.md`
- `docs/stories/old_stories/edset-01-backend-editor-settings.md`
- `docs/stories/old_stories/edset-02-editor-settings-ui.md`
- `frontend/src/lib/stores/editorSettings.js`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Configuration & Context (docs/SPECIFICATION.md:L834)
- Reference Files (docs/stories/markdown-toolbar-01-backend-config.md:L62)
- edset-01-backend-editor-settings.md (docs/stories/old_stories/edset-01-backend-editor-settings.md:L1)
- Story 1: Backend Editor Settings Config (docs/stories/old_stories/edset-01-backend-editor-settings.md:L1)
- Scenario 3: Validation rejects bad values (docs/stories/old_stories/edset-01-backend-editor-settings.md:L113)
- Developer Notes (docs/stories/old_stories/edset-01-backend-editor-settings.md:L13)
- Scenario 4: Concurrent access (docs/stories/old_stories/edset-01-backend-editor-settings.md:L133)
- Tasks / Subtasks (docs/stories/old_stories/edset-01-backend-editor-settings.md:L143)
- Definition of Done (docs/stories/old_stories/edset-01-backend-editor-settings.md:L166)
- Technical Considerations (docs/stories/old_stories/edset-01-backend-editor-settings.md:L24)
- Risks & Edge Cases (docs/stories/old_stories/edset-01-backend-editor-settings.md:L31)
- Reference Files (docs/stories/old_stories/edset-01-backend-editor-settings.md:L37)
- Acceptance Criteria (docs/stories/old_stories/edset-01-backend-editor-settings.md:L44)
- BDD Test Scenarios (docs/stories/old_stories/edset-01-backend-editor-settings.md:L75)
- Scenario 1: Default settings (docs/stories/old_stories/edset-01-backend-editor-settings.md:L77)
- Description (docs/stories/old_stories/edset-01-backend-editor-settings.md:L9)
- Scenario 2: Persistence round-trip (docs/stories/old_stories/edset-01-backend-editor-settings.md:L97)
- edset-02-editor-settings-ui.md (docs/stories/old_stories/edset-02-editor-settings-ui.md:L1)
- Story 2: Editor Settings Store, Monaco Integration, and Settings UI (docs/stories/old_stories/edset-02-editor-settings-ui.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/edset-02-editor-settings-ui.md:L103)
- Scenario 1: Store initialization (docs/stories/old_stories/edset-02-editor-settings-ui.md:L105)
- Scenario 2: Real-time Monaco update (docs/stories/old_stories/edset-02-editor-settings-ui.md:L124)
- Developer Notes (docs/stories/old_stories/edset-02-editor-settings-ui.md:L13)
- Scenario 3: Settings UI controls (docs/stories/old_stories/edset-02-editor-settings-ui.md:L141)
- Architecture (docs/stories/old_stories/edset-02-editor-settings-ui.md:L15)
- Scenario 4: lineHeight fix (docs/stories/old_stories/edset-02-editor-settings-ui.md:L159)
- Tasks / Subtasks (docs/stories/old_stories/edset-02-editor-settings-ui.md:L169)
- Definition of Done (docs/stories/old_stories/edset-02-editor-settings-ui.md:L197)
- Technical Considerations (docs/stories/old_stories/edset-02-editor-settings-ui.md:L45)
- Risks & Edge Cases (docs/stories/old_stories/edset-02-editor-settings-ui.md:L52)
- Reference Files (docs/stories/old_stories/edset-02-editor-settings-ui.md:L59)
- Acceptance Criteria (docs/stories/old_stories/edset-02-editor-settings-ui.md:L66)
- Description (docs/stories/old_stories/edset-02-editor-settings-ui.md:L9)
- initEditorSettings() (frontend/src/lib/stores/editorSettings.js:L29)
- updateEditorSetting() (frontend/src/lib/stores/editorSettings.js:L51)
- DefaultEditorSettings() (frontend/wailsjs/go/main/App.js:L13)
- PickFile() (frontend/wailsjs/go/main/App.js:L269)
- SetEditorSettings() (frontend/wailsjs/go/main/App.js:L361)
- SetMonoFont() (frontend/wailsjs/go/main/App.js:L377)
- GetEditorSettings() (frontend/wailsjs/go/main/App.js:L77)

# Depends on
- [mashedConfig](/modules/mashedconfig.md)

# Inferred
- [applyTheme](/modules/applytheme.md)
- [GetConfig](/modules/getconfig.md)
- [mashedConfig](/modules/mashedconfig.md)
- [SetTheme](/modules/settheme.md)

# Features
- no feature plan names these files

---
type: Module
title: SetEditorSettings
description: "Graphify community 117: docs/SPECIFICATION.md, docs/plans/markdown-toolbar-settings.md, docs/stories/markdown-toolbar-01-backend-config.md, docs/stories/markdown-toolbar-02-frontend-store.md, docs/sto"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: markdown-toolbar-settings, resource: docs/plans/markdown-toolbar-settings.md, last_modified: "2026-04-23T10:53:10+10:00", digest: b031e8aa3523f352 }
  - { id: markdown-toolbar-01-backend-config, resource: docs/stories/markdown-toolbar-01-backend-config.md, last_modified: "2026-04-23T10:53:25+10:00", digest: 14619a436e67dd63 }
  - { id: markdown-toolbar-02-frontend-store, resource: docs/stories/markdown-toolbar-02-frontend-store.md, last_modified: "2026-04-23T11:02:33+10:00", digest: fa89a5180f4ab303 }
  - { id: markdown-toolbar-07-app-hydration, resource: docs/stories/markdown-toolbar-07-app-hydration.md, last_modified: "2026-04-23T11:09:52+10:00", digest: dcf0f3341a4bef15 }
  - { id: edset-01-backend-editor-settings, resource: docs/stories/old_stories/edset-01-backend-editor-settings.md, last_modified: "2026-04-12T10:43:48+10:00", digest: dadc9968678e6077 }
  - { id: edset-02-editor-settings-ui, resource: docs/stories/old_stories/edset-02-editor-settings-ui.md, last_modified: "2026-04-12T10:43:48+10:00", digest: e8a511023c499c44 }
  - { id: editorSettings, resource: frontend/src/lib/stores/editorSettings.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8f55b394c63663b4 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/plans/markdown-toolbar-settings.md`
- `docs/stories/markdown-toolbar-01-backend-config.md`
- `docs/stories/markdown-toolbar-02-frontend-store.md`
- `docs/stories/markdown-toolbar-07-app-hydration.md`
- `docs/stories/old_stories/edset-01-backend-editor-settings.md`
- `docs/stories/old_stories/edset-02-editor-settings-ui.md`
- `frontend/src/lib/stores/editorSettings.js`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Configuration & Context (docs/SPECIFICATION.md:L834)
- Resolved decisions (docs/plans/markdown-toolbar-settings.md:L420)
- markdown-toolbar-01-backend-config.md (docs/stories/markdown-toolbar-01-backend-config.md:L1)
- Story 01: Backend Config — MarkdownMenuSettings and Wails bindings (docs/stories/markdown-toolbar-01-backend-config.md:L1)
- Description (docs/stories/markdown-toolbar-01-backend-config.md:L10)
- BDD Test Scenarios (docs/stories/markdown-toolbar-01-backend-config.md:L101)
- Developer Notes (docs/stories/markdown-toolbar-01-backend-config.md:L14)
- Tasks / Subtasks (docs/stories/markdown-toolbar-01-backend-config.md:L145)
- Architecture (docs/stories/markdown-toolbar-01-backend-config.md:L16)
- Definition of Done (docs/stories/markdown-toolbar-01-backend-config.md:L166)
- Technical Considerations (docs/stories/markdown-toolbar-01-backend-config.md:L32)
- Wails Binding Regeneration (docs/stories/markdown-toolbar-01-backend-config.md:L37)
- Default Values (authoritative — DO NOT change without updating the plan) (docs/stories/markdown-toolbar-01-backend-config.md:L45)
- Risks & Edge Cases (docs/stories/markdown-toolbar-01-backend-config.md:L56)
- Reference Files (docs/stories/markdown-toolbar-01-backend-config.md:L62)
- Acceptance Criteria (docs/stories/markdown-toolbar-01-backend-config.md:L66)
- Developer Notes (docs/stories/markdown-toolbar-02-frontend-store.md:L14)
- Architecture (docs/stories/markdown-toolbar-02-frontend-store.md:L16)
- Behavior contract (docs/stories/markdown-toolbar-02-frontend-store.md:L41)
- Defaults authority (docs/stories/markdown-toolbar-02-frontend-store.md:L47)
- Technical Considerations (docs/stories/markdown-toolbar-02-frontend-store.md:L50)
- Risks & Edge Cases (docs/stories/markdown-toolbar-02-frontend-store.md:L54)
- Reference Files (docs/stories/markdown-toolbar-02-frontend-store.md:L59)
- Risks & Edge Cases (docs/stories/markdown-toolbar-07-app-hydration.md:L50)
- edset-01-backend-editor-settings.md (docs/stories/old_stories/edset-01-backend-editor-settings.md:L1)
- Story 1: Backend Editor Settings Config (docs/stories/old_stories/edset-01-backend-editor-settings.md:L1)
- Scenario 3: Validation rejects bad values (docs/stories/old_stories/edset-01-backend-editor-settings.md:L113)
- Developer Notes (docs/stories/old_stories/edset-01-backend-editor-settings.md:L13)
- Scenario 4: Concurrent access (docs/stories/old_stories/edset-01-backend-editor-settings.md:L133)
- Tasks / Subtasks (docs/stories/old_stories/edset-01-backend-editor-settings.md:L143)
- Architecture (docs/stories/old_stories/edset-01-backend-editor-settings.md:L15)
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
- Developer Notes (docs/stories/old_stories/edset-02-editor-settings-ui.md:L13)
- Architecture (docs/stories/old_stories/edset-02-editor-settings-ui.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/edset-02-editor-settings-ui.md:L169)
- Definition of Done (docs/stories/old_stories/edset-02-editor-settings-ui.md:L197)
- Technical Considerations (docs/stories/old_stories/edset-02-editor-settings-ui.md:L45)
- Risks & Edge Cases (docs/stories/old_stories/edset-02-editor-settings-ui.md:L52)
- Reference Files (docs/stories/old_stories/edset-02-editor-settings-ui.md:L59)
- Acceptance Criteria (docs/stories/old_stories/edset-02-editor-settings-ui.md:L66)
- Description (docs/stories/old_stories/edset-02-editor-settings-ui.md:L9)
- editorSettings.js (frontend/src/lib/stores/editorSettings.js:L1)
- editorSettings (frontend/src/lib/stores/editorSettings.js:L23)
- initEditorSettings() (frontend/src/lib/stores/editorSettings.js:L29)
- updateEditorSetting() (frontend/src/lib/stores/editorSettings.js:L51)
- defaults (frontend/src/lib/stores/editorSettings.js:L7)
- DefaultEditorSettings() (frontend/wailsjs/go/main/App.js:L13)
- DefaultMarkdownMenuSettings() (frontend/wailsjs/go/main/App.js:L17)
- PickFile() (frontend/wailsjs/go/main/App.js:L269)
- SetEditorSettings() (frontend/wailsjs/go/main/App.js:L361)
- SetMarkdownMenuSettings() (frontend/wailsjs/go/main/App.js:L373)
- SetMonoFont() (frontend/wailsjs/go/main/App.js:L377)
- GetEditorSettings() (frontend/wailsjs/go/main/App.js:L77)
- GetMarkdownMenuSettings() (frontend/wailsjs/go/main/App.js:L89)

# Depends on
- [App.js](/modules/app-js.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)

# Inferred
- [applyTheme](/modules/applytheme.md)
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- [mashedConfig](/modules/mashedconfig.md)
- [SetTheme](/modules/settheme.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)

# Features
- no feature plan names these files

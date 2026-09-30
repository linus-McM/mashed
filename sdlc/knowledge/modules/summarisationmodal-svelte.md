---
type: Module
title: SummarisationModal.svelte
description: "Graphify community 8: docs/stories/old_stories/review-04-refactor-plan-agent.md, docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md, docs/stories/old_stories/review-scoped-03-wired-sco"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: review-04-refactor-plan-agent, resource: docs/stories/old_stories/review-04-refactor-plan-agent.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 209b70908e44c963 }
  - { id: review-scoped-01-backend-scoped-advice, resource: docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 24ac699c8ee46519 }
  - { id: review-scoped-03-wired-scoped-flow, resource: docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 0702d010516c97b3 }
  - { id: SummarisationModal, resource: frontend/src/views/SummarisationModal.svelte, last_modified: "2026-04-22T20:32:12+10:00", digest: 1b621514b92c549f }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
  - { id: plan, resource: sdlc/repo-health-remediation/plan.md, last_modified: "2026-09-30T01:21:31+10:00", digest: a05fe85f9adc1fbd }
---

# Files
- `docs/stories/old_stories/review-04-refactor-plan-agent.md`
- `docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md`
- `docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md`
- `frontend/src/views/SummarisationModal.svelte`
- `frontend/wailsjs/go/main/App.js`
- `sdlc/repo-health-remediation/plan.md`

# Symbols
- Story 4: Refactor Plan Agent (docs/stories/old_stories/review-04-refactor-plan-agent.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/review-04-refactor-plan-agent.md:L123)
- Scenario 1: Plan creation happy path (docs/stories/old_stories/review-04-refactor-plan-agent.md:L125)
- Developer Notes (docs/stories/old_stories/review-04-refactor-plan-agent.md:L13)
- Scenario 2: Input validation (docs/stories/old_stories/review-04-refactor-plan-agent.md:L145)
- Architecture (docs/stories/old_stories/review-04-refactor-plan-agent.md:L15)
- Scenario 3: Frontend interaction (docs/stories/old_stories/review-04-refactor-plan-agent.md:L161)
- Tasks / Subtasks (docs/stories/old_stories/review-04-refactor-plan-agent.md:L182)
- Backend: SpawnRefactorPlan (docs/stories/old_stories/review-04-refactor-plan-agent.md:L20)
- Definition of Done (docs/stories/old_stories/review-04-refactor-plan-agent.md:L204)
- Frontend Changes (docs/stories/old_stories/review-04-refactor-plan-agent.md:L60)
- Technical Considerations (docs/stories/old_stories/review-04-refactor-plan-agent.md:L70)
- Risks & Edge Cases (docs/stories/old_stories/review-04-refactor-plan-agent.md:L78)
- Reference Files (docs/stories/old_stories/review-04-refactor-plan-agent.md:L86)
- Description (docs/stories/old_stories/review-04-refactor-plan-agent.md:L9)
- Acceptance Criteria (docs/stories/old_stories/review-04-refactor-plan-agent.md:L92)
- review-scoped-01-backend-scoped-advice.md (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L1)
- Story 1: StreamScopedAdvice Backend Method (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L1)
- Scenario 2: Context prepend (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L115)
- Scenario 3: Validation and error handling (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L136)
- Tasks / Subtasks (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L153)
- Definition of Done (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L172)
- Acceptance Criteria (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L57)
- BDD Test Scenarios (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L89)
- Description (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L9)
- Scenario 1: Scoped diff assembly (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L91)
- review-scoped-03-wired-scoped-flow.md (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L1)
- Story 3: Wire Selection to Scoped Advice and Enriched Plan (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L1)
- Import Changes (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L115)
- Technical Considerations (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L124)
- Developer Notes (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L13)
- Risks & Edge Cases (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L131)
- Reference Files (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L137)
- Acceptance Criteria (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L143)
- Architecture (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L15)
- BDD Test Scenarios (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L177)
- Scenario 1: Scoped advice call (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L179)
- Scenario 2: Incremental re-generation (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L202)
- State Changes (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L21)
- Scenario 3: Enriched plan (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L226)
- Scenario 4: End-to-end flow (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L244)
- Tasks / Subtasks (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L268)
- Definition of Done (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L291)
- Function Changes (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L30)
- Description (docs/stories/old_stories/review-scoped-03-wired-scoped-flow.md:L9)
- SummarisationModal.svelte (frontend/src/views/SummarisationModal.svelte:L1)
- close() (frontend/src/views/SummarisationModal.svelte:L127)
- handleKeydown() (frontend/src/views/SummarisationModal.svelte:L131)
- openFile() (frontend/src/views/SummarisationModal.svelte:L135)
- toggleFile() (frontend/src/views/SummarisationModal.svelte:L139)
- handleCardKey() (frontend/src/views/SummarisationModal.svelte:L156)
- buildAdditionalContext() (frontend/src/views/SummarisationModal.svelte:L163)
- buildEnrichedAdvice() (frontend/src/views/SummarisationModal.svelte:L178)
- getAdvice() (frontend/src/views/SummarisationModal.svelte:L193)
- SpawnPRReview() (frontend/wailsjs/go/main/App.js:L429)
- SpawnRefactorPlan() (frontend/wailsjs/go/main/App.js:L433)
- StreamScopedAdvice() (frontend/wailsjs/go/main/App.js:L457)
- PR 2: Broken features and races (branch `repo-health/pr2-features`) (sdlc/repo-health-remediation/plan.md:L493)

# Depends on
- [App.js](/modules/app-js.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [runtime.js](/modules/runtime-js.md)
- [StreamAdvice](/modules/streamadvice.md)
- [svelte](/modules/svelte.md)

# Inferred
- [StreamAdvice](/modules/streamadvice.md)
- [svelte](/modules/svelte.md)
- [themeInit.js](/modules/themeinit-js.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

---
type: Module
title: StreamAdvice
description: "Graphify community 117: docs/SPECIFICATION.md, docs/stories/old_stories/review-02-review-backend.md, docs/stories/old_stories/review-03-summarisation-modal.md, docs/stories/old_stories/review-04-refac"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: review-02-review-backend, resource: docs/stories/old_stories/review-02-review-backend.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 897b72d9f41644f3 }
  - { id: review-03-summarisation-modal, resource: docs/stories/old_stories/review-03-summarisation-modal.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 81afe9ff7b51b451 }
  - { id: review-04-refactor-plan-agent, resource: docs/stories/old_stories/review-04-refactor-plan-agent.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 209b70908e44c963 }
  - { id: review-06-gitpanel-integration, resource: docs/stories/old_stories/review-06-gitpanel-integration.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 05598baa95b4f568 }
  - { id: review-backlog, resource: docs/stories/old_stories/review-backlog.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 225e9d8bb7cf6c45 }
  - { id: review-scoped-01-backend-scoped-advice, resource: docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 24ac699c8ee46519 }
  - { id: reviewEvents, resource: frontend/src/types/reviewEvents.ts, last_modified: "2026-04-22T19:23:56+10:00", digest: 49fb1a6f6b36384f }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/review-02-review-backend.md`
- `docs/stories/old_stories/review-03-summarisation-modal.md`
- `docs/stories/old_stories/review-04-refactor-plan-agent.md`
- `docs/stories/old_stories/review-06-gitpanel-integration.md`
- `docs/stories/old_stories/review-backlog.md`
- `docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md`
- `frontend/src/types/reviewEvents.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 9. UI Views (docs/SPECIFICATION.md:L776)
- Setup View (`Setup.svelte`) (docs/SPECIFICATION.md:L778)
- Notification Feed (`NotificationFeed.svelte`) (docs/SPECIFICATION.md:L781)
- Agent Detail (`AgentDetail.svelte`) (docs/SPECIFICATION.md:L788)
- Workflow Builder (`WorkflowBuilder.svelte`) (docs/SPECIFICATION.md:L797)
- Settings (`Settings.svelte`) (docs/SPECIFICATION.md:L812)
- Summarisation (`SummarisationModal.svelte`) (docs/SPECIFICATION.md:L820)
- About (`AboutModal.svelte`) (docs/SPECIFICATION.md:L823)
- review-02-review-backend.md (docs/stories/old_stories/review-02-review-backend.md:L1)
- Story 2: Code Review Summary & Advice Streaming Backend (docs/stories/old_stories/review-02-review-backend.md:L1)
- Acceptance Criteria (docs/stories/old_stories/review-02-review-backend.md:L101)
- Developer Notes (docs/stories/old_stories/review-02-review-backend.md:L13)
- BDD Test Scenarios (docs/stories/old_stories/review-02-review-backend.md:L136)
- Scenario 1: Streaming summary happy path (docs/stories/old_stories/review-02-review-backend.md:L138)
- Architecture (docs/stories/old_stories/review-02-review-backend.md:L15)
- Scenario 2: Advice streaming (docs/stories/old_stories/review-02-review-backend.md:L155)
- Scenario 3: Edge cases (docs/stories/old_stories/review-02-review-backend.md:L173)
- Tasks / Subtasks (docs/stories/old_stories/review-02-review-backend.md:L195)
- Definition of Done (docs/stories/old_stories/review-02-review-backend.md:L226)
- Data Flow (docs/stories/old_stories/review-02-review-backend.md:L55)
- Technical Considerations (docs/stories/old_stories/review-02-review-backend.md:L71)
- Risks & Edge Cases (docs/stories/old_stories/review-02-review-backend.md:L85)
- Description (docs/stories/old_stories/review-02-review-backend.md:L9)
- Reference Files (docs/stories/old_stories/review-02-review-backend.md:L93)
- review-03-summarisation-modal.md (docs/stories/old_stories/review-03-summarisation-modal.md:L1)
- Story 3: Summarisation Modal Frontend (docs/stories/old_stories/review-03-summarisation-modal.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/review-03-summarisation-modal.md:L123)
- Scenario 1: Full summarisation flow (docs/stories/old_stories/review-03-summarisation-modal.md:L125)
- Developer Notes (docs/stories/old_stories/review-03-summarisation-modal.md:L13)
- Scenario 2: Advice interaction (docs/stories/old_stories/review-03-summarisation-modal.md:L146)
- Architecture (docs/stories/old_stories/review-03-summarisation-modal.md:L15)
- Scenario 3: Modal lifecycle (docs/stories/old_stories/review-03-summarisation-modal.md:L167)
- Tasks / Subtasks (docs/stories/old_stories/review-03-summarisation-modal.md:L190)
- Definition of Done (docs/stories/old_stories/review-03-summarisation-modal.md:L222)
- Component Structure (docs/stories/old_stories/review-03-summarisation-modal.md:L26)
- Data Flow (docs/stories/old_stories/review-03-summarisation-modal.md:L43)
- Technical Considerations (docs/stories/old_stories/review-03-summarisation-modal.md:L53)
- Risks & Edge Cases (docs/stories/old_stories/review-03-summarisation-modal.md:L64)
- Reference Files (docs/stories/old_stories/review-03-summarisation-modal.md:L73)
- Acceptance Criteria (docs/stories/old_stories/review-03-summarisation-modal.md:L80)
- Description (docs/stories/old_stories/review-03-summarisation-modal.md:L9)
- review-04-refactor-plan-agent.md (docs/stories/old_stories/review-04-refactor-plan-agent.md:L1)
- review-06-gitpanel-integration.md (docs/stories/old_stories/review-06-gitpanel-integration.md:L1)
- Story 6: GitPanel Integration & Wiring (docs/stories/old_stories/review-06-gitpanel-integration.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/review-06-gitpanel-integration.md:L122)
- Scenario 1: Button rendering (docs/stories/old_stories/review-06-gitpanel-integration.md:L124)
- Developer Notes (docs/stories/old_stories/review-06-gitpanel-integration.md:L13)
- Scenario 2: Modal lifecycle (docs/stories/old_stories/review-06-gitpanel-integration.md:L146)
- Architecture (docs/stories/old_stories/review-06-gitpanel-integration.md:L15)
- Scenario 3: Event forwarding (docs/stories/old_stories/review-06-gitpanel-integration.md:L168)
- Tasks / Subtasks (docs/stories/old_stories/review-06-gitpanel-integration.md:L179)
- Definition of Done (docs/stories/old_stories/review-06-gitpanel-integration.md:L196)
- Frontend Changes (docs/stories/old_stories/review-06-gitpanel-integration.md:L26)
- Technical Considerations (docs/stories/old_stories/review-06-gitpanel-integration.md:L66)
- Risks & Edge Cases (docs/stories/old_stories/review-06-gitpanel-integration.md:L75)
- Reference Files (docs/stories/old_stories/review-06-gitpanel-integration.md:L81)
- Acceptance Criteria (docs/stories/old_stories/review-06-gitpanel-integration.md:L88)
- Description (docs/stories/old_stories/review-06-gitpanel-integration.md:L9)
- review-backlog.md (docs/stories/old_stories/review-backlog.md:L1)
- Sprint Backlog: Code Review Summarisation Panel (docs/stories/old_stories/review-backlog.md:L1)
- Dependency Graph (docs/stories/old_stories/review-backlog.md:L17)
- Sprint Backlog (docs/stories/old_stories/review-backlog.md:L3)
- Recommended Sprint Order (docs/stories/old_stories/review-backlog.md:L35)
- Parallelization Opportunities (docs/stories/old_stories/review-backlog.md:L44)
- Risk Summary (docs/stories/old_stories/review-backlog.md:L49)
- Developer Notes (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L13)
- Architecture (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L15)
- Technical Considerations (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L31)
- Risks & Edge Cases (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L41)
- Reference Files (docs/stories/old_stories/review-scoped-01-backend-scoped-advice.md:L49)
- reviewEvents.ts (frontend/src/types/reviewEvents.ts:L1)
- ReviewFileSummary (frontend/src/types/reviewEvents.ts:L12)
- ReviewSummary (frontend/src/types/reviewEvents.ts:L21)
- ReviewSummaryProgressEvent (frontend/src/types/reviewEvents.ts:L28)
- ReviewSummaryDoneEvent (frontend/src/types/reviewEvents.ts:L37)
- ReviewAdviceProgressEvent (frontend/src/types/reviewEvents.ts:L47)
- GitCommitStreaming() (frontend/wailsjs/go/main/App.js:L137)
- ListAdviceModes() (frontend/wailsjs/go/main/App.js:L181)
- ReadFileDiff() (frontend/wailsjs/go/main/App.js:L297)
- StreamAdvice() (frontend/wailsjs/go/main/App.js:L449)
- StreamCodeReviewSummary() (frontend/wailsjs/go/main/App.js:L453)

# Depends on
- [SummarisationModal.svelte](/modules/summarisationmodal-svelte.md)

# Inferred
- [SummarisationModal.svelte](/modules/summarisationmodal-svelte.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)

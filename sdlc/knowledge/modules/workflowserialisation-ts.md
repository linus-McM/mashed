---
type: Module
title: workflowSerialisation.ts
description: "Graphify community 78: docs/stories/skills-cmd-03-canvas-integration.md, docs/stories/svelte-check-01-js-stores-to-ts.md, docs/stories/svelte-check-01-report.md, frontend/src/components/bmad/__tests__"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: skills-cmd-03-canvas-integration, resource: docs/stories/skills-cmd-03-canvas-integration.md, last_modified: "2026-04-12T11:45:37+10:00", digest: 7b560cdd8c7cab67 }
  - { id: svelte-check-01-js-stores-to-ts, resource: docs/stories/svelte-check-01-js-stores-to-ts.md, last_modified: "2026-04-22T17:06:36+10:00", digest: 3cdcd1064aabe570 }
  - { id: svelte-check-01-report, resource: docs/stories/svelte-check-01-report.md, last_modified: "2026-04-22T17:53:53+10:00", digest: d63f920199e68cfd }
  - { id: CanvasPane.drop.test, resource: frontend/src/components/bmad/__tests__/CanvasPane.drop.test.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 457c024a3dab690f }
  - { id: canvasPaneDropHandler, resource: frontend/src/components/bmad/canvasPaneDropHandler.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: e28c1bbafab2e8be }
  - { id: workflowSerialisation.test, resource: frontend/src/lib/__tests__/workflowSerialisation.test.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: b75e10f1d14592c0 }
  - { id: workflowSerialisation, resource: frontend/src/lib/workflowSerialisation.ts, last_modified: "2026-04-28T12:36:05+10:00", digest: 36a9064bc7d2060a }
  - { id: workflow, resource: frontend/src/types/workflow.ts, last_modified: "2026-04-23T13:05:14+10:00", digest: dde86e2d10f68917 }
---

# Files
- `docs/stories/skills-cmd-03-canvas-integration.md`
- `docs/stories/svelte-check-01-js-stores-to-ts.md`
- `docs/stories/svelte-check-01-report.md`
- `frontend/src/components/bmad/__tests__/CanvasPane.drop.test.ts`
- `frontend/src/components/bmad/canvasPaneDropHandler.ts`
- `frontend/src/lib/__tests__/workflowSerialisation.test.ts`
- `frontend/src/lib/workflowSerialisation.ts`
- `frontend/src/types/workflow.ts`

# Symbols
- Description (docs/stories/skills-cmd-03-canvas-integration.md:L9)
- Technical Considerations (docs/stories/svelte-check-01-js-stores-to-ts.md:L30)
- Deferred follow-ups (docs/stories/svelte-check-01-report.md:L61)
- CanvasPane.drop.test.ts (frontend/src/components/bmad/__tests__/CanvasPane.drop.test.ts:L1)
- DroppedCommandNode (frontend/src/components/bmad/__tests__/CanvasPane.drop.test.ts:L21)
- canvasPaneDropHandler.ts (frontend/src/components/bmad/canvasPaneDropHandler.ts:L1)
- MashedAssetPayload (frontend/src/components/bmad/canvasPaneDropHandler.ts:L18)
- HandleMashedAssetDropArgs (frontend/src/components/bmad/canvasPaneDropHandler.ts:L25)
- handleMashedAssetDrop() (frontend/src/components/bmad/canvasPaneDropHandler.ts:L32)
- workflowSerialisation.test.ts (frontend/src/lib/__tests__/workflowSerialisation.test.ts:L1)
- makeWorkflowFixture() (frontend/src/lib/__tests__/workflowSerialisation.test.ts:L244)
- makeMixedCanvas() (frontend/src/lib/__tests__/workflowSerialisation.test.ts:L30)
- noopInferEdgeLabel() (frontend/src/lib/__tests__/workflowSerialisation.test.ts:L73)
- workflowSerialisation.ts (frontend/src/lib/workflowSerialisation.ts:L1)
- recordOr() (frontend/src/lib/workflowSerialisation.ts:L100)
- snapshotCanvas() (frontend/src/lib/workflowSerialisation.ts:L112)
- canvasNodesToWorkflowNodes() (frontend/src/lib/workflowSerialisation.ts:L146)
- canvasEdgesToWorkflowEdges() (frontend/src/lib/workflowSerialisation.ts:L175)
- COMMAND_NODE_SENTINEL_PHRASE (frontend/src/lib/workflowSerialisation.ts:L192)
- COMMAND_NODE_SENTINEL_DEFAULT_BODY (frontend/src/lib/workflowSerialisation.ts:L193)
- workflowNodesToCanvasNodes() (frontend/src/lib/workflowSerialisation.ts:L209)
- workflowEdgesToCanvasEdges() (frontend/src/lib/workflowSerialisation.ts:L242)
- isWorkflowShape() (frontend/src/lib/workflowSerialisation.ts:L281)
- serialise() (frontend/src/lib/workflowSerialisation.ts:L303)
- deserialise() (frontend/src/lib/workflowSerialisation.ts:L324)
- SavedWorkflowNode (frontend/src/lib/workflowSerialisation.ts:L36)
- SavedWorkflowEdge (frontend/src/lib/workflowSerialisation.ts:L50)
- SerialisedWorkflowNode (frontend/src/lib/workflowSerialisation.ts:L64)
- SerialisedWorkflowEdge (frontend/src/lib/workflowSerialisation.ts:L77)
- coordOrZero() (frontend/src/lib/workflowSerialisation.ts:L90)
- stringOr() (frontend/src/lib/workflowSerialisation.ts:L95)
- ProcessRegistryEntry (frontend/src/types/workflow.ts:L102)
- Position (frontend/src/types/workflow.ts:L21)
- Workflow (frontend/src/types/workflow.ts:L30)
- CanvasNode (frontend/src/types/workflow.ts:L76)
- CanvasEdge (frontend/src/types/workflow.ts:L84)

# Depends on
- [svelte](/modules/svelte.md)
- [vitest](/modules/vitest.md)

# Inferred
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Features
- no feature plan names these files

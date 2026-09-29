---
type: Module
title: ProcessNode.status.test.ts
description: "Graphify community 46: docs/stories/bmad-interactive-07-registry-entries.md, frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts, frontend/src/components/bmad/__tests__/CommandNode.s"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: bmad-interactive-07-registry-entries, resource: docs/stories/bmad-interactive-07-registry-entries.md, last_modified: "2026-04-20T15:36:58+10:00", digest: 47d47db81112ff4e }
  - { id: CommandNode.breadcrumb.test, resource: frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: d5f2905d83e01211 }
  - { id: CommandNode.status.test, resource: frontend/src/components/bmad/__tests__/CommandNode.status.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: 3fa251016ae15b6d }
  - { id: MultiFileLoaderNode.test, resource: frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: 05eb5149cdc50787 }
  - { id: ProcessNode.multiInput.test, resource: frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: b50d732d5ea73c1a }
  - { id: ProcessNode.status.test, resource: frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: cb84331adc8f4a2d }
  - { id: mountSvelte, resource: frontend/src/components/bmad/__tests__/mountSvelte.ts, last_modified: "2026-04-22T20:01:25+10:00", digest: becfbd6c142f9b1b }
  - { id: workflow, resource: frontend/src/types/workflow.ts, last_modified: "2026-04-23T13:05:14+10:00", digest: dde86e2d10f68917 }
---

# Files
- `docs/stories/bmad-interactive-07-registry-entries.md`
- `frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts`
- `frontend/src/components/bmad/__tests__/CommandNode.status.test.ts`
- `frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts`
- `frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts`
- `frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts`
- `frontend/src/components/bmad/__tests__/mountSvelte.ts`
- `frontend/src/types/workflow.ts`

# Symbols
- Exact registry shapes (docs/stories/bmad-interactive-07-registry-entries.md:L40)
- CommandNode.breadcrumb.test.ts (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L1)
- CommandNodeProps (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L24)
- BreadcrumbOverrides (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L30)
- makeNodeProps() (frontend/src/components/bmad/__tests__/CommandNode.breadcrumb.test.ts:L38)
- CommandNode.status.test.ts (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L1)
- StatusToken (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L18)
- CommandNodeProps (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L20)
- StatusOverrides (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L26)
- baseProps() (frontend/src/components/bmad/__tests__/CommandNode.status.test.ts:L32)
- MultiFileLoaderNode.test.ts (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L1)
- Entry (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L19)
- MultiFileLoaderProps (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L24)
- MultiFileLoaderOverrides (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L30)
- makeProps() (frontend/src/components/bmad/__tests__/MultiFileLoaderNode.test.ts:L34)
- ProcessNode.multiInput.test.ts (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L1)
- TestProcess (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L22)
- TestNodeData (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L30)
- ProcessNodeProps (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L37)
- MultiInputOverrides (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L43)
- makeProps() (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L50)
- collectBreadcrumbRows() (frontend/src/components/bmad/__tests__/ProcessNode.multiInput.test.ts:L70)
- ProcessNode.status.test.ts (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L1)
- StatusToken (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L18)
- Phase (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L19)
- AgentRole (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L20)
- TestProcess (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L33)
- TestNodeData (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L41)
- ProcessNodeProps (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L51)
- StatusOverrides (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L57)
- baseProps() (frontend/src/components/bmad/__tests__/ProcessNode.status.test.ts:L71)
- MountedComponent (frontend/src/components/bmad/__tests__/mountSvelte.ts:L21)
- mountComponent() (frontend/src/components/bmad/__tests__/mountSvelte.ts:L48)
- CanvasNodeData (frontend/src/types/workflow.ts:L49)

# Depends on
- [mountSvelte.ts](/modules/mountsvelte-ts.md)
- [ProcessNode.svelte](/modules/processnode-svelte.md)
- [svelte](/modules/svelte.md)
- [vitest](/modules/vitest.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files

/**
 * @vitest-environment jsdom
 */
// Coverage companion for ProcessNode.svelte: status-row, artifact-indicators,
// story-badge, phase/role branches. Works with the multi-input loop shipped
// in story breadcrumbs-04.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

import { mountComponent, type MountedComponent } from './mountSvelte';

// `vi.mock` factory is hoisted but may await a dynamic import — keeps the
// xyflow stub centralised in `mountSvelte.ts`.
vi.mock('@xyflow/svelte', async () => (await import('./mountSvelte')).xyflowHandleStub());

import ProcessNode from '../ProcessNode.svelte';

type StatusToken = 'pending' | 'running' | 'complete' | 'failed' | 'skipped';
type Phase = 'analysis' | 'planning' | 'solutioning' | 'implementation' | 'support';
type AgentRole =
  | 'analyst'
  | 'pm'
  | 'ux-designer'
  | 'architect'
  | 'developer'
  | 'tech-writer'
  | 'qa';

// Test-local shape matches what ProcessNode.svelte *reads* off `data.process`
// — not the full `ProcessDef` (which carries id/skillName/description/...
// required fields the component ignores). The Svelte constructor API accepts
// any props object; svelte-check verifies our scaffolding, not prop parity.
interface TestProcess {
  name?: string;
  phase?: string;
  agentRole?: string;
  inputs?: string[];
  outputs?: string[];
}

interface TestNodeData {
  process?: TestProcess | null;
  status?: StatusToken;
  config?: Record<string, unknown>;
  artifactStatus?: { found: string[]; missing: string[] };
  storyId?: string;
  storyStatus?: string;
  label?: string;
}

interface ProcessNodeProps {
  data: TestNodeData;
  id: string;
  selected: boolean;
}

interface StatusOverrides {
  status?: StatusToken;
  phase?: string;
  agentRole?: string;
  inputs?: string[];
  outputs?: string[];
  artifactStatus?: { found: string[]; missing: string[] };
  storyId?: string;
  storyStatus?: string;
  label?: string;
  config?: Record<string, unknown>;
  selected?: boolean;
}

function baseProps(overrides: StatusOverrides = {}): ProcessNodeProps {
  return {
    data: {
      process: {
        name: 'plan',
        phase: overrides.phase ?? 'planning',
        agentRole: overrides.agentRole ?? 'pm',
        inputs: overrides.inputs ?? [],
        outputs: overrides.outputs ?? [],
      },
      status: overrides.status ?? 'pending',
      config: overrides.config ?? {},
      artifactStatus: overrides.artifactStatus,
      storyId: overrides.storyId,
      storyStatus: overrides.storyStatus,
      label: overrides.label,
    },
    id: 'p-x',
    selected: overrides.selected ?? false,
  };
}

describe('ProcessNode — branch coverage', () => {
  let container: HTMLDivElement;
  let node: MountedComponent | null = null;

  beforeEach(() => {
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    node?.$destroy();
    container.remove();
    node = null;
  });

  const statuses: StatusToken[] = ['pending', 'running', 'complete', 'failed', 'skipped'];
  it.each(statuses)('renders status=%s status row', (status) => {
    node = mountComponent(ProcessNode, container, baseProps({ status }));
    expect(container.querySelector('.status-row')).not.toBeNull();
  });

  it('applies selected class when selected=true', () => {
    node = mountComponent(ProcessNode, container, baseProps({ selected: true }));
    expect(container.querySelector('.process-node.selected')).not.toBeNull();
  });

  it('applies running class when status=running', () => {
    node = mountComponent(ProcessNode, container, baseProps({ status: 'running' }));
    expect(container.querySelector('.process-node.running')).not.toBeNull();
  });

  const phases: Phase[] = ['analysis', 'planning', 'solutioning', 'implementation', 'support'];
  it.each(phases)('accepts phase=%s without throwing', (phase) => {
    node = mountComponent(ProcessNode, container, baseProps({ phase }));
    expect(container.querySelector('.phase-bar')).not.toBeNull();
  });

  it('falls back for unknown phase', () => {
    node = mountComponent(ProcessNode, container, baseProps({ phase: 'unknown-xyz' }));
    expect(container.querySelector('.phase-bar')).not.toBeNull();
  });

  const roles: AgentRole[] = [
    'analyst',
    'pm',
    'ux-designer',
    'architect',
    'developer',
    'tech-writer',
    'qa',
  ];
  it.each(roles)('renders role=%s icon', (agentRole) => {
    node = mountComponent(ProcessNode, container, baseProps({ agentRole }));
    expect(container.querySelector('.role-icon')).not.toBeNull();
  });

  it('renders artifact indicators when complete with found + missing', () => {
    node = mountComponent(
      ProcessNode,
      container,
      baseProps({
        status: 'complete',
        artifactStatus: { found: ['prd'], missing: ['architecture'] },
      }),
    );
    const icons = container.querySelectorAll('.artifact-icon');
    expect(icons.length).toBe(2);
    expect(container.querySelector('.artifact-icon.found')).not.toBeNull();
    expect(container.querySelector('.artifact-icon.missing')).not.toBeNull();
  });

  it('omits artifact indicators when status is not complete', () => {
    node = mountComponent(
      ProcessNode,
      container,
      baseProps({ status: 'pending', artifactStatus: { found: ['prd'], missing: [] } }),
    );
    expect(container.querySelector('.artifact-indicators')).toBeNull();
  });

  it('omits artifact indicators when found + missing are both empty', () => {
    node = mountComponent(
      ProcessNode,
      container,
      baseProps({ status: 'complete', artifactStatus: { found: [], missing: [] } }),
    );
    expect(container.querySelector('.artifact-indicators')).toBeNull();
  });

  it('renders story badge when storyId set', () => {
    node = mountComponent(
      ProcessNode,
      container,
      baseProps({ storyId: 'S-12', storyStatus: 'in_progress' }),
    );
    const badge = container.querySelector('.story-badge-id');
    expect(badge?.textContent).toBe('S-12');
  });

  it('omits story badge when storyId missing', () => {
    node = mountComponent(ProcessNode, container, baseProps({}));
    expect(container.querySelector('.story-badge')).toBeNull();
  });

  it('renders data.label when process.name missing', () => {
    const props: ProcessNodeProps = {
      data: {
        process: { inputs: [], outputs: [] },
        label: 'custom-label',
        status: 'pending',
        config: {},
      },
      id: 'p-l',
      selected: false,
    };
    node = mountComponent(ProcessNode, container, props);
    expect(container.querySelector('.node-label')?.textContent).toBe('custom-label');
  });

  it('falls back to "Process" when no name or label', () => {
    const props: ProcessNodeProps = {
      data: {
        process: { inputs: [], outputs: [] },
        status: 'pending',
        config: {},
      },
      id: 'p-fb',
      selected: false,
    };
    node = mountComponent(ProcessNode, container, props);
    expect(container.querySelector('.node-label')?.textContent).toBe('Process');
  });

  it('renders outputs with unresolved breadcrumbs when no paths set', () => {
    node = mountComponent(
      ProcessNode,
      container,
      baseProps({ outputs: ['draft'], config: {} }),
    );
    const rows = container.querySelectorAll('.breadcrumb-row.unresolved');
    expect(rows.length).toBeGreaterThanOrEqual(1);
  });
});

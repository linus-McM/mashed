/**
 * @vitest-environment jsdom
 */
// Coverage companion for ProcessNode.svelte: status-row, artifact-indicators,
// story-badge, phase/role branches. Works with the multi-input loop shipped
// in story breadcrumbs-04.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

vi.mock('@xyflow/svelte', () => ({
  Handle: class {
    $$ = {
      fragment: { c() {}, m() {}, p() {}, d() {}, l() {} },
      on_mount: [],
      on_destroy: [],
      after_update: [],
    };
    constructor(_opts) {}
    $set(_props) {}
    $destroy() { this.$$.fragment = null; this.$$.on_destroy = []; }
    $on(_event, _fn) { return () => {}; }
  },
  Position: { Left: 'left', Right: 'right', Top: 'top', Bottom: 'bottom' },
}));

import ProcessNode from '../ProcessNode.svelte';

function mount(container, props) {
  return new ProcessNode({ target: container, props });
}

function baseProps(overrides = {}) {
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
  let container;
  let node = null;

  beforeEach(() => {
    container = document.createElement('div');
    document.body.appendChild(container);
  });

  afterEach(() => {
    node?.$destroy();
    container.remove();
    node = null;
  });

  it.each(['pending', 'running', 'complete', 'failed', 'skipped'])(
    'renders status=%s status row',
    (status) => {
      node = mount(container, baseProps({ status }));
      expect(container.querySelector('.status-row')).not.toBeNull();
    },
  );

  it('applies selected class when selected=true', () => {
    node = mount(container, baseProps({ selected: true }));
    expect(container.querySelector('.process-node.selected')).not.toBeNull();
  });

  it('applies running class when status=running', () => {
    node = mount(container, baseProps({ status: 'running' }));
    expect(container.querySelector('.process-node.running')).not.toBeNull();
  });

  it.each(['analysis', 'planning', 'solutioning', 'implementation', 'support'])(
    'accepts phase=%s without throwing',
    (phase) => {
      node = mount(container, baseProps({ phase }));
      expect(container.querySelector('.phase-bar')).not.toBeNull();
    },
  );

  it('falls back for unknown phase', () => {
    node = mount(container, baseProps({ phase: 'unknown-xyz' }));
    expect(container.querySelector('.phase-bar')).not.toBeNull();
  });

  it.each(['analyst', 'pm', 'ux-designer', 'architect', 'developer', 'tech-writer', 'qa'])(
    'renders role=%s icon',
    (agentRole) => {
      node = mount(container, baseProps({ agentRole }));
      expect(container.querySelector('.role-icon')).not.toBeNull();
    },
  );

  it('renders artifact indicators when complete with found + missing', () => {
    node = mount(container, baseProps({
      status: 'complete',
      artifactStatus: { found: ['prd'], missing: ['architecture'] },
    }));
    const icons = container.querySelectorAll('.artifact-icon');
    expect(icons.length).toBe(2);
    expect(container.querySelector('.artifact-icon.found')).not.toBeNull();
    expect(container.querySelector('.artifact-icon.missing')).not.toBeNull();
  });

  it('omits artifact indicators when status is not complete', () => {
    node = mount(container, baseProps({
      status: 'pending',
      artifactStatus: { found: ['prd'], missing: [] },
    }));
    expect(container.querySelector('.artifact-indicators')).toBeNull();
  });

  it('omits artifact indicators when found + missing are both empty', () => {
    node = mount(container, baseProps({
      status: 'complete',
      artifactStatus: { found: [], missing: [] },
    }));
    expect(container.querySelector('.artifact-indicators')).toBeNull();
  });

  it('renders story badge when storyId set', () => {
    node = mount(container, baseProps({ storyId: 'S-12', storyStatus: 'in_progress' }));
    const badge = container.querySelector('.story-badge-id');
    expect(badge?.textContent).toBe('S-12');
  });

  it('omits story badge when storyId missing', () => {
    node = mount(container, baseProps({}));
    expect(container.querySelector('.story-badge')).toBeNull();
  });

  it('renders data.label when process.name missing', () => {
    node = new ProcessNode({
      target: container,
      props: {
        data: { process: { inputs: [], outputs: [] }, label: 'custom-label', status: 'pending', config: {} },
        id: 'p-l',
        selected: false,
      },
    });
    expect(container.querySelector('.node-label')?.textContent).toBe('custom-label');
  });

  it('falls back to "Process" when no name or label', () => {
    node = new ProcessNode({
      target: container,
      props: {
        data: { process: { inputs: [], outputs: [] }, status: 'pending', config: {} },
        id: 'p-fb',
        selected: false,
      },
    });
    expect(container.querySelector('.node-label')?.textContent).toBe('Process');
  });

  it('renders outputs with unresolved breadcrumbs when no paths set', () => {
    node = mount(container, baseProps({
      outputs: ['draft'],
      config: {},
    }));
    const rows = container.querySelectorAll('.breadcrumb-row.unresolved');
    expect(rows.length).toBeGreaterThanOrEqual(1);
  });
});

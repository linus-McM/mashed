/**
 * @vitest-environment jsdom
 */
// frontend/src/components/bmad/__tests__/CommandNode.status.test.js
// Coverage companion for CommandNode.svelte status-row branches.
// Exercises running/complete/failed/skipped/pending so all branches are hit.

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

import CommandNode from '../CommandNode.svelte';

function mount(container, props) {
  return new CommandNode({ target: container, props });
}

function baseProps(overrides = {}) {
  return {
    data: {
      config: {
        commandName: 'cmd',
        commandDescription: overrides.desc ?? 'desc',
      },
      status: overrides.status ?? 'pending',
    },
    id: 'cmd-x',
    selected: overrides.selected ?? false,
  };
}

describe('CommandNode — status branch coverage', () => {
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
    'renders status=%s without throwing',
    (status) => {
      node = mount(container, baseProps({ status }));
      const row = container.querySelector('.status-row');
      expect(row).not.toBeNull();
    },
  );

  it('applies selected class when selected=true', () => {
    node = mount(container, baseProps({ selected: true }));
    expect(container.querySelector('.command-node.selected')).not.toBeNull();
  });

  it('omits description when commandDescription empty', () => {
    node = mount(container, baseProps({ desc: '' }));
    expect(container.querySelector('.description')).toBeNull();
  });

  it('falls back to data.label when commandName missing', () => {
    node = new CommandNode({
      target: container,
      props: {
        data: { config: {}, label: 'fallback-label', status: 'pending' },
        id: 'cmd-fb',
        selected: false,
      },
    });
    expect(container.querySelector('.name')?.textContent).toBe('fallback-label');
  });
});

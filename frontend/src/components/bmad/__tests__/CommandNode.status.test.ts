/**
 * @vitest-environment jsdom
 */
// Coverage companion for CommandNode.svelte status-row branches.
// Exercises running/complete/failed/skipped/pending so all branches are hit.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

import type { CanvasNodeData } from '../../../types/workflow';
import { mountComponent, type MountedComponent } from './mountSvelte';

// `vi.mock` factory is hoisted but may await a dynamic import — keeps the
// xyflow stub centralised in `mountSvelte.ts`.
vi.mock('@xyflow/svelte', async () => (await import('./mountSvelte')).xyflowHandleStub());

import CommandNode from '../CommandNode.svelte';

type StatusToken = 'pending' | 'running' | 'complete' | 'failed' | 'skipped';

interface CommandNodeProps {
  data: CanvasNodeData;
  id: string;
  selected: boolean;
}

interface StatusOverrides {
  status?: StatusToken;
  desc?: string;
  selected?: boolean;
}

function baseProps(overrides: StatusOverrides = {}): CommandNodeProps {
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
  it.each(statuses)('renders status=%s without throwing', (status) => {
    node = mountComponent(CommandNode, container, baseProps({ status }));
    const row = container.querySelector('.status-row');
    expect(row).not.toBeNull();
  });

  it('applies selected class when selected=true', () => {
    node = mountComponent(CommandNode, container, baseProps({ selected: true }));
    expect(container.querySelector('.command-node.selected')).not.toBeNull();
  });

  it('omits description when commandDescription empty', () => {
    node = mountComponent(CommandNode, container, baseProps({ desc: '' }));
    expect(container.querySelector('.description')).toBeNull();
  });

  it('falls back to data.label when commandName missing', () => {
    const props: CommandNodeProps = {
      data: { config: {}, label: 'fallback-label', status: 'pending' },
      id: 'cmd-fb',
      selected: false,
    };
    node = mountComponent(CommandNode, container, props);
    expect(container.querySelector('.name')?.textContent).toBe('fallback-label');
  });
});

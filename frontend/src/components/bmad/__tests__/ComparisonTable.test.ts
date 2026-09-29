/**
 * @vitest-environment jsdom
 */
import { describe, it, expect } from 'vitest';

import ComparisonTable from '../ComparisonTable.svelte';
import { makeMount } from './mountSvelte';

const COLUMNS = ['Option', 'Latency', 'Cost'];
const ROWS = [
  ['A', '10ms', '$1'],
  ['B', '20ms', '$2'],
  ['C', '30ms', '$3'],
  ['D', '40ms', '$4'],
];

describe('ComparisonTable', () => {
  const mount = makeMount();
  const render = (heading = 'Trade-offs', columns = COLUMNS, rows = ROWS) =>
    mount(ComparisonTable, { heading, columns, rows });

  it('renders the heading', () => {
    expect(render().textContent).toContain('Trade-offs');
  });

  it('renders thead with one th per column', () => {
    const ths = render().querySelectorAll('thead th');
    expect(ths.length).toBe(3);
    expect(ths[0].textContent).toContain('Option');
    expect(ths[1].textContent).toContain('Latency');
    expect(ths[2].textContent).toContain('Cost');
  });

  it('renders tbody with one tr per row and cells in column order', () => {
    const trs = render().querySelectorAll('tbody tr');
    expect(trs.length).toBe(4);
    const firstCells = trs[0].querySelectorAll('td');
    expect(firstCells[0].textContent).toContain('A');
    expect(firstCells[1].textContent).toContain('10ms');
    expect(firstCells[2].textContent).toContain('$1');
  });

  it('distinguishes even rows from odd rows (zebra class or style)', () => {
    const trs = render().querySelectorAll<HTMLElement>('tbody tr');
    const [odd, even] = [trs[0], trs[1]];
    const classDiffers = odd.className !== even.className;
    const bgDiffers =
      window.getComputedStyle(odd).backgroundColor !==
      window.getComputedStyle(even).backgroundColor;
    expect(classDiffers || bgDiffers).toBe(true);
  });
});

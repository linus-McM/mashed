import { describe, it, expect, vi } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { runSave } from '../../lib/saveFeedback';

// R7 (UI): the markdown editor announces a rejected save and keeps the buffer.

describe('MarkdownEditor save rejection (R7)', () => {
  it('reports the backend reason and keeps the markdown buffer', async () => {
    const markdown = '# draft';
    const write = vi.fn().mockRejectedValue(new Error('read file: "/opt/x.md": path outside allowed roots'));

    const result = await runSave(() => markdown, write);

    expect(result.status).toBe('error');
    expect(result.message).toMatch(/outside/i);
    expect(write).toHaveBeenCalledWith('# draft');
    expect(markdown).toBe('# draft');
  });

  it('renders the save error in an aria-live region', () => {
    const src = readFileSync(resolve(__dirname, '../MarkdownEditor.svelte'), 'utf8');
    expect(src).toMatch(/runSave\(/);
    expect(src).toMatch(/class="save-status error"[^>]*role="status"[^>]*aria-live="polite"[^>]*>\{saveError\}/);
  });
});

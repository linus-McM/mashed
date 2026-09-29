import { describe, it, expect, vi } from 'vitest';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { runSave, describeSaveError } from '../../lib/saveFeedback';

// R7 (UI): when the backend rejects a save because the path is outside the
// allowed roots, the editor announces the reason and keeps the buffer.

describe('MonacoEditor save rejection (R7)', () => {
  it('reports a rejected save with the reason and never touches the content', async () => {
    let content = 'unsaved edits';
    const getContent = vi.fn(() => content);
    const write = vi.fn().mockRejectedValue('write file: "/etc/x": path outside allowed roots');

    const result = await runSave(getContent, write);

    expect(result.status).toBe('error');
    expect(result.message).toMatch(/outside/i);
    expect(result.message).toMatch(/home|dev/i);
    expect(getContent).toHaveBeenCalledOnce();
    expect(content).toBe('unsaved edits');
  });

  it('maps a write-protected rejection to its own message', () => {
    expect(describeSaveError('write file: "~/.zshrc": path is write-protected')).toMatch(/write-protected/i);
  });

  it('returns saved with no message on success', async () => {
    const result = await runSave(() => 'x', vi.fn().mockResolvedValue(undefined));
    expect(result).toEqual({ status: 'saved', message: '' });
  });

  it('renders the save error in an aria-live region', () => {
    const src = readFileSync(resolve(__dirname, '../MonacoEditor.svelte'), 'utf8');
    expect(src).toMatch(/runSave\(/);
    expect(src).toMatch(/class="save-status error"[^>]*role="status"[^>]*aria-live="polite"[^>]*>\{saveError\}/);
  });
});

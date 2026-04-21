/**
 * @vitest-environment jsdom
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { tick } from 'svelte';

import CodeBlock from '../CodeBlock.svelte';
import { makeMount } from './mountSvelte';

describe('CodeBlock', () => {
  const mount = makeMount();
  let writeText: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    writeText = vi.fn().mockResolvedValue(undefined);
    // jsdom has no clipboard by default — install a stub we can assert against.
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText },
    });
  });

  const render = (content: string, copyable = true, lang = 'sh') =>
    mount(CodeBlock, { content, copyable, lang });

  it('renders code content verbatim', () => {
    const el = render('rm -rf ~');
    expect(el.querySelector('[data-testid="code-block-content"]')?.textContent).toContain('rm -rf ~');
  });

  it('AC11_copy_writes_verbatim — clicking copy calls clipboard.writeText with exact content', async () => {
    const el = render('rm -rf ~');
    const btn = el.querySelector<HTMLButtonElement>('[data-testid="code-copy-button"]');
    expect(btn).not.toBeNull();
    btn!.click();
    await tick();
    await Promise.resolve();
    expect(writeText).toHaveBeenCalledWith('rm -rf ~');
  });

  it('AC11_copy_label_flips_to_copied — label reads "Copied" after successful copy', async () => {
    const el = render('rm -rf ~');
    const btn = el.querySelector<HTMLButtonElement>('[data-testid="code-copy-button"]')!;
    btn.click();
    await tick();
    await Promise.resolve();
    await tick();
    expect(btn.textContent).toContain('Copied');
  });

  it('does NOT render the copy button when copyable=false', () => {
    const el = render('rm -rf ~', false);
    expect(el.querySelector('[data-testid="code-copy-button"]')).toBeNull();
  });
});

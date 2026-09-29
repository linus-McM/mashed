/**
 * @vitest-environment jsdom
 */
import { describe, it, expect } from 'vitest';

import HintBanner from '../HintBanner.svelte';
import { makeMount } from './mountSvelte';

const TONES = ['info', 'warn', 'error', 'success'] as const;
type Tone = typeof TONES[number];

describe('HintBanner', () => {
  const mount = makeMount();
  const render = (tone: Tone, content = 'hint text') =>
    mount(HintBanner, { tone, content });

  for (const tone of TONES) {
    it(`renders tone="${tone}" with distinguishing data-tone attribute`, () => {
      const el = render(tone);
      const banner = el.querySelector<HTMLElement>('[data-testid="hint-banner"]');
      expect(banner?.getAttribute('data-tone')).toBe(tone);
    });

    it(`renders tone="${tone}" with an icon element`, () => {
      const el = render(tone);
      expect(el.querySelector('[data-testid="hint-banner-icon"]')).not.toBeNull();
    });

    it(`renders tone="${tone}" mono uppercase tonal label (Brief §2 + §6.c)`, () => {
      const el = render(tone);
      const label = el.querySelector<HTMLElement>('[data-testid="hint-banner-label"]');
      expect(label?.textContent).toBe(tone.toUpperCase());
    });
  }

  it('renders body content', () => {
    const el = render('info', 'check the logs');
    expect(el.textContent).toContain('check the logs');
  });
});

// Story 07: App Hydration — initMarkdownMenuSettings on mount.
//
// Option B: focused unit tests on the hydration helper + source-grep over
// App.svelte. Full App.svelte mount requires mocking Wails runtime, EventsOn,
// GetConfig/GetDevDir/GetNotifications/GetEditorSettings/ListLocalFonts,
// SetActiveContext, WriteConsoleLog, every store module, and lucide-svelte —
// well over five mocks. The narrower approach asserts the contract that
// matters: `App.svelte` contains the exact hydration call and does NOT
// introduce a redundant `GetMarkdownMenuSettings` round-trip, while the
// hydration helper itself is exercised against the same inputs App supplies.

import { describe, it, expect, beforeEach } from 'vitest';
import { get } from 'svelte/store';
import { readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import {
  markdownMenuSettings,
  initMarkdownMenuSettings,
  type MarkdownMenuSettings,
} from './lib/stores/markdownMenuSettings';

const DEFAULTS: MarkdownMenuSettings = {
  bold: true,
  italic: true,
  strikethrough: true,
  code: true,
  link: true,
  latex: false,
};

// Resolve App.svelte via import.meta.url so the test is cwd-independent.
// `fileURLToPath` decodes correctly on every platform; the bare `.pathname`
// trick fails under vitest/jsdom which rewrites the URL origin.
const APP_SVELTE_PATH = join(
  dirname(fileURLToPath(import.meta.url)),
  'App.svelte',
);
const APP_SVELTE_SOURCE = readFileSync(APP_SVELTE_PATH, 'utf8');

beforeEach(() => {
  // Reset store between cases.
  initMarkdownMenuSettings(null);
});

describe('Story 07 AC-1: persisted settings hydrate the store', () => {
  it('initMarkdownMenuSettings({bold:false, italic:true, strikethrough:true, code:true, link:true, latex:true}) -> store reads those values', () => {
    const persisted: MarkdownMenuSettings = {
      bold: false,
      italic: true,
      strikethrough: true,
      code: true,
      link: true,
      latex: true,
    };

    initMarkdownMenuSettings(persisted);

    expect(get(markdownMenuSettings)).toEqual(persisted);
  });
});

describe('Story 07 AC-2: missing markdownMenu falls back to defaults', () => {
  it('initMarkdownMenuSettings(undefined) -> store reads DEFAULTS', () => {
    // Start from a non-default state to prove the reset.
    markdownMenuSettings.set({ ...DEFAULTS, bold: false, latex: true });

    initMarkdownMenuSettings(undefined);

    expect(get(markdownMenuSettings)).toEqual(DEFAULTS);
  });
});

describe('Story 07 AC-4: App.svelte uses cfg.markdownMenu, not a second Wails call', () => {
  it('App.svelte source contains `initMarkdownMenuSettings(cfg.markdownMenu)` inside onMount', () => {
    // Exact call shape from the story spec.
    expect(APP_SVELTE_SOURCE).toMatch(/initMarkdownMenuSettings\(cfg\.markdownMenu\)/);

    // Sanity: the call sits inside an onMount block. We assert that the
    // substring "onMount" appears before the hydration call in the source.
    const onMountIdx = APP_SVELTE_SOURCE.indexOf('onMount(');
    const callIdx = APP_SVELTE_SOURCE.indexOf(
      'initMarkdownMenuSettings(cfg.markdownMenu)',
    );
    expect(onMountIdx).toBeGreaterThanOrEqual(0);
    expect(callIdx).toBeGreaterThan(onMountIdx);
  });

  it('App.svelte imports initMarkdownMenuSettings from the store module', () => {
    expect(APP_SVELTE_SOURCE).toMatch(
      /import\s*\{\s*initMarkdownMenuSettings\s*\}\s*from\s*'\.\/lib\/stores\/markdownMenuSettings'/,
    );
  });

  it('App.svelte does NOT import or call GetMarkdownMenuSettings', () => {
    expect(APP_SVELTE_SOURCE).not.toMatch(/GetMarkdownMenuSettings/);
  });

  it('hydration call is wrapped in try/catch (AC-3)', () => {
    // The exact pattern from the story spec. Guards against a regression where
    // someone removes the try/catch and a throwing init takes down the rest
    // of onMount (themes, fonts, editor settings).
    expect(APP_SVELTE_SOURCE).toMatch(
      /try\s*\{\s*initMarkdownMenuSettings\(cfg\.markdownMenu\);\s*\}\s*catch\s*\{\s*\}/,
    );
  });

  it('hydration call is synchronous — no `await` in front', () => {
    // Matches any whitespace before the call that is NOT an `await`.
    // The exact code we added uses `try { initMarkdownMenuSettings(...); } catch {}`.
    expect(APP_SVELTE_SOURCE).not.toMatch(
      /await\s+initMarkdownMenuSettings/,
    );
  });
});

/**
 * @vitest-environment jsdom
 */
// Tests for Settings.svelte — Story 04 (Settings layout refactor: 2-column panel grid).
//
// Mocks every Wails binding + runtime + theme/font stores so the view can mount
// inside jsdom without touching Go or the filesystem. The assertions target
// AC-1..AC-6 from docs/stories/markdown-toolbar-04-settings-layout-refactor.md.
//
// Limitation: jsdom does not evaluate CSS `@media` queries, so AC-3 (narrow
// viewport collapses to single column) is verified by grepping the compiled
// <style> block for the `@media (max-width: 1100px)` rule rather than by
// computed-style assertion. End-to-end visual verification is deferred to
// manual / Playwright checks per the story AC table.

import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { writable } from 'svelte/store';

// --- Wails binding mocks ---------------------------------------------------

vi.mock('../../wailsjs/go/main/App.js', () => ({
  GetConfig: vi.fn(async () => ({
    vscodiumExtPath: '',
    monoFont: 'Geist Mono',
    fontSize: 13,
    sidebarWidth: 280,
  })),
  SetTheme: vi.fn(async () => undefined),
  SetImportedTheme: vi.fn(async () => undefined),
  SetVSCodiumExtPath: vi.fn(async () => undefined),
  PickDirectory: vi.fn(async () => ''),
  ListVSCodiumThemes: vi.fn(async () => []),
  ListLocalFonts: vi.fn(async () => []),
  SetMonoFont: vi.fn(async () => undefined),
  SetFontSize: vi.fn(async () => undefined),
  SetSidebarWidth: vi.fn(async () => undefined),
}));

vi.mock('../../wailsjs/runtime/runtime.js', () => ({
  BrowserOpenURL: vi.fn(),
  EventsOn: vi.fn(() => () => undefined),
  EventsOff: vi.fn(),
}));

// --- Store mocks -----------------------------------------------------------

vi.mock('../lib/stores/theme.js', () => {
  const allThemes = writable({
    dark: {
      label: 'Dark',
      css: {
        '--bg-deepest': '#07080a',
        '--bg-surface': '#0d0f12',
        '--border-subtle': '#1e2530',
        '--accent-green': '#00e57a',
        '--accent-purple': '#9d6fff',
        '--text-dim': '#4a5a6a',
      },
    },
  });
  const themeIds = writable(['dark']);
  const currentThemeId = writable('dark');
  return {
    allThemes,
    themeIds,
    currentThemeId,
    applyTheme: vi.fn(),
    builtInThemeIds: ['dark'],
  };
});

vi.mock('../lib/themeInit.js', () => ({
  activateImportedTheme: vi.fn(async () => undefined),
  removeImportedTheme: vi.fn(async () => undefined),
  makeThemeId: vi.fn((path: string, ext: string) => `${ext}::${path}`),
}));

vi.mock('../lib/stores/font.js', () => ({
  applyFont: vi.fn(),
  registerLocalFonts: vi.fn(),
}));

vi.mock('../lib/stores/editorSettings.js', () => {
  const editorSettings = writable({
    cursorStyle: 'line',
    cursorBlinking: 'blink',
    wordWrap: 'off',
    lineNumbers: 'on',
    renderLineHighlight: 'line',
    renderWhitespace: 'none',
    minimapEnabled: true,
    tabSize: 2,
    insertSpaces: true,
    bracketPairColorization: true,
    fontLigatures: false,
    scrollBeyondLastLine: false,
    smoothScrolling: false,
  });
  return {
    editorSettings,
    updateEditorSetting: vi.fn(),
  };
});

vi.mock('../lib/stores/uiAdapterSettings', () => ({
  uiAdapterEnabled: writable(false),
  uiAdapterTimeoutMs: writable(3000),
  ollamaModel: writable('gemma3:4b'),
  ollamaEnabled: writable(false),
  uiAdapterUntrustedExpanded: writable(false),
  ollamaReachable: writable(null),
  ollamaModels: writable([]),
  hydrate: vi.fn(async () => undefined),
  refreshModels: vi.fn(async () => undefined),
  setEnabled: vi.fn(async () => true),
  setTimeoutMs: vi.fn(async () => true),
  setModel: vi.fn(async () => true),
  setOllamaEnabled: vi.fn(async () => true),
  setUntrustedExpanded: vi.fn(async () => true),
  validOllamaModelName: vi.fn(() => true),
}));

// --- Import AFTER mocks so the component picks up the stubs ---------------

import Settings from './Settings.svelte';

type Mounted = { $destroy(): void };
type SvelteInit = new (opts: { target: HTMLElement; props: object }) => Mounted;

function mount(): { target: HTMLElement; instance: Mounted } {
  const target = document.createElement('div');
  document.body.appendChild(target);
  const instance = new (Settings as unknown as SvelteInit)({ target, props: {} });
  return { target, instance };
}

// --- Settings.svelte source text (for style-block assertions) -------------

const here = dirname(fileURLToPath(import.meta.url));
const settingsSource = readFileSync(resolve(here, 'Settings.svelte'), 'utf-8');
const styleBlock = (() => {
  const m = settingsSource.match(/<style>([\s\S]*)<\/style>/);
  return m ? m[1] : '';
})();

// ---------------------------------------------------------------------------

describe('Settings layout — Story 04 (AC-1, AC-5)', () => {
  let mounted: { target: HTMLElement; instance: Mounted } | null = null;

  beforeEach(() => {
    mounted = null;
  });

  afterEach(() => {
    mounted?.instance.$destroy();
    mounted?.target.remove();
  });

  it('AC-1 renders a .col-settings with exactly two .settings-col children', () => {
    mounted = mount();
    const grid = mounted.target.querySelector('.col-settings');
    expect(grid).not.toBeNull();
    const cols = mounted.target.querySelectorAll('.col-settings > .settings-col');
    expect(cols.length).toBe(2);
    expect(mounted.target.querySelector('.settings-col.settings-col-1')).not.toBeNull();
    expect(mounted.target.querySelector('.settings-col.settings-col-2')).not.toBeNull();
  });

  it('AC-1 column 1 contains Font, Sidebar Width, and Editor panels', () => {
    mounted = mount();
    const col1 = mounted.target.querySelector('.settings-col-1');
    expect(col1).not.toBeNull();
    const titles = Array.from(col1!.querySelectorAll('.section-title')).map(
      (n) => n.textContent?.trim(),
    );
    expect(titles).toEqual(['Font', 'Sidebar Width', 'Editor']);
  });

  it('AC-1 column 2 contains Theme Extensions (and UI AST adapter if present)', () => {
    mounted = mount();
    const col2 = mounted.target.querySelector('.settings-col-2');
    expect(col2).not.toBeNull();
    const titles = Array.from(col2!.querySelectorAll('.section-title')).map(
      (n) => n.textContent?.trim(),
    );
    expect(titles[0]).toBe('Theme Extensions');
    // UI AST adapter is present in the current Settings.svelte — guard future drift.
    expect(titles).toContain('UI AST adapter');
  });

  it('regression: every .settings-section is wrapped in exactly one .settings-panel', () => {
    mounted = mount();
    const sections = mounted.target.querySelectorAll('.col-settings .settings-section');
    const panels = mounted.target.querySelectorAll('.col-settings .settings-panel');
    expect(sections.length).toBeGreaterThan(0);
    // Panel count must equal section count — regression guard for story 05.
    expect(panels.length).toBe(sections.length);
    // Each section's parent must be a .settings-panel (inline wrapper convention).
    for (const section of sections) {
      expect(section.parentElement?.classList.contains('settings-panel')).toBe(true);
    }
  });
});

describe('Settings layout — Story 04 style-block assertions (AC-2, AC-3, AC-6)', () => {
  it('AC-2 .settings-panel rule uses design tokens for bg, border, radius, padding', () => {
    // Extract the .settings-panel block (non-media, non-override one).
    const rule = styleBlock.match(/\.settings-panel\s*\{([^}]*)\}/);
    expect(rule).not.toBeNull();
    const body = rule![1];
    expect(body).toMatch(/background:\s*var\(--bg-surface\)/);
    expect(body).toMatch(/border:\s*1px solid var\(--border-subtle\)/);
    expect(body).toMatch(/border-radius:\s*var\(--radius-md\)/);
    expect(body).toMatch(/padding:\s*var\(--sp-lg\)/);
  });

  it('AC-5 .settings-panel .settings-section override zeroes margin-bottom', () => {
    expect(styleBlock).toMatch(
      /\.settings-panel\s+\.settings-section\s*\{[^}]*margin-bottom:\s*0[^}]*\}/,
    );
  });

  it('AC-1 .col-settings uses grid-template-columns: 1fr 1fr with align-items: start', () => {
    const rule = styleBlock.match(/\.col-settings\s*\{([^}]*)\}/);
    expect(rule).not.toBeNull();
    const body = rule![1];
    expect(body).toMatch(/display:\s*grid/);
    expect(body).toMatch(/grid-template-columns:\s*1fr\s+1fr/);
    expect(body).toMatch(/gap:\s*var\(--sp-lg\)/);
    expect(body).toMatch(/align-items:\s*start/);
  });

  it('AC-3 narrow viewport fallback: @media (max-width: 1100px) collapses the grid', () => {
    // jsdom does not apply @media rules, so assert the rule exists in source.
    expect(styleBlock).toMatch(
      /@media\s*\(\s*max-width:\s*1100px\s*\)\s*\{[\s\S]*?\.col-settings\s*\{[\s\S]*?grid-template-columns:\s*1fr\s*;?[\s\S]*?\}[\s\S]*?\}/,
    );
  });

  it('AC-6 no hardcoded hex colors or rgb()/rgba() in the new layout rules', () => {
    // Scope to the three new rules only — legacy section rules are out of scope
    // for this story and already use tokens where required.
    const newRulesStart = styleBlock.indexOf('/* Right column: other settings');
    expect(newRulesStart).toBeGreaterThan(-1);
    const mediaEnd = styleBlock.indexOf('@media (max-width: 1100px)');
    expect(mediaEnd).toBeGreaterThan(-1);
    // Span covers .col-settings, .settings-col, .settings-panel, override, and @media.
    const newSpan = styleBlock.slice(
      newRulesStart,
      mediaEnd + '@media (max-width: 1100px) { .col-settings { grid-template-columns: 1fr; } }'.length,
    );
    // No hex colors (#xxx / #xxxxxx) in the new CSS.
    expect(newSpan).not.toMatch(/#[0-9a-fA-F]{3,8}\b/);
    // No rgb()/rgba() either.
    expect(newSpan).not.toMatch(/rgba?\s*\(/);
  });
});

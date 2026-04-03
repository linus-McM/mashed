import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import {
  convertVSCodeTheme,
  validateConvertedTheme,
  getMonacoBase,
} from './themeConverter.js';

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

/** A Dracula-like dark theme with full colors + tokenColors */
const DRACULA_THEME = {
  name: 'Dracula',
  type: 'dark',
  colors: {
    'editor.background': '#282a36',
    'editor.foreground': '#f8f8f2',
    'editorWidget.background': '#21222c',
    'sideBar.background': '#21222c',
    'list.activeSelectionBackground': '#44475a',
    'editorWidget.border': '#191a21',
    'panel.border': '#191a21',
    'focusBorder': '#6272a4',
    'terminal.ansiGreen': '#50fa7b',
    'terminal.ansiYellow': '#f1fa8c',
    'terminal.ansiRed': '#ff5555',
    'terminal.ansiBlue': '#6272a4',
    'terminal.ansiMagenta': '#ff79c6',
    'terminal.ansiCyan': '#8be9fd',
    'terminal.ansiBlack': '#282a36',
    'terminal.ansiWhite': '#f8f8f2',
    'editorLineNumber.foreground': '#6272a4',
    'editorWhitespace.foreground': '#424450',
    'editor.lineHighlightBackground': '#44475a75',
    'terminal.background': '#282a36',
    'terminal.foreground': '#f8f8f2',
    'terminalCursor.foreground': '#f8f8f2',
    'editor.selectionBackground': '#44475a',
    'editorCursor.foreground': '#f8f8f2',
  },
  tokenColors: [
    {
      scope: ['comment', 'punctuation.definition.comment'],
      settings: { foreground: '#6272A4', fontStyle: 'italic' },
    },
    {
      scope: ['string', 'string.quoted'],
      settings: { foreground: '#F1FA8C' },
    },
    {
      scope: ['keyword', 'storage.type'],
      settings: { foreground: '#FF79C6', fontStyle: 'bold' },
    },
    {
      scope: 'constant.numeric',
      settings: { foreground: '#BD93F9' },
    },
    {
      scope: 'entity.name.function',
      settings: { foreground: '#50FA7B' },
    },
    {
      scope: 'variable',
      settings: { foreground: '#F8F8F2' },
    },
  ],
};

/** A minimal light theme */
const LIGHT_THEME = {
  name: 'Quiet Light',
  type: 'light',
  colors: {
    'editor.background': '#fafafa',
    'editor.foreground': '#333333',
  },
  tokenColors: [
    { scope: 'comment', settings: { foreground: '#aaaaaa', fontStyle: 'italic' } },
    { scope: 'keyword', settings: { foreground: '#7928a1' } },
  ],
};

/** High-contrast black theme */
const HC_BLACK_THEME = {
  name: 'High Contrast',
  uiTheme: 'hc-black',
  colors: {
    'editor.background': '#000000',
    'editor.foreground': '#ffffff',
  },
  tokenColors: [],
};

/** High-contrast light theme */
const HC_LIGHT_THEME = {
  name: 'High Contrast Light',
  type: 'hc-light',
  colors: {},
  tokenColors: [],
};

// ---------------------------------------------------------------------------
// Scenario 1: Full Dracula dark theme conversion
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — Dracula dark theme', () => {
  const result = convertVSCodeTheme(DRACULA_THEME, 'imported-dracula');

  it('has the correct label', () => {
    assert.equal(result.label, 'Dracula');
  });

  it('maps --bg-deepest from editor.background', () => {
    assert.equal(result.css['--bg-deepest'], '#282a36');
  });

  it('maps --bg-surface from editorWidget.background', () => {
    assert.equal(result.css['--bg-surface'], '#21222c');
  });

  it('maps --accent-green from terminal.ansiGreen', () => {
    assert.equal(result.css['--accent-green'], '#50fa7b');
  });

  it('maps --accent-green-dim as dimmed ansiGreen', () => {
    // 0x50=80 -> 80*0.4=32 -> 0x20, 0xfa=250 -> 250*0.4=100 -> 0x64, 0x7b=123 -> 123*0.4=49 -> 0x31
    assert.equal(result.css['--accent-green-dim'], '#206431');
  });

  it('maps --text-primary from editor.foreground', () => {
    assert.equal(result.css['--text-primary'], '#f8f8f2');
  });

  it('maps --text-dim from editorLineNumber.foreground', () => {
    assert.equal(result.css['--text-dim'], '#6272a4');
  });

  it('produces all 16 CSS variables', () => {
    const keys = Object.keys(result.css);
    assert.equal(keys.length, 16);
  });

  it('sets Monaco base to vs-dark', () => {
    assert.equal(result.monaco.base, 'vs-dark');
  });

  it('sets Monaco inherit to true', () => {
    assert.equal(result.monaco.inherit, true);
  });

  it('generates Monaco token rules for comment', () => {
    const commentRule = result.monaco.rules.find(r => r.token === 'comment');
    assert.ok(commentRule, 'Expected a comment rule');
    assert.equal(commentRule.foreground, '6272A4');
    assert.equal(commentRule.fontStyle, 'italic');
  });

  it('generates Monaco token rules for keyword with bold', () => {
    const kw = result.monaco.rules.find(r => r.token === 'keyword');
    assert.ok(kw);
    assert.equal(kw.foreground, 'FF79C6');
    assert.equal(kw.fontStyle, 'bold');
  });

  it('generates Monaco token rules for string', () => {
    const str = result.monaco.rules.find(r => r.token === 'string');
    assert.ok(str);
    assert.equal(str.foreground, 'F1FA8C');
  });

  it('strips # prefix from Monaco foreground values', () => {
    for (const rule of result.monaco.rules) {
      if (rule.foreground) {
        assert.ok(!rule.foreground.startsWith('#'), `Rule ${rule.token} foreground should not start with #`);
      }
    }
  });

  it('generates xterm theme with correct green', () => {
    assert.equal(result.xterm.green, '#50fa7b');
  });

  it('generates xterm theme with background', () => {
    assert.equal(result.xterm.background, '#282a36');
  });

  it('generates xterm theme with foreground', () => {
    assert.equal(result.xterm.foreground, '#f8f8f2');
  });

  it('generates xterm theme with cursor', () => {
    assert.equal(result.xterm.cursor, '#f8f8f2');
  });

  it('generates xterm theme with all 8 ANSI colors', () => {
    for (const key of ['black', 'red', 'green', 'yellow', 'blue', 'magenta', 'cyan', 'white']) {
      assert.ok(result.xterm[key], `xterm.${key} should be defined`);
    }
  });
});

// ---------------------------------------------------------------------------
// Scenario 2: Token specificity ordering
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — token specificity', () => {
  it('more-specific scope wins over general scope', () => {
    const theme = {
      name: 'Specificity Test',
      type: 'dark',
      colors: {},
      tokenColors: [
        { scope: 'comment', settings: { foreground: '#aaaaaa' } },
        { scope: 'comment.block.documentation', settings: { foreground: '#bbbbbb' } },
      ],
    };
    const result = convertVSCodeTheme(theme, 'spec-test');
    const commentRule = result.monaco.rules.find(r => r.token === 'comment');
    assert.ok(commentRule, 'Expected a comment rule');
    assert.equal(commentRule.foreground, 'bbbbbb');
  });

  it('scope as comma-separated string produces rules', () => {
    const theme = {
      name: 'Comma Scope Test',
      type: 'dark',
      colors: {},
      tokenColors: [
        { scope: 'keyword.control, storage.type', settings: { foreground: '#FF79C6' } },
      ],
    };
    const result = convertVSCodeTheme(theme, 'comma-test');
    const kw = result.monaco.rules.find(r => r.token === 'keyword');
    assert.ok(kw, 'Expected a keyword rule from comma-separated scope');
    assert.equal(kw.foreground, 'FF79C6');
  });

  it('scope as array produces rules', () => {
    const theme = {
      name: 'Array Scope Test',
      type: 'dark',
      colors: {},
      tokenColors: [
        { scope: ['string.quoted', 'string'], settings: { foreground: '#E5C07B' } },
      ],
    };
    const result = convertVSCodeTheme(theme, 'array-test');
    const str = result.monaco.rules.find(r => r.token === 'string');
    assert.ok(str);
    assert.equal(str.foreground, 'E5C07B');
  });
});

// ---------------------------------------------------------------------------
// Scenario 3: Hex normalization
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — hex normalization', () => {
  it('normalizes #RGB to #RRGGBB in CSS vars', () => {
    const theme = {
      name: 'Short Hex',
      type: 'dark',
      colors: {
        'editor.background': '#28a',
      },
      tokenColors: [],
    };
    const result = convertVSCodeTheme(theme, 'short-hex');
    assert.equal(result.css['--bg-deepest'], '#2288aa');
  });

  it('strips alpha from #RRGGBBAA in background CSS vars', () => {
    const theme = {
      name: 'Alpha Hex',
      type: 'dark',
      colors: {
        'editor.background': '#282a36ff',
      },
      tokenColors: [],
    };
    const result = convertVSCodeTheme(theme, 'alpha-hex');
    assert.equal(result.css['--bg-deepest'], '#282a36');
  });

  it('dimColor handles #RGB input', () => {
    // Testing indirectly via accent-green-dim
    const theme = {
      name: 'DimColor Short',
      type: 'dark',
      colors: {
        'terminal.ansiGreen': '#5f7',
      },
      tokenColors: [],
    };
    const result = convertVSCodeTheme(theme, 'dim-short');
    // #5f7 -> #55ff77, dimmed by 0.4: R=85*0.4=34=0x22, G=255*0.4=102=0x66, B=119*0.4=48=0x30
    assert.equal(result.css['--accent-green-dim'], '#226630');
  });

  it('dimColor handles #RRGGBBAA input', () => {
    const theme = {
      name: 'DimColor Alpha',
      type: 'dark',
      colors: {
        'terminal.ansiGreen': '#50fa7bff',
      },
      tokenColors: [],
    };
    const result = convertVSCodeTheme(theme, 'dim-alpha');
    // normalized to #50fa7b, then dimmed by 0.4
    assert.equal(result.css['--accent-green-dim'], '#206431');
  });
});

// ---------------------------------------------------------------------------
// Scenario 4: Light theme
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — light theme', () => {
  const result = convertVSCodeTheme(LIGHT_THEME, 'quiet-light');

  it('uses vs base', () => {
    assert.equal(result.monaco.base, 'vs');
  });

  it('maps --bg-deepest from editor.background', () => {
    assert.equal(result.css['--bg-deepest'], '#fafafa');
  });

  it('uses light fallback for --bg-surface when not in colors', () => {
    // editorWidget.background not in LIGHT_THEME.colors, should fall back to #ffffff
    assert.equal(result.css['--bg-surface'], '#ffffff');
  });

  it('uses light fallback for --accent-green when not in colors', () => {
    assert.equal(result.css['--accent-green'], '#059a50');
  });

  it('produces 16 CSS variables', () => {
    assert.equal(Object.keys(result.css).length, 16);
  });
});

// ---------------------------------------------------------------------------
// Scenario 5: High-contrast themes
// ---------------------------------------------------------------------------
describe('getMonacoBase', () => {
  it('maps dark type to vs-dark', () => {
    assert.equal(getMonacoBase({ type: 'dark' }), 'vs-dark');
  });

  it('maps light type to vs', () => {
    assert.equal(getMonacoBase({ type: 'light' }), 'vs');
  });

  it('maps vs to vs', () => {
    assert.equal(getMonacoBase({ type: 'vs' }), 'vs');
  });

  it('maps hc-black type to hc-black', () => {
    assert.equal(getMonacoBase({ type: 'hc-black' }), 'hc-black');
  });

  it('maps hc-black uiTheme to hc-black', () => {
    assert.equal(getMonacoBase({ uiTheme: 'hc-black' }), 'hc-black');
  });

  it('maps hc-light type to hc-light', () => {
    assert.equal(getMonacoBase({ type: 'hc-light' }), 'hc-light');
  });

  it('defaults to vs-dark for unknown type', () => {
    assert.equal(getMonacoBase({}), 'vs-dark');
  });
});

describe('convertVSCodeTheme — hc-black theme', () => {
  const result = convertVSCodeTheme(HC_BLACK_THEME, 'hc-test');

  it('sets Monaco base to hc-black', () => {
    assert.equal(result.monaco.base, 'hc-black');
  });

  it('uses dark defaults for missing CSS vars (hc-black is dark)', () => {
    // hc-black is treated as dark, so accent-green falls back to dark default
    assert.equal(result.css['--accent-green'], '#00e57a');
  });
});

describe('convertVSCodeTheme — hc-light theme', () => {
  const result = convertVSCodeTheme(HC_LIGHT_THEME, 'hc-light-test');

  it('sets Monaco base to hc-light', () => {
    assert.equal(result.monaco.base, 'hc-light');
  });
});

// ---------------------------------------------------------------------------
// Scenario 6: Missing data graceful handling
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — missing tokenColors', () => {
  const theme = { name: 'No Tokens', type: 'dark', colors: { 'editor.background': '#1e1e1e' } };
  const result = convertVSCodeTheme(theme, 'no-tokens');

  it('returns empty rules array', () => {
    assert.deepEqual(result.monaco.rules, []);
  });

  it('sets inherit to true', () => {
    assert.equal(result.monaco.inherit, true);
  });
});

describe('convertVSCodeTheme — missing colors object', () => {
  const theme = {
    name: 'No Colors',
    type: 'dark',
    tokenColors: [
      { scope: 'comment', settings: { foreground: '#aaa' } },
    ],
  };
  const result = convertVSCodeTheme(theme, 'no-colors');

  it('uses all fallback defaults for CSS vars', () => {
    assert.equal(result.css['--bg-deepest'], '#07080a');
    assert.equal(result.css['--text-primary'], '#c8d4e0');
  });

  it('still generates token rules', () => {
    assert.ok(result.monaco.rules.length > 0);
  });

  it('produces an empty Monaco colors object', () => {
    assert.deepEqual(result.monaco.colors, {});
  });
});

describe('convertVSCodeTheme — tokenColor entries with no scope (global)', () => {
  const theme = {
    name: 'Global Token',
    type: 'dark',
    colors: {},
    tokenColors: [
      { settings: { foreground: '#d4d4d4', background: '#1e1e1e' } },
      { scope: 'comment', settings: { foreground: '#608b4e' } },
    ],
  };
  const result = convertVSCodeTheme(theme, 'global-token');

  it('skips entries with no scope without crashing', () => {
    assert.ok(result.monaco.rules.length >= 1);
    const commentRule = result.monaco.rules.find(r => r.token === 'comment');
    assert.ok(commentRule);
    assert.equal(commentRule.foreground, '608b4e');
  });
});

describe('convertVSCodeTheme — tokenColor with empty settings', () => {
  const theme = {
    name: 'Empty Settings',
    type: 'dark',
    colors: {},
    tokenColors: [
      { scope: 'comment', settings: {} },
      { scope: 'keyword', settings: { foreground: '#c586c0' } },
    ],
  };
  const result = convertVSCodeTheme(theme, 'empty-settings');

  it('skips entries with empty settings gracefully', () => {
    // comment has no foreground, should not produce a rule (or produce one without foreground)
    const kwRule = result.monaco.rules.find(r => r.token === 'keyword');
    assert.ok(kwRule);
    assert.equal(kwRule.foreground, 'c586c0');
  });
});

// ---------------------------------------------------------------------------
// Scenario 7: Validation
// ---------------------------------------------------------------------------
describe('validateConvertedTheme', () => {
  it('returns true for a fully converted theme', () => {
    const result = convertVSCodeTheme(DRACULA_THEME, 'dracula');
    assert.equal(validateConvertedTheme(result), true);
  });

  it('returns false when --bg-deepest is missing', () => {
    const result = convertVSCodeTheme(DRACULA_THEME, 'dracula');
    delete result.css['--bg-deepest'];
    assert.equal(validateConvertedTheme(result), false);
  });

  it('returns false when --text-primary is missing', () => {
    const result = convertVSCodeTheme(DRACULA_THEME, 'dracula');
    delete result.css['--text-primary'];
    assert.equal(validateConvertedTheme(result), false);
  });

  it('returns false when css property is missing entirely', () => {
    assert.equal(validateConvertedTheme({ label: 'x', monaco: {}, xterm: {} }), false);
  });

  it('returns false when theme is null/undefined', () => {
    assert.equal(validateConvertedTheme(null), false);
    assert.equal(validateConvertedTheme(undefined), false);
  });
});

// ---------------------------------------------------------------------------
// Scenario 8: Alpha stripping on background vars specifically
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — alpha stripping on background CSS vars', () => {
  const theme = {
    name: 'Alpha BG',
    type: 'dark',
    colors: {
      'editor.background': '#282a36ff',
      'editorWidget.background': '#21222cee',
      'sideBar.background': '#21222cdd',
      'list.activeSelectionBackground': '#44475acc',
      // Non-background colors should keep alpha? Actually CSS vars are all normalized.
      // The spec says strip alpha from bg vars specifically.
      'editor.foreground': '#f8f8f2',
    },
    tokenColors: [],
  };
  const result = convertVSCodeTheme(theme, 'alpha-bg');

  it('strips alpha from --bg-deepest', () => {
    assert.equal(result.css['--bg-deepest'], '#282a36');
  });

  it('strips alpha from --bg-surface', () => {
    assert.equal(result.css['--bg-surface'], '#21222c');
  });

  it('strips alpha from --bg-elevated', () => {
    assert.equal(result.css['--bg-elevated'], '#21222c');
  });

  it('strips alpha from --bg-active', () => {
    assert.equal(result.css['--bg-active'], '#44475a');
  });
});

// ---------------------------------------------------------------------------
// Scenario 9: Monaco colors passthrough
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — Monaco colors', () => {
  it('passes through VSCode colors to Monaco colors object', () => {
    const result = convertVSCodeTheme(DRACULA_THEME, 'dracula');
    assert.equal(result.monaco.colors['editor.background'], '#282a36');
    assert.equal(result.monaco.colors['editor.foreground'], '#f8f8f2');
  });
});

// ---------------------------------------------------------------------------
// Scenario 10: themeId fallback for label
// ---------------------------------------------------------------------------
describe('convertVSCodeTheme — label fallback', () => {
  it('uses themeId when name is missing', () => {
    const theme = { type: 'dark', colors: {}, tokenColors: [] };
    const result = convertVSCodeTheme(theme, 'my-custom-theme');
    assert.equal(result.label, 'my-custom-theme');
  });
});

export function defineTheme(monaco) {
  monaco.editor.defineTheme('conductor-dark', {
    base: 'vs-dark',
    inherit: true,
    rules: [
      { token: 'keyword', foreground: '9d6fff', fontStyle: 'bold' },
      { token: 'string', foreground: '00e57a' },
      { token: 'number', foreground: 'f0a500' },
      { token: 'type', foreground: '00c4b3' },
      { token: 'function', foreground: '3d9eff' },
      { token: 'comment', foreground: '5a6a7a', fontStyle: 'italic' },
      { token: 'variable', foreground: 'c8d4e0' },
      { token: 'operator', foreground: 'c8d4e0' },
      { token: 'delimiter', foreground: '8899aa' },
    ],
    colors: {
      'editor.background': '#07080a',
      'editor.foreground': '#c8d4e0',
      'editorCursor.foreground': '#00e57a',
      'editor.selectionBackground': '#1a3a5c',
      'editor.lineHighlightBackground': '#0d1117',
      'editorLineNumber.foreground': '#3a4a5a',
      'editorLineNumber.activeForeground': '#8899aa',
      'editorGutter.background': '#07080a',
      'diffEditor.insertedTextBackground': '#00e57a18',
      'diffEditor.removedTextBackground': '#ff4a4a18',
      'diffEditor.insertedLineBackground': '#00e57a0d',
      'diffEditor.removedLineBackground': '#ff4a4a0d',
      'editorWidget.background': '#0d1117',
      'editorWidget.border': '#1a2332',
      'input.background': '#0d1117',
      'input.border': '#1a2332',
      'scrollbar.shadow': '#00000000',
      'scrollbarSlider.background': '#1a233280',
      'scrollbarSlider.hoverBackground': '#2a3342',
      'scrollbarSlider.activeBackground': '#3a4352',
    },
  });
}

export const EDITOR_FONT = "'Geist Mono', 'JetBrains Mono', monospace";

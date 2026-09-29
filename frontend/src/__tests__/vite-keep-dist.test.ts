// @vitest-environment node
// (vite.config.js pulls in esbuild, which rejects jsdom's TextEncoder.)
import { describe, it, expect } from 'vitest';
import { mkdtempSync, existsSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import config from '../../vite.config.js';

// R21: Vite's emptyOutDir deletes frontend/dist/.gitkeep on every build;
// the keepDistPlaceholder plugin puts it back so the tree stays clean and
// `go build` keeps working on a fresh clone.

type Hooked = {
  name: string;
  configResolved?: (c: { root: string; build: { outDir: string } }) => void;
  closeBundle?: () => void | Promise<void>;
};

describe('keepDistPlaceholder vite plugin (R21)', () => {
  it('recreates .gitkeep in the build outDir after the bundle closes', async () => {
    const plugins = ((config as { plugins?: unknown[] }).plugins ?? []).flat() as Hooked[];
    const plugin = plugins.find((p) => p && p.name === 'keep-dist-placeholder');
    expect(plugin, 'plugin registered in vite.config.js').toBeTruthy();

    const out = mkdtempSync(join(tmpdir(), 'dist-'));
    try {
      plugin!.configResolved?.({ root: out, build: { outDir: out } });
      await plugin!.closeBundle?.();
      expect(existsSync(join(out, '.gitkeep'))).toBe(true);
    } finally {
      rmSync(out, { recursive: true, force: true });
    }
  });
});

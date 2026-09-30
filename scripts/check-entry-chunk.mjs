#!/usr/bin/env node
// Fail when the built frontend entry chunk is 500 kB or larger (repo health
// R31, as amended 2026-09-30: Monaco and its workers are lazy chunks and are
// exempt; the entry chunk the app boots from must stay small).
// Run after `npm run build`:  node scripts/check-entry-chunk.mjs
import { readFileSync, statSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const LIMIT = 500 * 1024;
const dist = join(dirname(fileURLToPath(import.meta.url)), '..', 'frontend', 'dist');
const html = readFileSync(join(dist, 'index.html'), 'utf8');
const entries = [...html.matchAll(/<script[^>]+type="module"[^>]+src="\/?([^"]+\.js)"/g)].map((m) => m[1]);
if (entries.length === 0) {
  console.error('check-entry-chunk: no module entry script found in frontend/dist/index.html');
  process.exit(1);
}
let failed = false;
for (const rel of entries) {
  const size = statSync(join(dist, rel)).size;
  const kb = (size / 1024).toFixed(0);
  if (size >= LIMIT) {
    console.error(`check-entry-chunk: ${rel} is ${kb} kB (limit 500 kB); lazy-load more views`);
    failed = true;
  } else {
    console.log(`check-entry-chunk: ${rel} ${kb} kB (ok)`);
  }
}
process.exit(failed ? 1 : 0);

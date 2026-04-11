#!/usr/bin/env node
// uiqa-07 — orphan token linter.
// Walks frontend/src/{views,components} and fails if any .svelte file contains
// forbidden raw font-size / padding / margin / gap literals that should be
// design tokens. Pure Node, no deps.

import { readFileSync, readdirSync, statSync, existsSync } from 'node:fs';
import { join, resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const FRONTEND_SRC = resolve(__dirname, '..', 'src');
const TARGET_DIRS = [
  resolve(FRONTEND_SRC, 'views'),
  resolve(FRONTEND_SRC, 'components'),
];

const FORBIDDEN = [
  { label: 'font-size:10px', re: /font-size:\s*10px/ },
  { label: 'font-size:12px', re: /font-size:\s*12px/ },
  { label: 'font-size:15px', re: /font-size:\s*15px/ },
  { label: 'font-size:18px', re: /font-size:\s*18px/ },
  { label: 'spacing:3px', re: /(?:padding|margin|gap)[^:]*:\s*[^;]*\b3px\b/ },
  { label: 'spacing:5px', re: /(?:padding|margin|gap)[^:]*:\s*[^;]*\b5px\b/ },
  { label: 'spacing:6px', re: /(?:padding|margin|gap)[^:]*:\s*[^;]*\b6px\b/ },
];

function walk(dir, out = []) {
  if (!existsSync(dir)) return out;
  for (const entry of readdirSync(dir)) {
    const p = join(dir, entry);
    const st = statSync(p);
    if (st.isDirectory()) {
      if (entry === '__tests__' || entry === 'node_modules') continue;
      walk(p, out);
    } else if (entry.endsWith('.svelte')) {
      out.push(p);
    }
  }
  return out;
}

const files = TARGET_DIRS.flatMap((d) => walk(d));
const offenders = [];

for (const file of files) {
  const src = readFileSync(file, 'utf8');
  const lines = src.split('\n');
  lines.forEach((line, idx) => {
    for (const { label, re } of FORBIDDEN) {
      if (re.test(line)) {
        offenders.push(`${file}:${idx + 1}  [${label}]  ${line.trim()}`);
      }
    }
  });
}

if (offenders.length > 0) {
  console.error(`\nuiqa-07 token linter: ${offenders.length} orphan literal(s) found\n`);
  for (const o of offenders) console.error('  ' + o);
  console.error('\nReplace with design tokens from frontend/src/style.css');
  process.exit(1);
}

console.log('uiqa-07 token linter: clean');
process.exit(0);

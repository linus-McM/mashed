// Svelte compiler config — wires `@sveltejs/vite-plugin-svelte`'s
// `vitePreprocess` helper so `<script lang="ts">` blocks are transpiled
// via esbuild (the same toolchain Vite already uses). Without a
// preprocessor the parser rejects TS-only syntax (`import type`,
// annotations) with "Unexpected token".
//
// `vitePreprocess` has zero runtime cost and no extra dependency — it is
// re-exported from `@sveltejs/vite-plugin-svelte`, which is already in
// the dependency graph. Both `vite.config.js` (build + dev) and
// `vitest.config.js` (tests) auto-load this config via the plugin's
// default behaviour.
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
  preprocess: vitePreprocess(),
};

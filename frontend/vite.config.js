import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'
import {mkdirSync, writeFileSync} from 'node:fs'
import {resolve} from 'node:path'

// keepDistPlaceholder recreates dist/.gitkeep after each build. emptyOutDir
// deletes it, but the Go side embeds `all:frontend/dist` and a fresh clone
// must still `go build` without a Node toolchain (R21).
function keepDistPlaceholder() {
  let outDir = 'dist'
  return {
    name: 'keep-dist-placeholder',
    configResolved(config) {
      outDir = resolve(config.root, config.build.outDir)
    },
    closeBundle() {
      mkdirSync(outDir, {recursive: true})
      writeFileSync(resolve(outDir, '.gitkeep'), '')
    },
  }
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte(), keepDistPlaceholder()],
  build: {
    // Monaco core is ~2.6MB minified — irreducible for a full code editor.
    // Set limit above that so new unexpected bloat still triggers warnings.
    chunkSizeWarningLimit: 2600,
  }
})

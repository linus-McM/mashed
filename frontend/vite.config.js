import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte()],
  build: {
    // Monaco core is ~2.6MB minified — irreducible for a full code editor.
    // Set limit above that so new unexpected bloat still triggers warnings.
    chunkSizeWarningLimit: 2600,
  }
})

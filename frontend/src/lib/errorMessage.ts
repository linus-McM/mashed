/**
 * Normalise an unknown thrown value into a human-readable string.
 *
 * Story svelte-check-02 introduces this single helper so every `catch (e)`
 * block in the frontend can narrow `unknown` without per-site boilerplate.
 *
 * Resolution order:
 *   1. `Error` instance  → `.message`
 *   2. plain `string`    → returned as-is
 *   3. anything else     → `JSON.stringify(e)`, with `String(e)` as a
 *                          last-ditch fallback for circular references or
 *                          other serialisation failures.
 *
 * The helper never throws.
 */
export function errorMessage(e: unknown): string {
  if (e instanceof Error) return e.message;
  if (typeof e === 'string') return e;
  try {
    const s = JSON.stringify(e);
    // `JSON.stringify(undefined)` returns the literal `undefined` (not a
    // string). Fall through to `String(e)` for any non-string result.
    if (typeof s === 'string') return s;
  } catch {
    // fall through
  }
  return String(e);
}

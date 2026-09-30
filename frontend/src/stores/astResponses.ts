// Fresh store per modal mount so responses never leak between suspensions.
import { writable, type Writable } from 'svelte/store';

export function makeAstResponses(): Writable<Record<string, string>> {
  return writable<Record<string, string>>({});
}

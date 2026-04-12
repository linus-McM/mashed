/**
 * skills-watch-02: Unit tests for the in-flight guard pattern used by the
 * bmad:assets:changed event handler.
 *
 * Tests the core concurrency logic without mounting Svelte components —
 * the guard is extracted as a pure async controller.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';

/**
 * Minimal in-flight guard matching the pattern in WorkflowBuilder.svelte.
 * Accepts a fetcher function and returns a trigger + inspection helpers.
 */
function createInflightGuard(fetcher) {
  let fetching = false;
  let dirty = false;
  let fetchCount = 0;

  async function trigger() {
    if (fetching) {
      dirty = true;
      return;
    }
    fetching = true;
    try {
      await fetcher();
      fetchCount++;
    } finally {
      fetching = false;
      if (dirty) {
        dirty = false;
        // Fire-and-forget — matches WorkflowBuilder's refetchMashedAssets
        trigger();
      }
    }
  }

  return {
    trigger,
    get fetching() { return fetching; },
    get fetchCount() { return fetchCount; },
  };
}

describe('In-flight guard (AC-4: concurrent event coalescing)', () => {
  let resolvers;
  let fetcher;
  let guard;

  beforeEach(() => {
    resolvers = [];
    fetcher = vi.fn(() => new Promise((resolve) => { resolvers.push(resolve); }));
    guard = createInflightGuard(fetcher);
  });

  it('runs the fetcher on first trigger', async () => {
    const p = guard.trigger();
    expect(guard.fetching).toBe(true);
    expect(fetcher).toHaveBeenCalledTimes(1);
    resolvers[0]();
    await p;
    expect(guard.fetchCount).toBe(1);
  });

  it('marks dirty and defers when triggered during in-flight fetch', async () => {
    const p1 = guard.trigger();
    expect(guard.fetching).toBe(true);

    // Trigger again while first is in flight — should NOT start a new fetch
    const p2 = guard.trigger();
    expect(fetcher).toHaveBeenCalledTimes(1);

    // Resolve the first fetch — guard should auto-trigger a second
    resolvers[0]();
    await p1;

    // The second (dirty) fetch is now in flight
    expect(fetcher).toHaveBeenCalledTimes(2);
    resolvers[1]();
    // Allow microtask queue to flush
    await new Promise((r) => setTimeout(r, 0));
    expect(guard.fetchCount).toBe(2);
  });

  it('coalesces multiple rapid events into one extra fetch (AC-4)', async () => {
    const p1 = guard.trigger();

    // Fire 5 more events while fetch is in flight
    guard.trigger();
    guard.trigger();
    guard.trigger();
    guard.trigger();
    guard.trigger();

    // Only 1 fetch should be running
    expect(fetcher).toHaveBeenCalledTimes(1);

    // Resolve — exactly one more fetch should happen (dirty coalesced)
    resolvers[0]();
    await p1;
    expect(fetcher).toHaveBeenCalledTimes(2);

    // Resolve the coalesced fetch
    resolvers[1]();
    await new Promise((r) => setTimeout(r, 0));
    expect(guard.fetchCount).toBe(2);
  });

  it('preserves previous value on fetch error (AC-3 pattern)', async () => {
    const errorFetcher = vi.fn()
      .mockRejectedValueOnce(new Error('network'))
      .mockResolvedValueOnce('ok');

    const errorGuard = createInflightGuard(errorFetcher);

    // First call fails — fetchCount stays 0
    await expect(errorGuard.trigger()).rejects.toThrow('network');
    expect(errorGuard.fetchCount).toBe(0);

    // Second call succeeds
    await errorGuard.trigger();
    expect(errorGuard.fetchCount).toBe(1);
  });
});

describe('Event cleanup pattern (AC-5)', () => {
  it('EventsOn returns a cancel function that removes the listener', () => {
    // Simulates the Wails EventsOn/EventsOff contract
    const listeners = new Map();

    function EventsOn(name, cb) {
      listeners.set(name, cb);
      return () => { listeners.delete(name); };
    }

    const handler = vi.fn();
    const cancel = EventsOn('bmad:assets:changed', handler);

    expect(listeners.has('bmad:assets:changed')).toBe(true);

    // Simulate destroy
    cancel();
    expect(listeners.has('bmad:assets:changed')).toBe(false);
  });
});

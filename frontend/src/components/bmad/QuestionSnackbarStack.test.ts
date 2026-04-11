import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  truncate,
  getBorderColor,
  timeAgo,
  upsertQuestion,
  upsertIdle,
  dismissQuestion,
  dismissIdle,
  isQuestionEntry,
  partitionForDisplay,
  DEFAULT_BORDER_COLOR,
  MAX_VISIBLE,
  REPO_COLORS_KEY,
  type QuestionEventLike,
  type IdleEventLike,
  type SnackbarEntry,
} from './questionSnackbarUtils';

// jsdom 29 ships without a full localStorage implementation on some configs;
// provide a minimal in-memory shim so the tests run in isolation regardless
// of the environment.
const memoryStore: Record<string, string> = {};
const fakeStorage: Storage = {
  get length() {
    return Object.keys(memoryStore).length;
  },
  clear() {
    for (const key of Object.keys(memoryStore)) delete memoryStore[key];
  },
  getItem(key: string) {
    return Object.prototype.hasOwnProperty.call(memoryStore, key) ? memoryStore[key] : null;
  },
  setItem(key: string, value: string) {
    memoryStore[key] = String(value);
  },
  removeItem(key: string) {
    delete memoryStore[key];
  },
  key(index: number) {
    return Object.keys(memoryStore)[index] ?? null;
  },
};
// Override whatever the environment provides so behaviour is deterministic.
Object.defineProperty(globalThis, 'localStorage', {
  value: fakeStorage,
  writable: true,
  configurable: true,
});

// ---------- Helpers ----------

function makeEvent(overrides: Partial<QuestionEventLike> = {}): QuestionEventLike {
  return {
    kind: 'question',
    execId: 'exec-1',
    nodeId: 'node-A',
    repoPath: '/Users/dev/my-project',
    repoName: 'my-project',
    question: 'What file should I edit?',
    options: [],
    tmuxTarget: 'mashed:1.0',
    timestamp: 1_700_000_000_000,
    questionId: 'q-1',
    ...overrides,
  };
}

function makeIdle(overrides: Partial<IdleEventLike> = {}): IdleEventLike {
  return {
    kind: 'idle',
    execId: 'exec-1',
    nodeId: 'node-A',
    repoPath: '/Users/dev/my-project',
    repoName: 'my-project',
    tmuxTarget: 'mashed:1.0',
    timestamp: 1_700_000_000_000,
    ...overrides,
  };
}

/** Safely read .question from a SnackbarEntry in tests, narrowing via type guard. */
function asQuestion(entry: SnackbarEntry): QuestionEventLike {
  if (!isQuestionEntry(entry)) {
    throw new Error(`expected question entry, got idle for nodeId=${entry.nodeId}`);
  }
  return entry;
}

describe('QuestionSnackbarStack utilities', () => {
  // ──────────────────────────────────────────────────────────────
  // Constants
  // ──────────────────────────────────────────────────────────────
  describe('constants', () => {
    it('exports DEFAULT_BORDER_COLOR as "#1e2530"', () => {
      expect(DEFAULT_BORDER_COLOR).toBe('#1e2530');
    });

    it('exports MAX_VISIBLE as 5', () => {
      expect(MAX_VISIBLE).toBe(5);
    });

    it('exports REPO_COLORS_KEY as "mashed:repoBorderColors"', () => {
      expect(REPO_COLORS_KEY).toBe('mashed:repoBorderColors');
    });
  });

  // ──────────────────────────────────────────────────────────────
  // truncate()
  // ──────────────────────────────────────────────────────────────
  describe('truncate()', () => {
    it('returns short strings unchanged', () => {
      expect(truncate('hello', 80)).toBe('hello');
    });

    it('returns a string with exactly `max` chars unchanged', () => {
      const text = 'a'.repeat(80);
      expect(truncate(text, 80)).toBe(text);
    });

    it('appends a single ellipsis character when truncating a 120-char string', () => {
      const text = 'a'.repeat(120);
      const result = truncate(text, 80);
      // 80 content codepoints + 1 ellipsis codepoint
      expect([...result].length).toBe(81);
      expect(result.endsWith('…')).toBe(true);
    });

    it('preserves the first 80 characters when truncating', () => {
      const text = 'abcdefghij'.repeat(12); // 120 chars
      const result = truncate(text, 80);
      expect(result.slice(0, 80)).toBe(text.slice(0, 80));
    });

    it('handles empty string', () => {
      expect(truncate('', 80)).toBe('');
    });

    it('coerces non-string input to empty string', () => {
      // @ts-expect-error — testing runtime safety
      expect(truncate(null, 80)).toBe('');
      // @ts-expect-error — testing runtime safety
      expect(truncate(undefined, 80)).toBe('');
    });
  });

  // ──────────────────────────────────────────────────────────────
  // getBorderColor()
  // ──────────────────────────────────────────────────────────────
  describe('getBorderColor()', () => {
    beforeEach(() => {
      localStorage.clear();
    });

    afterEach(() => {
      localStorage.clear();
      vi.restoreAllMocks();
    });

    it('returns the stored colour for a known repo', () => {
      localStorage.setItem(
        REPO_COLORS_KEY,
        JSON.stringify({ 'my-repo': '#ff6b6b' }),
      );
      expect(getBorderColor('my-repo')).toBe('#ff6b6b');
    });

    it('returns DEFAULT_BORDER_COLOR when the repo is not in the map', () => {
      localStorage.setItem(
        REPO_COLORS_KEY,
        JSON.stringify({ 'other-repo': '#abcdef' }),
      );
      expect(getBorderColor('unknown-repo')).toBe(DEFAULT_BORDER_COLOR);
    });

    it('returns DEFAULT_BORDER_COLOR when localStorage has no entry', () => {
      expect(getBorderColor('my-repo')).toBe(DEFAULT_BORDER_COLOR);
    });

    it('returns DEFAULT_BORDER_COLOR when the stored JSON is malformed', () => {
      localStorage.setItem(REPO_COLORS_KEY, '{not-valid-json');
      expect(getBorderColor('my-repo')).toBe(DEFAULT_BORDER_COLOR);
    });

    it('returns DEFAULT_BORDER_COLOR when repoName is empty', () => {
      localStorage.setItem(
        REPO_COLORS_KEY,
        JSON.stringify({ '': '#ff0000' }),
      );
      expect(getBorderColor('')).toBe(DEFAULT_BORDER_COLOR);
    });
  });

  // ──────────────────────────────────────────────────────────────
  // timeAgo()
  // ──────────────────────────────────────────────────────────────
  describe('timeAgo()', () => {
    const now = 1_700_000_000_000;

    it('returns "just now" for under 60 seconds', () => {
      expect(timeAgo(now - 5_000, now)).toBe('just now');
      expect(timeAgo(now - 59_000, now)).toBe('just now');
    });

    it('returns "Nm ago" for minutes', () => {
      expect(timeAgo(now - 2 * 60_000, now)).toBe('2m ago');
      expect(timeAgo(now - 59 * 60_000, now)).toBe('59m ago');
    });

    it('returns "Nh ago" for hours', () => {
      expect(timeAgo(now - 3 * 60 * 60_000, now)).toBe('3h ago');
    });

    it('returns "Nd ago" for days', () => {
      expect(timeAgo(now - 2 * 24 * 60 * 60_000, now)).toBe('2d ago');
    });
  });

  // ──────────────────────────────────────────────────────────────
  // upsertQuestion()
  // ──────────────────────────────────────────────────────────────
  describe('upsertQuestion()', () => {
    it('appends a new event when nodeId is not in the queue', () => {
      const queue: SnackbarEntry[] = [];
      const evt = makeEvent({ nodeId: 'n1', question: 'First?' });
      const next = upsertQuestion(queue, evt);
      expect(next).toHaveLength(1);
      expect(next[0].nodeId).toBe('n1');
    });

    it('replaces the existing event when a new one arrives for the same nodeId', () => {
      const first = makeEvent({ nodeId: 'n1', question: 'First question?' });
      const second = makeEvent({ nodeId: 'n1', question: 'Second question?' });
      const next = upsertQuestion([first], second);
      expect(next).toHaveLength(1);
      expect(asQuestion(next[0]).question).toBe('Second question?');
    });

    it('keeps events for other nodeIds when replacing', () => {
      const a = makeEvent({ nodeId: 'n1', question: 'A' });
      const b = makeEvent({ nodeId: 'n2', question: 'B' });
      const aNew = makeEvent({ nodeId: 'n1', question: 'A2' });
      const next = upsertQuestion([a, b], aNew);
      expect(next).toHaveLength(2);
      const n1 = next.find((e) => e.nodeId === 'n1');
      const n2 = next.find((e) => e.nodeId === 'n2');
      expect(n1 ? asQuestion(n1).question : undefined).toBe('A2');
      expect(n2 ? asQuestion(n2).question : undefined).toBe('B');
    });

    it('does not mutate the input queue', () => {
      const queue = [makeEvent({ nodeId: 'n1' })];
      const copy = [...queue];
      upsertQuestion(queue, makeEvent({ nodeId: 'n2' }));
      expect(queue).toEqual(copy);
    });

    it('replaces an existing idle entry for the same nodeId (question wins)', () => {
      const idle = makeIdle({ nodeId: 'n1' });
      const question = makeEvent({ nodeId: 'n1', question: 'Real question?' });
      const next = upsertQuestion([idle], question);
      expect(next).toHaveLength(1);
      expect(next[0].kind).toBe('question');
      expect(asQuestion(next[0]).question).toBe('Real question?');
    });
  });

  // ──────────────────────────────────────────────────────────────
  // dismissQuestion()
  // ──────────────────────────────────────────────────────────────
  describe('dismissQuestion()', () => {
    it('removes an event matching nodeId', () => {
      const a = makeEvent({ nodeId: 'n1' });
      const b = makeEvent({ nodeId: 'n2' });
      const next = dismissQuestion([a, b], 'n1');
      expect(next).toHaveLength(1);
      expect(next[0].nodeId).toBe('n2');
    });

    it('returns the same-length array when nodeId is not present', () => {
      const a = makeEvent({ nodeId: 'n1' });
      const next = dismissQuestion([a], 'missing');
      expect(next).toHaveLength(1);
    });

    it('does not mutate the input queue', () => {
      const queue = [makeEvent({ nodeId: 'n1' })];
      const copy = [...queue];
      dismissQuestion(queue, 'n1');
      expect(queue).toEqual(copy);
    });

    it('also removes idle entries (dismissal is by nodeId, not kind)', () => {
      const a = makeIdle({ nodeId: 'n1' });
      const b = makeEvent({ nodeId: 'n2' });
      const next = dismissQuestion([a, b], 'n1');
      expect(next).toHaveLength(1);
      expect(next[0].nodeId).toBe('n2');
    });
  });

  // ──────────────────────────────────────────────────────────────
  // upsertIdle()
  // ──────────────────────────────────────────────────────────────
  describe('upsertIdle()', () => {
    it('appends a new idle entry when nodeId is not in the queue', () => {
      const queue: SnackbarEntry[] = [];
      const next = upsertIdle(queue, makeIdle({ nodeId: 'n1' }));
      expect(next).toHaveLength(1);
      expect(next[0].kind).toBe('idle');
      expect(next[0].nodeId).toBe('n1');
    });

    it('normalises the kind field to "idle" even when the caller forgets', () => {
      const queue: SnackbarEntry[] = [];
      // @ts-expect-error — testing runtime normalisation
      const next = upsertIdle(queue, { nodeId: 'n1', repoPath: '/tmp' });
      expect(next[0].kind).toBe('idle');
    });

    it('replaces an existing idle entry for the same nodeId', () => {
      const first = makeIdle({ nodeId: 'n1', tmuxTarget: 't-old' });
      const second = makeIdle({ nodeId: 'n1', tmuxTarget: 't-new' });
      const next = upsertIdle([first], second);
      expect(next).toHaveLength(1);
      expect((next[0] as IdleEventLike).tmuxTarget).toBe('t-new');
    });

    it('does NOT replace an existing question entry (questions win)', () => {
      const question = makeEvent({ nodeId: 'n1', question: 'Real question?' });
      const idle = makeIdle({ nodeId: 'n1' });
      const next = upsertIdle([question], idle);
      expect(next).toHaveLength(1);
      expect(next[0].kind).toBe('question');
      expect(asQuestion(next[0]).question).toBe('Real question?');
    });

    it('keeps other nodeIds untouched', () => {
      const a = makeEvent({ nodeId: 'n1' });
      const b = makeIdle({ nodeId: 'n2' });
      const next = upsertIdle([a, b], makeIdle({ nodeId: 'n3' }));
      expect(next).toHaveLength(3);
    });

    it('does not mutate the input queue', () => {
      const queue: SnackbarEntry[] = [makeIdle({ nodeId: 'n1' })];
      const copy = [...queue];
      upsertIdle(queue, makeIdle({ nodeId: 'n2' }));
      expect(queue).toEqual(copy);
    });
  });

  // ──────────────────────────────────────────────────────────────
  // dismissIdle()
  // ──────────────────────────────────────────────────────────────
  describe('dismissIdle()', () => {
    it('removes an idle entry matching nodeId', () => {
      const idle = makeIdle({ nodeId: 'n1' });
      const next = dismissIdle([idle], 'n1');
      expect(next).toHaveLength(0);
    });

    it('leaves a question entry untouched even when nodeId matches', () => {
      const question = makeEvent({ nodeId: 'n1' });
      const next = dismissIdle([question], 'n1');
      expect(next).toHaveLength(1);
      expect(next[0].kind).toBe('question');
    });

    it('removes only the matching idle entry among a mixed queue', () => {
      const a = makeIdle({ nodeId: 'n1' });
      const b = makeEvent({ nodeId: 'n2' });
      const c = makeIdle({ nodeId: 'n3' });
      const next = dismissIdle([a, b, c], 'n1');
      expect(next).toHaveLength(2);
      expect(next.find((e) => e.nodeId === 'n1')).toBeUndefined();
    });
  });

  // ──────────────────────────────────────────────────────────────
  // partitionForDisplay()
  // ──────────────────────────────────────────────────────────────
  describe('partitionForDisplay()', () => {
    function queueOf(n: number): SnackbarEntry[] {
      return Array.from({ length: n }, (_, i) =>
        makeEvent({ nodeId: `n${i}`, questionId: `q${i}` }),
      );
    }

    it('returns all events when queue length is <= MAX_VISIBLE', () => {
      const result = partitionForDisplay(queueOf(3));
      expect(result.visible).toHaveLength(3);
      expect(result.overflow).toBe(0);
    });

    it('returns MAX_VISIBLE events when queue length equals MAX_VISIBLE', () => {
      const result = partitionForDisplay(queueOf(MAX_VISIBLE));
      expect(result.visible).toHaveLength(MAX_VISIBLE);
      expect(result.overflow).toBe(0);
    });

    it('caps visible at MAX_VISIBLE and reports overflow when queue > MAX_VISIBLE', () => {
      const result = partitionForDisplay(queueOf(7));
      expect(result.visible).toHaveLength(MAX_VISIBLE);
      expect(result.overflow).toBe(2);
    });

    it('handles an empty queue', () => {
      const result = partitionForDisplay([]);
      expect(result.visible).toHaveLength(0);
      expect(result.overflow).toBe(0);
    });
  });

  // ──────────────────────────────────────────────────────────────
  // Integration — simulated snackbar lifecycle
  // ──────────────────────────────────────────────────────────────
  describe('snackbar lifecycle integration', () => {
    it('handles upsert then dismiss correctly', () => {
      let queue: SnackbarEntry[] = [];
      queue = upsertQuestion(queue, makeEvent({ nodeId: 'n1' }));
      queue = upsertQuestion(queue, makeEvent({ nodeId: 'n2' }));
      expect(queue).toHaveLength(2);

      queue = dismissQuestion(queue, 'n1');
      expect(queue).toHaveLength(1);
      expect(queue[0].nodeId).toBe('n2');
    });

    it('enforces overflow across bursts of events', () => {
      let queue: SnackbarEntry[] = [];
      for (let i = 0; i < 7; i++) {
        queue = upsertQuestion(queue, makeEvent({ nodeId: `n${i}` }));
      }
      const { visible, overflow } = partitionForDisplay(queue);
      expect(visible).toHaveLength(MAX_VISIBLE);
      expect(overflow).toBe(2);
    });

    it('models the idle → question takeover flow for one node', () => {
      let queue: SnackbarEntry[] = [];

      // 1. Claude reaches the prompt — idle snackbar appears.
      queue = upsertIdle(queue, makeIdle({ nodeId: 'n1' }));
      expect(queue).toHaveLength(1);
      expect(queue[0].kind).toBe('idle');

      // 2. A structured question is detected on the next poll —
      // the idle entry must be replaced, not duplicated.
      queue = upsertQuestion(
        queue,
        makeEvent({ nodeId: 'n1', question: 'Which file?' }),
      );
      expect(queue).toHaveLength(1);
      expect(queue[0].kind).toBe('question');
      expect(asQuestion(queue[0]).question).toBe('Which file?');

      // 3. A LATE idle event arrives for the same node — must NOT
      // clobber the question. This is the conservative policy that
      // prevents detector-ordering races from degrading the UX.
      queue = upsertIdle(queue, makeIdle({ nodeId: 'n1' }));
      expect(queue).toHaveLength(1);
      expect(queue[0].kind).toBe('question');

      // 4. User answers the question → question is dismissed.
      queue = dismissQuestion(queue, 'n1');
      expect(queue).toHaveLength(0);
    });
  });
});

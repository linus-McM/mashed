// Pure helpers and shared types for the NotificationFeed repo → agents tree.
// Extracted verbatim from views/NotificationFeed.svelte (spec R31); no
// behaviour change.

import type { StatusToken } from '../../types/status';

/**
 * Notification entry as it flows through the feed. Mirrors
 * `domain.NotificationEvent` (internal/domain/types.go) with the extra
 * per-render `subAgents` nesting the feed builds client-side.
 */
export type NotificationEntry = {
  agentId: string;
  agentName?: string;
  model?: string;
  repoName?: string;
  repoPath?: string;
  repoBranch?: string;
  eventType: StatusToken | string;
  summary?: string;
  timestamp?: string;
  tokensUsed?: number;
  tokensMax?: number;
  tokenSamples?: number[];
  priority?: number;
  tmuxTarget?: string;
  pid?: number;
  isSubAgent?: boolean;
  parentAgentId?: string;
  subAgentName?: string;
  subAgentDesc?: string;
  subAgentStatus?: 'running' | 'done' | string;
  subAgentResult?: string;
  subAgents?: NotificationEntry[];
};

/** Repo descriptor emitted by the backend `repos` event. */
export type RepoDescriptor = {
  name: string;
  path: string;
  branch: string;
};

/** Grouped repo → agents tree rendered by the feed. */
export type RepoGroup = {
  name: string;
  path: string;
  branch: string;
  agents: NotificationEntry[];
  worstStatus: string;
};

/** Cached git status per repo path — mirrors `main.RepoStatusInfo`. */
export type RepoStatusSnapshot = {
  dirty?: boolean;
  openPRs?: number;
  ahead?: number;
  behind?: number;
  protected?: boolean;
};

/** Transient per-repo action state (committing, pushing, etc). */
export type RepoAction = {
  action: string | null;
  result: string | null;
  error: string | null;
};

export const IDLE_ACTION: RepoAction = { action: null, result: null, error: null };

export type CommitLine ={ step: string; output: string };

/** Streaming-commit UI state keyed by repoPath. */
export type CommitPanel = {
  lines: CommitLine[];
  error: string | null;
  explanation: string | null;
  done: boolean;
  visible: boolean;
};

// Sort-priority tables. `Partial<Record<...>>` because unknown event types
// fall through to the `?? 7`/`?? 5` default.
const AGENT_PRIORITY: Partial<Record<string, number>> = {
  needs_response: 0,
  error: 1,
  running: 2,
  open: 3,
  started: 4,
  finished: 5,
  completed: 6,
};
const REPO_STATUS_PRIORITY: Partial<Record<string, number>> = {
  needs_response: 0,
  error: 1,
  running: 2,
  started: 3,
  completed: 4,
};

export function applyRepoOrder(groups: RepoGroup[], order: string[]): RepoGroup[] {
  if (!order || order.length === 0) return groups;
  const byName = new Map<string, RepoGroup>(groups.map(r => [r.name, r]));
  const result: RepoGroup[] = [];
  // Add repos in saved order first
  for (const name of order) {
    const hit = byName.get(name);
    if (hit) {
      result.push(hit);
      byName.delete(name);
    }
  }
  // Append any new repos not in saved order
  for (const r of byName.values()) {
    result.push(r);
  }
  return result;
}

export function buildRepoTree(events: NotificationEntry[], scannedRepos: RepoDescriptor[]): RepoGroup[] {
  const repoMap = new Map<string, RepoGroup>();

  // Seed with all scanned repos so they always show a panel
  for (const r of (scannedRepos || [])) {
    if (!repoMap.has(r.name)) {
      repoMap.set(r.name, {
        name: r.name,
        path: r.path,
        branch: r.branch,
        agents: [],
        worstStatus: 'idle',
      });
    }
  }

  for (const evt of events) {
    const repoKey = evt.repoName || 'unknown';
    if (!repoMap.has(repoKey)) {
      repoMap.set(repoKey, {
        name: repoKey,
        path: evt.repoPath || '',
        branch: evt.repoBranch || '',
        agents: [],
        worstStatus: 'running',
      });
    }
    const repo = repoMap.get(repoKey);
    if (!repo) continue;

    // Check if this is a sub-agent (ID contains "-sub-")
    const isSubAgent = !!(evt.agentId && evt.agentId.includes('-sub-'));

    if (isSubAgent) {
      // Find parent agent and nest under it
      const parentId = evt.agentId.split('-sub-')[0];
      const parent = repo.agents.find(a => a.agentId === parentId);
      if (!parent) {
        // Parent not found, show as top-level
        repo.agents.push({ ...evt, subAgents: [] });
      } else {
        if (!parent.subAgents) parent.subAgents = [];
        parent.subAgents.push(evt);
      }
    } else {
      // Top-level agent
      const existing = repo.agents.find(a => a.agentId === evt.agentId);
      if (existing) {
        Object.assign(existing, evt);
      } else {
        repo.agents.push({ ...evt, subAgents: [] });
      }
    }

    // Track worst status for repo header
    if (evt.eventType === 'needs_response' || evt.eventType === 'error') {
      repo.worstStatus = evt.eventType;
    }
  }

  // Sort agents within each repo: running/active on top, then by priority
  for (const repo of repoMap.values()) {
    repo.agents.sort((a, b) => {
      const pa = AGENT_PRIORITY[a.eventType] ?? 7;
      const pb = AGENT_PRIORITY[b.eventType] ?? 7;
      return pa - pb;
    });
  }

  // Sort repos: repos with attention-needed first, then alphabetical
  return Array.from(repoMap.values()).sort((a, b) => {
    const pa = REPO_STATUS_PRIORITY[a.worstStatus] ?? 5;
    const pb = REPO_STATUS_PRIORITY[b.worstStatus] ?? 5;
    if (pa !== pb) return pa - pb;
    return a.name.localeCompare(b.name);
  });
}

export function repoTokens(repo: RepoGroup): number {
  let sum = 0;
  for (const a of repo.agents) {
    sum += a.tokensUsed || 0;
    if (a.subAgents) {
      for (const s of a.subAgents) sum += s.tokensUsed || 0;
    }
  }
  return sum;
}

export function formatTokens(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M';
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'K';
  return String(n);
}

export function formatElapsed(ts: string | number | undefined | null): string {
  if (!ts) return '';
  const diff = Date.now() - new Date(ts).getTime();
  const secs = Math.floor(diff / 1000);
  if (secs < 60) return secs + 's';
  const mins = Math.floor(secs / 60);
  if (mins < 60) return mins + 'm';
  const hrs = Math.floor(mins / 60);
  return hrs + 'h ' + (mins % 60) + 'm';
}

/** Payload of `git:commit:progress` events. */
export type CommitProgressEvent = {
  repoPath: string;
  step?: string;
  output?: string;
  error?: string;
  explanation?: string;
  done?: boolean;
};

export type SessionModalRepo = { path: string; name: string; branch: string };
export type BranchModalRepo = { path: string; branch: string };
export type SwitchModalRepo = { path: string; branch: string; color: string };
export type MergeModalRepo = { path: string; branch: string };
export type ForcePushRepo = { path: string; message: string };

// Event detail types for parent → child component events.
export type SessionSpawnDetail = { command: string; model: string; repoPath: string };
export type BranchEventDetail = { branch?: string };
export type MergeEventDetail = { targetBranch?: string };

/**
 * Canonical status-token → theme-token colour map used when NotificationFeed
 * needs to render a status colour inline (e.g. ambient sparkline tint, any
 * future per-repo chip). Values MUST be theme tokens (`var(--…)`) — cerebrum
 * 2026-04-10 hard-bans raw hex for status colours. The map is `Record` (not
 * `Partial<Record>`) so svelte-check fails if a new `StatusToken` lands in
 * `types/status.ts` without a colour assignment here.
 *
 * StatusBadge keeps its own parallel map with label-specific tuning; the
 * single duplication is worth the reduced coupling between the feed-layout
 * file and the badge-render file.
 */
export const STATUS_COLORS: Record<StatusToken, string> = {
  running:        'var(--accent-green)',
  open:           'var(--accent-teal)',
  finished:       'var(--accent-amber)',
  needs_response: 'var(--accent-red)',
  waiting:        'var(--accent-red)',
  error:          'var(--accent-red)',
  completed:      'var(--accent-blue)',
  started:        'var(--accent-purple)',
  blocked:        'var(--accent-red)',
  done:           'var(--accent-blue)',
  queued:         'var(--text-dim)',
  terminal:       'var(--text-dim)',
};

import { writable, type Writable } from 'svelte/store';
import { ListRepoSessions } from '../../../wailsjs/go/main/App.js';
import type { Session, SessionState } from '../../types/session';

/**
 * Map from `repoPath` to its latest known list of sessions.
 *
 * Empty-map is the "not yet loaded" state — all consumer components pull
 * their repo's slice with `$repoSessions[repoPath] ?? []`.
 */
export const repoSessions: Writable<SessionState> = writable({});

/**
 * Refresh sessions for a single repo. Treats a `null`/`undefined` response
 * from Wails as "empty list" so the store never holds a non-array value.
 */
export async function refreshSessions(repoPath: string): Promise<Session[]> {
  const result = await ListRepoSessions(repoPath);
  const sessions: Session[] = result ?? [];
  repoSessions.update((cur) => ({ ...cur, [repoPath]: sessions }));
  return sessions;
}

/** Add a session to a repo's list, skipping if one with the same name exists. */
export function addSession(repoPath: string, session: Session): void {
  repoSessions.update((cur) => {
    const list = cur[repoPath] ?? [];
    if (list.some((s) => s.sessionName === session.sessionName)) return cur;
    return { ...cur, [repoPath]: [...list, session] };
  });
}

/** Remove a session by name from a specific repo. */
export function removeSession(repoPath: string, sessionName: string): void {
  repoSessions.update((cur) => {
    const list = (cur[repoPath] ?? []).filter((s) => s.sessionName !== sessionName);
    return { ...cur, [repoPath]: list };
  });
}

/** Remove a session by name from every repo it appears in. */
export function removeSessionByName(sessionName: string): void {
  repoSessions.update((cur) => {
    const updated: SessionState = { ...cur };
    for (const rp of Object.keys(updated)) {
      updated[rp] = (updated[rp] ?? []).filter((s) => s.sessionName !== sessionName);
    }
    return updated;
  });
}

/**
 * Construct a fresh `Session` client-side. Used by `NewSessionModal` at the
 * moment of optimistic insert, before the Wails spawn binding returns.
 */
export function makeSession(
  target: string,
  repoPath: string,
  repoName: string,
  sessionType: string,
  model: string,
): Session {
  return {
    sessionName: target,
    paneTarget: target,
    repoPath,
    repoName,
    sessionType,
    model,
    spawnedAt: new Date().toISOString(),
    isAlive: true,
  };
}

import { writable } from 'svelte/store';
import { ListRepoSessions } from '../../../wailsjs/go/main/App.js';

export const repoSessions = writable({});

export async function refreshSessions(repoPath) {
    const sessions = await ListRepoSessions(repoPath);
    repoSessions.update(cur => ({ ...cur, [repoPath]: sessions || [] }));
    return sessions || [];
}

export function addSession(repoPath, session) {
    repoSessions.update(cur => {
        const list = cur[repoPath] || [];
        if (list.some(s => s.sessionName === session.sessionName)) return cur;
        return { ...cur, [repoPath]: [...list, session] };
    });
}

export function removeSession(repoPath, sessionName) {
    repoSessions.update(cur => {
        const list = (cur[repoPath] || []).filter(s => s.sessionName !== sessionName);
        return { ...cur, [repoPath]: list };
    });
}

export function removeSessionByName(sessionName) {
    repoSessions.update(cur => {
        const updated = { ...cur };
        for (const rp of Object.keys(updated)) {
            updated[rp] = updated[rp].filter(s => s.sessionName !== sessionName);
        }
        return updated;
    });
}

export function makeSession(target, repoPath, repoName, sessionType, model) {
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

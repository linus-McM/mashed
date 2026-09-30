// Source-scanning tests read a split view (spec R31) as its parent file plus
// the child components its markup/CSS moved into, concatenated.
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const FRONTEND_SRC = resolve(__dirname, '..');
const src = (p: string) => resolve(FRONTEND_SRC, p);

export const FEED_FILES = [
  'views/NotificationFeed.svelte',
  'components/feed/RepoHeader.svelte',
  'components/feed/AgentList.svelte',
  'components/feed/RepoActions.svelte',
  'components/feed/CommitOutputPanel.svelte',
].map(src);

export const REPO_TREE = src('lib/feed/repoTree.ts');

export const AGENT_DETAIL_FILES = [
  'views/AgentDetail.svelte',
  'components/agent/FileStrip.svelte',
  'components/agent/CommitOutputPanel.svelte',
  'components/agent/SubAgentPanel.svelte',
  'components/agent/SessionTabs.svelte',
].map(src);

export const SETTINGS_FILES = [
  'views/Settings.svelte',
  'components/settings/EditorSettings.svelte',
  'components/settings/VSCodiumThemes.svelte',
  'components/settings/UIAdapterSettings.svelte',
].map(src);

export const WORKFLOW_BUILDER_FILES = [
  'views/WorkflowBuilder.svelte',
  'components/bmad/BuilderToolbar.svelte',
  'components/bmad/TerminalModal.svelte',
].map(src);

/** Reads files and joins them with newlines, optionally mapping each first. */
export function readJoined(files: string[], each: (s: string) => string = (s) => s): string {
  return files.map((f) => each(readFileSync(f, 'utf8'))).join('\n');
}

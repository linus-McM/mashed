/** Status-to-color mapping shared across sprint-aware components. */
export const storyStatusColors = {
  'backlog': 'var(--text-muted)',
  'ready-for-dev': 'var(--accent-blue, #3d9eff)',
  'in-progress': 'var(--accent-amber, #f0a500)',
  'review': 'var(--accent-purple, #9d6fff)',
  'done': 'var(--accent-green, #00e57a)',
};

export const storyStatusLabels = {
  'backlog': 'BACKLOG',
  'ready-for-dev': 'READY',
  'in-progress': 'IN PROG',
  'review': 'REVIEW',
  'done': 'DONE',
};

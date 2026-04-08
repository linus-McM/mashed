/**
 * Shared utility functions for control-flow node components.
 */

/**
 * Build a human-readable condition summary from a node config object.
 * @param {{ field?: string, operator?: string, value?: string, contains?: string }} config
 * @returns {string}
 */
export function formatConditionSummary(config) {
  if (config.field && config.operator) {
    return `${config.field} ${config.operator} ${config.value ?? ''}`;
  }
  if (config.contains) {
    return `contains: ${config.contains}`;
  }
  return '';
}

/**
 * Build the iteration counter display string.
 * Shows "N / max" during execution, or "max: N" when idle.
 * @param {number|null|undefined} iterationCount - current iteration (null/undefined if not running)
 * @param {number} maxIterations - configured maximum
 * @returns {string}
 */
export function formatIterationDisplay(iterationCount, maxIterations) {
  if (iterationCount != null) {
    return `${iterationCount} / ${maxIterations}`;
  }
  if (maxIterations) {
    return `max: ${maxIterations}`;
  }
  return '';
}

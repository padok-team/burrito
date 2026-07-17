import type { Layer } from './types';

export const getApplyTooltip = (layer: Layer) => {
  if (layer.isPR) {
    return 'Manual apply is not allowed on pull request layers';
  }
  if (layer.manualSyncStatus !== 'none') {
    return 'Run in progress...';
  }
  if (!layer.hasValidPlan) {
    return 'No valid plan available. Run a plan first before applying.';
  }
  return 'Apply';
};

export const getSyncTooltip = (layer: Layer) => {
  if (layer.manualSyncStatus !== 'none') {
    return 'Run in progress...';
  }
  if (layer.autoApply) {
    return 'Sync';
  }
  return 'Plan';
};

export type Layers = {
  results: Layer[];
};

export type Layer = {
  namespace: string;
  name: string;
  state: LayerState;
  repository: string;
  branch: string;
  path: string;
  runCount: number;
  lastRunAt: string;
  lastRun: Run;
  latestRuns: Run[];
  lastResult: string;
  isRunning: boolean;
  manualSyncStatus: ManualSyncStatus;
  isPR: boolean;
  hasValidPlan: boolean;
  autoApply: boolean;
};

export type LayerState = 'success' | 'warning' | 'error' | 'disabled';
export type ManualSyncStatus = 'none' | 'annotated' | 'pending';

export type Run = {
  id: string;
  commit: string;
  date: string;
  action: string;
};

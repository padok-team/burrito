export type Layers = {
  results: Layer[];
};

export type Layer = {
  id: string;
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
  isPR: boolean;
};

export type LayerState = "success" | "warning" | "error" | "disabled";

export type Run = {
  id: string;
  commit: string;
  date: string;
  action: string;
};

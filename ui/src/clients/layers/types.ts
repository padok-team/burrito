export type Layers = {
  results: Layer[];
  next?: string;
};

export type InfiniteLayerQueryResult = {
  results: Layer[];
  pageParams: string;
}

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
  isPR: boolean;
};

export type LayerState = "success" | "warning" | "error" | "disabled";

export type Run = {
  id: string;
  commit: string;
  date: string;
  action: string;
};

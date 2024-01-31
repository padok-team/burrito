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
  latestRuns: Run[];
  lastResult: string;
  isRunning: boolean;
  isPR: boolean;
};

export type LayerState = "success" | "warning" | "error" | "disabled";

export type Run = {
  commit: string;
  date: string;
  action: string;
};

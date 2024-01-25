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
  lastResult: string;
  isRunning: boolean;
  isPR: boolean;
};

export type LayerState = "success" | "warning" | "error" | "disabled";

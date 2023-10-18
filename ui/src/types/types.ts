export type Layer = {
  namespace: string;
  name: string;
  state: LayerState;
  repository: string;
  branch: string;
  path: string;
  lastResult: string;
};

export type LayerState = "success" | "warning" | "error" | "disabled";

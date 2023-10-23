import { Repository } from "@/clients/repositories/types";

export type Layers = {
  results: Layer[];
};

export type Layer = {
  namespace: string;
  name: string;
  state: LayerState;
  repository: Repository;
  branch: string;
  path: string;
  lastResult: string;
  isRunning: boolean;
  isPR: boolean;
};

export type LayerState = "success" | "warning" | "error" | "disabled";

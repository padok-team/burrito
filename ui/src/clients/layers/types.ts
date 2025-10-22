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
  autoApply: boolean;
  openTofu: boolean;
  terraform: boolean;
  terragrunt: boolean;
  isPR: boolean;
};

export type LayerState = 'success' | 'warning' | 'error' | 'disabled';
export type ManualSyncStatus = 'none' | 'annotated' | 'pending';

export type Run = {
  id: string;
  commit: string;
  author: string;
  message: string;
  date: string;
  action: string;
};

export type StateGraphNode = {
  id: string;
  addr: string;
  mode: string;
  type: string;
  name: string;
  module?: string;
  provider: string;
  instances_count: number;
  instances?: Array<StateGraphResourceInstance>;
};

export type StateGraphResourceInstance = {
  addr: string;
  dependencies?: string[];
  attributes?: Record<string, unknown>;
  created_at?: string;
};

export type StateGraphEdge = {
  from: string;
  to: string;
};

export type StateGraph = {
  nodes: StateGraphNode[];
  edges: StateGraphEdge[];
};

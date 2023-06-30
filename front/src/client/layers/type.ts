export enum LayerStatus {
  PlanNeeded = 'PlanNeeded',
  Idle = 'Idle',
  ApplyNeeded = 'ApplyNeeded',
  FailureGracePeriod = 'FailureGracePeriod',
}

export type LayerSummary = {
  id: string;
  name: string;
  repoURL: string;
  path: string;
  branch: string;
  status: LayerStatus;
};

export type Layer = {
  id: string;
  name: string;
  repoURL: string;
  path: string;
  branch: string;
  status: LayerStatus;
  lastPlanCommit: string;
  lastApplyCommit: string;
  lastRelevantCommit: string;
  Resources: Resource[];
};

export type Resource = {
  address: string;
  type: string;
  status: string;
  depends_on: string[];
};

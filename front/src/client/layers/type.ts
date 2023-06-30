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
  status: string;
};

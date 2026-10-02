import { Layer } from '@/clients/layers/types';

export type PullRequests = {
  results: PullRequest[];
};

export type PullRequest = {
  uid: string;
  name: string;
  namespace: string;
  id: string;
  repository: string;
  url: string;
  branch: string;
  base: string;
  state: PullRequestState;
  lastDiscoveredCommit: string;
  lastCommentedCommit: string;
  layers: Layer[];
};

// Mirrors TerraformPullRequest.status.state ('' until the controller sets it)
export type PullRequestState =
  'DiscoveryNeeded' | 'Planning' | 'CommentNeeded' | 'Idle' | '';

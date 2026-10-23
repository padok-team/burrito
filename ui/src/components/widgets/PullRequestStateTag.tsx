import React from 'react';

import { PullRequestState } from '@/clients/pulls/types';

export interface PullRequestStateTagProps {
  state: PullRequestState;
}

const styles: Record<PullRequestState, string> = {
  Idle: 'bg-status-success-default text-nuances-black',
  Planning: 'bg-blue-400 text-nuances-white',
  DiscoveryNeeded: 'bg-status-warning-default text-nuances-black',
  CommentNeeded: 'bg-status-warning-default text-nuances-black',
  '': 'bg-nuances-50 text-nuances-200'
};

const PullRequestStateTag: React.FC<PullRequestStateTagProps> = ({ state }) => (
  <div
    className={`
      flex
      px-3 py-1
      items-center
      rounded-full
      text-sm
      font-semibold
      leading-5
      ${styles[state] ?? styles['']}
    `}
  >
    {state || 'Unknown'}
  </div>
);

export default PullRequestStateTag;

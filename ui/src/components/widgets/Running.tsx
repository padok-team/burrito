import React from 'react';

import SyncIcon from '@/assets/icons/SyncIcon';

interface RunningProps {
  pending?: boolean;
}

const Running: React.FC<RunningProps> = ({ pending }) => {
  return (
    <div className={`flex items-center gap-2 text-blue-500 fill-blue-500`}>
      <span className="text-sm font-semibold">{pending ? 'Sync scheduled' : 'Running'}</span>
      <SyncIcon className="animate-spin-slow" height={16} width={16} />
    </div>
  );
};

export default Running;

import React from 'react';
import { useQuery } from '@tanstack/react-query';

import { fetchLayersStatus } from '@/clients/layers/client';
import { reactQueryKeys } from '@/clients/reactQueryConfig';

interface StatusBarProps {
  className?: string;
  compact?: boolean;
}

const StatusBar: React.FC<StatusBarProps> = ({ className = '', compact = false }) => {
  const statusQuery = useQuery({
    queryKey: reactQueryKeys.layersStatus,
    queryFn: fetchLayersStatus,
    refetchInterval: 30000,
  });

  if (statusQuery.isLoading) {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <div className="w-16 h-6 bg-nuances-200 rounded animate-pulse" />
        <div className="w-20 h-6 bg-nuances-200 rounded animate-pulse" />
      </div>
    );
  }

  if (statusQuery.isError || !statusQuery.data) {
    return (
      <div className={`text-status-error-default text-sm ${className}`}>
        Status unavailable
      </div>
    );
  }

  const { total, ok, error, applyNeeded, planNeeded, disabled, running } = statusQuery.data;
  const totalIssues = error + applyNeeded + planNeeded;

  if (compact) {
    return (
      <div className={`flex items-center gap-2 text-sm ${className}`}>
        <div className="bg-nuances-50 text-nuances-200 px-3 py-1 rounded-full font-semibold">
          Total: {total}
        </div>
        {ok > 0 && (
          <div className="bg-status-success-default text-nuances-black px-3 py-1 rounded-full font-semibold">
            {ok} OK
          </div>
        )}
        {running > 0 && (
          <div className="bg-blue-500 text-nuances-white px-3 py-1 rounded-full font-semibold">
            {running} running
          </div>
        )}
        {totalIssues > 0 && (
          <div className="bg-status-warning-default text-nuances-black px-3 py-1 rounded-full font-semibold">
            {totalIssues} need attention
          </div>
        )}
        {error > 0 && (
          <div className="bg-status-error-default text-nuances-white px-3 py-1 rounded-full font-semibold">
            {error} errors
          </div>
        )}
      </div>
    );
  }

  return (
    <div className={`flex items-center gap-3 ${className}`}>
      <div className="bg-nuances-50 text-nuances-200 px-3 py-1 rounded-full text-sm font-semibold">
        Total: {total}
      </div>
      
      {ok > 0 && (
        <div className="bg-status-success-default text-nuances-black px-3 py-1 rounded-full text-sm font-semibold">
          OK: {ok}
        </div>
      )}

      {running > 0 && (
        <div className="bg-blue-500 text-nuances-white px-3 py-1 rounded-full text-sm font-semibold">
          Running: {running}
        </div>
      )}

      {applyNeeded > 0 && (
        <div className="bg-status-warning-default text-nuances-black px-3 py-1 rounded-full text-sm font-semibold">
          Apply Needed: {applyNeeded}
        </div>
      )}

      {planNeeded > 0 && (
        <div className="bg-status-warning-default text-nuances-black px-3 py-1 rounded-full text-sm font-semibold">
          Plan Needed: {planNeeded}
        </div>
      )}

      {error > 0 && (
        <div className="bg-status-error-default text-nuances-white px-3 py-1 rounded-full text-sm font-semibold">
          Error: {error}
        </div>
      )}

      {disabled > 0 && (
        <div className="bg-nuances-50 text-nuances-200 px-3 py-1 rounded-full text-sm font-semibold">
          Disabled: {disabled}
        </div>
      )}
    </div>
  );
};

export default StatusBar;

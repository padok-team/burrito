import React, { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';

import { fetchLayersStatus } from '@/clients/layers/client';
import { reactQueryKeys } from '@/clients/reactQueryConfig';

interface StatusBarProps {
  className?: string;
  variant?: 'light' | 'dark';
}

interface StatusItemProps {
  label: string;
  count: number;
  variant: 'success' | 'warning' | 'error' | 'disabled' | 'apply-needed' | 'plan-needed' | 'running';
  theme?: 'light' | 'dark';
}

const StatusItem: React.FC<StatusItemProps> = ({ label, count, variant, theme = 'light' }) => {
  const styles = useMemo(() => {
    const variantStyles = {
      success: `bg-status-success-default text-nuances-black`,
      warning: `bg-status-warning-default text-nuances-black`,
      error: `bg-status-error-default text-nuances-white`,
      disabled: theme === 'light' ? `bg-nuances-200 text-nuances-400` : `bg-nuances-400 text-nuances-50`,
      'apply-needed': `bg-status-warning-default text-nuances-black`,
      'plan-needed': `bg-status-warning-default text-nuances-black`,
      'running': `bg-blue-500 text-nuances-white`
    };

    return `flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium leading-4 ${variantStyles[variant]}`;
  }, [variant, theme]);

  // Handle singular/plural for Errors
  let displayLabel = label;
  if (label === 'Errors' && count === 1) {
    displayLabel = 'Error';
  }
    
  return (
    <div className={styles}>
      <span>{count.toLocaleString()}</span>
      <span>{displayLabel.toLowerCase()}</span>
    </div>
  );
};

const OptimizedStatusBar: React.FC<StatusBarProps> = ({ className = '', variant = 'light' }) => {
  const statusQuery = useQuery({
    queryKey: reactQueryKeys.layersStatus,
    queryFn: fetchLayersStatus,
    refetchInterval: 30000, // Refetch every 30 seconds for real-time updates
    staleTime: 10000, // Consider data stale after 10 seconds
    gcTime: 60000, // Keep in cache for 1 minute
  });

  const statusItems = useMemo(() => {
    if (!statusQuery.data) return [];

    const { total, ok, error, applyNeeded, planNeeded, running } = statusQuery.data;

    return [
      { label: 'Total', count: total, variant: 'disabled' as const },
      { label: 'OK', count: ok, variant: 'success' as const },
      { label: 'Running', count: running, variant: 'running' as const },
      { label: 'Out of Sync', count: applyNeeded + planNeeded, variant: 'warning' as const },
      { label: 'Errors', count: error, variant: 'error' as const },
    ].filter(item => item.count > 0 || item.label === 'Total');
  }, [statusQuery.data]);

  if (statusQuery.isLoading) {
    return (
      <div className={`flex items-center gap-2 px-3 py-2 ${className}`}>
        <div className="flex items-center gap-2">
          <div className="w-1 h-1 rounded-full bg-nuances-200 animate-pulse" />
          <div className="w-12 h-3 bg-nuances-200 rounded animate-pulse" />
          <div className="w-6 h-3 bg-nuances-200 rounded animate-pulse" />
        </div>
      </div>
    );
  }

  if (statusQuery.isError) {
    return (
      <div className={`flex items-center gap-2 px-3 py-2 text-status-error-default ${className}`}>
        <div className="w-1 h-1 rounded-full bg-status-error-default" />
        <span className="text-xs font-medium">Status error</span>
      </div>
    );
  }

  return (
    <div className={`flex items-center gap-2 px-3 py-2 ${className}`}>
      {statusItems.map((item) => (
        <StatusItem
          key={item.label}
          label={item.label}
          count={item.count}
          variant={item.variant}
          theme={variant}
        />
      ))}
      {statusQuery.isFetching && (
        <div className="flex items-center gap-1 text-nuances-400">
          <div className="w-0.5 h-0.5 rounded-full bg-nuances-400 animate-pulse" />
          <span className="text-xs">•••</span>
        </div>
      )}
    </div>
  );
};

export default OptimizedStatusBar;

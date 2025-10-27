import React, { useMemo } from 'react';
import { Layer } from '@/clients/layers/types';
import Tag from '../widgets/Tag';
import { formatHumanDate } from '@/utils/locale';
import { getLayerType } from '@/utils/layer';
import { twMerge } from 'tailwind-merge';
import Running from '../widgets/Running';

interface LayerStatusProps {
  className?: string;
  variant: 'health' | 'lastOperation' | 'details';
  theme?: 'light' | 'dark';
  layer?: Layer;
  syncPending?: boolean;
}

const operationResult: Array<{ value: string; label: string }> = [
  { value: 'apply', label: 'Applied' },
  { value: 'plan', label: 'Planned' },
];

const titleOptions: Record<LayerStatusProps['variant'], string> = {
  health: 'Status',
  lastOperation: 'Last Operation',
  details: 'Details',
};

const LayerStatus: React.FC<LayerStatusProps> = ({
  className,
  variant = 'health',
  theme = 'light',
  layer,
  syncPending
}) => {
  const title = titleOptions[variant] || 'Status';
  const lastRunAtText = useMemo(() => formatHumanDate(layer?.lastRunAt), [layer?.lastRunAt]);
  const lastRunAtTitle = useMemo(() => {
    if (!layer?.lastRunAt) return undefined;
    const d = new Date(String(layer.lastRunAt));
    return isNaN(d.getTime()) ? undefined : d.toLocaleString();
  }, [layer?.lastRunAt]);
  const styles = {
    base: {
      light: `bg-nuances-white
        shadow-light`,

      dark: `bg-nuances-400
        shadow-dark`
    },

    isRunning: {
      light: `outline-blue-400`,
      dark: `outline-blue-500`
    }
  };
  const mutedText = theme === 'light' ? 'text-gray-500' : 'text-nuances-200';

  return (
    <div className={className}>
      <div className={twMerge(`h-full py-3 px-6 rounded-2xl ${styles.base[theme]}`, layer?.isRunning && variant === 'health' ? `outline-solid outline-4 ${styles.isRunning[theme]}` : '')}>
        <h2 className="text-sm uppercase font-semibold text-primary-600">{title}</h2>
        { variant === 'health' &&
        <div className="flex flex-col items-center justify-center h-full text-lg min-w-16">
          {layer &&
            <Tag variant={layer.state} className="text-lg mb-4" />
          }
          {(layer?.isRunning || syncPending) && (
            <Running pending={!!syncPending && !layer?.isRunning} />
          )}
        </div>
        }
        {layer && variant==='lastOperation' &&
        <div className="flex">
          <div className="flex flex-col min-w-40">
            <span className={twMerge("text-lg py-1 font-bold",mutedText)}> {layer.lastRun?.action && operationResult.find(op => op.value === layer.lastRun.action)?.label}</span>
            <span className={twMerge('text-sm', mutedText)} title={lastRunAtTitle}> {lastRunAtText}</span>
            <span className={twMerge('text-sm', mutedText)}>Auto Apply is <span className="font-semibold">{layer.autoApply ? 'enabled' : 'disabled'}</span></span>
          </div>
          <div className="flex flex-col min-w-32">
            <div className="grid grid-cols-[20%_80%] mt-4 gap-x-8">
              <span className={twMerge('text-sm text-right', mutedText)}>Author:</span>
              <span className={twMerge('text-sm truncate pr-8', mutedText)} title={layer.lastRun?.author}>{layer.lastRun?.author}</span>
              <span className={twMerge('text-sm text-right', mutedText)}>Message:</span>
              <span className={twMerge('text-sm truncate pr-8', mutedText)} title={layer.lastRun?.message}>{layer.lastRun?.message}</span>
              <span className={twMerge('text-sm text-right', mutedText)}>Commit:</span>
              <span className={twMerge('text-sm truncate pr-8', mutedText)} title={layer.lastRun?.commit}>{layer.lastRun?.commit}</span>
            </div>
          </div>
        </div>
        }
        {layer && variant==='details' &&
          <div className="flex flex-col min-w-58">
            <div className="grid grid-cols-[30%_70%] mt-4 gap-x-4">
              <span className={twMerge('text-sm text-right', mutedText)}>Type:</span>
              <span className={twMerge('text-sm truncate pr-8', mutedText)}>{getLayerType(layer)}</span>
              {layer.terragrunt &&
                <span className={twMerge('text-sm text-right', mutedText)}>Terragrunt Enabled</span>
              }
              <span className={twMerge('text-sm text-right', mutedText)}>Git ref:</span>
              <span className={twMerge('text-sm truncate pr-8', mutedText)}>{layer.branch}</span>
              <span className={twMerge('text-sm text-right', mutedText)}>Code path:</span>
              <span className={twMerge('text-sm truncate pr-8', mutedText)} title={layer.path}>{layer.path || '/'}</span>
              </div>
          </div>
        }
      </div>
    </div>
  );
};

export default LayerStatus;

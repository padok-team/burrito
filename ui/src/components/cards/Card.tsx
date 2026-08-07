import React, { useEffect, useState } from 'react';
import { twMerge } from 'tailwind-merge';
import { Tooltip } from 'react-tooltip';

import Running from '@/components/widgets/Running';
import Tag from '@/components/widgets/Tag';
import ModalLogsTerminal from '@/components/tools/ModalLogsTerminal';
import CodeBranchIcon from '@/assets/icons/CodeBranchIcon';
import ChiliLight from '@/assets/illustrations/ChiliLight';
import ChiliDark from '@/assets/illustrations/ChiliDark';

import type { Layer } from '@/clients/layers/types';
import GenericIconButton from '../buttons/GenericIconButton';
import { applyLayer, syncLayer } from '@/clients/layers/client';
import { getApplyTooltip, getSyncTooltip } from '@/clients/layers/utils';
import SyncIcon from '@/assets/icons/SyncIcon';
import PlayIcon from '@/assets/icons/PlayIcon';

export interface CardProps {
  className?: string;
  variant?: 'light' | 'dark';
  layer: Layer;
}

const Card: React.FC<CardProps> = ({
  className,
  variant = 'light',
  layer,
  layer: {
    name,
    namespace,
    state,
    repository,
    branch,
    path,
    lastResult,
    isRunning,
    isPR
  }
}) => {
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

  const getTag = () => {
    return (
      <div className="flex items-center">
        <Tag variant={state} />
        {state === 'error' &&
          (variant === 'light' ? (
            <ChiliLight
              className="absolute translate-x-16 rotate-[-21deg]"
              height={40}
              width={40}
            />
          ) : (
            <ChiliDark
              className="absolute translate-x-16 rotate-[-21deg]"
              height={40}
              width={40}
            />
          ))}
      </div>
    );
  };

  const syncSelectedLayer = async (layer: Layer) => {
    const sync = await syncLayer(layer.namespace, layer.name);
    if (sync.status === 200) {
      setIsManualActionPending(true);
    }
  };

  const applySelectedLayer = async (layer: Layer) => {
    const apply = await applyLayer(layer.namespace, layer.name);
    if (apply.status === 200) {
      setIsManualActionPending(true);
    }
  };

  const [isManualActionPending, setIsManualActionPending] = useState(
    layer.manualSyncStatus !== 'none'
  );

  useEffect(() => {
    setIsManualActionPending(layer.manualSyncStatus !== 'none');
  }, [layer.manualSyncStatus, layer.lastRunAt]);

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        items-start
        rounded-2xl
        p-6
        gap-4
        ${styles.base[variant]}`,
        isRunning && `outline-solid outline-4 ${styles.isRunning[variant]}`,
        className
      )}
    >
      <div
        className={`
          flex
          items-center
          justify-between
          self-stretch
          gap-4
        `}
      >
        <span
          className={`
            text-lg
            font-black
            leading-6
            truncate
            ${variant === 'light' ? 'text-nuances-black' : 'text-nuances-50'}
          `}
        >
          {name}
        </span>
        {isRunning ? (
          <Running />
        ) : isPR ? (
          <CodeBranchIcon
            className={`
              ${variant === 'light' ? 'fill-nuances-black' : 'fill-nuances-50'}
            `}
          />
        ) : null}
      </div>
      <div className="grid grid-cols-[min-content_1fr] items-start gap-x-7 gap-y-2">
        {[
          ['Namespace', namespace],
          ['State', getTag()],
          ['Repository', repository],
          ['Branch', branch],
          ['Path', path],
          ['Last result', lastResult]
        ].map(([label, value], index) => (
          <React.Fragment key={index}>
            <span
              className={`
                text-base
                font-normal
                truncate
                ${variant === 'light' ? 'text-primary-600' : 'text-nuances-300'}
              `}
            >
              {label}
            </span>
            <div
              className={`
                text-base
                font-semibold
                truncate
                ${
                  variant === 'light' ? 'text-nuances-black' : 'text-nuances-50'
                }
              `}
            >
              <span
                data-tooltip-id="card-tooltip"
                data-tooltip-content={
                  label === 'Path' || label === 'Last result'
                    ? (value as string)
                    : null
                }
              >
                {value}
              </span>
            </div>
          </React.Fragment>
        ))}
      </div>
      <div className="flex gap-4">
        {layer.latestRuns.length > 0 && (
          <ModalLogsTerminal layer={layer} variant={variant} />
        )}
        <GenericIconButton
          variant={variant}
          Icon={SyncIcon}
          disabled={isManualActionPending}
          onClick={() => syncSelectedLayer(layer)}
          tooltip={getSyncTooltip(layer)}
        />
        {!layer.autoApply && (
          <GenericIconButton
            variant={variant}
            Icon={PlayIcon}
            disabled={layer.isPR || isManualActionPending}
            onClick={() => applySelectedLayer(layer)}
            tooltip={getApplyTooltip(layer)}
          />
        )}
      </div>
      <Tooltip
        opacity={1}
        id="card-tooltip"
        variant={variant === 'light' ? 'dark' : 'light'}
      />
    </div>
  );
};

export default Card;

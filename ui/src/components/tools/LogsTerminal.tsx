import React, { useState, useEffect } from 'react';
import { twMerge } from 'tailwind-merge';
import { useQuery } from '@tanstack/react-query';
import { Tooltip } from 'react-tooltip';

import { fetchAttempts } from '@/clients/runs/client';
import { fetchLogs } from '@/clients/logs/client';
import { reactQueryKeys } from '@/clients/reactQueryConfig';

import AttemptsDropdown from '@/components/dropdowns/AttemptsDropdown';
import AttemptButton from '@/components/buttons/AttemptButton';

import SyncIcon from '@/assets/icons/SyncIcon';
import CopyIcon from '@/assets/icons/CopyIcon';
import DownloadAltIcon from '@/assets/icons/DownloadAltIcon';

import { Layer } from '@/clients/layers/types';

export interface LogsTerminalProps {
  className?: string;
  variant?: 'light' | 'dark';
  layer: Layer;
  run: string;
}

const LogsTerminal: React.FC<LogsTerminalProps> = ({
  className,
  variant = 'light',
  layer: { namespace, name },
  run
}) => {
  const styles = {
    light: `bg-nuances-50
      text-nuances-black
      fill-nuances-black
      border-primary-500`,
    dark: `bg-nuances-400
      text-nuances-50
      fill-nuances-50
      border-nuances-black`
  };

  const [selectedAttempts, setSelectedAttempts] = useState<number[]>([]);
  const [activeAttempt, setActiveAttempt] = useState<number | null>(null);
  /* TODO: The states above should eventually be moved to search params.
  An unresolved bug in the useSearchParams hook is causing rerenders which
  triggers the useEffect and resets the selectedAttempts and activeAttempt states. */
  const [displayLogsCopiedTooltip, setDisplayLogsCopiedTooltip] =
    useState<boolean>(false);

  const attemptsQuery = useQuery({
    queryKey: reactQueryKeys.attempts(namespace, name, run),
    queryFn: () => fetchAttempts(namespace, name, run)
  });

  const logsQuery = useQuery({
    queryKey: reactQueryKeys.logs(namespace, name, run, activeAttempt),
    queryFn: () => fetchLogs(namespace, name, run, activeAttempt),
    enabled: activeAttempt !== null && !attemptsQuery.isFetching
  });

  useEffect(() => {
    setActiveAttempt(null);
    setSelectedAttempts([]);

    if (attemptsQuery.data && attemptsQuery.data.count > 0) {
      setActiveAttempt(attemptsQuery.data.count - 1);
      setSelectedAttempts([attemptsQuery.data.count - 1]);
    }
  }, [attemptsQuery.data]);

  const handleCopy = () => {
    if (logsQuery.isSuccess) {
      navigator.clipboard.writeText(logsQuery.data.results.join('')); // TODO: check if this works properly
      setDisplayLogsCopiedTooltip(true);
    }
  };

  const handleClose = (attempt: number) => {
    setSelectedAttempts(selectedAttempts.filter((a) => a !== attempt));
    setActiveAttempt(activeAttempt === attempt ? null : activeAttempt);
  };

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        rounded-2xl
        overflow-auto
        ${styles[variant]}`,
        className
      )}
    >
      <div className="flex flex-row justify-between items-center gap-16 p-4 overflow-auto shrink-0">
        <div className="flex flex-row items-center gap-4">
          <span className="text-lg font-black">{name}</span>
          <span className="text-base font-semibold">{namespace}</span>
          <AttemptsDropdown
            className={
              variant === 'light'
                ? 'bg-primary-300 text-primary-600 fill-primary-600'
                : 'bg-nuances-300 text-nuances-400 fill-nuances-400'
            }
            variant={variant}
            runId={run}
            namespace={namespace}
            layer={name}
            selectedAttempts={selectedAttempts}
            setSelectedAttempts={setSelectedAttempts}
          />
        </div>
        <div className="flex flex-row items-center gap-4">
          <SyncIcon
            className={`
              cursor-pointer
              ${logsQuery.isRefetching && 'animate-spin-slow'}
            `}
            height={30}
            width={30}
            onClick={() => logsQuery.refetch()}
          />
          <div
            data-tooltip-id={'terminal-tooltip'}
            data-tooltip-content={'Copied to clipboard'}
            onMouseLeave={() => setDisplayLogsCopiedTooltip(false)}
          >
            <CopyIcon
              className="cursor-pointer"
              height={30}
              width={30}
              onClick={handleCopy}
            />
          </div>
          <DownloadAltIcon
            className="cursor-not-allowed" // TODO: add download functionality
            height={30}
            width={30}
          />
        </div>
      </div>
      <hr
        className={`
          h-px
          w-full
          ${variant === 'light' ? 'border-primary-600' : 'border-nuances-300'}
        `}
      />
      <div className="flex flex-row items-center gap-1 p-4 shrink-0 overflow-auto">
        {selectedAttempts
          .sort((a, b) => b - a)
          .map((attempt) => (
            <AttemptButton
              key={attempt}
              variant={variant}
              attempt={attempt + 1}
              isActive={attempt === activeAttempt}
              onClick={() => setActiveAttempt(attempt)}
              onClose={() => handleClose(attempt)}
            />
          ))}
      </div>
      <div className="pb-4 overflow-auto">
        <table>
          <tbody>
            {activeAttempt !== null &&
              logsQuery.isSuccess &&
              logsQuery.data.results.map((log, i) => (
                <tr key={i}>
                  <td
                    className={`
                      text-sm
                      px-4
                      ${
                        variant === 'light'
                          ? 'text-primary-600'
                          : 'text-nuances-300'
                      }
                    `}
                  >
                    {i + 1}
                  </td>
                  <td>{log}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>
      <Tooltip
        id="terminal-tooltip"
        isOpen={displayLogsCopiedTooltip}
        opacity={1}
        variant={variant === 'light' ? 'dark' : 'light'}
      />
    </div>
  );
};

export default LogsTerminal;

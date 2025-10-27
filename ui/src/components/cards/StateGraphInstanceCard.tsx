import React, { useEffect, useRef, useState } from 'react';
import { twMerge } from 'tailwind-merge';

import AngleDownIcon from '@/assets/icons/AngleDownIcon';
import CopyIcon from '@/assets/icons/CopyIcon';

import { StateGraphResourceInstance } from '@/clients/layers/types';
import type { PlanAction } from '@/utils/terraformPlan';

export interface StateGraphInstanceCardProps {
  className?: string;
  variant?: 'light' | 'dark';
  instance: StateGraphResourceInstance;
  defaultExpanded?: boolean;
  tone?: 'current' | 'future';
  planAction?: PlanAction | null;
  badge?: React.ReactNode;
  onDependencyClick?: (addr: string) => void;
  isDependencyAvailable?: (addr: string) => boolean;
}

const StateGraphInstanceCard: React.FC<StateGraphInstanceCardProps> = ({
  className,
  variant = 'light',
  instance: { addr, dependencies, attributes, created_at },
  defaultExpanded = false,
  tone = 'current',
  planAction = null,
  badge,
  onDependencyClick,
  isDependencyAvailable
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const [copyStatus, setCopyStatus] = useState<'idle' | 'copied'>('idle');
  const copyResetRef = useRef<number | null>(null);
  const handleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  const styles = {
    current: {
      light: `bg-primary-100
        text-nuances-black
        outline-primary-500
        hover:bg-nuances-white`,

      dark: `bg-nuances-400
        text-nuances-50
        outline-nuances-50
        hover:bg-nuances-400`
    },
    future: {
      create: {
        light: `bg-emerald-50
          text-emerald-900
          outline-emerald-400
          border border-emerald-200
          hover:bg-emerald-100`,
        dark: `bg-emerald-900/30
          text-emerald-100
          outline-emerald-500
          border border-emerald-800
          hover:bg-emerald-900/40`
      },
      update: {
        light: `bg-amber-50
          text-amber-900
          outline-amber-400
          border border-amber-200
          hover:bg-amber-100`,
        dark: `bg-amber-900/30
          text-amber-100
          outline-amber-500
          border border-amber-800
          hover:bg-amber-900/40`
      },
      replace: {
        light: `bg-violet-50
          text-violet-900
          outline-violet-400
          border border-violet-200
          hover:bg-violet-100`,
        dark: `bg-violet-900/30
          text-violet-100
          outline-violet-500
          border border-violet-800
          hover:bg-violet-900/40`
      }
    }
  };

  const futureKey: 'create' | 'update' | 'replace' =
    tone === 'future' && planAction && planAction !== 'delete'
      ? (planAction as 'create' | 'update' | 'replace')
      : 'update';
  const cardClass =
    tone === 'future'
      ? styles.future[futureKey as 'create' | 'update' | 'replace'][variant]
      : styles.current[variant];
  const headerColor =
    tone === 'future'
      ? {
          create: 'text-emerald-700 dark:text-emerald-200',
          update: 'text-amber-700 dark:text-amber-200',
          replace: 'text-violet-700 dark:text-violet-200'
        }[futureKey as 'create' | 'update' | 'replace']
      : 'text-primary-600';
  const iconColor =
    tone === 'future'
      ? {
          create: 'fill-emerald-600',
          update: 'fill-amber-600',
          replace: 'fill-violet-600'
        }[futureKey as 'create' | 'update' | 'replace']
      : 'fill-primary-600';
  const mutedTextClass =
    variant === 'light' ? 'text-gray-500' : 'text-nuances-200';

  const propertiesClass =
    variant === 'light'
      ? 'bg-gray-50 text-gray-900'
      : 'bg-nuances-black/70 text-nuances-50';
  const dependencyLinkClass = twMerge(
    'underline text-left focus-visible:outline-solid focus-visible:outline-1 focus-visible:outline-offset-2 rounded-sm transition-colors cursor-pointer',
    variant === 'light'
      ? 'text-primary-600 hover:text-primary-400 focus-visible:outline-primary-600'
      : 'text-primary-300 hover:text-primary-100 focus-visible:outline-nuances-50'
  );
  const dependencyDisabledClass = twMerge(
    'cursor-not-allowed focus-visible:outline-none hover:text-current no-underline',
    variant === 'light'
      ? 'text-nuances-300 hover:text-nuances-300'
      : 'text-nuances-200 hover:text-nuances-200'
  );
  const copyButtonStyles = {
    light: `bg-nuances-white text-nuances-black border border-nuances-300 hover:bg-primary-100`,
    dark: `bg-nuances-400 text-nuances-50 border border-nuances-200/60 hover:bg-nuances-400`
  } as const;
  const copyButtonClass = twMerge(
    'absolute top-2 right-2 inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-semibold transition-colors focus-visible:outline-solid focus-visible:outline-1 focus-visible:outline-offset-2',
    copyButtonStyles[variant],
    variant === 'light'
      ? 'focus-visible:outline-primary-600 focus-visible:outline-offset-2'
      : 'focus-visible:outline-nuances-50 focus-visible:outline-offset-2'
  );

  useEffect(() => {
    return () => {
      if (copyResetRef.current) {
        window.clearTimeout(copyResetRef.current);
      }
    };
  }, []);

  const handleCopyAttributes = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.stopPropagation();
    if (!attributes) return;
    const text = JSON.stringify(attributes, null, 2);

    const setCopied = () => {
      setCopyStatus('copied');
      if (copyResetRef.current) {
        window.clearTimeout(copyResetRef.current);
      }
      copyResetRef.current = window.setTimeout(() => setCopyStatus('idle'), 1500);
    };

    if (navigator?.clipboard?.writeText) {
      navigator.clipboard
        .writeText(text)
        .then(setCopied)
        .catch((error) => {
          console.error('Failed to copy attributes', error);
        });
      return;
    }

    try {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.setAttribute('readonly', '');
      textarea.style.position = 'absolute';
      textarea.style.left = '-9999px';
      document.body.appendChild(textarea);
      textarea.select();
      const successful = document.execCommand('copy');
      document.body.removeChild(textarea);
      if (successful) {
        setCopied();
      }
    } catch (error) {
      console.error('Failed to copy attributes', error);
    }
  };

  return (
    <div
      onClick={handleExpand}
      className={twMerge(
        'flex flex-col gap-2 p-4 rounded-lg outline-solid outline-1 cursor-pointer transition-colors',
        cardClass,
        className
      )}
    >
      <div className="flex justify-between items-center">
        <div className="flex items-center gap-2">
          <p className={twMerge('text-sm uppercase font-semibold', headerColor)}>
            {addr}
          </p>
          {badge}
        </div>
        <AngleDownIcon
          className={twMerge(
            iconColor,
            'transition-transform duration-300',
            isExpanded && 'rotate-180'
          )}
          height={16}
          width={16}
        />
      </div>

      {isExpanded && attributes && (
        <div>
          {created_at && (
            <p className={twMerge('text-sm mb-2', mutedTextClass)}>
              Created at:{' '}
              <span className="font-medium">
                {new Date(created_at).toLocaleString()}
              </span>
            </p>
          )}
          {dependencies && dependencies.length > 0 && (
            <div className={twMerge('text-sm mb-2', mutedTextClass)}>
              <p>Dependencies:</p>
              <ul className="ml-4 list-disc">
                {dependencies.map((dep) => {
                  const isAvailable = isDependencyAvailable
                    ? isDependencyAvailable(dep)
                    : !!onDependencyClick;
                  return (
                  <li key={dep}>
                    <button
                      type="button"
                      className={twMerge(
                        dependencyLinkClass,
                        !isAvailable && dependencyDisabledClass
                      )}
                      disabled={!isAvailable}
                      onClick={(event) => {
                        event.stopPropagation();
                        if (isAvailable && onDependencyClick) {
                          onDependencyClick(dep);
                        }
                      }}
                    >
                      {dep}
                    </button>
                  </li>
                )})}
              </ul>
            </div>
          )}
          <div>
            <h3 className={twMerge('text-sm mb-2', mutedTextClass)}>
              Attributes:
            </h3>
          </div>
          <div className="relative mt-2">
            <button
              type="button"
              onClick={handleCopyAttributes}
              className={copyButtonClass}
              aria-label={copyStatus === 'copied' ? 'Attributes copied' : 'Copy attributes to clipboard'}
            >
              <CopyIcon className="h-4 w-4 fill-current" />
              {copyStatus === 'copied' && <span>Copied</span>}
            </button>
            <div
              onClick={(e) => e.stopPropagation()}
              className={twMerge(
                propertiesClass,
                'text-sm shadow-light p-3 rounded-md overflow-auto max-h-48 font-mono whitespace-pre'
              )}
              role="region"
              aria-label="attributes-json"
            >
              <pre className="m-0">{JSON.stringify(attributes, null, 2)}</pre>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default StateGraphInstanceCard;

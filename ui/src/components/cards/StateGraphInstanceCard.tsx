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
        light: `outline-terraform-create
          border-2 border-terraform-create`,
        dark: `outline-terraform-create
          border-2 border-terraform-create`,
      },
      update: {
        light: `outline-terraform-update
          border-2 border-terraform-update`,
        dark: `outline-terraform-update
          border-2 border-terraform-update`,
      },
      replace: {
        light: `outline-terraform-replace
          border-2 border-terraform-replace`,
        dark: `outline-terraform-replace
          border-2 border-terraform-replace`,
      }
    }
  };

  const futureKey: 'create' | 'update' | 'replace' =
    tone === 'future' && planAction && planAction !== 'delete'
      ? (planAction as 'create' | 'update' | 'replace')
      : 'update';
  const cardClass =
    tone === 'future'
      ? twMerge(styles.future[futureKey as 'create' | 'update' | 'replace'][variant],styles.current[variant])
      : styles.current[variant];
  const headerColor =
    tone === 'future'
      ? {
          create: 'text-terraform-create dark:text-terraform-create',
          update: 'text-terraform-update dark:text-terraform-update',
          replace: 'text-terraform-replace dark:text-terraform-replace'
        }[futureKey as 'create' | 'update' | 'replace']
      : 'text-primary-600';
  const iconColor =
    tone === 'future'
      ? {
          create: 'fill-terraform-create',
          update: 'fill-terraform-update',
          replace: 'fill-terraform-replace'
        }[futureKey as 'create' | 'update' | 'replace']
      : 'fill-primary-600';
  const mutedTextClass =
    variant === 'light' ? 'text-gray-500' : 'text-nuances-200';

  const propertiesClass =
    variant === 'light'
      ? 'bg-gray-50 text-gray-900'
      : 'bg-nuances-black/70 text-nuances-50';
  const futureBadgeStyles: Record<'create' | 'update' | 'replace', Record<'light' | 'dark', string>> = {
    create: {
      light: 'text-terraform-create border border-terraform-create',
      dark: 'text-terraform-create border border-terraform-create/60'
    },
    update: {
      light: 'text-terraform-update border border-terraform-update',
      dark: 'text-terraform-update border border-terraform-update/60'
    },
    replace: {
      light: 'text-terraform-replace border border-terraform-replace',
      dark: 'text-terraform-replace border border-terraform-replace/60'
    }
  };
  const dependencyLinkClass = twMerge(
    'underline text-left focus-visible:outline-solid focus-visible:outline-1 focus-visible:outline-offset-2 rounded-sm transition-colors cursor-pointer',
    variant === 'light'
      ? 'text-primary-600 hover:text-primary-400 focus-visible:outline-primary-600'
      : 'text-primary-300 hover:text-primary-100 focus-visible:outline-nuances-50'
  );
  const computedBadge =
    badge ??
    (tone === 'future' &&
    planAction &&
    futureBadgeStyles[futureKey as 'create' | 'update' | 'replace']
      ? (
          <span
            className={twMerge(
              'text-[10px] uppercase font-semibold px-2 py-0.5 rounded-full border inline-flex items-center gap-1',
              futureBadgeStyles[futureKey as 'create' | 'update' | 'replace'][variant]
            )}
          >
            {planAction}
          </span>
        )
      : null);
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
      <div className="flex justify-between items-center gap-2">
        <div className="flex items-center gap-2 min-w-0">
          <p
            className={twMerge(
              'text-sm uppercase font-semibold truncate',
              headerColor
            )}
            title={addr}
          >
            {addr}
          </p>
          {computedBadge && <span className="flex-shrink-0">{computedBadge}</span>}
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
                'text-xs shadow-light p-3 rounded-md overflow-auto max-h-64 font-mono whitespace-pre'
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

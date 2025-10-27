import React, { useState } from 'react';
import { twMerge } from 'tailwind-merge';

import AngleDownIcon from '@/assets/icons/AngleDownIcon';

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
}

const StateGraphInstanceCard: React.FC<StateGraphInstanceCardProps> = ({
  className,
  variant = 'light',
  instance: { addr, dependencies, attributes, created_at },
  defaultExpanded = false,
  tone = 'current',
  planAction = null,
  badge
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const handleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  const styles = {
    current: {
      light: `bg-primary-100
        text-nuances-black
        outline-primary-500
        hover:bg-nuances-white`,

      dark: `bg-nuances-black
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
            <p className="text-sm text-gray-500 mb-2">
              Created at:{' '}
              <span className="font-medium">
                {new Date(created_at).toLocaleString()}
              </span>
            </p>
          )}
          {dependencies && dependencies.length > 0 && (
            <div className="text-sm text-gray-500 mb-2">
              <p>Dependencies:</p>
              <ul className="ml-4 list-disc">
                {dependencies.map((dep) => (
                  <li key={dep}>{dep}</li>
                ))}
              </ul>
            </div>
          )}
          <div>
            <h3 className="text-sm text-gray-500 mb-2">Attributes:</h3>
          </div>
          <div
            onClick={(e) => e.stopPropagation()}
            className="mt-2 bg-gray-50 dark:bg-nuances-black/70 text-sm shadow-light text-gray-900 dark:text-nuances-50 p-3 rounded-md overflow-auto max-h-48 font-mono whitespace-pre"
            role="region"
            aria-label="attributes-json"
          >
            <pre className="m-0">{JSON.stringify(attributes, null, 2)}</pre>
          </div>
        </div>
      )}
    </div>
  );
};

export default StateGraphInstanceCard;

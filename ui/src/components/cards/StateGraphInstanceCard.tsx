import React, { useState } from 'react';
import { twMerge } from 'tailwind-merge';

import AngleDownIcon from '@/assets/icons/AngleDownIcon';

import { StateGraphResourceInstance } from '@/clients/layers/types';

export interface StateGraphInstanceCardProps {
  className?: string;
  variant?: 'light' | 'dark';
  instance: StateGraphResourceInstance
  defaultExpanded?: boolean;
}

const StateGraphInstanceCard: React.FC<StateGraphInstanceCardProps> = ({
  className,
  variant = 'light',
  instance: { addr, dependencies, attributes, created_at },
  defaultExpanded = false
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const handleExpand = () => {
    setIsExpanded(!isExpanded);
  };

  const styles = {
    base: {
      light: `bg-primary-100
        text-nuances-black
        outline-primary-500
        hover:bg-nuances-white`,

      dark: `bg-nuances-black
        text-nuances-50
        outline-nuances-50
        hover:bg-nuances-400`
    },
  };

return (
    <div 
        onClick={handleExpand} 
        className={twMerge(
            'flex flex-col gap-2 p-4 rounded-lg outline-solid outline-1 cursor-pointer',
            styles.base[variant],
            className
        )}
    >
        <div className="flex justify-between items-center">
            <p className="text-sm uppercase font-semibold text-primary-600">{addr}</p>
            <AngleDownIcon
                className={twMerge(
                    'fill-primary-600 transition-transform duration-300',
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
                        Created at: <span className="font-medium">{new Date(created_at).toLocaleString()}</span>
                    </p>
                )}
                {dependencies && dependencies.length > 0 && (
                    <p className="text-sm text-gray-500 mb-2">
                        Dependencies:
                        <ul>
                            {dependencies.map((dep) => (
                                <li key={dep} className="ml-4 list-disc">{dep}</li>
                            ))}
                        </ul>
                    </p>
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

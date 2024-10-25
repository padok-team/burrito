import React from 'react';
import { twMerge } from 'tailwind-merge';

import WindowIcon from '@/assets/icons/WindowIcon';

export interface LogsButtonProps
  extends React.HTMLAttributes<HTMLButtonElement> {
  className?: string;
  variant?: 'light' | 'dark';
}

const LogsButton = React.forwardRef<HTMLButtonElement, LogsButtonProps>(
  ({ className, variant = 'light', ...props }, ref) => {
    return (
      <button
        className={twMerge(
          `${
            variant === 'light'
              ? 'hover:bg-primary-300 '
              : 'hover:bg-nuances-black'
          }
          rounded-full
          cursor-pointer
          transition-colors
          duration-300`,
          className
        )}
        ref={ref}
        {...props}
      >
        <WindowIcon className="p-2 fill-blue-500" width={40} height={40} />
      </button>
    );
  }
);

export default LogsButton;

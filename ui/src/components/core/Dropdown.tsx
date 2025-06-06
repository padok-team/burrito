import React from 'react';
import { twMerge } from 'tailwind-merge';

import AngleDownIcon from '@/assets/icons/AngleDownIcon';

export interface DropdownProps {
  className?: string;
  variant?: 'light' | 'dark';
  label: string;
  filled?: boolean;
  disabled?: boolean;
}

const Dropdown = React.forwardRef<HTMLDivElement, DropdownProps>(
  (
    { className, variant = 'light', label, filled, disabled, ...props },
    ref
  ) => {
    const styles = {
      base: {
        light: `bg-primary-400
        text-primary-600
        fill-primary-600`,

        dark: `bg-nuances-400
        text-nuances-300
        fill-nuances-300`
      },

      filled: {
        light: `text-nuances-black`,
        dark: `text-nuances-50`
      },

      disabled: `bg-nuances-50
        text-nuances-200
        fill-nuances-200
        hover:outline-0
        focus:outline-0
        cursor-default`
    };

    return (
      <div
        className={twMerge(
          `relative
          flex
          flex-row
          items-center
          justify-center
          h-8
          p-2
          gap-2
          rounded-lg
          text-base
          font-medium
          whitespace-nowrap
          cursor-pointer
          outline-primary-600
          outline-offset-0
          hover:outline-solid
          hover:outline-1
          focus:outline-solid
          focus:outline-2
          ${styles.base[variant]}`,
          className,
          filled && styles.filled[variant],
          disabled && styles.disabled
        )}
        tabIndex={0}
        ref={ref}
        {...props}
      >
        {label}
        <AngleDownIcon />
      </div>
    );
  }
);

export default Dropdown;

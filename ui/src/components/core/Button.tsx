import React from 'react';
import { twMerge } from 'tailwind-merge';

import LoaderIcon from '@/assets/icons/LoaderIcon';

export interface ButtonProps {
  className?: string;
  theme?: 'light' | 'dark';
  variant?: 'primary' | 'secondary' | 'tertiary';
  children?: React.ReactNode;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  isLoading?: boolean;
  disabled?: boolean;
  label?: string;
  onClick?: () => void;
}

const Button: React.FC<ButtonProps> = ({
  className,
  variant = 'primary',
  theme = 'light',
  children,
  leftIcon,
  rightIcon,
  isLoading,
  disabled,
  label,
  onClick
}) => {
  const styles = {
    light: {
      base: {
        primary: `bg-nuances-black
          text-nuances-white
          hover:bg-nuances-400
          active:bg-nuances-400
          focus-visible:outline-solid
          focus-visible:outline-1
          focus-visible:outline-offset-[3px]
          focus-visible:outline-nuances-black
          fill-nuances-white`,

        secondary: `bg-nuances-white
          text-nuances-black
          border
          border-nuances-black
          hover:bg-nuances-50
          active:bg-nuances-50
          focus-visible:outline-solid
          focus-visible:outline-1
          focus-visible:outline-offset-[3px]
          focus-visible:outline-nuances-white
          fill-nuances-black`,

        tertiary: `
          text-primary-600
          underline
          hover:text-primary-400
          hover:fill-primary-400
          active:text-primary-400
          active:fill-primary-400
          focus-visible:outline-hidden
          fill-primary-600`
      },

      disabled: {
        primary: `bg-nuances-50
          text-nuances-300
          fill-nuances-300
          active: bg-nuances-50
          hover:bg-nuances-50
          `,

        secondary: `bg-nuances-50
          text-nuances-300
          fill-nuances-300
          active: bg-nuances-50
          hover:bg-nuances-50`,

        tertiary: `
          underline
          text-nuances-200
          fill-nuances-200
          active:text-nuances-200
          active:fill-nuances-200
          hover:text-nuances-200
          hover:fill-nuances-200`
      }
    },
    dark: {
      base: {
        primary: `bg-nuances-black
          text-nuances-white
          hover:bg-nuances-400
          active:bg-nuances-400
          focus-visible:outline-solid
          focus-visible:outline-1
          focus-visible:outline-offset-[3px]
          focus-visible:outline-nuances-black
          fill-nuances-white`,

        secondary: `bg-nuances-white
          text-nuances-black
          border
          border-nuances-black
          hover:bg-nuances-50
          active:bg-nuances-50
          focus-visible:outline-solid
          focus-visible:outline-1
          focus-visible:outline-offset-[3px]
          focus-visible:outline-nuances-white
          fill-nuances-black`,

        tertiary: `bg-nuances-black
          text-primary-600
          underline
          hover:text-primary-400
          hover:fill-primary-400
          active:text-primary-400
          active:fill-primary-400
          focus-visible:outline-hidden
          fill-primary-600`
      },

      disabled: {
        primary: `bg-nuances-50
          text-nuances-300
          fill-nuances-300
          hover:bg-nuances-50
          active:bg-nuances-50`,

        secondary: `bg-nuances-50
          text-nuances-300
          fill-nuances-300
          hover:bg-nuances-50
          active:bg-nuances-50`,

        tertiary: `bg-nuances-black
          underline
          text-nuances-300
          fill-nuances-300
          hover:text-nuances-300
          hover:fill-nuances-300
          active:text-nuances-300
          active:fill-nuances-300`
      }
    }
  };

  return (
    <button
      className={twMerge(
        `relative
        px-4
        py-2
        rounded-md`,
        styles[theme].base[variant],
        disabled && styles[theme].disabled[variant],
        className
      )}
      tabIndex={0}
      disabled={disabled}
      onClick={onClick}
    >
      <div className="flex justify-center items-center gap-2">
        {leftIcon && (
          <span className={`${isLoading && 'invisible'}`}>{leftIcon}</span>
        )}
        <div className={`font-semibold text-base ${isLoading && 'invisible'}`}>
          {children}
        </div>
        {isLoading && (
          <div
            className={`
              absolute
              inset-0
              flex
              justify-center
              items-center
              z-10
            `}
          >
            <LoaderIcon className="animate-spin" />
          </div>
        )}
        {rightIcon && (
          <span className={`${isLoading && 'invisible'}`}>{rightIcon}</span>
        )}
        {label && (
          <span
            className={`font-semibold text-base ${isLoading && 'invisible'}`}
          >
            {label}
          </span>
        )}
      </div>
    </button>
  );
};

export default Button;

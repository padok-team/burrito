import React, { useRef, useId } from 'react';
import { twMerge } from 'tailwind-merge';

import ExclamationTriangleIcon from '@/assets/icons/ExclamationTriangleIcon';

export interface InputProps {
  className?: string;
  variant?: 'light' | 'dark';
  label?: string;
  type?: React.HTMLInputTypeAttribute;
  placeholder?: string;
  value?: string;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  caption?: string;
  error?: boolean;
  disabled?: boolean;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const Input: React.FC<InputProps> = ({
  className,
  variant = 'light',
  label,
  type,
  placeholder,
  value,
  leftIcon,
  rightIcon,
  caption,
  error,
  disabled,
  onChange
}) => {
  const inputId = useId();
  const inputRef = useRef<HTMLInputElement>(null);

  const handleEscape = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (inputRef.current && event.key === 'Escape') {
      // event.preventDefault();
      inputRef.current.blur();
    }
  };

  const styles = {
    base: {
      light: `bg-primary-400
        text-nuances-black
        caret-nuances-black
        placeholder-primary-600
        fill-primary-600`,

      dark: `bg-nuances-400
        text-nuances-50
        caret-nuances-50
        placeholder-nuances-300
        fill-nuances-300`
    },

    error: `outline
      outline-1
      outline-status-error-default`,

    disabled: `bg-nuances-50
      placeholder-nuances-200
      fill-nuances-200
      outline-0
      hover:outline-0
      focus:outline-0
      active:outline-0`
  };

  return (
    <div className="w-full">
      {label && (
        <label
          className={twMerge(
            `block
            font-normal
            text-base
            mb-2
            ${variant === 'light' ? 'text-nuances-black' : 'text-nuances-50'}`,
            disabled && 'text-nuances-300'
          )}
          htmlFor={inputId}
        >
          {label}
        </label>
      )}
      <div className="relative flex items-center justify-center w-full">
        {leftIcon && (
          <div
            className={twMerge(
              `absolute
              left-0
              translate-x-4
              pointer-events-none
              ${styles.base[variant]}`,
              disabled && styles.disabled
            )}
          >
            {leftIcon}
          </div>
        )}
        <input
          className={twMerge(
            `px-4
            py-2
            h-10
            w-full
            rounded-lg
            font-medium
            text-base
            outline-primary-600
            outline-offset-0
            hover:outline-solid
            hover:outline-1
            focus:outline-solid
            focus:outline-2
            active:outline-solid
            active:outline-2
            ${styles.base[variant]}`,
            leftIcon && 'pl-12',
            rightIcon && 'pr-12',
            error && styles.error,
            error && 'pr-12',
            disabled && styles.disabled,
            className
          )}
          id={inputId}
          type={type}
          placeholder={placeholder}
          disabled={disabled}
          value={value}
          onChange={onChange}
          onKeyDown={handleEscape}
          ref={inputRef}
        />
        {rightIcon && !error && (
          <div
            className={twMerge(
              `absolute
              right-0
              -translate-x-4
              pointer-events-none
              ${styles.base[variant]}`,
              disabled && styles.disabled
            )}
          >
            {rightIcon}
          </div>
        )}
        {error && (
          <div
            className={twMerge(
              `absolute
              right-0
              -translate-x-4
              pointer-events-none
              fill-status-error-default`,
              disabled && styles.disabled
            )}
          >
            <ExclamationTriangleIcon />
          </div>
        )}
      </div>
      {caption && (
        <span
          className={twMerge(
            `block
            font-normal
            text-sm
            mt-2
            ${variant === 'light' ? 'text-primary-600' : 'text-nuances-300'}`,
            error && 'text-status-error-default',
            disabled && 'text-nuances-300'
          )}
        >
          {caption}
        </span>
      )}
    </div>
  );
};

export default Input;

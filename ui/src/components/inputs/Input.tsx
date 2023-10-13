import React from "react";
import { twMerge } from "tailwind-merge";

export interface InputProps {
  className?: string;
  variant?: "light" | "dark";
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
  variant,
  label,
  type,
  placeholder,
  value,
  leftIcon, // TODO: Implement
  rightIcon, // TODO: Implement
  caption,
  error,
  disabled,
  onChange,
}) => {
  const styles = {
    base: {
      light: `bg-primary-400
        text-nuances-black
        placeholder-primary-600`,

      dark: `bg-nuances-400
        text-nuances-50
        placeholder-nuances-300`,
    },

    error: `outline
      outline-1
      outline-status-error-default`,

    disabled: `bg-nuances-50
      placeholder-nuances-200
      hover:outline-0
      focus:outline-0
      active:outline-0`,
  };

  variant = variant ?? "light";

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        justify-center
        items-start
        gap-1`,
        className
      )}
    >
      {label && (
        <label
          className={twMerge(
            `font-normal
            text-base
            ${variant === "light" ? "text-nuances-black" : "text-nuances-50"}`,
            disabled && "text-nuances-300"
          )}
        >
          {label}
        </label>
      )}
      <input
        className={twMerge(
          `flex
          px-4
          py-2
          h-10
          rounded-lg
          font-medium
          text-base
          outline-primary-600
          hover:outline
          hover:outline-1
          focus:outline
          focus:outline-2
          active:outline
          active:outline-2
          ${styles.base[variant]}`,
          error && styles.error,
          disabled && styles.disabled
        )}
        type={type}
        placeholder={placeholder}
        disabled={disabled}
        value={value}
        onChange={onChange}
      />
      {caption && (
        <span
          className={twMerge(
            `font-normal
            text-sm
            ${variant === "light" ? "text-primary-600" : "text-nuances-300"}`,
            error && "text-status-error-default",
            disabled && "text-nuances-300"
          )}
        >
          {caption}
        </span>
      )}
    </div>
  );
};

export default Input;

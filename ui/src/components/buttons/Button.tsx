import React from "react";
import { twMerge } from "tailwind-merge";

import LoaderIcon from "@/assets/icons/LoaderIcon";

export interface ButtonProps {
  className?: string;
  children?: React.ReactNode;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  variant?: "primary" | "secondary" | "tertiary";
  isLoading?: boolean;
  disabled?: boolean;
  onClick?: () => void;
}

const Button: React.FC<ButtonProps> = ({
  className,
  children,
  leftIcon,
  rightIcon,
  variant,
  isLoading,
  disabled,
  onClick,
}) => {
  const variantClasses = {
    primary: `bg-nuances-black
      text-nuances-white
      hover:bg-nuances-400
      active:bg-nuances-400
      fill-nuances-white`,

    secondary: `
      bg-nuances-white
      text-nuances-black
      border
      border-nuances-black
      hover:bg-nuances-50
      active:bg-nuances-50
      fill-nuances-black`,

    tertiary: `bg-nuances-white
      text-primary-600
      underline
      hover:text-primary-400
      hover:fill-primary-400
      active:text-primary-400
      active:fill-primary-400
      fill-primary-600`,
  };

  const variantClassesDisabled = {
    primary: `bg-nuances-50
      text-nuances-300
      fill-nuances-300`,

    secondary: `bg-nuances-50
      text-nuances-300
      fill-nuances-300`,

    tertiary: `bg-nuances-white
      text-nuances-300
      underline
      fill-nuances-300`,
  };

  variant = variant ?? "primary";

  return (
    <button
      className={twMerge(
        `relative px-4 py-2 rounded-md ${
          !disabled ? variantClasses[variant] : variantClassesDisabled[variant]
        }`,
        className
      )}
      disabled={disabled}
      onClick={onClick}
    >
      <div className="flex justify-center items-center gap-2">
        {leftIcon && (
          <span className={`${isLoading && "invisible"}`}>{leftIcon}</span>
        )}
        <div className={`font-semibold text-base ${isLoading && "invisible"}`}>
          {children}
        </div>
        {isLoading && (
          <div className="absolute inset-0 flex justify-center items-center z-10">
            <LoaderIcon className="animate-spin" />
          </div>
        )}
        {rightIcon && (
          <span className={`${isLoading && "invisible"}`}>{rightIcon}</span>
        )}
      </div>
    </button>
  );
};

export default Button;

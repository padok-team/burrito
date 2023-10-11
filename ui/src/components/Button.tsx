import React from "react";

import LoaderIcon from "./icons/LoaderIcon";

export interface ButtonProps {
  children?: React.ReactNode;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  variant?: "primary" | "secondary" | "tertiary";
  isLoading?: boolean;
  disabled?: boolean;
  onClick?: () => void;
}

const Button: React.FC<ButtonProps> = ({
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
      active:text-primary-400
      fill-primary-600`,
  };

  const variantClassesDisabled = {
    primary: `bg-nuances-50
      text-nuances-300`,

    secondary: `bg-nuances-50
      text-nuances-300`,

    tertiary: `bg-nuances-white
      text-nuances-300
      underline`,
  };

  variant = variant ?? "primary";

  return (
    <button
      className={`Button relative px-4 py-2 rounded-md ${
        !disabled ? variantClasses[variant] : variantClassesDisabled[variant]
      }`}
      disabled={disabled}
      onClick={onClick}
    >
      {leftIcon && <span>{leftIcon}</span>}
      <div className={`font-semibold text-base ${isLoading && "invisible"}`}>
        {children}
      </div>
      {isLoading && (
        <div className="absolute inset-0 flex justify-center items-center z-10">
          <LoaderIcon className="w-6 h-6 animate-spin" />
        </div>
      )}
      {rightIcon && <span>{rightIcon}</span>}
    </button>
  );
};

export default Button;

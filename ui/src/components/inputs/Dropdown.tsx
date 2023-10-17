import React from "react";
import { twMerge } from "tailwind-merge";

import AngleDownIcon from "@/assets/icons/AngleDownIcon";

export interface DropdownProps {
  className?: string;
  variant?: "light" | "dark";
  label: string;
  children: React.ReactNode;
  filled?: boolean;
  disabled?: boolean;
}

const Dropdown: React.FC<DropdownProps> = ({
  className,
  variant = "light",
  label,
  children,
  filled,
  disabled,
}) => {
  const styles = {
    base: {
      light: `bg-primary-400
        text-primary-600
        fill-primary-600
        `,

      dark: `bg-nuances-400
        text-nuances-300
        fill-nuances-300`,
    },

    filled: {
      light: `text-nuances-black`,
      dark: `text-nuances-50`,
    },

    disabled: `bg-nuances-50
      text-nuances-200
      fill-nuances-200
      hover:outline-0
      focus:outline-0`,
  };

  return (
    <div
      className={twMerge(
        `flex
        flex-row
        items-center
        justify-center
        h-8
        p-2
        gap-2
        rounded-lg
        text-base
        font-medium
        outline-primary-600
        outline-offset-0
        hover:outline
        hover:outline-1
        focus:outline
        focus:outline-2
        ${styles.base[variant]}`,
        filled && styles.filled[variant],
        disabled && styles.disabled,
        className
      )}
      tabIndex={0}
    >
      {label}
      <AngleDownIcon />
    </div>
  );
};

export default Dropdown;

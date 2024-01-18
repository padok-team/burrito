import React from "react";
import { twMerge } from "tailwind-merge";

import CheckIcon from "@/assets/icons/CheckIcon";
import MinusIcon from "@/assets/icons/MinusIcon";

export interface CheckboxProps {
  className?: string;
  variant?: "light" | "dark";
  label: string;
  checked?: boolean;
  defaultChecked?: boolean;
  disabled?: boolean;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const Checkbox: React.FC<CheckboxProps> = ({
  className,
  variant = "light",
  label,
  checked,
  defaultChecked,
  disabled,
  onChange,
}) => {
  const styles = {
    base: {
      light: `bg-nuances-white
        border-primary-500
        hover:border-primary-600
        focus:outline
        focus:outline-1
        focus:outline-offset-2
        focus:outline-primary-600
        checked:bg-primary-600
        checked:border-primary-600
        checked:hover:border-primary-600`,

      dark: `bg-nuances-400
        border-nuances-300
        hover:border-nuances-100
        focus:outline
        focus:outline-1
        focus:outline-offset-2
        focus:outline-nuances-100
        checked:bg-primary-600
        checked:border-primary-600
        checked:hover:border-primary-600`,
    },

    disabled: `bg-nuances-50
      border-nuances-200
      hover:border-nuances-200`,
  };

  return (
    <label
      className={twMerge(
        `relative
        inline-flex
        items-center
        gap-4
        cursor-pointer`,
        className
      )}
    >
      <input
        className={twMerge(
          `peer
          appearance-none
          cursor-pointer
          h-5
          min-h-[20px]
          w-5
          min-w-[20px]
          border
          rounded
          ${styles.base[variant]}`,
          disabled && styles.disabled
        )}
        type="checkbox"
        tabIndex={0}
        checked={checked}
        defaultChecked={defaultChecked}
        disabled={disabled}
        onChange={onChange}
      />
      <CheckIcon
        className={`
          absolute
          left-0.5
          fill-nuances-white
          pointer-events-none
          hidden
          peer-checked:block
          peer-checked:peer-hover:hidden
        `}
        height={16}
        width={16}
      />
      <MinusIcon
        className={`
          absolute
          left-0.5
          fill-nuances-white
          pointer-events-none
          hidden
          peer-checked:hidden
          peer-checked:peer-hover:block
        `}
        height={16}
        width={16}
      />
      <span
        className={twMerge(
          `text-base
          font-normal
          ${variant === "light" ? "text-nuances-black" : "text-nuances-50"}`,
          disabled && "text-nuances-200"
        )}
      >
        {label}
      </span>
    </label>
  );
};

export default Checkbox;

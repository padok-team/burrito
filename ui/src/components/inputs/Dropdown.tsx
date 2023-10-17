import React, { useState, useRef, useEffect } from "react";
import { twMerge } from "tailwind-merge";

import Box from "@/components/misc/Box";

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
  const [open, setOpen] = useState(false);
  const divRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleFocus = () => {
      setOpen(true);
    };

    const handleBlur = () => {
      setOpen(false);
    };

    const divElement = divRef.current;

    if (divElement) {
      divElement.addEventListener("focus", handleFocus);
      divElement.addEventListener("blur", handleBlur);
    }

    return () => {
      if (divElement) {
        divElement.removeEventListener("focus", handleFocus);
        divElement.removeEventListener("blur", handleBlur);
      }
    };
  }, []);

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
      ref={divRef}
    >
      {label}
      <AngleDownIcon />

      {open && (
        <Box
          className={`absolute left-0 top-10 items-center justify-center`}
          variant={variant}
        >
          {children}
        </Box>
      )}
    </div>
  );
};

export default Dropdown;

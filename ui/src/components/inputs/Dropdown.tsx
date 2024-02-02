import React, { useState, useRef } from "react";
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
  const toggleRef = useRef<HTMLDivElement>(null);

  const handleClick = (event: React.MouseEvent<HTMLDivElement>) => {
    if (toggleRef.current) {
      if (event.target === toggleRef.current) {
        setOpen(!open);
      }
    }
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    if (toggleRef.current) {
      if (event.key === "Escape") {
        toggleRef.current.blur();
        setOpen(false);
      } else if (event.key === " " || event.key === "Enter") {
        if (toggleRef.current === document.activeElement) {
          event.preventDefault();
          setOpen(!open);
        }
      }
    }
  };

  const handleBlur = (event: React.FocusEvent<HTMLDivElement>) => {
    if (!event.currentTarget.contains(event.relatedTarget)) {
      setOpen(false);
    }
  };

  const styles = {
    base: {
      light: `bg-primary-400
        text-primary-600
        fill-primary-600`,

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
        whitespace-nowrap
        cursor-pointer
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
      onMouseDown={handleClick}
      onKeyDown={handleKeyDown}
      onBlur={handleBlur}
      ref={toggleRef}
    >
      {label}
      <AngleDownIcon className="pointer-events-none" />

      {open && (
        <Box
          className={`
            absolute
            left-0
            top-10
            items-center
            justify-center
            cursor-auto
          `}
          variant={variant}
        >
          {children}
        </Box>
      )}
    </div>
  );
};

export default Dropdown;

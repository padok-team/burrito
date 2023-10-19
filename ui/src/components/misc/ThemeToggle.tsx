import React, { useContext } from "react";
import { twMerge } from "tailwind-merge";

import { ThemeContext } from "@/contexts/ThemeContext";

import Box from "@/components/misc/Box";
import Toggle from "@/components/buttons/Toggle";

export interface ThemeToggleProps {
  className?: string;
  variant?: "light" | "dark";
}

const ThemeToggle: React.FC<ThemeToggleProps> = ({
  className,
  variant = "light",
}) => {
  const { theme, setTheme } = useContext(ThemeContext);
  return (
    <Box
      variant={variant}
      className={twMerge("items-center justify-center p-4 gap-4", className)}
    >
      <span
        className={`
          text-base
          font-normal
          ${variant === "light" ? "text-nuances-black" : "text-nuances-50"}
        `}
      >
        {variant === "dark" ? "Disable" : "Enable"} Dark Mode
      </span>
      <Toggle
        checked={theme === "dark"}
        onChange={() => setTheme(theme === "dark" ? "light" : "dark")}
      />
    </Box>
  );
};

export default ThemeToggle;

import React, { useContext } from "react";
import { twMerge } from "tailwind-merge";

import { ThemeContext } from "@/contexts/ThemeContext";

import Box from "@/components/core/Box";
import Toggle from "@/components/core/Toggle";

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
      <Toggle
        className={`
          text-base
          font-normal
          ${variant === "light" ? "text-nuances-black" : "text-nuances-50"}
        `}
        checked={theme === "dark"}
        onChange={() => setTheme(theme === "dark" ? "light" : "dark")}
        label={`${variant === "dark" ? "Disable" : "Enable"} Dark Mode`}
      />
    </Box>
  );
};

export default ThemeToggle;

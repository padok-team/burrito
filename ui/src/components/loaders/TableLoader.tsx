import React from "react";
import { twMerge } from "tailwind-merge";

export interface TableLoaderProps {
  className?: string;
  variant?: "light" | "dark";
}

const TableLoader: React.FC<TableLoaderProps> = ({
  className,
  variant = "light",
}) => {
  const styles = {
    light: `bg-[linear-gradient(270deg,_#D8EBFF_0%,_#ECF5FF_100%)]`,
    dark: `bg-[linear-gradient(270deg,_#252525_0%,_rgba(68,_67,_67,_0.24)_100%)]`,
  };

  return (
    <div
      className={twMerge(
        `h-4
        rounded-full
        animate-pulse
        ${styles[variant]}`,
        className
      )}
    ></div>
  );
};

export default TableLoader;

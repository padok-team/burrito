import React from "react";
import { twMerge } from "tailwind-merge";

import WindowIcon from "@/assets/icons/WindowIcon";

export interface LogsButtonProps {
  className?: string;
  variant?: "light" | "dark";
  onClick?: () => void;
}

const LogsButton: React.FC<LogsButtonProps> = ({
  className,
  variant = "light",
  onClick,
}) => {
  return (
    <div
      className={twMerge(
        `${
          variant === "light"
            ? "hover:bg-primary-300 "
            : "hover:bg-nuances-black"
        }
        rounded-full
        cursor-pointer
        transition-colors
        duration-300`,
        className
      )}
      onClick={onClick}
    >
      <WindowIcon className="p-2 fill-blue-500" width={40} height={40} />
    </div>
  );
};

export default LogsButton;

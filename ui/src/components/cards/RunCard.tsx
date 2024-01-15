import React, { useState } from "react";
import { twMerge } from "tailwind-merge";

import SyncIcon from "@/assets/icons/SyncIcon";
import AngleDownIcon from "@/assets/icons/AngleDownIcon";

import { Layer } from "@/clients/layers/types";

export interface RunCardProps {
  className?: string;
  variant?: "light" | "dark";
  isActive?: boolean;
  onClick?: () => void;
  layer: Layer;
}

const RunCard: React.FC<RunCardProps> = ({
  className,
  variant = "light",
  isActive,
  onClick,
  layer: { name, namespace, isRunning },
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const styles = {
    base: {
      light: `bg-primary-100
        text-nuances-black
        outline-primary-500
        hover:bg-nuances-white`,

      dark: `bg-nuances-black
        text-nuances-50
        outline-nuances-50
        hover:bg-nuances-400`,
    },

    isActive: {
      light: `bg-nuances-white
        shadow-light`,

      dark: `bg-nuances-400
        shadow-dark`,
    },

    isRunning: {
      light: `outline-blue-400`,

      dark: `outline-blue-500`,
    },
  };

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        gap-4
        p-4
        truncate
        cursor-pointer
        outline
        outline-1
        rounded-2xl
        transition-shadow
        duration-700
        ${styles.base[variant]}`,
        isActive && `outline-0 ${styles.isActive[variant]}`,
        isRunning && `outline-4 ${styles.isRunning[variant]}`,
        className
      )}
      onClick={onClick}
    >
      <div className="flex flex-col gap-2">
        <div className="flex justify-between">
          <span className="text-lg font-black">{name}</span>
          {isRunning && (
            <div className="flex items-center gap-2 text-blue-500 fill-blue-500">
              <span className="text-sm font-semibold">Running</span>
              <SyncIcon className="animate-spin-slow" height={16} width={16} />
            </div>
          )}
        </div>
        <span className="text-base font-semibold">{namespace}</span>
      </div>
      <div
        className="flex gap-2 cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <span
          className={
            variant === "light" ? "text-primary-600" : "text-nuances-300"
          }
        >
          Runs
        </span>
        <div className="flex items-center">
          {/* TODO - Replace with actual value */}
          <span className="font-semibold">34</span>
          <AngleDownIcon
            className={`
              fill-blue-500
              ${isExpanded && "transform -rotate-180"}
              transition-transform
              duration-500
            `}
            height={20}
            width={20}
          />
        </div>
      </div>
      <div
        // TODO - adjust animation with the number of runs
        className={`
          -mt-4
          overflow-hidden
          transition-all
          duration-500
          ${isExpanded ? "max-h-[152px] opacity-100" : "max-h-0 opacity-0"}
        `}
      >
        <div className="flex flex-col gap-1 pt-4">
          {/* TODO - Replace with actual runs */}
          <span>Run 1 - 12/01/2024</span>
          <span>Run 2 - 12/01/2024</span>
          <span>Run 3 - 12/01/2024</span>
          <span>Run 4 - 12/01/2024</span>
          <span>Run 5 - 12/01/2024</span>
          <span>Run 6 - 12/01/2024</span>
          <span>Run 7 - 12/01/2024</span>
          <span>Run 8 - 12/01/2024</span>
          <span>Run 9 - 12/01/2024</span>
        </div>
      </div>
    </div>
  );
};

export default RunCard;

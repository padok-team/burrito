import React, { useState } from "react";
import { twMerge } from "tailwind-merge";

import Running from "@/components/widgets/Running";

import AngleDownIcon from "@/assets/icons/AngleDownIcon";

import { Layer } from "@/clients/layers/types";

export interface RunCardProps {
  className?: string;
  variant?: "light" | "dark";
  isActive?: boolean;
  onClick?: () => void;
  handleActive?: (layer: Layer, run: string) => void;
  layer: Layer;
}

const RunCard: React.FC<RunCardProps> = ({
  className,
  variant = "light",
  isActive,
  onClick,
  handleActive,
  layer,
  layer: { name, namespace, runCount, latestRuns, isRunning },
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleActiveRun = (
    event: React.MouseEvent<HTMLSpanElement>,
    layer: Layer,
    run: string
  ) => {
    event.stopPropagation();
    handleActive && handleActive(layer, run);
  };

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
        <div className="flex justify-between gap-4">
          <span className="text-lg font-black truncate">{name}</span>
          {isRunning && <Running />}
        </div>
        <span className="text-base font-semibold truncate">{namespace}</span>
      </div>
      <div
        className="flex items-center gap-2 w-fit cursor-pointer"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <span
          className={
            variant === "light" ? "text-primary-600" : "text-nuances-300"
          }
        >
          Runs
        </span>
        <span className="font-semibold">{runCount}</span>
        <AngleDownIcon
          className={`
            -ml-2
            fill-blue-500
            ${isExpanded && "transform -rotate-180"}
            transition-transform
            duration-500
          `}
          height={20}
          width={20}
        />
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
          {latestRuns.map((run, index) => (
            <span
              key={index}
              className="cursor-pointer hover:underline"
              onClick={(event) => handleActiveRun(event, layer, run.id)}
            >
              {run.commit} - {run.date} - {run.action}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
};

export default RunCard;

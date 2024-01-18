import React from "react";
import { twMerge } from "tailwind-merge";

import Dropdown from "@/components/core/Dropdown";
import AttemptButton from "@/components/buttons/AttemptButton";
import CopyIcon from "@/assets/icons/CopyIcon";
import DownloadAltIcon from "@/assets/icons/DownloadAltIcon";

import { Layer } from "@/clients/layers/types";

export interface LogsTerminalProps {
  className?: string;
  variant?: "light" | "dark";
  layer: Layer;
}

const LogsTerminal: React.FC<LogsTerminalProps> = ({
  className,
  variant = "light",
  layer: { namespace, name },
}) => {
  const styles = {
    light: `bg-nuances-50
      text-nuances-black
      fill-nuances-black
      border-primary-500`,
    dark: `bg-nuances-400
      text-nuances-50
      fill-nuances-50
      border-nuances-black`,
  };

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        w-[800px]
        rounded-2xl
        ${styles[variant]}`,
        className
      )}
    >
      <div className="flex flex-row justify-between items-center p-4">
        <div className="flex flex-row items-center gap-4">
          <span className="text-lg font-black">{name}</span>
          <span className="text-base font-semibold">{namespace}</span>
          <Dropdown label="Latest attempt" variant={variant}>
            <></>
          </Dropdown>
        </div>
        <div className="flex flex-row items-center gap-4">
          <CopyIcon height={30} width={30} />
          <DownloadAltIcon height={30} width={30} />
        </div>
      </div>
      <hr
        className={`
          h-[1px]
          w-full
          ${variant === "light" ? "border-primary-600" : "border-nuances-300"}
        `}
      />
      <div className="flex flex-row items-center gap-1 px-4 pt-4">
        <AttemptButton
          variant={variant}
          attempt={1}
          isActive={true}
          onClick={() => console.log("active")}
          onClose={() => console.log("close")}
        />
        <AttemptButton
          variant={variant}
          attempt={2}
          isActive={false}
          onClick={() => console.log("active")}
          onClose={() => console.log("close")}
        />
      </div>
    </div>
  );
};

export default LogsTerminal;

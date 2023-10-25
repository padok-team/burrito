import React from "react";
import { twMerge } from "tailwind-merge";
import { Tooltip } from "react-tooltip";

import Tag from "@/components/tags/Tag";
import SyncIcon from "@/assets/icons/SyncIcon";
import CodeBranchIcon from "@/assets/icons/CodeBranchIcon";
import ChiliLight from "@/assets/illustrations/ChiliLight";
import ChiliDark from "@/assets/illustrations/ChiliDark";

import { Layer } from "@/clients/layers/types";

export interface CardProps {
  className?: string;
  variant?: "light" | "dark";
  layer: Layer;
}

const Card: React.FC<CardProps> = ({
  className,
  variant = "light",
  layer: {
    name,
    namespace,
    state,
    repository,
    branch,
    path,
    lastResult,
    isRunning,
    isPR,
  },
}) => {
  const getTag = () => {
    return (
      <div className="flex items-center">
        <Tag variant={state} />
        {state === "error" &&
          (variant === "light" ? (
            <ChiliLight
              className="absolute translate-x-16 rotate-[-21deg]"
              height={40}
              width={40}
            />
          ) : (
            <ChiliDark
              className="absolute translate-x-16 rotate-[-21deg]"
              height={40}
              width={40}
            />
          ))}
      </div>
    );
  };

  return (
    <div
      className={twMerge(
        `flex
        flex-col
        items-start
        rounded-2xl
        p-6
        gap-4
        ${variant === "light" ? "bg-nuances-white" : "bg-nuances-400"}
        ${variant === "light" ? "shadow-light" : "shadow-dark"}
        ${
          isRunning &&
          `outline outline-4 ${
            variant === "light" ? "outline-blue-400" : "outline-blue-500"
          }`
        }`,
        className
      )}
    >
      <div
        className={`
          flex
          items-center
          justify-between
          self-stretch
          gap-3
        `}
      >
        <span
          className={`
            text-lg
            font-black
            leading-6
            truncate
            ${variant === "light" ? "text-nuances-black" : "text-nuances-50"}
          `}
        >
          {name}
        </span>
        {isRunning ? (
          <div className="flex items-center gap-2 text-blue-500 fill-blue-500">
            <span className="text-sm font-semibold">Running</span>
            <SyncIcon className="animate-spin-slow" height={16} width={16} />
          </div>
        ) : isPR ? (
          <CodeBranchIcon
            className={`
              ${variant === "light" ? "fill-nuances-black" : "fill-nuances-50"}
            `}
          />
        ) : null}
      </div>
      <div className="grid grid-cols-[min-content_1fr] items-start gap-x-7 gap-y-2">
        {[
          ["Namespace", namespace],
          ["State", getTag()],
          ["Repository", repository],
          ["Branch", branch],
          ["Path", path],
          ["Last result", lastResult],
        ].map(([label, value], index) => (
          <React.Fragment key={index}>
            <span
              className={`
                text-base
                font-normal
                truncate
                ${variant === "light" ? "text-primary-600" : "text-nuances-300"}
              `}
            >
              {label}
            </span>
            <div
              className={`
                text-base
                font-semibold
                truncate
                ${
                  variant === "light" ? "text-nuances-black" : "text-nuances-50"
                }
              `}
            >
              <span
                data-tooltip-id="card-tooltip"
                data-tooltip-content={
                  label === "Path" || label === "Last result"
                    ? (value as string)
                    : null
                }
              >
                {value}
              </span>
            </div>
          </React.Fragment>
        ))}
      </div>
      <Tooltip
        className="!opacity-100"
        id="card-tooltip"
        variant={variant === "light" ? "dark" : "light"}
      />
    </div>
  );
};

export default Card;

import React from "react";
import { twMerge } from "tailwind-merge";

import Tag from "@/components/tags/Tag";
import SyncIcon from "@/assets/icons/SyncIcon";
import Chili from "@/assets/illustrations/Chili";

import { LayerState } from "@/types/types";

export interface CardProps {
  className?: string;
  variant?: "light" | "dark";
  title: string;
  isRunning?: boolean;
  namespace: string;
  state: LayerState;
  repository: string;
  branch: string;
  path: string;
  lastResult: string;
}

const Card: React.FC<CardProps> = ({
  className,
  variant = "light",
  title,
  isRunning,
  namespace,
  state,
  repository,
  branch,
  path,
  lastResult,
}) => {
  const getTag = () => {
    return (
      <div className="flex items-center">
        <Tag variant={state} />
        {state === "error" && (
          <Chili
            className="absolute translate-x-16 rotate-[-21deg]"
            height={40}
            width={40}
          />
        )}
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
        className={`flex
          items-center
          justify-between
          self-stretch`}
      >
        <span
          className={`text-lg
            font-black
            leading-6
            text-ellipsis
            whitespace-nowrap
            overflow-hidden
            ${variant === "light" ? "text-nuances-black" : "text-nuances-50"}`}
        >
          {title}
        </span>
        {isRunning && (
          <div className="flex items-center gap-2 text-blue-500 fill-blue-500">
            <span className="text-sm font-semibold">Running</span>
            <SyncIcon height={16} width={16} />
          </div>
        )}
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
              className={`text-base
                font-normal
                text-ellipsis
                whitespace-nowrap
                overflow-hidden
                ${
                  variant === "light" ? "text-primary-600" : "text-nuances-300"
                }`}
            >
              {label}
            </span>
            <span
              className={`text-base
                font-semibold
                text-ellipsis
                whitespace-nowrap
                overflow-hidden
                ${
                  variant === "light" ? "text-nuances-black" : "text-nuances-50"
                }`}
            >
              {value}
            </span>
          </React.Fragment>
        ))}
      </div>
    </div>
  );
};

export default Card;

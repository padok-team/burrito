import React, { useContext } from "react";

import { ThemeContext } from "@/contexts/ThemeContext";

import Button from "@/components/buttons/Button";
import Input from "@/components/inputs/Input";
import Dropdown from "@/components/inputs/Dropdown";
import Toggle from "@/components/buttons/Toggle";

import SearchIcon from "@/assets/icons/SearchIcon";

const Logs: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  return (
    <div className="relative flex flex-col flex-grow h-screen gap-3 overflow-auto">
      <div
        className={`
          sticky
          top-0
          z-10
          flex
          flex-col
          p-6
          pb-3
          gap-6
          ${theme === "light" ? "bg-primary-100" : "bg-nuances-black"}
        `}
      >
        <div className="flex justify-between">
          <h1
            className={`
              text-[32px]
              font-extrabold
              leading-[130%]
              ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
            `}
          >
            Logs
          </h1>
          <Button variant={theme === "light" ? "primary" : "secondary"}>
            Refresh logs
          </Button>
        </div>
        <Input
          variant={theme}
          className="w-full"
          placeholder="Search into logs"
          leftIcon={<SearchIcon />}
        />
        <div className="flex flex-row items-center justify-between gap-8">
          <div className="flex flex-row items-center gap-4">
            <span
              className={`
                text-base
                font-semibold
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
            >
              {`
                0 logs
              `}
            </span>
            <span
              className={`
                border-l
                h-6
                ${
                  theme === "light"
                    ? "border-primary-600"
                    : "border-nuances-200"
                }
              `}
            ></span>
            <span
              className={`
                text-base
                font-medium
                ${theme === "light" ? "text-primary-600" : "text-nuances-200"}
              `}
            >
              Filter by
            </span>
            <div className="flex flex-row items-center gap-2">
              <Dropdown variant={theme} label="Repository">
                <></>
              </Dropdown>
              <Dropdown variant={theme} label="Date">
                <></>
              </Dropdown>
            </div>
            <Toggle
              className={`
                text-sm
                font-medium
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
              label="Show running logs"
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default Logs;

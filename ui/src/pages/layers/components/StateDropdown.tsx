import React from "react";

import Box from "@/components/misc/Box";
import Checkbox from "@/components/checkboxes/Checkbox";

import { LayerState } from "@/types/types";

export interface StateDropdownProps {
  variant?: "light" | "dark";
  filter: LayerState[];
  onChange: (filter: LayerState[]) => void;
}

const StateDropdown: React.FC<StateDropdownProps> = ({
  variant,
  filter,
  onChange,
}) => {
  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    state: LayerState
  ) => {
    if (e.target.checked) {
      onChange([...filter, state]);
    } else {
      onChange(filter.filter((s) => s !== state));
    }
  };

  return (
    <Box
      variant={variant}
      className="flex-col items-center justify-center gap-2"
    >
      <span className="self-start mx-4 mt-2">State</span>
      <hr
        className={`
          h-[1px]
          w-full
          ${variant === "light" ? "bg-primary-600" : "bg-nuances-300"}
        `}
      />
      <div className="flex flex-col self-start mx-4 mb-2 gap-2">
        <Checkbox
          variant={variant}
          label="OK"
          checked={filter.includes("success")}
          onChange={(e) => handleChange(e, "success")}
        />
        <Checkbox
          variant={variant}
          label="OutOfSync"
          checked={filter.includes("warning")}
          onChange={(e) => handleChange(e, "warning")}
        />
        <Checkbox
          variant={variant}
          label="Error"
          checked={filter.includes("error")}
          onChange={(e) => handleChange(e, "error")}
        />
      </div>
    </Box>
  );
};

export default StateDropdown;

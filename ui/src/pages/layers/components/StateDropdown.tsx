import React from "react";

import Box from "@/components/misc/Box";
import Checkbox from "@/components/checkboxes/Checkbox";

export interface StateDropdownProps {
  variant?: "light" | "dark";
}

const StateDropdown: React.FC<StateDropdownProps> = ({ variant }) => {
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
        <Checkbox variant={variant} label="OK" />
        <Checkbox variant={variant} label="OutOfSync" />
        <Checkbox variant={variant} label="Error" />
      </div>
    </Box>
  );
};

export default StateDropdown;

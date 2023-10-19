import React from "react";

import Box from "@/components/misc/Box";
import Input from "@/components/inputs/Input";
import Checkbox from "@/components/checkboxes/Checkbox";

export interface RepositoryDropdownProps {
  variant?: "light" | "dark";
}

const RepositoryDropdown: React.FC<RepositoryDropdownProps> = ({ variant }) => {
  return (
    <Box
      variant={variant}
      className="flex-col items-center justify-center gap-2"
    >
      <span className="self-start mx-4 mt-2">Repository</span>
      <hr
        className={`
          h-[1px]
          w-full
          ${variant === "light" ? "bg-primary-600" : "bg-nuances-300"}
        `}
      />
      <Input
        variant={variant}
        className="w-[200px] mx-2"
        placeholder="Search repository"
      />
      <hr
        className={`
          h-[1px]
          w-full
          ${variant === "light" ? "bg-primary-600" : "bg-nuances-300"}
        `}
      />
      <div className="flex flex-col self-start mx-4 mb-2 gap-2">
        <Checkbox variant={variant} label="Burrito" />
        <Checkbox variant={variant} label="Burrito-1" />
        <Checkbox variant={variant} label="Burrito-2" />
      </div>
    </Box>
  );
};

export default RepositoryDropdown;

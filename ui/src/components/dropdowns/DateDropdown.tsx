import React from "react";

import Box from "@/components/core/Box";
import Checkbox from "@/components/core/Checkbox";

export interface DateDropdownProps {
  variant?: "light" | "dark";
  filter: "ascending" | "descending" | null;
  onChange: (filter: "ascending" | "descending" | null) => void;
}

const DateDropdown: React.FC<DateDropdownProps> = ({
  variant,
  filter,
  onChange,
}) => {
  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    state: "ascending" | "descending" | null
  ) => {
    if (e.target.checked) {
      onChange(state);
    } else {
      onChange(null);
    }
  };

  return (
    <Box
      variant={variant}
      className={`
        flex-col
        items-center
        justify-center
        z-10
        gap-2
        ${variant === "light" ? "text-primary-600" : "text-nuances-300"}
      `}
    >
      <span className="self-start mx-4 mt-2">Date</span>
      <hr
        className={`
          h-[1px]
          w-full
          ${variant === "light" ? "border-primary-600" : "border-nuances-300"}
        `}
      />
      <div className="flex flex-col self-start mx-4 mb-2 gap-2 whitespace-nowrap">
        <Checkbox
          variant={variant}
          label="Recent to old"
          checked={filter === "descending"}
          onChange={(e) => handleChange(e, "descending")}
        />
        <Checkbox
          variant={variant}
          label="Old to recent"
          checked={filter === "ascending"}
          onChange={(e) => handleChange(e, "ascending")}
        />
      </div>
    </Box>
  );
};

export default DateDropdown;

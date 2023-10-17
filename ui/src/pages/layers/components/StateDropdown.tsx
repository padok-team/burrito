import React from "react";

import Box from "@/components/misc/Box";
import Checkbox from "@/components/checkboxes/Checkbox";

const StateDropdown: React.FC = () => {
  return (
    <Box className="flex-col items-center justify-center gap-2">
      <span className="self-start mx-4 mt-2">State</span>
      <hr className="h-[1px] w-full bg-primary-600" />
      <div className="flex flex-col self-start mx-4 mb-2 gap-2">
        <Checkbox label="OK" />
        <Checkbox label="OutOfSync" />
        <Checkbox label="Error" />
      </div>
    </Box>
  );
};

export default StateDropdown;

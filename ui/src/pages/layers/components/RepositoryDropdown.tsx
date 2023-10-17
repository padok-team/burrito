import React from "react";

import Box from "@/components/misc/Box";
import Input from "@/components/inputs/Input";
import Checkbox from "@/components/checkboxes/Checkbox";

const RepositoryDropdown: React.FC = () => {
  return (
    <Box className="flex-col items-center justify-center gap-2">
      <span className="self-start mx-4 mt-2">Repository</span>
      <hr className="h-[1px] w-full bg-primary-600" />
      <Input className="w-[200px] mx-2" placeholder="Search repository" />
      <hr className="h-[1px] w-full bg-primary-600" />
      <div className="flex flex-col self-start mx-4 mb-2 gap-2">
        <Checkbox label="Burrito" />
        <Checkbox label="Burrito-1" />
        <Checkbox label="Burrito-2" />
      </div>
    </Box>
  );
};

export default RepositoryDropdown;

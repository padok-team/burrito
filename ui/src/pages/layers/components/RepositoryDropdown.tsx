import React, { useState } from "react";

import Box from "@/components/misc/Box";
import Input from "@/components/inputs/Input";
import Checkbox from "@/components/checkboxes/Checkbox";

export interface RepositoryDropdownProps {
  variant?: "light" | "dark";
  filter: string[];
  onChange: (filter: string[]) => void;
}

const RepositoryDropdown: React.FC<RepositoryDropdownProps> = ({
  variant,
  filter,
  onChange,
}) => {
  const testData: string[] = [
    "burrito-1",
    "burrito-2",
    "burrito-3",
    "burrito-4",
  ];

  const [repositories, setRepositories] = useState<string[]>(testData);
  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    repository: string
  ) => {
    if (e.target.checked) {
      onChange([...filter, repository]);
    } else {
      onChange(filter.filter((r) => r !== repository));
    }
  };

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
      <div className="flex flex-col self-start max-h-52 w-full overflow-scroll px-4 py-1 mb-2 gap-2">
        {repositories.map((repository) => (
          <Checkbox
            key={repository}
            variant={variant}
            label={repository}
            checked={filter.includes(repository)}
            onChange={(e) => handleChange(e, repository)}
          />
        ))}
      </div>
    </Box>
  );
};

export default RepositoryDropdown;

import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { fetchRepositories } from "@/clients/repositories/client";
import { reactQueryKeys } from "@/clients/reactQueryConfig";

import Box from "@/components/core/Box";
import Input from "@/components/core/Input";
import Checkbox from "@/components/core/Checkbox";

export interface RepositoriesDropdownProps {
  variant?: "light" | "dark";
  filter: string[];
  onChange: (filter: string[]) => void;
}

const RepositoriesDropdown: React.FC<RepositoriesDropdownProps> = ({
  variant,
  filter,
  onChange,
}) => {
  const [search, setSearch] = useState<string>("");

  const repositoriesQuery = useQuery({
    queryKey: reactQueryKeys.repositories,
    queryFn: fetchRepositories,
    select: (data) => ({
      ...data,
      results: data.results.filter((r) =>
        r.name.toLowerCase().includes(search.toLowerCase())
      ),
    }),
  });

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

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  };

  return (
    <Box
      variant={variant}
      className="flex-col items-center justify-center gap-2"
    >
      <span
        className={`
          self-start
          mx-4
          mt-2
          ${variant === "light" ? "text-primary-600" : "text-nuances-300"}
        `}
      >
        Repositories
      </span>
      <hr
        className={`
          h-[1px]
          w-full
          ${variant === "light" ? "border-primary-600" : "border-nuances-300"}
        `}
      />
      <Input
        variant={variant}
        className="w-full mx-2"
        placeholder="Search repositories"
        value={search}
        onChange={handleSearch}
      />
      <hr
        className={`
          h-[1px]
          w-full
          ${variant === "light" ? "border-primary-600" : "border-nuances-300"}
        `}
      />
      <div
        className={`
          flex
          flex-col
          self-start
          max-h-52
          w-full
          overflow-scroll
          px-4
          py-1
          mb-2
          gap-2
        `}
      >
        {repositoriesQuery.isLoading && <span>Loading...</span>}
        {repositoriesQuery.isError && <span>An error occurred.</span>}
        {repositoriesQuery.isSuccess &&
          (repositoriesQuery.data.results.length !== 0 ? (
            repositoriesQuery.data.results.map((repository) => (
              <Checkbox
                key={repository.name}
                variant={variant}
                label={repository.name}
                checked={filter.includes(repository.name)}
                onChange={(e) => handleChange(e, repository.name)}
              />
            ))
          ) : (
            <span>No repositories found.</span>
          ))}
      </div>
    </Box>
  );
};

export default RepositoriesDropdown;

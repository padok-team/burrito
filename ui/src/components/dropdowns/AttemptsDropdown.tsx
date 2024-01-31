import React from "react";
import { useQuery } from "@tanstack/react-query";

import { fetchAttempts } from "@/clients/runs/client";
import { reactQueryKeys } from "@/clients/reactQueryConfig";

import Box from "@/components/core/Box";
import Checkbox from "@/components/core/Checkbox";

export interface AttemptsDropdownProps {
  variant?: "light" | "dark";
  runId: string;
  select: number[];
  onChange: (select: number[]) => void;
}

const AttemptsDropdown: React.FC<AttemptsDropdownProps> = ({
  variant,
  runId,
  select,
  onChange,
}) => {
  const attemptsQuery = useQuery({
    queryKey: reactQueryKeys.attempts(runId),
    queryFn: () => fetchAttempts(runId),
  });

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement>,
    index: number
  ) => {
    if (e.target.checked) {
      onChange([...select, index]);
    } else {
      onChange(select.filter((i) => i !== index));
    }
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
        Attempts
      </span>
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
        {attemptsQuery.isLoading && <span>Loading...</span>}
        {attemptsQuery.isError && <span>An error occurred.</span>}
        {attemptsQuery.isSuccess &&
          (attemptsQuery.data.count !== 0 ? (
            Array.from(Array(attemptsQuery.data.count)).map((_, index) => (
              <Checkbox
                key={index}
                variant={variant}
                label={`Attempt ${index + 1}`}
                checked={select.includes(index)}
                onChange={(e) => handleChange(e, index)}
              />
            ))
          ) : (
            <span>No attempts found.</span>
          ))}
      </div>
    </Box>
  );
};

export default AttemptsDropdown;

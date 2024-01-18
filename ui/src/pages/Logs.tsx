import React, { useContext, useState } from "react";
import { useQuery } from "@tanstack/react-query";

import { fetchLayers } from "@/clients/layers/client";
import { reactQueryKeys } from "@/clients/reactQueryConfig";

import { ThemeContext } from "@/contexts/ThemeContext";

import Button from "@/components/core/Button";
import Input from "@/components/core/Input";
import Dropdown from "@/components/core/Dropdown";
import Toggle from "@/components/core/Toggle";
import RunCard from "@/components/cards/RunCard";

import RepositoryDropdown from "@/components/dropdowns/RepositoryDropdown";
import DateDropdown from "@/components/dropdowns/DateDropdown";

import SearchIcon from "@/assets/icons/SearchIcon";

import { Layer } from "@/clients/layers/types";

const Logs: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const [search, setSearch] = useState<string>("");
  const [activeLayer, setActiveLayer] = useState<Layer | null>(null);
  const [repositoryFilter, setRepositoryFilter] = useState<string[]>([]);
  const [dateFilter, setDateFilter] = useState<
    "ascending" | "descending" | null
  >(null);
  const [hidePRFilter, setHidePRFilter] = useState<boolean>(true);

  const layersQuery = useQuery({
    queryKey: reactQueryKeys.layers,
    queryFn: fetchLayers,
    select: (data) => ({
      ...data,
      results: data.results
        .filter((layer) =>
          layer.name.toLowerCase().includes(search.toLowerCase())
        )
        .filter(
          (layer) =>
            repositoryFilter.length === 0 ||
            repositoryFilter.includes(layer.repository)
        )
        .filter((layer) => !hidePRFilter || !layer.isPR),
    }),
  });

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
          value={search}
          onChange={(e) => setSearch(e.target.value)}
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
                ${
                  layersQuery.isSuccess ? layersQuery.data.results.length : 0
                } layers
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
              <Dropdown
                variant={theme}
                label="Repository"
                filled={repositoryFilter.length !== 0}
              >
                <RepositoryDropdown
                  variant={theme}
                  filter={repositoryFilter}
                  onChange={setRepositoryFilter}
                />
              </Dropdown>
              <Dropdown
                variant={theme}
                label="Date"
                filled={dateFilter !== null}
              >
                <DateDropdown
                  variant={theme}
                  filter={dateFilter}
                  onChange={setDateFilter}
                />
              </Dropdown>
            </div>
            <Toggle
              className={`
                text-sm
                font-medium
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
              checked={hidePRFilter}
              onChange={() => setHidePRFilter(!hidePRFilter)}
              label="Hide Pull Requests"
            />
          </div>
        </div>
      </div>
      {layersQuery.isLoading && <></>}
      {layersQuery.isError && <></>}
      {layersQuery.isSuccess && (
        <div className="flex flex-col gap-4 p-6">
          {layersQuery.data.results.map((layer, index) => (
            <RunCard
              key={index}
              variant={theme}
              isActive={activeLayer?.name === layer.name}
              onClick={() => setActiveLayer(layer)}
              layer={layer}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default Logs;

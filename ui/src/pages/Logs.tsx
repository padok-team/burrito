import React, { useContext, useCallback, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams, useSearchParams, useNavigate } from "react-router-dom";

import { fetchLayers } from "@/clients/layers/client";
import { reactQueryKeys } from "@/clients/reactQueryConfig";

import { ThemeContext } from "@/contexts/ThemeContext";

import Button from "@/components/core/Button";
import Input from "@/components/core/Input";
import Dropdown from "@/components/core/Dropdown";
import RepositoriesDropdown from "@/components/dropdowns/RepositoriesDropdown";
import DateDropdown from "@/components/dropdowns/DateDropdown";
import Toggle from "@/components/core/Toggle";
import RunCardLoader from "@/components/loaders/RunCardLoader";
import RunCard from "@/components/cards/RunCard";
import LogsTerminal from "@/components/tools/LogsTerminal";

import SearchIcon from "@/assets/icons/SearchIcon";

import { Layer } from "@/clients/layers/types";

const Logs: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const { layerId, runId } = useParams();
  const [searchParams, setSerchParams] = useSearchParams();
  const navigate = useNavigate();

  const search = useMemo<string>(
    () => searchParams.get("search") || "",
    [searchParams]
  );

  const setSearch = useCallback(
    (search: string) => {
      searchParams.set("search", search);
      setSerchParams(searchParams);
    },
    [searchParams, setSerchParams]
  );

  const repositoryFilter = useMemo<string[]>(() => {
    const param = searchParams.get("repositories");
    return param ? param.split(",") : [];
  }, [searchParams]);

  const setRepositoryFilter = useCallback(
    (repositoryFilter: string[]) => {
      searchParams.set("repositories", repositoryFilter.join(","));
      setSerchParams(searchParams);
    },
    [searchParams, setSerchParams]
  );

  const dateFilter = useMemo<"ascending" | "descending" | null>(() => {
    const param = searchParams.get("date");
    return param === "ascending"
      ? "ascending"
      : param === "descending"
      ? "descending"
      : null;
  }, [searchParams]);

  const setDateFilter = useCallback(
    (dateFilter: "ascending" | "descending" | null) => {
      searchParams.set("date", dateFilter || "");
      setSerchParams(searchParams);
    },
    [searchParams, setSerchParams]
  );

  const hidePRFilter = useMemo<boolean>(
    () => searchParams.get("hidepr") !== "false",
    [searchParams]
  );

  const setHidePRFilter = useCallback(
    (hidePRFilter: boolean) => {
      searchParams.set("hidepr", hidePRFilter.toString());
      setSerchParams(searchParams);
    },
    [searchParams, setSerchParams]
  );

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
        .filter((layer) => !hidePRFilter || !layer.isPR)
        .sort((a, b) =>
          dateFilter === "ascending"
            ? new Date(a.lastRunAt).getTime() - new Date(b.lastRunAt).getTime()
            : dateFilter === "descending"
            ? new Date(b.lastRunAt).getTime() - new Date(a.lastRunAt).getTime()
            : 0
        ),
    }),
  });

  const handleActive = (layer: Layer, run?: string) => {
    navigate(
      `/logs/${layer.id}${
        run
          ? `/${run}`
          : layer.latestRuns.length > 0
          ? `/${layer.latestRuns[0].id}`
          : ""
      }`
    );
  };

  return (
    <div className="flex flex-col flex-1 h-screen min-w-0">
      <div
        className={`
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
          <Button
            variant={theme === "light" ? "primary" : "secondary"}
            isLoading={layersQuery.isRefetching}
            onClick={() => layersQuery.refetch()}
          >
            Refresh layers
          </Button>
        </div>
        <Input
          variant={theme}
          className="w-full"
          placeholder="Search into layers"
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
                label="Repositories"
                filled={repositoryFilter.length !== 0}
              >
                <RepositoriesDropdown
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
      <div
        className={`
          flex
          flex-row
          gap-6
          p-6
          ${layersQuery.isSuccess ? "overflow-auto" : "overflow-hidden"}
        `}
      >
        <div className="flex flex-col w-1/3 h-fit gap-6">
          {layersQuery.isLoading ? (
            Array.from({ length: 100 }).map((_, index) => (
              <RunCardLoader key={index} variant={theme} />
            ))
          ) : layersQuery.isError ? (
            <span
              className={`
                text-lg
                font-semibold
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
            >
              An error has occurred
            </span>
          ) : layersQuery.isSuccess ? (
            layersQuery.data.results.length > 0 ? (
              layersQuery.data.results.map((layer, index) => (
                <RunCard
                  key={index}
                  variant={theme}
                  isActive={layerId === layer.id}
                  onClick={() => handleActive(layer)}
                  handleActive={handleActive}
                  layer={layer}
                />
              ))
            ) : (
              <span
                className={`
                text-lg
                font-semibold
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
              >
                No layers found
              </span>
            )
          ) : (
            <></>
          )}
        </div>
        {layersQuery.isSuccess &&
          layersQuery.data.results.length > 0 &&
          (layerId &&
          layersQuery.data.results.some((layer) => layer.id === layerId) ? (
            runId ? (
              ((activeLayerObject) =>
                activeLayerObject && (
                  <LogsTerminal
                    className="flex-1 min-w-0 sticky top-0"
                    layer={activeLayerObject}
                    run={runId}
                    variant={theme}
                  />
                ))(
                layersQuery.data.results.find((layer) => layer.id === layerId)
              )
            ) : (
              <div className="flex items-center justify-center flex-1 min-w-0 sticky top-0">
                <span
                  className={`
                  text-xl
                  font-black
                  ${
                    theme === "light" ? "text-nuances-black" : "text-nuances-50"
                  }
                `}
                >
                  There is no run for this layer...
                </span>
              </div>
            )
          ) : (
            <div className="flex items-center justify-center flex-1 min-w-0 sticky top-0">
              <span
                className={`
                text-xl
                font-black
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
              >
                Select a layer...
              </span>
            </div>
          ))}
      </div>
    </div>
  );
};

export default Logs;

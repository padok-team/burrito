import React, { useState, useContext, useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";

import { fetchLayers } from "@/clients/layers/client";
import { reactQueryKeys } from "@/clients/reactQueryConfig";

import { ThemeContext } from "@/contexts/ThemeContext";

import Button from "@/components/core/Button";
import Input from "@/components/core/Input";
import Dropdown from "@/components/core/Dropdown";
import Toggle from "@/components/core/Toggle";
import NavigationButton from "@/components/navigation/NavigationButton";
import Card from "@/components/cards/Card";
import Table from "@/components/tables/Table";

import StateDropdown from "@/components/dropdowns/StateDropdown";
import RepositoriesDropdown from "@/components/dropdowns/RepositoriesDropdown";

import { LayerState } from "@/clients/layers/types";

import SearchIcon from "@/assets/icons/SearchIcon";
import AppsIcon from "@/assets/icons/AppsIcon";
import BarsIcon from "@/assets/icons/BarsIcon";
import CardLoader from "@/components/loaders/CardLoader";

const Layers: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const [view, setView] = useState<"grid" | "table">("grid");
  const [searchParams, setSerchParams] = useSearchParams();

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

  const stateFilter = useMemo<LayerState[]>(() => {
    const param = searchParams.get("state");
    return (param ? param.split(",") : []) as LayerState[];
  }, [searchParams]);

  const setStateFilter = useCallback(
    (stateFilter: LayerState[]) => {
      searchParams.set("state", stateFilter.join(","));
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
            stateFilter.length === 0 || stateFilter.includes(layer.state)
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
    <div className="flex flex-col flex-1 h-screen min-w-0">
      <div
        className={`
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
            Layers
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
                label="State"
                filled={stateFilter.length !== 0}
              >
                <StateDropdown
                  variant={theme}
                  filter={stateFilter}
                  onChange={setStateFilter}
                />
              </Dropdown>
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
          <div className="flex flex-row items-center gap-2">
            <NavigationButton
              icon={<AppsIcon />}
              variant={theme}
              selected={view === "grid"}
              onClick={() => setView("grid")}
            />
            <NavigationButton
              icon={<BarsIcon />}
              variant={theme}
              selected={view === "table"}
              onClick={() => setView("table")}
            />
          </div>
        </div>
      </div>
      <div
        className={`
          relative
          ${layersQuery.isSuccess ? "overflow-auto" : "overflow-hidden"}
        `}
      >
        {view === "grid" ? (
          <div className="grid grid-cols-[repeat(auto-fit,_minmax(400px,_1fr))] p-6 gap-6">
            {layersQuery.isLoading ? (
              Array.from({ length: 100 }).map((_, index) => (
                <CardLoader key={index} variant={theme} />
              ))
            ) : layersQuery.isError ? (
              <span
                className={`
                  text-lg
                  font-semibold
                  ${
                    theme === "light" ? "text-nuances-black" : "text-nuances-50"
                  }
                `}
              >
                An error has occurred.
              </span>
            ) : layersQuery.isSuccess ? (
              layersQuery.data.results.length > 0 ? (
                layersQuery.data.results.map((layer, index) => (
                  <Card key={index} variant={theme} layer={layer} />
                ))
              ) : (
                <span
                  className={`
                    text-lg
                    font-semibold
                    ${
                      theme === "light"
                        ? "text-nuances-black"
                        : "text-nuances-50"
                    }
                  `}
                >
                  No layers found
                </span>
              )
            ) : (
              <></>
            )}
          </div>
        ) : view === "table" ? (
          <div>
            {layersQuery.isLoading ? (
              <Table variant={theme} isLoading data={[]} />
            ) : layersQuery.isError ? (
              <span
                className={`
                  text-lg
                  font-semibold
                  ${
                    theme === "light" ? "text-nuances-black" : "text-nuances-50"
                  }
                `}
              >
                An error has occurred.
              </span>
            ) : layersQuery.isSuccess ? (
              layersQuery.data.results.length > 0 ? (
                <Table variant={theme} data={layersQuery.data.results} />
              ) : (
                <div className="p-6">
                  <span
                    className={`
                    text-lg
                    font-semibold
                    ${
                      theme === "light"
                        ? "text-nuances-black"
                        : "text-nuances-50"
                    }
                  `}
                  >
                    No layers found
                  </span>
                </div>
              )
            ) : (
              <></>
            )}
          </div>
        ) : (
          <></>
        )}
      </div>
    </div>
  );
};

export default Layers;
